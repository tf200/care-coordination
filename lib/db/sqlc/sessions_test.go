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
// Test: UpdateUserSession
// ============================================================

func TestUpdateUserSession(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) UpdateUserSessionParams
		wantErr  bool
		checkErr func(t *testing.T, err error)
		validate func(t *testing.T, q *Queries, params UpdateUserSessionParams)
	}{
		{
			name: "success",
			setup: func(t *testing.T, q *Queries) UpdateUserSessionParams {
				userID := CreateTestUser(t, q, CreateTestUserOptions{})
				sessionID := CreateTestSession(t, q, CreateTestSessionOptions{
					UserID: userID,
				})
				return UpdateUserSessionParams{
					ID:          sessionID,
					TokenHash:   "updated-hash",
					TokenFamily: "updated-family",
					ExpiresAt:   pgtype.Timestamptz{Time: time.Now().Add(2 * time.Hour).Round(time.Microsecond), Valid: true},
					UserAgent:   strPtr("Updated Agent"),
					IpAddress:   strPtr("1.2.3.4"),
				}
			},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, params UpdateUserSessionParams) {
				session, err := q.GetUserSession(context.Background(), params.TokenHash)
				require.NoError(t, err)
				assert.Equal(t, params.ID, session.ID)
				assert.Equal(t, params.TokenHash, session.TokenHash)
				assert.Equal(t, params.TokenFamily, session.TokenFamily)
				assert.WithinDuration(t, params.ExpiresAt.Time, session.ExpiresAt.Time, time.Second)
				assert.Equal(t, params.UserAgent, session.UserAgent)
				assert.Equal(t, params.IpAddress, session.IpAddress)
			},
		},
		{
			name: "not_found",
			setup: func(t *testing.T, q *Queries) UpdateUserSessionParams {
				return UpdateUserSessionParams{
					ID:          "non-existent-id",
					TokenHash:   "some-hash",
					TokenFamily: "some-family",
					ExpiresAt:   pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true},
				}
			},
			wantErr: false, // exec queries in pgx don't return error if 0 rows affected unless explicit
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				params := tt.setup(t, q)

				err := q.UpdateUserSession(ctx, params)

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
