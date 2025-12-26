package db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================
// Test: CreateEmployee
// ============================================================

func TestCreateEmployee(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) CreateEmployeeParams
		wantErr  bool
		checkErr func(t *testing.T, err error)
	}{
		{
			name: "success",
			setup: func(t *testing.T, q *Queries) CreateEmployeeParams {
				userID := CreateTestUser(t, q, CreateTestUserOptions{})
				locationID := CreateTestLocation(t, q, CreateTestLocationOptions{})
				return CreateEmployeeParams{
					ID:          generateTestID(),
					UserID:      userID,
					FirstName:   "John",
					LastName:    "Doe",
					Bsn:         "123456789",
					DateOfBirth: toPgDate(time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)),
					PhoneNumber: "+31612345678",
					Gender:      GenderEnumMale,
					LocationID:  locationID,
				}
			},
			wantErr: false,
		},
		{
			name: "duplicate_bsn",
			setup: func(t *testing.T, q *Queries) CreateEmployeeParams {
				userID := CreateTestUser(t, q, CreateTestUserOptions{})
				bsn := "DUPLICATE123"
				CreateTestEmployee(t, q, CreateTestEmployeeOptions{UserID: userID, Bsn: &bsn})
				// Return params that will fail
				userID2 := CreateTestUser(t, q, CreateTestUserOptions{})
				locationID := CreateTestLocation(t, q, CreateTestLocationOptions{})
				return CreateEmployeeParams{
					ID:          generateTestID(),
					UserID:      userID2,
					FirstName:   "Jane",
					LastName:    "Smith",
					Bsn:         bsn,
					DateOfBirth: toPgDate(time.Date(1992, 2, 2, 0, 0, 0, 0, time.UTC)),
					PhoneNumber: "+31687654321",
					Gender:      GenderEnumFemale,
					LocationID:  locationID,
				}
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.True(t, IsUniqueViolation(err), "expected unique violation, got: %v", err)
			},
		},
		{
			name: "duplicate_id",
			setup: func(t *testing.T, q *Queries) CreateEmployeeParams {
				userID := CreateTestUser(t, q, CreateTestUserOptions{})
				existingID := generateTestID()
				CreateTestEmployee(t, q, CreateTestEmployeeOptions{ID: &existingID, UserID: userID})
				userID2 := CreateTestUser(t, q, CreateTestUserOptions{})
				locationID := CreateTestLocation(t, q, CreateTestLocationOptions{})
				return CreateEmployeeParams{
					ID:          existingID,
					UserID:      userID2,
					FirstName:   "Alice",
					LastName:    "Brown",
					Bsn:         "987654321",
					DateOfBirth: toPgDate(time.Date(1985, 5, 5, 0, 0, 0, 0, time.UTC)),
					PhoneNumber: "+31611223344",
					Gender:      GenderEnumOther,
					LocationID:  locationID,
				}
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.True(t, IsUniqueViolation(err), "expected unique violation, got: %v", err)
			},
		},
		{
			name: "invalid_user_id",
			setup: func(t *testing.T, q *Queries) CreateEmployeeParams {
				locationID := CreateTestLocation(t, q, CreateTestLocationOptions{})
				return CreateEmployeeParams{
					ID:          generateTestID(),
					UserID:      "non-existent-user-id",
					FirstName:   "Bob",
					LastName:    "Wilson",
					Bsn:         "555666777",
					DateOfBirth: toPgDate(time.Date(1980, 10, 10, 0, 0, 0, 0, time.UTC)),
					PhoneNumber: "+31699887766",
					Gender:      GenderEnumMale,
					LocationID:  locationID,
				}
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.True(t, IsForeignKeyViolation(err), "expected FK violation, got: %v", err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				params := tt.setup(t, q)

				err := q.CreateEmployee(ctx, params)

				if tt.wantErr {
					require.Error(t, err)
					if tt.checkErr != nil {
						tt.checkErr(t, err)
					}
					return
				}

				require.NoError(t, err)
			})
		})
	}
}

// ============================================================
// Test: ListEmployees
// ============================================================

func TestListEmployees(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries)
		params   ListEmployeesParams
		validate func(t *testing.T, results []ListEmployeesRow)
	}{
		{
			name:   "empty",
			setup:  func(t *testing.T, q *Queries) {},
			params: ListEmployeesParams{Limit: 10, Offset: 0},
			validate: func(t *testing.T, results []ListEmployeesRow) {
				assert.Len(t, results, 0)
			},
		},
		{
			name: "with_data",
			setup: func(t *testing.T, q *Queries) {
				userID := CreateTestUser(t, q, CreateTestUserOptions{})
				CreateTestEmployee(t, q, CreateTestEmployeeOptions{UserID: userID})
			},
			params: ListEmployeesParams{Limit: 10, Offset: 0},
			validate: func(t *testing.T, results []ListEmployeesRow) {
				assert.Len(t, results, 1)
				assert.Equal(t, int64(1), results[0].TotalCount)
			},
		},
		{
			name: "with_pagination",
			setup: func(t *testing.T, q *Queries) {
				for i := 0; i < 5; i++ {
					userID := CreateTestUser(t, q, CreateTestUserOptions{})
					CreateTestEmployee(t, q, CreateTestEmployeeOptions{UserID: userID})
				}
			},
			params: ListEmployeesParams{Limit: 2, Offset: 0},
			validate: func(t *testing.T, results []ListEmployeesRow) {
				assert.Len(t, results, 2)
				assert.Equal(t, int64(5), results[0].TotalCount)
			},
		},
		{
			name: "with_search_by_first_name",
			setup: func(t *testing.T, q *Queries) {
				userID1 := CreateTestUser(t, q, CreateTestUserOptions{})
				CreateTestEmployee(
					t,
					q,
					CreateTestEmployeeOptions{UserID: userID1, FirstName: strPtr("Alice")},
				)
				userID2 := CreateTestUser(t, q, CreateTestUserOptions{})
				CreateTestEmployee(
					t,
					q,
					CreateTestEmployeeOptions{UserID: userID2, FirstName: strPtr("Bob")},
				)
			},
			params: ListEmployeesParams{Limit: 10, Offset: 0, Search: strPtr("Alice")},
			validate: func(t *testing.T, results []ListEmployeesRow) {
				assert.Len(t, results, 1)
				assert.Equal(t, "Alice", results[0].FirstName)
			},
		},
		{
			name: "with_search_by_last_name",
			setup: func(t *testing.T, q *Queries) {
				userID1 := CreateTestUser(t, q, CreateTestUserOptions{})
				CreateTestEmployee(
					t,
					q,
					CreateTestEmployeeOptions{UserID: userID1, LastName: strPtr("Smith")},
				)
				userID2 := CreateTestUser(t, q, CreateTestUserOptions{})
				CreateTestEmployee(
					t,
					q,
					CreateTestEmployeeOptions{UserID: userID2, LastName: strPtr("Johnson")},
				)
			},
			params: ListEmployeesParams{Limit: 10, Offset: 0, Search: strPtr("Smith")},
			validate: func(t *testing.T, results []ListEmployeesRow) {
				assert.Len(t, results, 1)
				assert.Equal(t, "Smith", results[0].LastName)
			},
		},
		{
			name: "with_search_full_name",
			setup: func(t *testing.T, q *Queries) {
				userID1 := CreateTestUser(t, q, CreateTestUserOptions{})
				CreateTestEmployee(
					t,
					q,
					CreateTestEmployeeOptions{
						UserID:    userID1,
						FirstName: strPtr("John"),
						LastName:  strPtr("Doe"),
					},
				)
				userID2 := CreateTestUser(t, q, CreateTestUserOptions{})
				CreateTestEmployee(
					t,
					q,
					CreateTestEmployeeOptions{
						UserID:    userID2,
						FirstName: strPtr("Jane"),
						LastName:  strPtr("Smith"),
					},
				)
			},
			params: ListEmployeesParams{Limit: 10, Offset: 0, Search: strPtr("John Doe")},
			validate: func(t *testing.T, results []ListEmployeesRow) {
				assert.Len(t, results, 1)
				assert.Equal(t, "John", results[0].FirstName)
				assert.Equal(t, "Doe", results[0].LastName)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				tt.setup(t, q)

				results, err := q.ListEmployees(ctx, tt.params)

				require.NoError(t, err)
				tt.validate(t, results)
			})
		})
	}
}
