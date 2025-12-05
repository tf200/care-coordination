package db

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================
// Test: CreateUser
// ============================================================

func TestCreateUser(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) CreateUserParams
		wantErr  bool
		checkErr func(t *testing.T, err error)
	}{
		{
			name: "success",
			setup: func(t *testing.T, q *Queries) CreateUserParams {
				return CreateUserParams{
					ID:           generateTestID(),
					Email:        "newuser@example.com",
					PasswordHash: "$2a$10$hashedpassword",
				}
			},
			wantErr: false,
		},
		{
			name: "duplicate_email",
			setup: func(t *testing.T, q *Queries) CreateUserParams {
				// Create first user
				email := "duplicate@example.com"
				CreateTestUser(t, q, CreateTestUserOptions{Email: &email})
				// Return params that will fail
				return CreateUserParams{
					ID:           generateTestID(),
					Email:        email,
					PasswordHash: "$2a$10$different",
				}
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.True(t, IsUniqueViolation(err), "expected unique violation, got: %v", err)
			},
		},
		{
			name: "duplicate_id",
			setup: func(t *testing.T, q *Queries) CreateUserParams {
				existingID := generateTestID()
				CreateTestUser(t, q, CreateTestUserOptions{ID: &existingID})
				return CreateUserParams{
					ID:           existingID,
					Email:        "different@example.com",
					PasswordHash: "$2a$10$hashedpassword",
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

				id, err := q.CreateUser(ctx, params)

				if tt.wantErr {
					require.Error(t, err)
					if tt.checkErr != nil {
						tt.checkErr(t, err)
					}
					return
				}

				require.NoError(t, err)
				assert.Equal(t, params.ID, id)

				// Verify user was created
				user, err := q.GetUserByEmail(ctx, params.Email)
				require.NoError(t, err)
				assert.Equal(t, params.ID, user.ID)
				assert.Equal(t, params.Email, user.Email)
			})
		})
	}
}

// ============================================================
// Test: GetUserByEmail
// ============================================================

func TestGetUserByEmail(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) string // returns email to query
		wantErr  bool
		checkErr func(t *testing.T, err error)
		validate func(t *testing.T, user User, email string)
	}{
		{
			name: "found",
			setup: func(t *testing.T, q *Queries) string {
				email := "findme@example.com"
				CreateTestUser(t, q, CreateTestUserOptions{Email: &email})
				return email
			},
			wantErr: false,
			validate: func(t *testing.T, user User, email string) {
				assert.Equal(t, email, user.Email)
				assert.NotEmpty(t, user.ID)
				assert.NotEmpty(t, user.PasswordHash)
				assert.True(t, user.CreatedAt.Valid)
			},
		},
		{
			name: "not_found",
			setup: func(t *testing.T, q *Queries) string {
				return "nonexistent@example.com"
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.True(t, errors.Is(err, pgx.ErrNoRows), "expected ErrNoRows, got: %v", err)
			},
		},
		{
			name: "case_sensitive",
			setup: func(t *testing.T, q *Queries) string {
				email := "CaseSensitive@Example.com"
				CreateTestUser(t, q, CreateTestUserOptions{Email: &email})
				return "casesensitive@example.com" // Different case
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.True(t, errors.Is(err, pgx.ErrNoRows), "expected ErrNoRows for case mismatch")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				email := tt.setup(t, q)

				user, err := q.GetUserByEmail(ctx, email)

				if tt.wantErr {
					require.Error(t, err)
					if tt.checkErr != nil {
						tt.checkErr(t, err)
					}
					return
				}

				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, user, email)
				}
			})
		})
	}
}

// ============================================================
// Test: CreateUserSession
// ============================================================

func TestCreateUserSession(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) CreateUserSessionParams
		wantErr  bool
		checkErr func(t *testing.T, err error)
	}{
		{
			name: "success",
			setup: func(t *testing.T, q *Queries) CreateUserSessionParams {
				userID := CreateTestUser(t, q, CreateTestUserOptions{})
				return CreateUserSessionParams{
					ID:          generateTestID(),
					UserID:      userID,
					TokenHash:   "test-token-hash",
					TokenFamily: "test-family",
					ExpiresAt:   pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true},
				}
			},
			wantErr: false,
		},
		{
			name: "invalid_user_id",
			setup: func(t *testing.T, q *Queries) CreateUserSessionParams {
				return CreateUserSessionParams{
					ID:          generateTestID(),
					UserID:      "non-existent-user-id",
					TokenHash:   "some-hash",
					TokenFamily: "some-family",
					ExpiresAt:   pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true},
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

				err := q.CreateUserSession(ctx, params)

				if tt.wantErr {
					require.Error(t, err)
					if tt.checkErr != nil {
						tt.checkErr(t, err)
					}
					return
				}

				require.NoError(t, err)

				// Verify session exists
				session, err := q.GetUserSession(ctx, params.TokenHash)
				require.NoError(t, err)
				assert.Equal(t, params.UserID, session.UserID)
			})
		})
	}
}

// ============================================================
// Test: GetUserSession
// ============================================================

func TestGetUserSession(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) string // returns tokenHash to query
		wantErr  bool
		checkErr func(t *testing.T, err error)
	}{
		{
			name: "found",
			setup: func(t *testing.T, q *Queries) string {
				userID := CreateTestUser(t, q, CreateTestUserOptions{})
				tokenHash := "find-this-token"
				CreateTestSession(t, q, CreateTestSessionOptions{
					UserID:    userID,
					TokenHash: &tokenHash,
				})
				return tokenHash
			},
			wantErr: false,
		},
		{
			name: "not_found",
			setup: func(t *testing.T, q *Queries) string {
				return "non-existent-token-hash"
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
				tokenHash := tt.setup(t, q)

				session, err := q.GetUserSession(ctx, tokenHash)

				if tt.wantErr {
					require.Error(t, err)
					if tt.checkErr != nil {
						tt.checkErr(t, err)
					}
					return
				}

				require.NoError(t, err)
				assert.Equal(t, tokenHash, session.TokenHash)
			})
		})
	}
}

// ============================================================
// Test: DeleteUserSession
// ============================================================

func TestDeleteUserSession(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, q *Queries) string // returns tokenHash to delete
	}{
		{
			name: "existing_session",
			setup: func(t *testing.T, q *Queries) string {
				userID := CreateTestUser(t, q, CreateTestUserOptions{})
				tokenHash := "delete-me-token"
				CreateTestSession(t, q, CreateTestSessionOptions{
					UserID:    userID,
					TokenHash: &tokenHash,
				})
				return tokenHash
			},
		},
		{
			name: "non_existent_session",
			setup: func(t *testing.T, q *Queries) string {
				return "non-existent-token"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				tokenHash := tt.setup(t, q)

				// Delete should never error (idempotent)
				err := q.DeleteUserSession(ctx, tokenHash)
				require.NoError(t, err)

				// Verify session is gone
				_, err = q.GetUserSession(ctx, tokenHash)
				assert.True(t, errors.Is(err, pgx.ErrNoRows), "session should not exist")
			})
		})
	}
}

// ============================================================
// Test: Factory Functions
// ============================================================

func TestFactoryFunctions(t *testing.T) {
	t.Run("CreateTestUser_generates_unique_ids", func(t *testing.T) {
		runTestWithTx(t, func(t *testing.T, q *Queries) {
			id1 := CreateTestUser(t, q, CreateTestUserOptions{})
			id2 := CreateTestUser(t, q, CreateTestUserOptions{})
			id3 := CreateTestUser(t, q, CreateTestUserOptions{})

			assert.NotEqual(t, id1, id2)
			assert.NotEqual(t, id2, id3)
			assert.NotEqual(t, id1, id3)
		})
	})

	t.Run("CreateFullClientDependencyChain_creates_all_deps", func(t *testing.T) {
		runTestWithTx(t, func(t *testing.T, q *Queries) {
			deps := CreateFullClientDependencyChain(t, q)

			assert.NotEmpty(t, deps.UserID)
			assert.NotEmpty(t, deps.EmployeeID)
			assert.NotEmpty(t, deps.LocationID)
			assert.NotEmpty(t, deps.RegistrationFormID)
			assert.NotEmpty(t, deps.IntakeFormID)

			// Verify we can create a client with these deps
			clientID := CreateTestClient(t, q, CreateTestClientOptions{
				RegistrationFormID: deps.RegistrationFormID,
				IntakeFormID:       deps.IntakeFormID,
				AssignedLocationID: deps.LocationID,
				CoordinatorID:      deps.EmployeeID,
			})
			assert.NotEmpty(t, clientID)
		})
	})
}
