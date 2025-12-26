package db

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================
// Test: CreatePermission
// ============================================================

func TestCreatePermission(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) CreatePermissionParams
		wantErr  bool
		checkErr func(t *testing.T, err error)
	}{
		{
			name: "success",
			setup: func(t *testing.T, q *Queries) CreatePermissionParams {
				return CreatePermissionParams{
					ID:          generateTestID(),
					Resource:    "clients",
					Action:      "read",
					Description: strPtr("Read client data"),
				}
			},
			wantErr: false,
		},
		{
			name: "success_without_description",
			setup: func(t *testing.T, q *Queries) CreatePermissionParams {
				return CreatePermissionParams{
					ID:       generateTestID(),
					Resource: "locations",
					Action:   "write",
				}
			},
			wantErr: false,
		},
		{
			name: "duplicate_id",
			setup: func(t *testing.T, q *Queries) CreatePermissionParams {
				existingID := generateTestID()
				CreateTestPermission(t, q, CreateTestPermissionOptions{ID: &existingID})
				return CreatePermissionParams{
					ID:       existingID,
					Resource: "different",
					Action:   "different",
				}
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.True(t, IsUniqueViolation(err), "expected unique violation, got: %v", err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				params := tt.setup(t, q)

				permission, err := q.CreatePermission(ctx, params)

				if tt.wantErr {
					require.Error(t, err)
					if tt.checkErr != nil {
						tt.checkErr(t, err)
					}
					return
				}

				require.NoError(t, err)
				assert.Equal(t, params.ID, permission.ID)
				assert.Equal(t, params.Resource, permission.Resource)
				assert.Equal(t, params.Action, permission.Action)
				assert.True(t, permission.CreatedAt.Valid)

				// Verify permission was created
				fetched, err := q.GetPermissionByID(ctx, params.ID)
				require.NoError(t, err)
				assert.Equal(t, params.ID, fetched.ID)
			})
		})
	}
}

// ============================================================
// Test: GetPermissionByID
// ============================================================

func TestGetPermissionByID(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) string
		wantErr  bool
		checkErr func(t *testing.T, err error)
		validate func(t *testing.T, permission Permission)
	}{
		{
			name: "found",
			setup: func(t *testing.T, q *Queries) string {
				return CreateTestPermission(t, q, CreateTestPermissionOptions{})
			},
			wantErr: false,
			validate: func(t *testing.T, permission Permission) {
				assert.NotEmpty(t, permission.ID)
				assert.NotEmpty(t, permission.Resource)
				assert.NotEmpty(t, permission.Action)
				assert.True(t, permission.CreatedAt.Valid)
			},
		},
		{
			name: "not_found",
			setup: func(t *testing.T, q *Queries) string {
				return "non-existent-id"
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.True(t, errors.Is(err, pgx.ErrNoRows), "expected ErrNoRows, got: %v", err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				id := tt.setup(t, q)

				permission, err := q.GetPermissionByID(ctx, id)

				if tt.wantErr {
					require.Error(t, err)
					if tt.checkErr != nil {
						tt.checkErr(t, err)
					}
					return
				}

				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, permission)
				}
			})
		})
	}
}

// ============================================================
// Test: ListPermissions
// ============================================================

func TestListPermissions(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries)
		params   ListPermissionsParams
		validate func(t *testing.T, results []ListPermissionsRow)
	}{
		{
			name:   "seeded_permissions_exist",
			setup:  func(t *testing.T, q *Queries) {},
			params: ListPermissionsParams{Limit: 100, Offset: 0},
			validate: func(t *testing.T, results []ListPermissionsRow) {
				// Migration seeds 15 permissions
				assert.GreaterOrEqual(t, len(results), 15)
			},
		},
		{
			name: "added_permissions_appear",
			setup: func(t *testing.T, q *Queries) {
				CreateTestPermission(t, q, CreateTestPermissionOptions{
					Resource: strPtr("test_res_1"),
					Action:   strPtr("read"),
				})
				CreateTestPermission(t, q, CreateTestPermissionOptions{
					Resource: strPtr("test_res_2"),
					Action:   strPtr("write"),
				})
			},
			params: ListPermissionsParams{Limit: 100, Offset: 0},
			validate: func(t *testing.T, results []ListPermissionsRow) {
				// Migration seeds + 2 new ones
				assert.GreaterOrEqual(t, len(results), 17)
			},
		},
		{
			name: "pagination",
			setup: func(t *testing.T, q *Queries) {
				for i := 0; i < 5; i++ {
					resource := fmt.Sprintf("paginated_resource_%d", i)
					CreateTestPermission(t, q, CreateTestPermissionOptions{
						Resource: &resource,
						Action:   strPtr("read"),
					})
				}
			},
			params: ListPermissionsParams{Limit: 5, Offset: 0},
			validate: func(t *testing.T, results []ListPermissionsRow) {
				assert.Len(t, results, 5)
				// Total count should include seeded + new ones
				assert.GreaterOrEqual(t, results[0].TotalCount, int64(20))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				tt.setup(t, q)

				results, err := q.ListPermissions(ctx, tt.params)

				require.NoError(t, err)
				tt.validate(t, results)
			})
		})
	}
}

// ============================================================
// Test: DeletePermission
// ============================================================

func TestDeletePermission(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, q *Queries) string
	}{
		{
			name: "existing",
			setup: func(t *testing.T, q *Queries) string {
				return CreateTestPermission(t, q, CreateTestPermissionOptions{})
			},
		},
		{
			name: "non_existent",
			setup: func(t *testing.T, q *Queries) string {
				return "non-existent-id"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				id := tt.setup(t, q)

				// Delete is idempotent
				err := q.DeletePermission(ctx, id)
				require.NoError(t, err)

				// Verify it's gone
				_, err = q.GetPermissionByID(ctx, id)
				assert.ErrorIs(t, err, pgx.ErrNoRows)
			})
		})
	}
}

// ============================================================
// Test: CreateRole
// ============================================================

func TestCreateRole(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) CreateRoleParams
		wantErr  bool
		checkErr func(t *testing.T, err error)
	}{
		{
			name: "success",
			setup: func(t *testing.T, q *Queries) CreateRoleParams {
				return CreateRoleParams{
					ID:          generateTestID(),
					Name:        "custom_role",
					Description: strPtr("Custom test role"),
				}
			},
			wantErr: false,
		},
		{
			name: "success_without_description",
			setup: func(t *testing.T, q *Queries) CreateRoleParams {
				return CreateRoleParams{
					ID:   generateTestID(),
					Name: "test_user_role",
				}
			},
			wantErr: false,
		},
		{
			name: "duplicate_id",
			setup: func(t *testing.T, q *Queries) CreateRoleParams {
				existingID := generateTestID()
				CreateTestRole(t, q, CreateTestRoleOptions{ID: &existingID})
				return CreateRoleParams{
					ID:   existingID,
					Name: "different",
				}
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.True(t, IsUniqueViolation(err), "expected unique violation, got: %v", err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				params := tt.setup(t, q)

				role, err := q.CreateRole(ctx, params)

				if tt.wantErr {
					require.Error(t, err)
					if tt.checkErr != nil {
						tt.checkErr(t, err)
					}
					return
				}

				require.NoError(t, err)
				assert.Equal(t, params.ID, role.ID)
				assert.Equal(t, params.Name, role.Name)
				assert.True(t, role.CreatedAt.Valid)

				// Verify role was created
				fetched, err := q.GetRoleByID(ctx, params.ID)
				require.NoError(t, err)
				assert.Equal(t, params.ID, fetched.ID)
			})
		})
	}
}

// ============================================================
// Test: GetRoleByID
// ============================================================

func TestGetRoleByID(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) string
		wantErr  bool
		checkErr func(t *testing.T, err error)
		validate func(t *testing.T, role Role)
	}{
		{
			name: "found",
			setup: func(t *testing.T, q *Queries) string {
				return CreateTestRole(t, q, CreateTestRoleOptions{})
			},
			wantErr: false,
			validate: func(t *testing.T, role Role) {
				assert.NotEmpty(t, role.ID)
				assert.NotEmpty(t, role.Name)
				assert.True(t, role.CreatedAt.Valid)
			},
		},
		{
			name: "not_found",
			setup: func(t *testing.T, q *Queries) string {
				return "non-existent-id"
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.True(t, errors.Is(err, pgx.ErrNoRows), "expected ErrNoRows, got: %v", err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				id := tt.setup(t, q)

				role, err := q.GetRoleByID(ctx, id)

				if tt.wantErr {
					require.Error(t, err)
					if tt.checkErr != nil {
						tt.checkErr(t, err)
					}
					return
				}

				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, role)
				}
			})
		})
	}
}

// ============================================================
// Test: GetRoleByName
// ============================================================

func TestGetRoleByName(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) string
		wantErr  bool
		checkErr func(t *testing.T, err error)
		validate func(t *testing.T, role Role)
	}{
		{
			name: "found",
			setup: func(t *testing.T, q *Queries) string {
				roleName := "test_find_role"
				CreateTestRole(t, q, CreateTestRoleOptions{Name: &roleName})
				return roleName
			},
			wantErr: false,
			validate: func(t *testing.T, role Role) {
				assert.Equal(t, "test_find_role", role.Name)
			},
		},
		{
			name: "not_found",
			setup: func(t *testing.T, q *Queries) string {
				return "nonexistent"
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.True(t, errors.Is(err, pgx.ErrNoRows), "expected ErrNoRows, got: %v", err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				name := tt.setup(t, q)

				role, err := q.GetRoleByName(ctx, name)

				if tt.wantErr {
					require.Error(t, err)
					if tt.checkErr != nil {
						tt.checkErr(t, err)
					}
					return
				}

				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, role)
				}
			})
		})
	}
}

// ============================================================
// Test: ListRoles
// ============================================================

func TestListRoles(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries)
		params   ListRolesParams
		validate func(t *testing.T, results []ListRolesRow)
	}{
		{
			name:   "seeded_roles_exist",
			setup:  func(t *testing.T, q *Queries) {},
			params: ListRolesParams{Limit: 100, Offset: 0},
			validate: func(t *testing.T, results []ListRolesRow) {
				// Migration seeds 3 roles (admin, coordinator, viewer)
				assert.GreaterOrEqual(t, len(results), 3)
			},
		},
		{
			name: "with_counts",
			setup: func(t *testing.T, q *Queries) {
				// Create role with permissions and users
				roleID := CreateTestRole(t, q, CreateTestRoleOptions{Name: strPtr("test_count_role")})
				permID1 := CreateTestPermission(t, q, CreateTestPermissionOptions{})
				permID2 := CreateTestPermission(t, q, CreateTestPermissionOptions{})
				AssignTestPermissionToRole(t, q, roleID, permID1)
				AssignTestPermissionToRole(t, q, roleID, permID2)

				userID1 := CreateTestUser(t, q, CreateTestUserOptions{})
				userID2 := CreateTestUser(t, q, CreateTestUserOptions{})
				AssignTestRoleToUser(t, q, userID1, roleID)
				AssignTestRoleToUser(t, q, userID2, roleID)
			},
			params: ListRolesParams{Limit: 100, Offset: 0},
			validate: func(t *testing.T, results []ListRolesRow) {
				// Find our test role among seeded and created roles
				var testRole *ListRolesRow
				for i := range results {
					if results[i].Name == "test_count_role" {
						testRole = &results[i]
						break
					}
				}
				require.NotNil(t, testRole, "test_count_role not found")
				assert.Equal(t, int64(2), testRole.PermissionCount)
				assert.Equal(t, int64(2), testRole.UserCount)
			},
		},
		{
			name: "pagination",
			setup: func(t *testing.T, q *Queries) {
				for i := 0; i < 5; i++ {
					name := fmt.Sprintf("paginated_role_%d", i)
					CreateTestRole(t, q, CreateTestRoleOptions{Name: &name})
				}
			},
			params: ListRolesParams{Limit: 3, Offset: 0},
			validate: func(t *testing.T, results []ListRolesRow) {
				assert.Len(t, results, 3)
				// Total includes seeded + new ones
				assert.GreaterOrEqual(t, results[0].TotalCount, int64(8))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				tt.setup(t, q)

				results, err := q.ListRoles(ctx, tt.params)

				require.NoError(t, err)
				tt.validate(t, results)
			})
		})
	}
}

// ============================================================
// Test: UpdateRole
// ============================================================

func TestUpdateRole(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) UpdateRoleParams
		validate func(t *testing.T, q *Queries, params UpdateRoleParams)
	}{
		{
			name: "update_name_only",
			setup: func(t *testing.T, q *Queries) UpdateRoleParams {
				id := CreateTestRole(t, q, CreateTestRoleOptions{Name: strPtr("old_name")})
				return UpdateRoleParams{
					ID:   id,
					Name: "new_name",
				}
			},
			validate: func(t *testing.T, q *Queries, params UpdateRoleParams) {
				role, err := q.GetRoleByID(context.Background(), params.ID)
				require.NoError(t, err)
				assert.Equal(t, params.Name, role.Name)
			},
		},
		{
			name: "update_full",
			setup: func(t *testing.T, q *Queries) UpdateRoleParams {
				id := CreateTestRole(t, q, CreateTestRoleOptions{})
				return UpdateRoleParams{
					ID:          id,
					Name:        "updated_role",
					Description: strPtr("Updated description"),
				}
			},
			validate: func(t *testing.T, q *Queries, params UpdateRoleParams) {
				role, err := q.GetRoleByID(context.Background(), params.ID)
				require.NoError(t, err)
				assert.Equal(t, params.Name, role.Name)
				assert.Equal(t, *params.Description, *role.Description)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				params := tt.setup(t, q)

				role, err := q.UpdateRole(ctx, params)

				require.NoError(t, err)
				assert.Equal(t, params.ID, role.ID)
				if tt.validate != nil {
					tt.validate(t, q, params)
				}
			})
		})
	}
}

// ============================================================
// Test: DeleteRole
// ============================================================

func TestDeleteRole(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, q *Queries) string
	}{
		{
			name: "existing",
			setup: func(t *testing.T, q *Queries) string {
				return CreateTestRole(t, q, CreateTestRoleOptions{})
			},
		},
		{
			name: "non_existent",
			setup: func(t *testing.T, q *Queries) string {
				return "non-existent-id"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				id := tt.setup(t, q)

				// Delete is idempotent
				err := q.DeleteRole(ctx, id)
				require.NoError(t, err)

				// Verify it's gone
				_, err = q.GetRoleByID(ctx, id)
				assert.ErrorIs(t, err, pgx.ErrNoRows)
			})
		})
	}
}

// ============================================================
// Test: AssignPermissionToRole
// ============================================================

func TestAssignPermissionToRole(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) AssignPermissionToRoleParams
		wantErr  bool
		checkErr func(t *testing.T, err error)
	}{
		{
			name: "success",
			setup: func(t *testing.T, q *Queries) AssignPermissionToRoleParams {
				roleID := CreateTestRole(t, q, CreateTestRoleOptions{})
				permID := CreateTestPermission(t, q, CreateTestPermissionOptions{})
				return AssignPermissionToRoleParams{
					RoleID:       roleID,
					PermissionID: permID,
				}
			},
			wantErr: false,
		},
		{
			name: "duplicate_assignment",
			setup: func(t *testing.T, q *Queries) AssignPermissionToRoleParams {
				roleID := CreateTestRole(t, q, CreateTestRoleOptions{})
				permID := CreateTestPermission(t, q, CreateTestPermissionOptions{})
				AssignTestPermissionToRole(t, q, roleID, permID)
				return AssignPermissionToRoleParams{
					RoleID:       roleID,
					PermissionID: permID,
				}
			},
			wantErr: false, // ON CONFLICT DO NOTHING
		},
		{
			name: "invalid_role_id",
			setup: func(t *testing.T, q *Queries) AssignPermissionToRoleParams {
				permID := CreateTestPermission(t, q, CreateTestPermissionOptions{})
				return AssignPermissionToRoleParams{
					RoleID:       "non-existent-role",
					PermissionID: permID,
				}
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.True(t, IsForeignKeyViolation(err), "expected foreign key violation, got: %v", err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				params := tt.setup(t, q)

				err := q.AssignPermissionToRole(ctx, params)

				if tt.wantErr {
					require.Error(t, err)
					if tt.checkErr != nil {
						tt.checkErr(t, err)
					}
					return
				}

				require.NoError(t, err)

				// Verify assignment
				perms, err := q.ListPermissionsForRole(ctx, params.RoleID)
				require.NoError(t, err)
				found := false
				for _, p := range perms {
					if p.ID == params.PermissionID {
						found = true
						break
					}
				}
				assert.True(t, found, "permission not found in role")
			})
		})
	}
}

// ============================================================
// Test: BatchAssignPermissionsToRole
// ============================================================

func TestBatchAssignPermissionsToRole(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) BatchAssignPermissionsToRoleParams
		wantErr  bool
		validate func(t *testing.T, q *Queries, params BatchAssignPermissionsToRoleParams)
	}{
		{
			name: "success",
			setup: func(t *testing.T, q *Queries) BatchAssignPermissionsToRoleParams {
				roleID := CreateTestRole(t, q, CreateTestRoleOptions{})
				perm1 := CreateTestPermission(t, q, CreateTestPermissionOptions{})
				perm2 := CreateTestPermission(t, q, CreateTestPermissionOptions{})
				perm3 := CreateTestPermission(t, q, CreateTestPermissionOptions{})
				return BatchAssignPermissionsToRoleParams{
					RoleID:        roleID,
					PermissionIds: []string{perm1, perm2, perm3},
				}
			},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, params BatchAssignPermissionsToRoleParams) {
				perms, err := q.ListPermissionsForRole(context.Background(), params.RoleID)
				require.NoError(t, err)
				assert.Len(t, perms, 3)
			},
		},
		{
			name: "empty_list",
			setup: func(t *testing.T, q *Queries) BatchAssignPermissionsToRoleParams {
				roleID := CreateTestRole(t, q, CreateTestRoleOptions{})
				return BatchAssignPermissionsToRoleParams{
					RoleID:        roleID,
					PermissionIds: []string{},
				}
			},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, params BatchAssignPermissionsToRoleParams) {
				perms, err := q.ListPermissionsForRole(context.Background(), params.RoleID)
				require.NoError(t, err)
				assert.Len(t, perms, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				params := tt.setup(t, q)

				err := q.BatchAssignPermissionsToRole(ctx, params)

				if tt.wantErr {
					require.Error(t, err)
					return
				}

				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, q, params)
				}
			})
		})
	}
}

// ============================================================
// Test: ListPermissionsForRole
// ============================================================

func TestListPermissionsForRole(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) string
		validate func(t *testing.T, perms []Permission)
	}{
		{
			name: "no_permissions",
			setup: func(t *testing.T, q *Queries) string {
				return CreateTestRole(t, q, CreateTestRoleOptions{})
			},
			validate: func(t *testing.T, perms []Permission) {
				assert.Len(t, perms, 0)
			},
		},
		{
			name: "multiple_permissions",
			setup: func(t *testing.T, q *Queries) string {
				roleID := CreateTestRole(t, q, CreateTestRoleOptions{})
				perm1 := CreateTestPermission(t, q, CreateTestPermissionOptions{
					Resource: strPtr("clients"),
					Action:   strPtr("read"),
				})
				perm2 := CreateTestPermission(t, q, CreateTestPermissionOptions{
					Resource: strPtr("clients"),
					Action:   strPtr("write"),
				})
				AssignTestPermissionToRole(t, q, roleID, perm1)
				AssignTestPermissionToRole(t, q, roleID, perm2)
				return roleID
			},
			validate: func(t *testing.T, perms []Permission) {
				assert.Len(t, perms, 2)
				// Verify ordering by resource, action
				assert.Equal(t, "clients", perms[0].Resource)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				roleID := tt.setup(t, q)

				perms, err := q.ListPermissionsForRole(ctx, roleID)

				require.NoError(t, err)
				tt.validate(t, perms)
			})
		})
	}
}

// ============================================================
// Test: RemovePermissionFromRole
// ============================================================

func TestRemovePermissionFromRole(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, q *Queries) RemovePermissionFromRoleParams
	}{
		{
			name: "existing_assignment",
			setup: func(t *testing.T, q *Queries) RemovePermissionFromRoleParams {
				roleID := CreateTestRole(t, q, CreateTestRoleOptions{})
				permID := CreateTestPermission(t, q, CreateTestPermissionOptions{})
				AssignTestPermissionToRole(t, q, roleID, permID)
				return RemovePermissionFromRoleParams{
					RoleID:       roleID,
					PermissionID: permID,
				}
			},
		},
		{
			name: "non_existent_assignment",
			setup: func(t *testing.T, q *Queries) RemovePermissionFromRoleParams {
				roleID := CreateTestRole(t, q, CreateTestRoleOptions{})
				permID := CreateTestPermission(t, q, CreateTestPermissionOptions{})
				return RemovePermissionFromRoleParams{
					RoleID:       roleID,
					PermissionID: permID,
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				params := tt.setup(t, q)

				err := q.RemovePermissionFromRole(ctx, params)
				require.NoError(t, err)

				// Verify removal
				perms, err := q.ListPermissionsForRole(ctx, params.RoleID)
				require.NoError(t, err)
				for _, p := range perms {
					assert.NotEqual(t, params.PermissionID, p.ID)
				}
			})
		})
	}
}

// ============================================================
// Test: DeleteAllPermissionsFromRole
// ============================================================

func TestDeleteAllPermissionsFromRole(t *testing.T) {
	t.Run("removes_all_permissions", func(t *testing.T) {
		runTestWithTx(t, func(t *testing.T, q *Queries) {
			ctx := context.Background()

			roleID := CreateTestRole(t, q, CreateTestRoleOptions{})
			perm1 := CreateTestPermission(t, q, CreateTestPermissionOptions{})
			perm2 := CreateTestPermission(t, q, CreateTestPermissionOptions{})
			AssignTestPermissionToRole(t, q, roleID, perm1)
			AssignTestPermissionToRole(t, q, roleID, perm2)

			// Verify we have permissions
			perms, err := q.ListPermissionsForRole(ctx, roleID)
			require.NoError(t, err)
			assert.Len(t, perms, 2)

			// Delete all
			err = q.DeleteAllPermissionsFromRole(ctx, roleID)
			require.NoError(t, err)

			// Verify all removed
			perms, err = q.ListPermissionsForRole(ctx, roleID)
			require.NoError(t, err)
			assert.Len(t, perms, 0)
		})
	})
}

// ============================================================
// Test: AssignRoleToUser
// ============================================================

func TestAssignRoleToUser(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) AssignRoleToUserParams
		wantErr  bool
		checkErr func(t *testing.T, err error)
	}{
		{
			name: "success",
			setup: func(t *testing.T, q *Queries) AssignRoleToUserParams {
				userID := CreateTestUser(t, q, CreateTestUserOptions{})
				roleID := CreateTestRole(t, q, CreateTestRoleOptions{})
				return AssignRoleToUserParams{
					UserID: userID,
					RoleID: roleID,
				}
			},
			wantErr: false,
		},
		{
			name: "update_existing_role",
			setup: func(t *testing.T, q *Queries) AssignRoleToUserParams {
				userID := CreateTestUser(t, q, CreateTestUserOptions{})
				role1 := CreateTestRole(t, q, CreateTestRoleOptions{Name: strPtr("role1")})
				role2 := CreateTestRole(t, q, CreateTestRoleOptions{Name: strPtr("role2")})
				AssignTestRoleToUser(t, q, userID, role1)
				return AssignRoleToUserParams{
					UserID: userID,
					RoleID: role2,
				}
			},
			wantErr: false,
		},
		{
			name: "invalid_user_id",
			setup: func(t *testing.T, q *Queries) AssignRoleToUserParams {
				roleID := CreateTestRole(t, q, CreateTestRoleOptions{})
				return AssignRoleToUserParams{
					UserID: "non-existent-user",
					RoleID: roleID,
				}
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.True(t, IsForeignKeyViolation(err), "expected foreign key violation, got: %v", err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				params := tt.setup(t, q)

				err := q.AssignRoleToUser(ctx, params)

				if tt.wantErr {
					require.Error(t, err)
					if tt.checkErr != nil {
						tt.checkErr(t, err)
					}
					return
				}

				require.NoError(t, err)

				// Verify assignment
				role, err := q.GetRoleForUser(ctx, params.UserID)
				require.NoError(t, err)
				assert.Equal(t, params.RoleID, role.ID)
			})
		})
	}
}

// ============================================================
// Test: GetRoleForUser
// ============================================================

func TestGetRoleForUser(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) string
		wantErr  bool
		checkErr func(t *testing.T, err error)
		validate func(t *testing.T, role Role)
	}{
		{
			name: "found",
			setup: func(t *testing.T, q *Queries) string {
				userID := CreateTestUser(t, q, CreateTestUserOptions{})
				roleID := CreateTestRole(t, q, CreateTestRoleOptions{})
				AssignTestRoleToUser(t, q, userID, roleID)
				return userID
			},
			wantErr: false,
			validate: func(t *testing.T, role Role) {
				assert.NotEmpty(t, role.Name)
			},
		},
		{
			name: "not_found",
			setup: func(t *testing.T, q *Queries) string {
				return CreateTestUser(t, q, CreateTestUserOptions{})
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.True(t, errors.Is(err, pgx.ErrNoRows), "expected ErrNoRows, got: %v", err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				userID := tt.setup(t, q)

				role, err := q.GetRoleForUser(ctx, userID)

				if tt.wantErr {
					require.Error(t, err)
					if tt.checkErr != nil {
						tt.checkErr(t, err)
					}
					return
				}

				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, role)
				}
			})
		})
	}
}

// ============================================================
// Test: RemoveRoleFromUser
// ============================================================

func TestRemoveRoleFromUser(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, q *Queries) string
	}{
		{
			name: "existing_assignment",
			setup: func(t *testing.T, q *Queries) string {
				userID := CreateTestUser(t, q, CreateTestUserOptions{})
				roleID := CreateTestRole(t, q, CreateTestRoleOptions{})
				AssignTestRoleToUser(t, q, userID, roleID)
				return userID
			},
		},
		{
			name: "non_existent_assignment",
			setup: func(t *testing.T, q *Queries) string {
				return CreateTestUser(t, q, CreateTestUserOptions{})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				userID := tt.setup(t, q)

				err := q.RemoveRoleFromUser(ctx, userID)
				require.NoError(t, err)

				// Verify removal
				_, err = q.GetRoleForUser(ctx, userID)
				assert.ErrorIs(t, err, pgx.ErrNoRows)
			})
		})
	}
}

// ============================================================
// Test: ListUsersWithRole
// ============================================================

func TestListUsersWithRole(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) string
		validate func(t *testing.T, users []ListUsersWithRoleRow)
	}{
		{
			name: "no_users",
			setup: func(t *testing.T, q *Queries) string {
				return CreateTestRole(t, q, CreateTestRoleOptions{})
			},
			validate: func(t *testing.T, users []ListUsersWithRoleRow) {
				assert.Len(t, users, 0)
			},
		},
		{
			name: "multiple_users",
			setup: func(t *testing.T, q *Queries) string {
				roleID := CreateTestRole(t, q, CreateTestRoleOptions{})
				email1 := "alice@example.com"
				email2 := "bob@example.com"
				user1 := CreateTestUser(t, q, CreateTestUserOptions{Email: &email1})
				user2 := CreateTestUser(t, q, CreateTestUserOptions{Email: &email2})
				AssignTestRoleToUser(t, q, user1, roleID)
				AssignTestRoleToUser(t, q, user2, roleID)
				return roleID
			},
			validate: func(t *testing.T, users []ListUsersWithRoleRow) {
				assert.Len(t, users, 2)
				// Verify ordering by email
				assert.Equal(t, "alice@example.com", users[0].Email)
				assert.Equal(t, "bob@example.com", users[1].Email)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				roleID := tt.setup(t, q)

				users, err := q.ListUsersWithRole(ctx, roleID)

				require.NoError(t, err)
				tt.validate(t, users)
			})
		})
	}
}
