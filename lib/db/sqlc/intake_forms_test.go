package db

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================
// Test: CreateIntakeForm
// ============================================================

func TestCreateIntakeForm(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) CreateIntakeFormParams
		wantErr  bool
		checkErr func(t *testing.T, err error)
		validate func(t *testing.T, q *Queries, params CreateIntakeFormParams)
	}{
		{
			name: "success_with_all_fields",
			setup: func(t *testing.T, q *Queries) CreateIntakeFormParams {
				userID := CreateTestUser(t, q, CreateTestUserOptions{})
				locationID := CreateTestLocation(t, q, CreateTestLocationOptions{})
				employeeID := CreateTestEmployee(t, q, CreateTestEmployeeOptions{UserID: userID})
				regFormID := CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{})

				return CreateIntakeFormParams{
					ID:                      generateTestID(),
					RegistrationFormID:      regFormID,
					IntakeDate:              toPgDate(time.Now()),
					IntakeTime:              toPgTime(time.Date(0, 1, 1, 10, 30, 0, 0, time.UTC)),
					LocationID:              locationID,
					CoordinatorID:           employeeID,
					FamilySituation:         strPtr("Single parent household"),
					MainProvider:            strPtr("Primary care physician"),
					Limitations:             strPtr("None identified"),
					FocusAreas:              strPtr("Mental health support"),
					Notes:                   strPtr("Initial intake notes"),
					EvaluationIntervalWeeks: int32Ptr(4),
				}
			},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, params CreateIntakeFormParams) {
				ctx := context.Background()
				form, err := q.GetIntakeForm(ctx, params.ID)
				require.NoError(t, err)
				assert.Equal(t, params.RegistrationFormID, form.RegistrationFormID)
				assert.Equal(t, params.LocationID, form.LocationID)
				assert.Equal(t, params.CoordinatorID, form.CoordinatorID)
				assert.Equal(t, *params.FamilySituation, *form.FamilySituation)
				assert.Equal(t, *params.EvaluationIntervalWeeks, *form.EvaluationIntervalWeeks)
				assert.Equal(t, IntakeStatusEnumPending, form.Status) // Default status
			},
		},
		{
			name: "success_minimal_fields",
			setup: func(t *testing.T, q *Queries) CreateIntakeFormParams {
				userID := CreateTestUser(t, q, CreateTestUserOptions{})
				locationID := CreateTestLocation(t, q, CreateTestLocationOptions{})
				employeeID := CreateTestEmployee(t, q, CreateTestEmployeeOptions{UserID: userID})
				regFormID := CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{})

				return CreateIntakeFormParams{
					ID:                 generateTestID(),
					RegistrationFormID: regFormID,
					IntakeDate:         toPgDate(time.Now()),
					IntakeTime:         toPgTime(time.Date(0, 1, 1, 14, 0, 0, 0, time.UTC)),
					LocationID:         locationID,
					CoordinatorID:      employeeID,
				}
			},
			wantErr: false,
		},
		{
			name: "duplicate_registration_form",
			setup: func(t *testing.T, q *Queries) CreateIntakeFormParams {
				userID := CreateTestUser(t, q, CreateTestUserOptions{})
				locationID := CreateTestLocation(t, q, CreateTestLocationOptions{})
				employeeID := CreateTestEmployee(t, q, CreateTestEmployeeOptions{UserID: userID})
				regFormID := CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{})

				// Create first intake form
				CreateTestIntakeForm(t, q, CreateTestIntakeFormOptions{
					RegistrationFormID: regFormID,
					LocationID:         locationID,
					CoordinatorID:      employeeID,
				})

				// Try to create second with same registration form
				return CreateIntakeFormParams{
					ID:                 generateTestID(),
					RegistrationFormID: regFormID,
					IntakeDate:         toPgDate(time.Now()),
					IntakeTime:         toPgTime(time.Date(0, 1, 1, 10, 0, 0, 0, time.UTC)),
					LocationID:         locationID,
					CoordinatorID:      employeeID,
				}
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.True(t, IsUniqueViolation(err), "expected unique violation, got: %v", err)
			},
		},
		{
			name: "invalid_registration_form_fk",
			setup: func(t *testing.T, q *Queries) CreateIntakeFormParams {
				userID := CreateTestUser(t, q, CreateTestUserOptions{})
				locationID := CreateTestLocation(t, q, CreateTestLocationOptions{})
				employeeID := CreateTestEmployee(t, q, CreateTestEmployeeOptions{UserID: userID})

				return CreateIntakeFormParams{
					ID:                 generateTestID(),
					RegistrationFormID: "nonexistent-reg-form",
					IntakeDate:         toPgDate(time.Now()),
					IntakeTime:         toPgTime(time.Date(0, 1, 1, 10, 0, 0, 0, time.UTC)),
					LocationID:         locationID,
					CoordinatorID:      employeeID,
				}
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.True(t, IsForeignKeyViolation(err), "expected FK violation, got: %v", err)
			},
		},
		{
			name: "invalid_location_fk",
			setup: func(t *testing.T, q *Queries) CreateIntakeFormParams {
				userID := CreateTestUser(t, q, CreateTestUserOptions{})
				locationID := CreateTestLocation(t, q, CreateTestLocationOptions{})
				employeeID := CreateTestEmployee(t, q, CreateTestEmployeeOptions{
					UserID:     userID,
					LocationID: &locationID,
				})
				regFormID := CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{})

				return CreateIntakeFormParams{
					ID:                 generateTestID(),
					RegistrationFormID: regFormID,
					IntakeDate:         toPgDate(time.Now()),
					IntakeTime:         toPgTime(time.Date(0, 1, 1, 10, 0, 0, 0, time.UTC)),
					LocationID:         "nonexistent-location",
					CoordinatorID:      employeeID,
				}
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.True(t, IsForeignKeyViolation(err), "expected FK violation, got: %v", err)
			},
		},
		{
			name: "invalid_coordinator_fk",
			setup: func(t *testing.T, q *Queries) CreateIntakeFormParams {
				locationID := CreateTestLocation(t, q, CreateTestLocationOptions{})
				regFormID := CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{})

				return CreateIntakeFormParams{
					ID:                 generateTestID(),
					RegistrationFormID: regFormID,
					IntakeDate:         toPgDate(time.Now()),
					IntakeTime:         toPgTime(time.Date(0, 1, 1, 10, 0, 0, 0, time.UTC)),
					LocationID:         locationID,
					CoordinatorID:      "nonexistent-coordinator",
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

				err := q.CreateIntakeForm(ctx, params)

				if tt.wantErr {
					require.Error(t, err)
					if tt.checkErr != nil {
						tt.checkErr(t, err)
					}
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
// Test: GetIntakeForm
// ============================================================

func TestGetIntakeForm(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) string // returns ID to query
		wantErr  bool
		checkErr func(t *testing.T, err error)
		validate func(t *testing.T, form IntakeForm)
	}{
		{
			name: "found",
			setup: func(t *testing.T, q *Queries) string {
				userID := CreateTestUser(t, q, CreateTestUserOptions{})
				locationID := CreateTestLocation(t, q, CreateTestLocationOptions{})
				employeeID := CreateTestEmployee(t, q, CreateTestEmployeeOptions{UserID: userID})
				regFormID := CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{})

				return CreateTestIntakeForm(t, q, CreateTestIntakeFormOptions{
					RegistrationFormID: regFormID,
					LocationID:         locationID,
					CoordinatorID:      employeeID,
				})
			},
			wantErr: false,
			validate: func(t *testing.T, form IntakeForm) {
				assert.NotEmpty(t, form.ID)
				assert.NotEmpty(t, form.RegistrationFormID)
				assert.NotEmpty(t, form.LocationID)
				assert.NotEmpty(t, form.CoordinatorID)
				assert.True(t, form.CreatedAt.Valid)
				assert.Equal(t, IntakeStatusEnumPending, form.Status)
			},
		},
		{
			name: "not_found",
			setup: func(t *testing.T, q *Queries) string {
				return "nonexistent-id"
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

				form, err := q.GetIntakeForm(ctx, id)

				if tt.wantErr {
					require.Error(t, err)
					if tt.checkErr != nil {
						tt.checkErr(t, err)
					}
					return
				}

				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, form)
				}
			})
		})
	}
}

// ============================================================
// Test: GetIntakeFormWithDetails
// ============================================================

func TestGetIntakeFormWithDetails(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) string // returns ID to query
		wantErr  bool
		checkErr func(t *testing.T, err error)
		validate func(t *testing.T, result GetIntakeFormWithDetailsRow)
	}{
		{
			name: "found_with_all_details",
			setup: func(t *testing.T, q *Queries) string {
				orgID := CreateTestReferringOrg(t, q, CreateTestReferringOrgOptions{
					Name: strPtr("Test Hospital"),
				})
				userID := CreateTestUser(t, q, CreateTestUserOptions{})
				locationID := CreateTestLocation(t, q, CreateTestLocationOptions{
					Name: strPtr("Main Clinic"),
				})
				employeeID := CreateTestEmployee(t, q, CreateTestEmployeeOptions{
					UserID:    userID,
					FirstName: strPtr("John"),
					LastName:  strPtr("Coordinator"),
				})
				regFormID := CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{
					FirstName:      strPtr("Alice"),
					LastName:       strPtr("Client"),
					ReferringOrgID: &orgID,
				})

				return CreateTestIntakeForm(t, q, CreateTestIntakeFormOptions{
					RegistrationFormID: regFormID,
					LocationID:         locationID,
					CoordinatorID:      employeeID,
					FamilySituation:    strPtr("Test family situation"),
				})
			},
			wantErr: false,
			validate: func(t *testing.T, result GetIntakeFormWithDetailsRow) {
				assert.NotEmpty(t, result.ID)
				assert.Equal(t, "Alice", *result.ClientFirstName)
				assert.Equal(t, "Client", *result.ClientLastName)
				assert.Equal(t, "Test Hospital", *result.OrgName)
				assert.Equal(t, "Main Clinic", *result.LocationName)
				assert.Equal(t, "John", *result.CoordinatorFirstName)
				assert.Equal(t, "Coordinator", *result.CoordinatorLastName)
				assert.Equal(t, "Test family situation", *result.FamilySituation)
				assert.False(t, result.HasClient) // No client created yet
			},
		},
		{
			name: "found_without_referring_org",
			setup: func(t *testing.T, q *Queries) string {
				userID := CreateTestUser(t, q, CreateTestUserOptions{})
				locationID := CreateTestLocation(t, q, CreateTestLocationOptions{})
				employeeID := CreateTestEmployee(t, q, CreateTestEmployeeOptions{UserID: userID})
				regFormID := CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{
					FirstName: strPtr("Bob"),
					// No referring org
				})

				return CreateTestIntakeForm(t, q, CreateTestIntakeFormOptions{
					RegistrationFormID: regFormID,
					LocationID:         locationID,
					CoordinatorID:      employeeID,
				})
			},
			wantErr: false,
			validate: func(t *testing.T, result GetIntakeFormWithDetailsRow) {
				assert.Equal(t, "Bob", *result.ClientFirstName)
				assert.Nil(t, result.OrgName)
			},
		},
		{
			name: "with_client_created",
			setup: func(t *testing.T, q *Queries) string {
				// Create full client chain - this creates intake form internally
				_, deps := CreateTestClientWithDependencies(t, q)
				return deps.IntakeFormID
			},
			wantErr: false,
			validate: func(t *testing.T, result GetIntakeFormWithDetailsRow) {
				assert.True(t, result.HasClient)
			},
		},
		{
			name: "not_found",
			setup: func(t *testing.T, q *Queries) string {
				return "nonexistent-id"
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

				result, err := q.GetIntakeFormWithDetails(ctx, id)

				if tt.wantErr {
					require.Error(t, err)
					if tt.checkErr != nil {
						tt.checkErr(t, err)
					}
					return
				}

				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			})
		})
	}
}

// ============================================================
// Test: GetIntakeStats
// ============================================================

func TestGetIntakeStats(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries)
		validate func(t *testing.T, stats GetIntakeStatsRow)
	}{
		{
			name:  "empty_database",
			setup: func(t *testing.T, q *Queries) {},
			validate: func(t *testing.T, stats GetIntakeStatsRow) {
				assert.Equal(t, int64(0), stats.TotalCount)
				assert.Equal(t, int64(0), stats.PendingCount)
				assert.Equal(t, float64(0), stats.ConversionPercentage)
			},
		},
		{
			name: "with_pending_forms",
			setup: func(t *testing.T, q *Queries) {
				// Create 3 intake forms (default status is 'pending')
				for i := 0; i < 3; i++ {
					userID := CreateTestUser(t, q, CreateTestUserOptions{})
					locationID := CreateTestLocation(t, q, CreateTestLocationOptions{})
					employeeID := CreateTestEmployee(t, q, CreateTestEmployeeOptions{UserID: userID})
					regFormID := CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{})

					CreateTestIntakeForm(t, q, CreateTestIntakeFormOptions{
						RegistrationFormID: regFormID,
						LocationID:         locationID,
						CoordinatorID:      employeeID,
					})
				}
			},
			validate: func(t *testing.T, stats GetIntakeStatsRow) {
				assert.Equal(t, int64(3), stats.TotalCount)
				assert.Equal(t, int64(3), stats.PendingCount)
				assert.Equal(t, float64(0), stats.ConversionPercentage)
			},
		},
		{
			name: "with_mixed_statuses",
			setup: func(t *testing.T, q *Queries) {
				ctx := context.Background()

				// Create 2 pending
				for i := 0; i < 2; i++ {
					userID := CreateTestUser(t, q, CreateTestUserOptions{})
					locationID := CreateTestLocation(t, q, CreateTestLocationOptions{})
					employeeID := CreateTestEmployee(t, q, CreateTestEmployeeOptions{UserID: userID})
					regFormID := CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{})

					CreateTestIntakeForm(t, q, CreateTestIntakeFormOptions{
						RegistrationFormID: regFormID,
						LocationID:         locationID,
						CoordinatorID:      employeeID,
					})
				}

				// Create 2 completed
				for i := 0; i < 2; i++ {
					userID := CreateTestUser(t, q, CreateTestUserOptions{})
					locationID := CreateTestLocation(t, q, CreateTestLocationOptions{})
					employeeID := CreateTestEmployee(t, q, CreateTestEmployeeOptions{UserID: userID})
					regFormID := CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{})

					intakeID := CreateTestIntakeForm(t, q, CreateTestIntakeFormOptions{
						RegistrationFormID: regFormID,
						LocationID:         locationID,
						CoordinatorID:      employeeID,
					})
					q.UpdateIntakeFormStatus(ctx, UpdateIntakeFormStatusParams{
						ID:     intakeID,
						Status: IntakeStatusEnumCompleted,
					})
				}
			},
			validate: func(t *testing.T, stats GetIntakeStatsRow) {
				assert.Equal(t, int64(4), stats.TotalCount)
				assert.Equal(t, int64(2), stats.PendingCount)
				assert.Equal(t, float64(50), stats.ConversionPercentage)
			},
		},
		{
			name: "all_completed",
			setup: func(t *testing.T, q *Queries) {
				ctx := context.Background()

				for i := 0; i < 3; i++ {
					userID := CreateTestUser(t, q, CreateTestUserOptions{})
					locationID := CreateTestLocation(t, q, CreateTestLocationOptions{})
					employeeID := CreateTestEmployee(t, q, CreateTestEmployeeOptions{UserID: userID})
					regFormID := CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{})

					intakeID := CreateTestIntakeForm(t, q, CreateTestIntakeFormOptions{
						RegistrationFormID: regFormID,
						LocationID:         locationID,
						CoordinatorID:      employeeID,
					})
					q.UpdateIntakeFormStatus(ctx, UpdateIntakeFormStatusParams{
						ID:     intakeID,
						Status: IntakeStatusEnumCompleted,
					})
				}
			},
			validate: func(t *testing.T, stats GetIntakeStatsRow) {
				assert.Equal(t, int64(3), stats.TotalCount)
				assert.Equal(t, int64(0), stats.PendingCount)
				assert.Equal(t, float64(100), stats.ConversionPercentage)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				tt.setup(t, q)

				stats, err := q.GetIntakeStats(ctx)

				require.NoError(t, err)
				tt.validate(t, stats)
			})
		})
	}
}

// ============================================================
// Test: ListIntakeForms
// ============================================================

func TestListIntakeForms(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries)
		params   ListIntakeFormsParams
		validate func(t *testing.T, results []ListIntakeFormsRow)
	}{
		{
			name:   "empty",
			setup:  func(t *testing.T, q *Queries) {},
			params: ListIntakeFormsParams{Limit: 10, Offset: 0},
			validate: func(t *testing.T, results []ListIntakeFormsRow) {
				assert.Len(t, results, 0)
			},
		},
		{
			name: "with_data",
			setup: func(t *testing.T, q *Queries) {
				for i := 0; i < 2; i++ {
					userID := CreateTestUser(t, q, CreateTestUserOptions{})
					locationID := CreateTestLocation(t, q, CreateTestLocationOptions{})
					employeeID := CreateTestEmployee(t, q, CreateTestEmployeeOptions{UserID: userID})
					regFormID := CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{})

					CreateTestIntakeForm(t, q, CreateTestIntakeFormOptions{
						RegistrationFormID: regFormID,
						LocationID:         locationID,
						CoordinatorID:      employeeID,
					})
				}
			},
			params: ListIntakeFormsParams{Limit: 10, Offset: 0},
			validate: func(t *testing.T, results []ListIntakeFormsRow) {
				assert.Len(t, results, 2)
				assert.Equal(t, int64(2), results[0].TotalCount)
			},
		},
		{
			name: "with_pagination",
			setup: func(t *testing.T, q *Queries) {
				for i := 0; i < 5; i++ {
					userID := CreateTestUser(t, q, CreateTestUserOptions{})
					locationID := CreateTestLocation(t, q, CreateTestLocationOptions{})
					employeeID := CreateTestEmployee(t, q, CreateTestEmployeeOptions{UserID: userID})
					regFormID := CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{})

					CreateTestIntakeForm(t, q, CreateTestIntakeFormOptions{
						RegistrationFormID: regFormID,
						LocationID:         locationID,
						CoordinatorID:      employeeID,
					})
				}
			},
			params: ListIntakeFormsParams{Limit: 2, Offset: 0},
			validate: func(t *testing.T, results []ListIntakeFormsRow) {
				assert.Len(t, results, 2)
				assert.Equal(t, int64(5), results[0].TotalCount)
			},
		},
		{
			name: "with_offset",
			setup: func(t *testing.T, q *Queries) {
				for i := 0; i < 5; i++ {
					userID := CreateTestUser(t, q, CreateTestUserOptions{})
					locationID := CreateTestLocation(t, q, CreateTestLocationOptions{})
					employeeID := CreateTestEmployee(t, q, CreateTestEmployeeOptions{UserID: userID})
					regFormID := CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{})

					CreateTestIntakeForm(t, q, CreateTestIntakeFormOptions{
						RegistrationFormID: regFormID,
						LocationID:         locationID,
						CoordinatorID:      employeeID,
					})
				}
			},
			params: ListIntakeFormsParams{Limit: 10, Offset: 3},
			validate: func(t *testing.T, results []ListIntakeFormsRow) {
				assert.Len(t, results, 2) // 5 total - 3 offset = 2 remaining
			},
		},
		{
			name: "with_search_by_client_name",
			setup: func(t *testing.T, q *Queries) {
				// Create Alice
				userID1 := CreateTestUser(t, q, CreateTestUserOptions{})
				locationID1 := CreateTestLocation(t, q, CreateTestLocationOptions{})
				employeeID1 := CreateTestEmployee(t, q, CreateTestEmployeeOptions{UserID: userID1})
				regFormID1 := CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{
					FirstName: strPtr("Alice"),
					LastName:  strPtr("Wonder"),
				})
				CreateTestIntakeForm(t, q, CreateTestIntakeFormOptions{
					RegistrationFormID: regFormID1,
					LocationID:         locationID1,
					CoordinatorID:      employeeID1,
				})

				// Create Bob
				userID2 := CreateTestUser(t, q, CreateTestUserOptions{})
				locationID2 := CreateTestLocation(t, q, CreateTestLocationOptions{})
				employeeID2 := CreateTestEmployee(t, q, CreateTestEmployeeOptions{UserID: userID2})
				regFormID2 := CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{
					FirstName: strPtr("Bob"),
					LastName:  strPtr("Builder"),
				})
				CreateTestIntakeForm(t, q, CreateTestIntakeFormOptions{
					RegistrationFormID: regFormID2,
					LocationID:         locationID2,
					CoordinatorID:      employeeID2,
				})
			},
			params: ListIntakeFormsParams{Limit: 10, Offset: 0, Column3: "Alice"},
			validate: func(t *testing.T, results []ListIntakeFormsRow) {
				assert.Len(t, results, 1)
				assert.Equal(t, "Alice", *results[0].FirstName)
			},
		},
		{
			name: "search_no_match",
			setup: func(t *testing.T, q *Queries) {
				userID := CreateTestUser(t, q, CreateTestUserOptions{})
				locationID := CreateTestLocation(t, q, CreateTestLocationOptions{})
				employeeID := CreateTestEmployee(t, q, CreateTestEmployeeOptions{UserID: userID})
				regFormID := CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{
					FirstName: strPtr("Charlie"),
				})
				CreateTestIntakeForm(t, q, CreateTestIntakeFormOptions{
					RegistrationFormID: regFormID,
					LocationID:         locationID,
					CoordinatorID:      employeeID,
				})
			},
			params: ListIntakeFormsParams{Limit: 10, Offset: 0, Column3: "Zzzzzz"},
			validate: func(t *testing.T, results []ListIntakeFormsRow) {
				assert.Len(t, results, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				tt.setup(t, q)

				results, err := q.ListIntakeForms(ctx, tt.params)

				require.NoError(t, err)
				tt.validate(t, results)
			})
		})
	}
}

// ============================================================
// Test: UpdateIntakeForm
// ============================================================

func TestUpdateIntakeForm(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) (string, UpdateIntakeFormParams)
		wantErr  bool
		validate func(t *testing.T, q *Queries, id string)
	}{
		{
			name: "update_family_situation",
			setup: func(t *testing.T, q *Queries) (string, UpdateIntakeFormParams) {
				userID := CreateTestUser(t, q, CreateTestUserOptions{})
				locationID := CreateTestLocation(t, q, CreateTestLocationOptions{})
				employeeID := CreateTestEmployee(t, q, CreateTestEmployeeOptions{UserID: userID})
				regFormID := CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{})

				id := CreateTestIntakeForm(t, q, CreateTestIntakeFormOptions{
					RegistrationFormID: regFormID,
					LocationID:         locationID,
					CoordinatorID:      employeeID,
					FamilySituation:    strPtr("Original situation"),
				})
				return id, UpdateIntakeFormParams{
					ID:              id,
					FamilySituation: strPtr("Updated situation"),
				}
			},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				form, err := q.GetIntakeForm(ctx, id)
				require.NoError(t, err)
				assert.Equal(t, "Updated situation", *form.FamilySituation)
			},
		},
		{
			name: "update_multiple_fields",
			setup: func(t *testing.T, q *Queries) (string, UpdateIntakeFormParams) {
				userID := CreateTestUser(t, q, CreateTestUserOptions{})
				locationID := CreateTestLocation(t, q, CreateTestLocationOptions{})
				employeeID := CreateTestEmployee(t, q, CreateTestEmployeeOptions{UserID: userID})
				regFormID := CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{})

				id := CreateTestIntakeForm(t, q, CreateTestIntakeFormOptions{
					RegistrationFormID: regFormID,
					LocationID:         locationID,
					CoordinatorID:      employeeID,
				})
				return id, UpdateIntakeFormParams{
					ID:              id,
					FamilySituation: strPtr("New family situation"),
					MainProvider:    strPtr("New provider"),
					Limitations:     strPtr("New limitations"),
					FocusAreas:      strPtr("New focus areas"),
					Notes:           strPtr("New notes"),
				}
			},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				form, err := q.GetIntakeForm(ctx, id)
				require.NoError(t, err)
				assert.Equal(t, "New family situation", *form.FamilySituation)
				assert.Equal(t, "New provider", *form.MainProvider)
				assert.Equal(t, "New limitations", *form.Limitations)
				assert.Equal(t, "New focus areas", *form.FocusAreas)
				assert.Equal(t, "New notes", *form.Notes)
			},
		},
		{
			name: "update_status",
			setup: func(t *testing.T, q *Queries) (string, UpdateIntakeFormParams) {
				userID := CreateTestUser(t, q, CreateTestUserOptions{})
				locationID := CreateTestLocation(t, q, CreateTestLocationOptions{})
				employeeID := CreateTestEmployee(t, q, CreateTestEmployeeOptions{UserID: userID})
				regFormID := CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{})

				id := CreateTestIntakeForm(t, q, CreateTestIntakeFormOptions{
					RegistrationFormID: regFormID,
					LocationID:         locationID,
					CoordinatorID:      employeeID,
				})
				return id, UpdateIntakeFormParams{
					ID:     id,
					Status: NullIntakeStatusEnum{IntakeStatusEnum: IntakeStatusEnumCompleted, Valid: true},
				}
			},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				form, err := q.GetIntakeForm(ctx, id)
				require.NoError(t, err)
				assert.Equal(t, IntakeStatusEnumCompleted, form.Status)
			},
		},
		{
			name: "update_evaluation_interval",
			setup: func(t *testing.T, q *Queries) (string, UpdateIntakeFormParams) {
				userID := CreateTestUser(t, q, CreateTestUserOptions{})
				locationID := CreateTestLocation(t, q, CreateTestLocationOptions{})
				employeeID := CreateTestEmployee(t, q, CreateTestEmployeeOptions{UserID: userID})
				regFormID := CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{})

				id := CreateTestIntakeForm(t, q, CreateTestIntakeFormOptions{
					RegistrationFormID: regFormID,
					LocationID:         locationID,
					CoordinatorID:      employeeID,
				})
				return id, UpdateIntakeFormParams{
					ID:                      id,
					EvaluationIntervalWeeks: int32Ptr(8),
				}
			},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				form, err := q.GetIntakeForm(ctx, id)
				require.NoError(t, err)
				assert.Equal(t, int32(8), *form.EvaluationIntervalWeeks)
			},
		},
		{
			name: "update_location",
			setup: func(t *testing.T, q *Queries) (string, UpdateIntakeFormParams) {
				userID := CreateTestUser(t, q, CreateTestUserOptions{})
				locationID := CreateTestLocation(t, q, CreateTestLocationOptions{})
				newLocationID := CreateTestLocation(t, q, CreateTestLocationOptions{Name: strPtr("New Location")})
				employeeID := CreateTestEmployee(t, q, CreateTestEmployeeOptions{UserID: userID})
				regFormID := CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{})

				id := CreateTestIntakeForm(t, q, CreateTestIntakeFormOptions{
					RegistrationFormID: regFormID,
					LocationID:         locationID,
					CoordinatorID:      employeeID,
				})
				return id, UpdateIntakeFormParams{
					ID:         id,
					LocationID: &newLocationID,
				}
			},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				form, err := q.GetIntakeForm(ctx, id)
				require.NoError(t, err)
				assert.NotEmpty(t, form.LocationID)
			},
		},
		{
			name: "partial_update_preserves_other_fields",
			setup: func(t *testing.T, q *Queries) (string, UpdateIntakeFormParams) {
				userID := CreateTestUser(t, q, CreateTestUserOptions{})
				locationID := CreateTestLocation(t, q, CreateTestLocationOptions{})
				employeeID := CreateTestEmployee(t, q, CreateTestEmployeeOptions{UserID: userID})
				regFormID := CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{})

				id := CreateTestIntakeForm(t, q, CreateTestIntakeFormOptions{
					RegistrationFormID: regFormID,
					LocationID:         locationID,
					CoordinatorID:      employeeID,
					FamilySituation:    strPtr("Original family situation"),
					Notes:              strPtr("Original notes"),
				})
				// Only update notes
				return id, UpdateIntakeFormParams{
					ID:    id,
					Notes: strPtr("Updated notes"),
				}
			},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				form, err := q.GetIntakeForm(ctx, id)
				require.NoError(t, err)
				assert.Equal(t, "Updated notes", *form.Notes)
				assert.Equal(t, "Original family situation", *form.FamilySituation) // Preserved
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				id, params := tt.setup(t, q)

				err := q.UpdateIntakeForm(ctx, params)

				if tt.wantErr {
					require.Error(t, err)
					return
				}

				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, q, id)
				}
			})
		})
	}
}

// ============================================================
// Test: UpdateIntakeFormStatus
// ============================================================

func TestUpdateIntakeFormStatus(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) string // returns ID
		status   IntakeStatusEnum
		wantErr  bool
		validate func(t *testing.T, q *Queries, id string)
	}{
		{
			name: "update_to_completed",
			setup: func(t *testing.T, q *Queries) string {
				userID := CreateTestUser(t, q, CreateTestUserOptions{})
				locationID := CreateTestLocation(t, q, CreateTestLocationOptions{})
				employeeID := CreateTestEmployee(t, q, CreateTestEmployeeOptions{UserID: userID})
				regFormID := CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{})

				return CreateTestIntakeForm(t, q, CreateTestIntakeFormOptions{
					RegistrationFormID: regFormID,
					LocationID:         locationID,
					CoordinatorID:      employeeID,
				})
			},
			status:  IntakeStatusEnumCompleted,
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				form, err := q.GetIntakeForm(ctx, id)
				require.NoError(t, err)
				assert.Equal(t, IntakeStatusEnumCompleted, form.Status)
			},
		},
		{
			name: "update_to_rejected",
			setup: func(t *testing.T, q *Queries) string {
				userID := CreateTestUser(t, q, CreateTestUserOptions{})
				locationID := CreateTestLocation(t, q, CreateTestLocationOptions{})
				employeeID := CreateTestEmployee(t, q, CreateTestEmployeeOptions{UserID: userID})
				regFormID := CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{})

				return CreateTestIntakeForm(t, q, CreateTestIntakeFormOptions{
					RegistrationFormID: regFormID,
					LocationID:         locationID,
					CoordinatorID:      employeeID,
				})
			},
			status:  IntakeStatusEnumRejected,
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				form, err := q.GetIntakeForm(ctx, id)
				require.NoError(t, err)
				assert.Equal(t, IntakeStatusEnumRejected, form.Status)
			},
		},
		{
			name: "update_to_pending",
			setup: func(t *testing.T, q *Queries) string {
				ctx := context.Background()
				userID := CreateTestUser(t, q, CreateTestUserOptions{})
				locationID := CreateTestLocation(t, q, CreateTestLocationOptions{})
				employeeID := CreateTestEmployee(t, q, CreateTestEmployeeOptions{UserID: userID})
				regFormID := CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{})

				id := CreateTestIntakeForm(t, q, CreateTestIntakeFormOptions{
					RegistrationFormID: regFormID,
					LocationID:         locationID,
					CoordinatorID:      employeeID,
				})
				// First change to completed
				q.UpdateIntakeFormStatus(ctx, UpdateIntakeFormStatusParams{
					ID:     id,
					Status: IntakeStatusEnumCompleted,
				})
				return id
			},
			status:  IntakeStatusEnumPending,
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				form, err := q.GetIntakeForm(ctx, id)
				require.NoError(t, err)
				assert.Equal(t, IntakeStatusEnumPending, form.Status)
			},
		},
		{
			name: "updates_updated_at_timestamp",
			setup: func(t *testing.T, q *Queries) string {
				userID := CreateTestUser(t, q, CreateTestUserOptions{})
				locationID := CreateTestLocation(t, q, CreateTestLocationOptions{})
				employeeID := CreateTestEmployee(t, q, CreateTestEmployeeOptions{UserID: userID})
				regFormID := CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{})

				return CreateTestIntakeForm(t, q, CreateTestIntakeFormOptions{
					RegistrationFormID: regFormID,
					LocationID:         locationID,
					CoordinatorID:      employeeID,
				})
			},
			status:  IntakeStatusEnumCompleted,
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				form, err := q.GetIntakeForm(ctx, id)
				require.NoError(t, err)
				assert.True(t, form.UpdatedAt.Valid)
				// UpdatedAt should be recent (within last few seconds)
				assert.WithinDuration(t, time.Now(), form.UpdatedAt.Time, 5*time.Second)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				id := tt.setup(t, q)

				err := q.UpdateIntakeFormStatus(ctx, UpdateIntakeFormStatusParams{
					ID:     id,
					Status: tt.status,
				})

				if tt.wantErr {
					require.Error(t, err)
					return
				}

				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, q, id)
				}
			})
		})
	}
}
