package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	db "care-cordination/lib/db/sqlc"
	dbmocks "care-cordination/lib/db/sqlc/mocks"
	loggermocks "care-cordination/lib/logger/mocks"
	"care-cordination/lib/token"
	tokenmocks "care-cordination/lib/token/mocks"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

// ============================================================
// Test Helpers
// ============================================================

func hashPassword(t *testing.T, password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)
	return string(hash)
}

func createTestRefreshClaims(tokenHash, tokenFamily string) *token.RefreshTokenClaims {
	return &token.RefreshTokenClaims{
		TokenHash:   tokenHash,
		Tokenfamily: tokenFamily,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}
}

// ============================================================
// Test: Login
// ============================================================

func TestLogin(t *testing.T) {
	tests := []struct {
		name      string
		req       *LoginRequest
		userAgent string
		ipAddress string
		setup     func(
			mockStore *dbmocks.MockStoreInterface,
			mockToken *tokenmocks.MockTokenManager,
			hashedPassword string,
		)
		wantErr     bool
		expectedErr error
		validate    func(t *testing.T, resp *LoginResponse)
	}{
		{
			name: "success",
			req: &LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			userAgent: "Mozilla/5.0",
			ipAddress: "127.0.0.1",
			setup: func(
				mockStore *dbmocks.MockStoreInterface,
				mockToken *tokenmocks.MockTokenManager,
				hashedPassword string,
			) {
				mockStore.EXPECT().
					GetUserByEmail(gomock.Any(), "test@example.com").
					Return(db.User{
						ID:           "user-123",
						Email:        "test@example.com",
						PasswordHash: hashedPassword,
					}, nil)

				mockStore.EXPECT().
					GetEmployeeByUserID(gomock.Any(), "user-123").
					Return(db.GetEmployeeByUserIDRow{ID: "employee-123"}, nil)

				mockToken.EXPECT().
					GenerateAccessToken("user-123", "employee-123", gomock.Any()).
					Return("access-token-123", nil)

				mockToken.EXPECT().
					GenerateRefreshToken("user-123", gomock.Any()).
					Return("refresh-token-123", createTestRefreshClaims("token-hash", "token-family"), nil)

				mockStore.EXPECT().
					CreateUserSession(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: false,
			validate: func(t *testing.T, resp *LoginResponse) {
				assert.Equal(t, "access-token-123", resp.AccessToken)
				assert.Equal(t, "refresh-token-123", resp.RefreshToken)
			},
		},
		{
			name: "user_not_found",
			req: &LoginRequest{
				Email:    "notfound@example.com",
				Password: "password123",
			},
			userAgent: "Mozilla/5.0",
			ipAddress: "127.0.0.1",
			setup: func(
				mockStore *dbmocks.MockStoreInterface,
				mockToken *tokenmocks.MockTokenManager,
				hashedPassword string,
			) {
				mockStore.EXPECT().
					GetUserByEmail(gomock.Any(), "notfound@example.com").
					Return(db.User{}, pgx.ErrNoRows)
			},
			wantErr:     true,
			expectedErr: ErrInvalidCredentials,
		},
		{
			name: "invalid_password",
			req: &LoginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			userAgent: "Mozilla/5.0",
			ipAddress: "127.0.0.1",
			setup: func(
				mockStore *dbmocks.MockStoreInterface,
				mockToken *tokenmocks.MockTokenManager,
				hashedPassword string,
			) {
				mockStore.EXPECT().
					GetUserByEmail(gomock.Any(), "test@example.com").
					Return(db.User{
						ID:           "user-123",
						Email:        "test@example.com",
						PasswordHash: hashedPassword,
					}, nil)
			},
			wantErr:     true,
			expectedErr: ErrInvalidCredentials,
		},
		{
			name: "employee_not_found",
			req: &LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			userAgent: "Mozilla/5.0",
			ipAddress: "127.0.0.1",
			setup: func(
				mockStore *dbmocks.MockStoreInterface,
				mockToken *tokenmocks.MockTokenManager,
				hashedPassword string,
			) {
				mockStore.EXPECT().
					GetUserByEmail(gomock.Any(), "test@example.com").
					Return(db.User{
						ID:           "user-123",
						Email:        "test@example.com",
						PasswordHash: hashedPassword,
					}, nil)

				mockStore.EXPECT().
					GetEmployeeByUserID(gomock.Any(), "user-123").
					Return(db.GetEmployeeByUserIDRow{}, pgx.ErrNoRows)
			},
			wantErr:     true,
			expectedErr: ErrInvalidCredentials,
		},
		{
			name: "access_token_generation_error",
			req: &LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			userAgent: "Mozilla/5.0",
			ipAddress: "127.0.0.1",
			setup: func(
				mockStore *dbmocks.MockStoreInterface,
				mockToken *tokenmocks.MockTokenManager,
				hashedPassword string,
			) {
				mockStore.EXPECT().
					GetUserByEmail(gomock.Any(), "test@example.com").
					Return(db.User{
						ID:           "user-123",
						Email:        "test@example.com",
						PasswordHash: hashedPassword,
					}, nil)

				mockStore.EXPECT().
					GetEmployeeByUserID(gomock.Any(), "user-123").
					Return(db.GetEmployeeByUserIDRow{ID: "employee-123"}, nil)

				mockToken.EXPECT().
					GenerateAccessToken("user-123", "employee-123", gomock.Any()).
					Return("", errors.New("token generation failed"))
			},
			wantErr:     true,
			expectedErr: ErrInternal,
		},
		{
			name: "refresh_token_generation_error",
			req: &LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			userAgent: "Mozilla/5.0",
			ipAddress: "127.0.0.1",
			setup: func(
				mockStore *dbmocks.MockStoreInterface,
				mockToken *tokenmocks.MockTokenManager,
				hashedPassword string,
			) {
				mockStore.EXPECT().
					GetUserByEmail(gomock.Any(), "test@example.com").
					Return(db.User{
						ID:           "user-123",
						Email:        "test@example.com",
						PasswordHash: hashedPassword,
					}, nil)

				mockStore.EXPECT().
					GetEmployeeByUserID(gomock.Any(), "user-123").
					Return(db.GetEmployeeByUserIDRow{ID: "employee-123"}, nil)

				mockToken.EXPECT().
					GenerateAccessToken("user-123", "employee-123", gomock.Any()).
					Return("access-token", nil)

				mockToken.EXPECT().
					GenerateRefreshToken("user-123", gomock.Any()).
					Return("", nil, errors.New("refresh token generation failed"))
			},
			wantErr:     true,
			expectedErr: ErrInternal,
		},
		{
			name: "session_creation_error",
			req: &LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			userAgent: "Mozilla/5.0",
			ipAddress: "127.0.0.1",
			setup: func(
				mockStore *dbmocks.MockStoreInterface,
				mockToken *tokenmocks.MockTokenManager,
				hashedPassword string,
			) {
				mockStore.EXPECT().
					GetUserByEmail(gomock.Any(), "test@example.com").
					Return(db.User{
						ID:           "user-123",
						Email:        "test@example.com",
						PasswordHash: hashedPassword,
					}, nil)

				mockStore.EXPECT().
					GetEmployeeByUserID(gomock.Any(), "user-123").
					Return(db.GetEmployeeByUserIDRow{ID: "employee-123"}, nil)

				mockToken.EXPECT().
					GenerateAccessToken("user-123", "employee-123", gomock.Any()).
					Return("access-token", nil)

				mockToken.EXPECT().
					GenerateRefreshToken("user-123", gomock.Any()).
					Return("refresh-token", createTestRefreshClaims("token-hash", "token-family"), nil)

				mockStore.EXPECT().
					CreateUserSession(gomock.Any(), gomock.Any()).
					Return(errors.New("database error"))
			},
			wantErr:     true,
			expectedErr: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := dbmocks.NewMockStoreInterface(ctrl)
			mockToken := tokenmocks.NewMockTokenManager(ctrl)
			mockLogger := loggermocks.NewMockLogger(ctrl)

			// Allow logger calls with any args
			mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			hashedPassword := hashPassword(t, "password123")
			tt.setup(mockStore, mockToken, hashedPassword)

			service := NewAuthService(mockStore, mockToken, mockLogger)

			resp, err := service.Login(context.Background(), tt.req, tt.userAgent, tt.ipAddress)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
				return
			}

			require.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, resp)
			}
		})
	}
}

// ============================================================
// Test: RefreshTokens
// ============================================================

func TestRefreshTokens(t *testing.T) {
	tests := []struct {
		name        string
		req         *RefreshTokensRequest
		userAgent   string
		ipAddress   string
		setup       func(mockStore *dbmocks.MockStoreInterface, mockToken *tokenmocks.MockTokenManager)
		wantErr     bool
		expectedErr error
		validate    func(t *testing.T, resp *RefreshTokensResponse)
	}{
		{
			name: "success",
			req: &RefreshTokensRequest{
				RefreshToken: "valid-refresh-token",
			},
			userAgent: "Mozilla/5.0",
			ipAddress: "127.0.0.1",
			setup: func(mockStore *dbmocks.MockStoreInterface, mockToken *tokenmocks.MockTokenManager) {
				mockToken.EXPECT().
					ValidateRefreshToken("valid-refresh-token").
					Return(createTestRefreshClaims("token-hash", "token-family"), nil)

				mockStore.EXPECT().
					GetUserSession(gomock.Any(), "token-hash").
					Return(db.Session{
						ID:          "session-123",
						UserID:      "user-123",
						TokenFamily: "token-family",
						TokenHash:   "token-hash",
						ExpiresAt:   pgtype.Timestamptz{Time: time.Now().Add(24 * time.Hour), Valid: true},
					}, nil)

				mockStore.EXPECT().
					GetEmployeeByUserID(gomock.Any(), "user-123").
					Return(db.GetEmployeeByUserIDRow{ID: "employee-123"}, nil)

				mockToken.EXPECT().
					GenerateAccessToken("user-123", "employee-123", gomock.Any()).
					Return("new-access-token", nil)

				mockToken.EXPECT().
					GenerateRefreshToken("user-123", gomock.Any()).
					Return("new-refresh-token", createTestRefreshClaims("new-token-hash", "token-family"), nil)

				mockStore.EXPECT().
					UpdateUserSession(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: false,
			validate: func(t *testing.T, resp *RefreshTokensResponse) {
				assert.Equal(t, "new-access-token", resp.AccessToken)
				assert.Equal(t, "new-refresh-token", resp.RefreshToken)
			},
		},
		{
			name: "invalid_refresh_token",
			req: &RefreshTokensRequest{
				RefreshToken: "invalid-token",
			},
			userAgent: "Mozilla/5.0",
			ipAddress: "127.0.0.1",
			setup: func(mockStore *dbmocks.MockStoreInterface, mockToken *tokenmocks.MockTokenManager) {
				mockToken.EXPECT().
					ValidateRefreshToken("invalid-token").
					Return(nil, errors.New("invalid token"))
			},
			wantErr:     true,
			expectedErr: ErrInvalidToken,
		},
		{
			name: "session_not_found",
			req: &RefreshTokensRequest{
				RefreshToken: "valid-refresh-token",
			},
			userAgent: "Mozilla/5.0",
			ipAddress: "127.0.0.1",
			setup: func(mockStore *dbmocks.MockStoreInterface, mockToken *tokenmocks.MockTokenManager) {
				mockToken.EXPECT().
					ValidateRefreshToken("valid-refresh-token").
					Return(createTestRefreshClaims("token-hash", "token-family"), nil)

				mockStore.EXPECT().
					GetUserSession(gomock.Any(), "token-hash").
					Return(db.Session{}, pgx.ErrNoRows)
			},
			wantErr:     true,
			expectedErr: ErrInvalidToken,
		},
		{
			name: "token_family_mismatch",
			req: &RefreshTokensRequest{
				RefreshToken: "valid-refresh-token",
			},
			userAgent: "Mozilla/5.0",
			ipAddress: "127.0.0.1",
			setup: func(mockStore *dbmocks.MockStoreInterface, mockToken *tokenmocks.MockTokenManager) {
				mockToken.EXPECT().
					ValidateRefreshToken("valid-refresh-token").
					Return(createTestRefreshClaims("token-hash", "token-family-A"), nil)

				mockStore.EXPECT().
					GetUserSession(gomock.Any(), "token-hash").
					Return(db.Session{
						ID:          "session-123",
						UserID:      "user-123",
						TokenFamily: "token-family-B", // Different family
						TokenHash:   "token-hash",
					}, nil)
			},
			wantErr:     true,
			expectedErr: ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := dbmocks.NewMockStoreInterface(ctrl)
			mockToken := tokenmocks.NewMockTokenManager(ctrl)
			mockLogger := loggermocks.NewMockLogger(ctrl)

			mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			tt.setup(mockStore, mockToken)

			service := NewAuthService(mockStore, mockToken, mockLogger)

			resp, err := service.RefreshTokens(context.Background(), tt.req, tt.userAgent, tt.ipAddress)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
				return
			}

			require.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, resp)
			}
		})
	}
}

// ============================================================
// Test: Logout
// ============================================================

func TestLogout(t *testing.T) {
	tests := []struct {
		name        string
		req         *LogoutRequest
		ctx         context.Context
		setup       func(mockStore *dbmocks.MockStoreInterface)
		wantErr     bool
		expectedErr error
	}{
		{
			name: "success",
			req: &LogoutRequest{
				RefreshToken: "token-hash-to-delete",
			},
			ctx: context.WithValue(context.Background(), "user_id", "user-123"),
			setup: func(mockStore *dbmocks.MockStoreInterface) {
				mockStore.EXPECT().
					DeleteUserSession(gomock.Any(), "token-hash-to-delete").
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "missing_user_id",
			req: &LogoutRequest{
				RefreshToken: "token-hash",
			},
			ctx:         context.Background(), // No user_id in context
			setup:       func(mockStore *dbmocks.MockStoreInterface) {},
			wantErr:     true,
			expectedErr: ErrInvalidToken,
		},
		{
			name: "delete_session_error",
			req: &LogoutRequest{
				RefreshToken: "token-hash",
			},
			ctx: context.WithValue(context.Background(), "user_id", "user-123"),
			setup: func(mockStore *dbmocks.MockStoreInterface) {
				mockStore.EXPECT().
					DeleteUserSession(gomock.Any(), "token-hash").
					Return(errors.New("database error"))
			},
			wantErr:     true,
			expectedErr: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := dbmocks.NewMockStoreInterface(ctrl)
			mockLogger := loggermocks.NewMockLogger(ctrl)

			mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			tt.setup(mockStore)

			service := NewAuthService(mockStore, nil, mockLogger)

			err := service.Logout(tt.ctx, tt.req)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
				return
			}

			require.NoError(t, err)
		})
	}
}
