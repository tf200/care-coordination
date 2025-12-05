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
