package db

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================
// Test: CreateReferringOrg
// ============================================================

func TestCreateReferringOrg(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) CreateReferringOrgParams
		wantErr  bool
		checkErr func(t *testing.T, err error)
	}{
		{
			name: "success",
			setup: func(t *testing.T, q *Queries) CreateReferringOrgParams {
				return CreateReferringOrgParams{
					ID:            generateTestID(),
					Name:          "Test Org Alpha",
					ContactPerson: "John Doe",
					PhoneNumber:   "+31612345678",
					Email:         "alpha@example.com",
				}
			},
			wantErr: false,
		},
		{
			name: "duplicate_id",
			setup: func(t *testing.T, q *Queries) CreateReferringOrgParams {
				id := generateTestID()
				CreateTestReferringOrg(t, q, CreateTestReferringOrgOptions{ID: &id})
				return CreateReferringOrgParams{
					ID:            id,
					Name:          "Different Org",
					ContactPerson: "Somebody Else",
					PhoneNumber:   "+31600000000",
					Email:         "different@example.com",
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

				err := q.CreateReferringOrg(ctx, params)

				if tt.wantErr {
					require.Error(t, err)
					if tt.checkErr != nil {
						tt.checkErr(t, err)
					}
					return
				}

				require.NoError(t, err)

				// Verify org was created
				org, err := q.GetReferringOrgByID(ctx, params.ID)
				require.NoError(t, err)
				assert.Equal(t, params.ID, org.ID)
				assert.Equal(t, params.Name, org.Name)
			})
		})
	}
}

// ============================================================
// Test: GetReferringOrgByID
// ============================================================

func TestGetReferringOrgByID(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) string // returns ID to query
		wantErr  bool
		checkErr func(t *testing.T, err error)
		validate func(t *testing.T, org ReferringOrg)
	}{
		{
			name: "found",
			setup: func(t *testing.T, q *Queries) string {
				return CreateTestReferringOrg(t, q, CreateTestReferringOrgOptions{})
			},
			wantErr: false,
			validate: func(t *testing.T, org ReferringOrg) {
				assert.NotEmpty(t, org.ID)
				assert.NotEmpty(t, org.Name)
				assert.NotEmpty(t, org.ContactPerson)
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

				org, err := q.GetReferringOrgByID(ctx, id)

				if tt.wantErr {
					require.Error(t, err)
					if tt.checkErr != nil {
						tt.checkErr(t, err)
					}
					return
				}

				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, org)
				}
			})
		})
	}
}

// ============================================================
// Test: ListReferringOrgs
// ============================================================

func TestListReferringOrgs(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) // create test data
		params   ListReferringOrgsParams
		validate func(t *testing.T, results []ListReferringOrgsRow)
	}{
		{
			name:   "empty",
			setup:  func(t *testing.T, q *Queries) {},
			params: ListReferringOrgsParams{Limit: 10, Offset: 0},
			validate: func(t *testing.T, results []ListReferringOrgsRow) {
				assert.Len(t, results, 0)
			},
		},
		{
			name: "with_search",
			setup: func(t *testing.T, q *Queries) {
				CreateTestReferringOrg(t, q, CreateTestReferringOrgOptions{Name: strPtr("Alpha Org")})
				CreateTestReferringOrg(t, q, CreateTestReferringOrgOptions{Name: strPtr("Beta Org")})
			},
			params: ListReferringOrgsParams{Limit: 10, Offset: 0, Search: strPtr("Alpha")},
			validate: func(t *testing.T, results []ListReferringOrgsRow) {
				assert.Len(t, results, 1)
				assert.Equal(t, "Alpha Org", results[0].Name)
				assert.Equal(t, int64(1), results[0].TotalCount)
			},
		},
		{
			name: "pagination",
			setup: func(t *testing.T, q *Queries) {
				for i := 0; i < 5; i++ {
					name := fmt.Sprintf("Org %d", i)
					CreateTestReferringOrg(t, q, CreateTestReferringOrgOptions{Name: &name})
				}
			},
			params: ListReferringOrgsParams{Limit: 2, Offset: 0},
			validate: func(t *testing.T, results []ListReferringOrgsRow) {
				assert.Len(t, results, 2)
				assert.Equal(t, int64(5), results[0].TotalCount)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				tt.setup(t, q)

				results, err := q.ListReferringOrgs(ctx, tt.params)

				require.NoError(t, err)
				tt.validate(t, results)
			})
		})
	}
}

// ============================================================
// Test: ListReferringOrgsWithCounts
// ============================================================

func TestListReferringOrgsWithCounts(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) // create test data
		params   ListReferringOrgsWithCountsParams
		validate func(t *testing.T, results []ListReferringOrgsWithCountsRow)
	}{
		{
			name: "with_counts",
			setup: func(t *testing.T, q *Queries) {
				orgID := CreateTestReferringOrg(t, q, CreateTestReferringOrgOptions{Name: strPtr("Counted Org")})

				// Create 2 in_care, 1 waiting_list, 3 discharged
				for i := 0; i < 2; i++ {
					status := ClientStatusEnumInCare
					now := time.Now()
					CreateTestClientWithDependenciesCustom(t, q, CreateTestClientOptions{
						Status:         &status,
						ReferringOrgID: &orgID,
						CareStartDate:  &now,
					})
				}
				for i := 0; i < 1; i++ {
					status := ClientStatusEnumWaitingList
					CreateTestClientWithDependenciesCustom(t, q, CreateTestClientOptions{
						Status:         &status,
						ReferringOrgID: &orgID,
					})
				}
				for i := 0; i < 3; i++ {
					status := ClientStatusEnumDischarged
					// Discharged clients need care_start_date and discharge_date/status
					now := time.Now()
					dischargeStatus := DischargeStatusEnumCompleted
					reason := DischargeReasonEnumTreatmentCompleted
					CreateTestClientWithDependenciesCustom(t, q, CreateTestClientOptions{
						Status:             &status,
						ReferringOrgID:     &orgID,
						CareStartDate:      &now,
						DischargeDate:      &now,
						DischargeStatus:    &dischargeStatus,
						ReasonForDischarge: &reason,
					})
				}
			},
			params: ListReferringOrgsWithCountsParams{Limit: 10, Offset: 0, Search: strPtr("Counted")},
			validate: func(t *testing.T, results []ListReferringOrgsWithCountsRow) {
				require.Len(t, results, 1)
				assert.Equal(t, int64(2), results[0].InCareCount)
				assert.Equal(t, int64(1), results[0].WaitingListCount)
				assert.Equal(t, int64(3), results[0].DischargedCount)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				tt.setup(t, q)

				results, err := q.ListReferringOrgsWithCounts(ctx, tt.params)

				require.NoError(t, err)
				tt.validate(t, results)
			})
		})
	}
}

// ============================================================
// Test: UpdateReferringOrg
// ============================================================

func TestUpdateReferringOrg(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) UpdateReferringOrgParams
		validate func(t *testing.T, q *Queries, params UpdateReferringOrgParams)
	}{
		{
			name: "partial_update",
			setup: func(t *testing.T, q *Queries) UpdateReferringOrgParams {
				id := CreateTestReferringOrg(t, q, CreateTestReferringOrgOptions{Name: strPtr("Old Name")})
				return UpdateReferringOrgParams{
					ID:   id,
					Name: strPtr("New Name"),
				}
			},
			validate: func(t *testing.T, q *Queries, params UpdateReferringOrgParams) {
				org, err := q.GetReferringOrgByID(context.Background(), params.ID)
				require.NoError(t, err)
				assert.Equal(t, *params.Name, org.Name)
				assert.NotEmpty(t, org.ContactPerson) // Should remain unchanged
			},
		},
		{
			name: "full_update",
			setup: func(t *testing.T, q *Queries) UpdateReferringOrgParams {
				id := CreateTestReferringOrg(t, q, CreateTestReferringOrgOptions{})
				return UpdateReferringOrgParams{
					ID:            id,
					Name:          strPtr("Full Update Org"),
					ContactPerson: strPtr("Jane Doe"),
					PhoneNumber:   strPtr("+31611111111"),
					Email:         strPtr("jane@example.com"),
				}
			},
			validate: func(t *testing.T, q *Queries, params UpdateReferringOrgParams) {
				org, err := q.GetReferringOrgByID(context.Background(), params.ID)
				require.NoError(t, err)
				assert.Equal(t, *params.Name, org.Name)
				assert.Equal(t, *params.ContactPerson, org.ContactPerson)
				assert.Equal(t, *params.PhoneNumber, org.PhoneNumber)
				assert.Equal(t, *params.Email, org.Email)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				params := tt.setup(t, q)

				err := q.UpdateReferringOrg(ctx, params)

				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, q, params)
				}
			})
		})
	}
}

// ============================================================
// Test: DeleteReferringOrg
// ============================================================

func TestDeleteReferringOrg(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, q *Queries) string // returns ID to delete
	}{
		{
			name: "existing",
			setup: func(t *testing.T, q *Queries) string {
				return CreateTestReferringOrg(t, q, CreateTestReferringOrgOptions{})
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
				err := q.DeleteReferringOrg(ctx, id)
				require.NoError(t, err)

				// Verify it's gone
				_, err = q.GetReferringOrgByID(ctx, id)
				assert.ErrorIs(t, err, pgx.ErrNoRows)
			})
		})
	}
}

// Helper to create client with dependencies and custom options
func CreateTestClientWithDependenciesCustom(t *testing.T, q *Queries, opts CreateTestClientOptions) string {
	deps := CreateFullClientDependencyChain(t, q)
	opts.RegistrationFormID = deps.RegistrationFormID
	opts.IntakeFormID = deps.IntakeFormID
	opts.AssignedLocationID = deps.LocationID
	opts.CoordinatorID = deps.EmployeeID
	return CreateTestClient(t, q, opts)
}
