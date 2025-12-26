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
// Test: CreateRegistrationForm
// ============================================================

func TestCreateRegistrationForm(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) CreateRegistrationFormParams
		wantErr  bool
		checkErr func(t *testing.T, err error)
		validate func(t *testing.T, q *Queries, params CreateRegistrationFormParams)
	}{
		{
			name: "success_with_referring_org",
			setup: func(t *testing.T, q *Queries) CreateRegistrationFormParams {
				orgID := CreateTestReferringOrg(t, q, CreateTestReferringOrgOptions{})
				return CreateRegistrationFormParams{
					ID:                 generateTestID(),
					FirstName:          "John",
					LastName:           "Doe",
					Bsn:                generateTestID()[:9],
					Gender:             GenderEnumMale,
					DateOfBirth:        toPgDate(time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)),
					RefferingOrgID:     &orgID,
					CareType:           CareTypeEnumProtectedLiving,
					RegistrationReason: "Test reason",
					AdditionalNotes:    strPtr("Test notes"),
					AttachmentIds:      []string{"att1", "att2"},
				}
			},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, params CreateRegistrationFormParams) {
				ctx := context.Background()
				form, err := q.GetRegistrationForm(ctx, params.ID)
				require.NoError(t, err)
				assert.Equal(t, params.FirstName, form.FirstName)
				assert.Equal(t, params.LastName, form.LastName)
				assert.Equal(t, params.Bsn, form.Bsn)
				assert.Equal(t, params.Gender, form.Gender)
				assert.Equal(t, params.RefferingOrgID, form.RefferingOrgID)
				assert.Equal(t, params.CareType, form.CareType)
				assert.Equal(t, params.AttachmentIds, form.AttachmentIds)
			},
		},
		{
			name: "success_without_referring_org",
			setup: func(t *testing.T, q *Queries) CreateRegistrationFormParams {
				return CreateRegistrationFormParams{
					ID:                 generateTestID(),
					FirstName:          "Jane",
					LastName:           "Smith",
					Bsn:                generateTestID()[:9],
					Gender:             GenderEnumFemale,
					DateOfBirth:        toPgDate(time.Date(1985, 6, 15, 0, 0, 0, 0, time.UTC)),
					RefferingOrgID:     nil, // No referring org
					CareType:           CareTypeEnumAmbulatoryCare,
					RegistrationReason: "Self-referral",
				}
			},
			wantErr: false,
		},
		{
			name: "duplicate_bsn",
			setup: func(t *testing.T, q *Queries) CreateRegistrationFormParams {
				bsn := "DUPLICATE1"
				CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{Bsn: &bsn})
				return CreateRegistrationFormParams{
					ID:                 generateTestID(),
					FirstName:          "Another",
					LastName:           "Person",
					Bsn:                bsn, // Duplicate
					Gender:             GenderEnumOther,
					DateOfBirth:        toPgDate(time.Date(1992, 2, 2, 0, 0, 0, 0, time.UTC)),
					CareType:           CareTypeEnumProtectedLiving,
					RegistrationReason: "Duplicate BSN test",
				}
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.True(t, IsUniqueViolation(err), "expected unique violation, got: %v", err)
			},
		},
		{
			name: "invalid_referring_org_fk",
			setup: func(t *testing.T, q *Queries) CreateRegistrationFormParams {
				fakeOrgID := "nonexistent-org-id"
				return CreateRegistrationFormParams{
					ID:                 generateTestID(),
					FirstName:          "Invalid",
					LastName:           "ForeignKey",
					Bsn:                generateTestID()[:9],
					Gender:             GenderEnumOther,
					DateOfBirth:        toPgDate(time.Date(1995, 5, 5, 0, 0, 0, 0, time.UTC)),
					RefferingOrgID:     &fakeOrgID,
					CareType:           CareTypeEnumProtectedLiving,
					RegistrationReason: "FK test",
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

				err := q.CreateRegistrationForm(ctx, params)

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
// Test: GetRegistrationForm
// ============================================================

func TestGetRegistrationForm(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) string // returns ID to query
		wantErr  bool
		checkErr func(t *testing.T, err error)
		validate func(t *testing.T, form RegistrationForm)
	}{
		{
			name: "found",
			setup: func(t *testing.T, q *Queries) string {
				return CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{})
			},
			wantErr: false,
			validate: func(t *testing.T, form RegistrationForm) {
				assert.NotEmpty(t, form.ID)
				assert.NotEmpty(t, form.FirstName)
				assert.NotEmpty(t, form.LastName)
				assert.NotEmpty(t, form.Bsn)
				assert.True(t, form.CreatedAt.Valid)
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

				form, err := q.GetRegistrationForm(ctx, id)

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
// Test: ListRegistrationForms
// ============================================================

func TestListRegistrationForms(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries)
		params   ListRegistrationFormsParams
		validate func(t *testing.T, results []ListRegistrationFormsRow)
	}{
		{
			name:   "empty",
			setup:  func(t *testing.T, q *Queries) {},
			params: ListRegistrationFormsParams{Limit: 10, Offset: 0},
			validate: func(t *testing.T, results []ListRegistrationFormsRow) {
				assert.Len(t, results, 0)
			},
		},
		{
			name: "with_data",
			setup: func(t *testing.T, q *Queries) {
				CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{
					FirstName: strPtr("Alice"),
					LastName:  strPtr("Wonder"),
				})
				CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{
					FirstName: strPtr("Bob"),
					LastName:  strPtr("Builder"),
				})
			},
			params: ListRegistrationFormsParams{Limit: 10, Offset: 0},
			validate: func(t *testing.T, results []ListRegistrationFormsRow) {
				assert.Len(t, results, 2)
				assert.Equal(t, int64(2), results[0].TotalCount)
			},
		},
		{
			name: "with_pagination",
			setup: func(t *testing.T, q *Queries) {
				for i := 0; i < 5; i++ {
					CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{})
				}
			},
			params: ListRegistrationFormsParams{Limit: 2, Offset: 0},
			validate: func(t *testing.T, results []ListRegistrationFormsRow) {
				assert.Len(t, results, 2)
				assert.Equal(t, int64(5), results[0].TotalCount)
			},
		},
		{
			name: "with_offset",
			setup: func(t *testing.T, q *Queries) {
				for i := 0; i < 5; i++ {
					CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{})
				}
			},
			params: ListRegistrationFormsParams{Limit: 10, Offset: 3},
			validate: func(t *testing.T, results []ListRegistrationFormsRow) {
				assert.Len(t, results, 2) // 5 total - 3 offset = 2 remaining
			},
		},
		{
			name: "with_search_first_name",
			setup: func(t *testing.T, q *Queries) {
				CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{
					FirstName: strPtr("Alice"),
					LastName:  strPtr("Wonder"),
				})
				CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{
					FirstName: strPtr("Bob"),
					LastName:  strPtr("Builder"),
				})
				CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{
					FirstName: strPtr("Charlie"),
					LastName:  strPtr("Brown"),
				})
			},
			params: ListRegistrationFormsParams{Limit: 10, Offset: 0, Search: strPtr("Alice")},
			validate: func(t *testing.T, results []ListRegistrationFormsRow) {
				assert.Len(t, results, 1)
				assert.Equal(t, "Alice", results[0].FirstName)
			},
		},
		{
			name: "with_search_last_name",
			setup: func(t *testing.T, q *Queries) {
				CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{
					FirstName: strPtr("Alice"),
					LastName:  strPtr("Wonder"),
				})
				CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{
					FirstName: strPtr("Bob"),
					LastName:  strPtr("Builder"),
				})
			},
			params: ListRegistrationFormsParams{Limit: 10, Offset: 0, Search: strPtr("Builder")},
			validate: func(t *testing.T, results []ListRegistrationFormsRow) {
				assert.Len(t, results, 1)
				assert.Equal(t, "Bob", results[0].FirstName)
			},
		},
		{
			name: "search_no_match",
			setup: func(t *testing.T, q *Queries) {
				CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{
					FirstName: strPtr("Alice"),
				})
			},
			params: ListRegistrationFormsParams{Limit: 10, Offset: 0, Search: strPtr("Zzzzzz")},
			validate: func(t *testing.T, results []ListRegistrationFormsRow) {
				assert.Len(t, results, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				tt.setup(t, q)

				results, err := q.ListRegistrationForms(ctx, tt.params)

				require.NoError(t, err)
				tt.validate(t, results)
			})
		})
	}
}

// ============================================================
// Test: GetRegistrationFormWithDetails
// ============================================================

func TestGetRegistrationFormWithDetails(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) string // returns ID to query
		wantErr  bool
		checkErr func(t *testing.T, err error)
		validate func(t *testing.T, result GetRegistrationFormWithDetailsRow)
	}{
		{
			name: "found_with_referring_org",
			setup: func(t *testing.T, q *Queries) string {
				orgID := CreateTestReferringOrg(t, q, CreateTestReferringOrgOptions{
					Name:          strPtr("Test Hospital"),
					ContactPerson: strPtr("Dr. Smith"),
					PhoneNumber:   strPtr("+31612345678"),
					Email:         strPtr("hospital@example.com"),
				})
				return CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{
					FirstName:      strPtr("Alice"),
					LastName:       strPtr("Wonder"),
					ReferringOrgID: &orgID,
				})
			},
			wantErr: false,
			validate: func(t *testing.T, result GetRegistrationFormWithDetailsRow) {
				assert.Equal(t, "Alice", result.FirstName)
				assert.Equal(t, "Wonder", result.LastName)
				assert.NotNil(t, result.OrgName)
				assert.Equal(t, "Test Hospital", *result.OrgName)
				assert.NotNil(t, result.OrgContactPerson)
				assert.Equal(t, "Dr. Smith", *result.OrgContactPerson)
				assert.NotNil(t, result.OrgPhoneNumber)
				assert.NotNil(t, result.OrgEmail)
				assert.False(t, result.IntakeCompleted)
				assert.False(t, result.HasClient)
			},
		},
		{
			name: "found_without_referring_org",
			setup: func(t *testing.T, q *Queries) string {
				return CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{
					FirstName: strPtr("Bob"),
					LastName:  strPtr("Builder"),
				})
			},
			wantErr: false,
			validate: func(t *testing.T, result GetRegistrationFormWithDetailsRow) {
				assert.Equal(t, "Bob", result.FirstName)
				assert.Nil(t, result.OrgName)
				assert.Nil(t, result.OrgContactPerson)
				assert.Nil(t, result.OrgPhoneNumber)
				assert.Nil(t, result.OrgEmail)
			},
		},
		{
			name: "with_intake_completed",
			setup: func(t *testing.T, q *Queries) string {
				// Create dependencies
				userID := CreateTestUser(t, q, CreateTestUserOptions{})
				locationID := CreateTestLocation(t, q, CreateTestLocationOptions{})
				employeeID := CreateTestEmployee(t, q, CreateTestEmployeeOptions{UserID: userID})

				regID := CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{
					FirstName: strPtr("Charlie"),
				})

				// Create intake form to mark intake_completed as true
				CreateTestIntakeForm(t, q, CreateTestIntakeFormOptions{
					RegistrationFormID: regID,
					LocationID:         locationID,
					CoordinatorID:      employeeID,
				})

				return regID
			},
			wantErr: false,
			validate: func(t *testing.T, result GetRegistrationFormWithDetailsRow) {
				assert.Equal(t, "Charlie", result.FirstName)
				assert.True(t, result.IntakeCompleted)
			},
		},
		{
			name: "with_client_created",
			setup: func(t *testing.T, q *Queries) string {
				// Create full client chain
				clientID, deps := CreateTestClientWithDependencies(t, q)
				_ = clientID
				return deps.RegistrationFormID
			},
			wantErr: false,
			validate: func(t *testing.T, result GetRegistrationFormWithDetailsRow) {
				assert.True(t, result.IntakeCompleted)
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

				result, err := q.GetRegistrationFormWithDetails(ctx, id)

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
// Test: GetRegistrationStats
// ============================================================

func TestGetRegistrationStats(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries)
		validate func(t *testing.T, stats GetRegistrationStatsRow)
	}{
		{
			name:  "empty_database",
			setup: func(t *testing.T, q *Queries) {},
			validate: func(t *testing.T, stats GetRegistrationStatsRow) {
				assert.Equal(t, int64(0), stats.TotalCount)
				assert.Equal(t, int64(0), stats.ApprovedCount)
				assert.Equal(t, int64(0), stats.InReviewCount)
			},
		},
		{
			name: "with_pending_forms",
			setup: func(t *testing.T, q *Queries) {
				// Create 3 registration forms (default status is 'pending')
				CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{})
				CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{})
				CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{})
			},
			validate: func(t *testing.T, stats GetRegistrationStatsRow) {
				assert.Equal(t, int64(3), stats.TotalCount)
				assert.Equal(t, int64(0), stats.ApprovedCount)
				assert.Equal(t, int64(0), stats.InReviewCount)
			},
		},
		{
			name: "with_mixed_statuses",
			setup: func(t *testing.T, q *Queries) {
				ctx := context.Background()

				// Create and update to approved
				id1 := CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{})
				q.UpdateRegistrationFormStatus(ctx, UpdateRegistrationFormStatusParams{
					ID:     id1,
					Status: NullRegistrationStatusEnum{RegistrationStatusEnum: RegistrationStatusEnumApproved, Valid: true},
				})

				// Create and update to in_review
				id2 := CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{})
				q.UpdateRegistrationFormStatus(ctx, UpdateRegistrationFormStatusParams{
					ID:     id2,
					Status: NullRegistrationStatusEnum{RegistrationStatusEnum: RegistrationStatusEnumInReview, Valid: true},
				})

				// Create and update to another approved
				id3 := CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{})
				q.UpdateRegistrationFormStatus(ctx, UpdateRegistrationFormStatusParams{
					ID:     id3,
					Status: NullRegistrationStatusEnum{RegistrationStatusEnum: RegistrationStatusEnumApproved, Valid: true},
				})

				// Create pending (default)
				CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{})
			},
			validate: func(t *testing.T, stats GetRegistrationStatsRow) {
				assert.Equal(t, int64(4), stats.TotalCount)
				assert.Equal(t, int64(2), stats.ApprovedCount)
				assert.Equal(t, int64(1), stats.InReviewCount)
			},
		},
		{
			name: "excludes_soft_deleted",
			setup: func(t *testing.T, q *Queries) {
				ctx := context.Background()

				// Create and soft delete
				id1 := CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{})
				q.SoftDeleteRegistrationForm(ctx, id1)

				// Create active form
				CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{})
			},
			validate: func(t *testing.T, stats GetRegistrationStatsRow) {
				// Only 1 should be counted (the non-deleted one)
				assert.Equal(t, int64(1), stats.TotalCount)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				tt.setup(t, q)

				stats, err := q.GetRegistrationStats(ctx)

				require.NoError(t, err)
				tt.validate(t, stats)
			})
		})
	}
}

// ============================================================
// Test: SoftDeleteRegistrationForm
// ============================================================

func TestSoftDeleteRegistrationForm(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) string // returns ID to delete
		wantErr  bool
		validate func(t *testing.T, q *Queries, id string)
	}{
		{
			name: "success",
			setup: func(t *testing.T, q *Queries) string {
				return CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{})
			},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				form, err := q.GetRegistrationForm(ctx, id)
				require.NoError(t, err)
				require.NotNil(t, form.IsDeleted, "expected is_deleted to be set")
				assert.True(t, *form.IsDeleted, "expected is_deleted to be true")
				assert.True(t, form.UpdatedAt.Valid, "expected updated_at to be set")
			},
		},
		{
			name: "nonexistent_id_no_error",
			setup: func(t *testing.T, q *Queries) string {
				return "nonexistent-id"
			},
			wantErr: false, // UPDATE on non-existent row doesn't error, just affects 0 rows
			validate: func(t *testing.T, q *Queries, id string) {
				// Nothing to validate - the row doesn't exist
			},
		},
		{
			name: "already_deleted",
			setup: func(t *testing.T, q *Queries) string {
				ctx := context.Background()
				id := CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{})
				// Delete once
				q.SoftDeleteRegistrationForm(ctx, id)
				return id
			},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				form, err := q.GetRegistrationForm(ctx, id)
				require.NoError(t, err)
				require.NotNil(t, form.IsDeleted)
				assert.True(t, *form.IsDeleted)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				id := tt.setup(t, q)

				err := q.SoftDeleteRegistrationForm(ctx, id)

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
// Test: UpdateRegistrationForm
// ============================================================

func TestUpdateRegistrationForm(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) (string, UpdateRegistrationFormParams)
		wantErr  bool
		validate func(t *testing.T, q *Queries, id string)
	}{
		{
			name: "update_first_name",
			setup: func(t *testing.T, q *Queries) (string, UpdateRegistrationFormParams) {
				id := CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{
					FirstName: strPtr("Original"),
				})
				return id, UpdateRegistrationFormParams{
					ID:        id,
					FirstName: strPtr("Updated"),
				}
			},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				form, err := q.GetRegistrationForm(ctx, id)
				require.NoError(t, err)
				assert.Equal(t, "Updated", form.FirstName)
			},
		},
		{
			name: "update_multiple_fields",
			setup: func(t *testing.T, q *Queries) (string, UpdateRegistrationFormParams) {
				id := CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{
					FirstName: strPtr("John"),
					LastName:  strPtr("Doe"),
				})
				return id, UpdateRegistrationFormParams{
					ID:                 id,
					FirstName:          strPtr("Jane"),
					LastName:           strPtr("Smith"),
					RegistrationReason: strPtr("Updated reason"),
					AdditionalNotes:    strPtr("New notes"),
				}
			},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				form, err := q.GetRegistrationForm(ctx, id)
				require.NoError(t, err)
				assert.Equal(t, "Jane", form.FirstName)
				assert.Equal(t, "Smith", form.LastName)
				assert.Equal(t, "Updated reason", form.RegistrationReason)
				assert.NotNil(t, form.AdditionalNotes)
				assert.Equal(t, "New notes", *form.AdditionalNotes)
			},
		},
		{
			name: "update_status",
			setup: func(t *testing.T, q *Queries) (string, UpdateRegistrationFormParams) {
				id := CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{})
				return id, UpdateRegistrationFormParams{
					ID:     id,
					Status: NullRegistrationStatusEnum{RegistrationStatusEnum: RegistrationStatusEnumApproved, Valid: true},
				}
			},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				form, err := q.GetRegistrationForm(ctx, id)
				require.NoError(t, err)
				assert.True(t, form.Status.Valid)
				assert.Equal(t, RegistrationStatusEnumApproved, form.Status.RegistrationStatusEnum)
			},
		},
		{
			name: "update_gender",
			setup: func(t *testing.T, q *Queries) (string, UpdateRegistrationFormParams) {
				id := CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{
					Gender: func() *GenderEnum { g := GenderEnumMale; return &g }(),
				})
				return id, UpdateRegistrationFormParams{
					ID:     id,
					Gender: NullGenderEnum{GenderEnum: GenderEnumFemale, Valid: true},
				}
			},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				form, err := q.GetRegistrationForm(ctx, id)
				require.NoError(t, err)
				assert.Equal(t, GenderEnumFemale, form.Gender)
			},
		},
		{
			name: "update_care_type",
			setup: func(t *testing.T, q *Queries) (string, UpdateRegistrationFormParams) {
				id := CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{
					CareType: func() *CareTypeEnum { c := CareTypeEnumProtectedLiving; return &c }(),
				})
				return id, UpdateRegistrationFormParams{
					ID:       id,
					CareType: NullCareTypeEnum{CareTypeEnum: CareTypeEnumAmbulatoryCare, Valid: true},
				}
			},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				form, err := q.GetRegistrationForm(ctx, id)
				require.NoError(t, err)
				assert.Equal(t, CareTypeEnumAmbulatoryCare, form.CareType)
			},
		},
		{
			name: "update_attachment_ids",
			setup: func(t *testing.T, q *Queries) (string, UpdateRegistrationFormParams) {
				id := CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{})
				return id, UpdateRegistrationFormParams{
					ID:            id,
					AttachmentIds: []string{"att1", "att2", "att3"},
				}
			},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				form, err := q.GetRegistrationForm(ctx, id)
				require.NoError(t, err)
				assert.Equal(t, []string{"att1", "att2", "att3"}, form.AttachmentIds)
			},
		},
		{
			name: "update_referring_org",
			setup: func(t *testing.T, q *Queries) (string, UpdateRegistrationFormParams) {
				id := CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{})
				orgID := CreateTestReferringOrg(t, q, CreateTestReferringOrgOptions{})
				return id, UpdateRegistrationFormParams{
					ID:             id,
					RefferingOrgID: &orgID,
				}
			},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				form, err := q.GetRegistrationForm(ctx, id)
				require.NoError(t, err)
				assert.NotNil(t, form.RefferingOrgID)
			},
		},
		{
			name: "partial_update_preserves_other_fields",
			setup: func(t *testing.T, q *Queries) (string, UpdateRegistrationFormParams) {
				id := CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{
					FirstName:          strPtr("Original"),
					LastName:           strPtr("Name"),
					RegistrationReason: strPtr("Original reason"),
				})
				// Only update first name
				return id, UpdateRegistrationFormParams{
					ID:        id,
					FirstName: strPtr("NewFirst"),
				}
			},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				form, err := q.GetRegistrationForm(ctx, id)
				require.NoError(t, err)
				assert.Equal(t, "NewFirst", form.FirstName)
				assert.Equal(t, "Name", form.LastName)                      // Preserved
				assert.Equal(t, "Original reason", form.RegistrationReason) // Preserved
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				id, params := tt.setup(t, q)

				err := q.UpdateRegistrationForm(ctx, params)

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
// Test: UpdateRegistrationFormStatus
// ============================================================

func TestUpdateRegistrationFormStatus(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) string // returns ID
		status   NullRegistrationStatusEnum
		wantErr  bool
		validate func(t *testing.T, q *Queries, id string)
	}{
		{
			name: "update_to_approved",
			setup: func(t *testing.T, q *Queries) string {
				return CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{})
			},
			status:  NullRegistrationStatusEnum{RegistrationStatusEnum: RegistrationStatusEnumApproved, Valid: true},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				form, err := q.GetRegistrationForm(ctx, id)
				require.NoError(t, err)
				assert.True(t, form.Status.Valid)
				assert.Equal(t, RegistrationStatusEnumApproved, form.Status.RegistrationStatusEnum)
			},
		},
		{
			name: "update_to_rejected",
			setup: func(t *testing.T, q *Queries) string {
				return CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{})
			},
			status:  NullRegistrationStatusEnum{RegistrationStatusEnum: RegistrationStatusEnumRejected, Valid: true},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				form, err := q.GetRegistrationForm(ctx, id)
				require.NoError(t, err)
				assert.True(t, form.Status.Valid)
				assert.Equal(t, RegistrationStatusEnumRejected, form.Status.RegistrationStatusEnum)
			},
		},
		{
			name: "update_to_in_review",
			setup: func(t *testing.T, q *Queries) string {
				return CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{})
			},
			status:  NullRegistrationStatusEnum{RegistrationStatusEnum: RegistrationStatusEnumInReview, Valid: true},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				form, err := q.GetRegistrationForm(ctx, id)
				require.NoError(t, err)
				assert.True(t, form.Status.Valid)
				assert.Equal(t, RegistrationStatusEnumInReview, form.Status.RegistrationStatusEnum)
			},
		},
		{
			name: "update_to_pending",
			setup: func(t *testing.T, q *Queries) string {
				ctx := context.Background()
				id := CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{})
				// First change to approved
				q.UpdateRegistrationFormStatus(ctx, UpdateRegistrationFormStatusParams{
					ID:     id,
					Status: NullRegistrationStatusEnum{RegistrationStatusEnum: RegistrationStatusEnumApproved, Valid: true},
				})
				return id
			},
			status:  NullRegistrationStatusEnum{RegistrationStatusEnum: RegistrationStatusEnumPending, Valid: true},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				form, err := q.GetRegistrationForm(ctx, id)
				require.NoError(t, err)
				assert.True(t, form.Status.Valid)
				assert.Equal(t, RegistrationStatusEnumPending, form.Status.RegistrationStatusEnum)
			},
		},
		{
			name: "updates_updated_at_timestamp",
			setup: func(t *testing.T, q *Queries) string {
				return CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{})
			},
			status:  NullRegistrationStatusEnum{RegistrationStatusEnum: RegistrationStatusEnumApproved, Valid: true},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				form, err := q.GetRegistrationForm(ctx, id)
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

				err := q.UpdateRegistrationFormStatus(ctx, UpdateRegistrationFormStatusParams{
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
