package token

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testAccessSecret  = "test-access-secret-key-32-bytes!"
	testRefreshSecret = "test-refresh-secret-key-32-byte!"
	testAccessTTL     = 15 * time.Minute
	testRefreshTTL    = 7 * 24 * time.Hour
)

func newTestTokenManager() TokenManager {
	return NewTokenManager(testAccessSecret, testRefreshSecret, testAccessTTL, testRefreshTTL, 5*time.Minute)
}

// ============================================================
// Test: GenerateAccessToken
// ============================================================

func TestGenerateAccessToken(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		employeeID string
		now        time.Time
		wantErr    bool
		validate   func(t *testing.T, tm TokenManager, token string, userID, employeeID string, now time.Time)
	}{
		{
			name:       "success",
			userID:     "user-123",
			employeeID: "emp-456",
			now:        time.Now(),
			wantErr:    false,
			validate: func(t *testing.T, tm TokenManager, token string, userID, employeeID string, now time.Time) {
				claims, err := tm.ValidateAccessToken(token)
				require.NoError(t, err)
				assert.Equal(t, userID, claims.Subject)
				assert.Equal(t, employeeID, claims.EmployeeID)
				assert.Equal(t, "care-coordination", claims.Issuer)
				assert.Contains(t, claims.Audience, "care-coordination")
				assert.WithinDuration(t, now, claims.IssuedAt.Time, time.Second)
				assert.WithinDuration(t, now.Add(testAccessTTL), claims.ExpiresAt.Time, time.Second)
			},
		},
		{
			name:       "empty_user_id",
			userID:     "",
			employeeID: "emp-456",
			now:        time.Now(),
			wantErr:    false,
			validate: func(t *testing.T, tm TokenManager, token string, userID, employeeID string, now time.Time) {
				claims, err := tm.ValidateAccessToken(token)
				require.NoError(t, err)
				assert.Equal(t, "", claims.Subject)
				assert.Equal(t, employeeID, claims.EmployeeID)
			},
		},
		{
			name:       "empty_employee_id",
			userID:     "user-123",
			employeeID: "",
			now:        time.Now(),
			wantErr:    false,
			validate: func(t *testing.T, tm TokenManager, token string, userID, employeeID string, now time.Time) {
				claims, err := tm.ValidateAccessToken(token)
				require.NoError(t, err)
				assert.Equal(t, userID, claims.Subject)
				assert.Equal(t, "", claims.EmployeeID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := newTestTokenManager()

			token, err := tm.GenerateAccessToken(tt.userID, tt.employeeID, tt.now)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, token)
			if tt.validate != nil {
				tt.validate(t, tm, token, tt.userID, tt.employeeID, tt.now)
			}
		})
	}
}

// ============================================================
// Test: GenerateRefreshToken
// ============================================================

func TestGenerateRefreshToken(t *testing.T) {
	tests := []struct {
		name     string
		userID   string
		now      time.Time
		wantErr  bool
		validate func(t *testing.T, tm TokenManager, token string, claims *RefreshTokenClaims, userID string, now time.Time)
	}{
		{
			name:    "success",
			userID:  "user-123",
			now:     time.Now(),
			wantErr: false,
			validate: func(t *testing.T, tm TokenManager, token string, claims *RefreshTokenClaims, userID string, now time.Time) {
				assert.NotEmpty(t, claims.Tokenfamily)
				assert.NotEmpty(t, claims.TokenHash)
				assert.Equal(t, userID, claims.Subject)
				assert.Equal(t, "care-coordination", claims.Issuer)
				assert.Contains(t, claims.Audience, "care-coordination")
				assert.WithinDuration(t, now, claims.IssuedAt.Time, time.Second)
				assert.WithinDuration(t, now.Add(testRefreshTTL), claims.ExpiresAt.Time, time.Second)

				// Validate the token string
				validatedClaims, err := tm.ValidateRefreshToken(token)
				require.NoError(t, err)
				assert.Equal(t, claims.Tokenfamily, validatedClaims.Tokenfamily)
				assert.Equal(t, claims.TokenHash, validatedClaims.TokenHash)
			},
		},
		{
			name:    "empty_user_id",
			userID:  "",
			now:     time.Now(),
			wantErr: false,
			validate: func(t *testing.T, tm TokenManager, token string, claims *RefreshTokenClaims, userID string, now time.Time) {
				assert.Equal(t, "", claims.Subject)
				assert.NotEmpty(t, claims.Tokenfamily)
				assert.NotEmpty(t, claims.TokenHash)
			},
		},
		{
			name:    "unique_token_family_and_hash",
			userID:  "user-123",
			now:     time.Now(),
			wantErr: false,
			validate: func(t *testing.T, tm TokenManager, token string, claims *RefreshTokenClaims, userID string, now time.Time) {
				// Generate another token and ensure they're different
				_, claims2, err := tm.GenerateRefreshToken(userID, now)
				require.NoError(t, err)
				assert.NotEqual(t, claims.Tokenfamily, claims2.Tokenfamily)
				assert.NotEqual(t, claims.TokenHash, claims2.TokenHash)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := newTestTokenManager()

			token, claims, err := tm.GenerateRefreshToken(tt.userID, tt.now)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, token)
			require.NotNil(t, claims)
			if tt.validate != nil {
				tt.validate(t, tm, token, claims, tt.userID, tt.now)
			}
		})
	}
}

// ============================================================
// Test: ValidateAccessToken
// ============================================================

func TestValidateAccessToken(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, tm TokenManager) string // returns token to validate
		wantErr  bool
		checkErr func(t *testing.T, err error)
		validate func(t *testing.T, claims *AccessTokenClaims)
	}{
		{
			name: "valid_token",
			setup: func(t *testing.T, tm TokenManager) string {
				token, err := tm.GenerateAccessToken("user-123", "emp-456", time.Now())
				require.NoError(t, err)
				return token
			},
			wantErr: false,
			validate: func(t *testing.T, claims *AccessTokenClaims) {
				assert.Equal(t, "user-123", claims.Subject)
				assert.Equal(t, "emp-456", claims.EmployeeID)
			},
		},
		{
			name: "expired_token",
			setup: func(t *testing.T, tm TokenManager) string {
				// Generate a token with a time in the past
				expiredTime := time.Now().Add(-2 * testAccessTTL)
				token, err := tm.GenerateAccessToken("user-123", "emp-456", expiredTime)
				require.NoError(t, err)
				return token
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, jwt.ErrTokenExpired)
			},
		},
		{
			name: "invalid_token_string",
			setup: func(t *testing.T, tm TokenManager) string {
				return "invalid.token.string"
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "empty_token",
			setup: func(t *testing.T, tm TokenManager) string {
				return ""
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "wrong_secret",
			setup: func(t *testing.T, tm TokenManager) string {
				// Create a token with a different secret
				wrongTm := NewTokenManager("wrong-secret-key-32-bytes-long!", testRefreshSecret, testAccessTTL, testRefreshTTL, 5*time.Minute)
				token, err := wrongTm.GenerateAccessToken("user-123", "emp-456", time.Now())
				require.NoError(t, err)
				return token
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, jwt.ErrSignatureInvalid)
			},
		},
		{
			name: "tampered_token",
			setup: func(t *testing.T, tm TokenManager) string {
				token, err := tm.GenerateAccessToken("user-123", "emp-456", time.Now())
				require.NoError(t, err)
				// Tamper with the token by modifying a character
				runes := []rune(token)
				if runes[len(runes)-5] == 'a' {
					runes[len(runes)-5] = 'b'
				} else {
					runes[len(runes)-5] = 'a'
				}
				return string(runes)
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := newTestTokenManager()
			tokenStr := tt.setup(t, tm)

			claims, err := tm.ValidateAccessToken(tokenStr)

			if tt.wantErr {
				require.Error(t, err)
				if tt.checkErr != nil {
					tt.checkErr(t, err)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, claims)
			if tt.validate != nil {
				tt.validate(t, claims)
			}
		})
	}
}

// ============================================================
// Test: ValidateRefreshToken
// ============================================================

func TestValidateRefreshToken(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, tm TokenManager) string // returns token to validate
		wantErr  bool
		checkErr func(t *testing.T, err error)
		validate func(t *testing.T, claims *RefreshTokenClaims)
	}{
		{
			name: "valid_token",
			setup: func(t *testing.T, tm TokenManager) string {
				token, _, err := tm.GenerateRefreshToken("user-123", time.Now())
				require.NoError(t, err)
				return token
			},
			wantErr: false,
			validate: func(t *testing.T, claims *RefreshTokenClaims) {
				assert.Equal(t, "user-123", claims.Subject)
				assert.NotEmpty(t, claims.Tokenfamily)
				assert.NotEmpty(t, claims.TokenHash)
			},
		},
		{
			name: "expired_token",
			setup: func(t *testing.T, tm TokenManager) string {
				// Generate a token with a time in the past
				expiredTime := time.Now().Add(-2 * testRefreshTTL)
				token, _, err := tm.GenerateRefreshToken("user-123", expiredTime)
				require.NoError(t, err)
				return token
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, jwt.ErrTokenExpired)
			},
		},
		{
			name: "invalid_token_string",
			setup: func(t *testing.T, tm TokenManager) string {
				return "invalid.token.string"
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "empty_token",
			setup: func(t *testing.T, tm TokenManager) string {
				return ""
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "wrong_secret",
			setup: func(t *testing.T, tm TokenManager) string {
				// Create a token with a different secret
				wrongTm := NewTokenManager(testAccessSecret, "wrong-refresh-secret-32-bytes!!", testAccessTTL, testRefreshTTL, 5*time.Minute)
				token, _, err := wrongTm.GenerateRefreshToken("user-123", time.Now())
				require.NoError(t, err)
				return token
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, jwt.ErrSignatureInvalid)
			},
		},
		{
			name: "access_token_used_as_refresh",
			setup: func(t *testing.T, tm TokenManager) string {
				// Try to use an access token for refresh validation
				token, err := tm.GenerateAccessToken("user-123", "emp-456", time.Now())
				require.NoError(t, err)
				return token
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				// Should fail because it's signed with a different secret
				assert.ErrorIs(t, err, jwt.ErrSignatureInvalid)
			},
		},
		{
			name: "tampered_token",
			setup: func(t *testing.T, tm TokenManager) string {
				token, _, err := tm.GenerateRefreshToken("user-123", time.Now())
				require.NoError(t, err)
				// Tamper with the token by modifying a character
				runes := []rune(token)
				if runes[len(runes)-5] == 'a' {
					runes[len(runes)-5] = 'b'
				} else {
					runes[len(runes)-5] = 'a'
				}
				return string(runes)
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := newTestTokenManager()
			tokenStr := tt.setup(t, tm)

			claims, err := tm.ValidateRefreshToken(tokenStr)

			if tt.wantErr {
				require.Error(t, err)
				if tt.checkErr != nil {
					tt.checkErr(t, err)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, claims)
			if tt.validate != nil {
				tt.validate(t, claims)
			}
		})
	}
}

// ============================================================
// Test: Cross-validation (access vs refresh tokens)
// ============================================================

func TestCrossTokenValidation(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, tm TokenManager) (accessToken, refreshToken string)
		test  func(t *testing.T, tm TokenManager, accessToken, refreshToken string)
	}{
		{
			name: "refresh_token_cannot_be_validated_as_access",
			setup: func(t *testing.T, tm TokenManager) (string, string) {
				access, err := tm.GenerateAccessToken("user-123", "emp-456", time.Now())
				require.NoError(t, err)
				refresh, _, err := tm.GenerateRefreshToken("user-123", time.Now())
				require.NoError(t, err)
				return access, refresh
			},
			test: func(t *testing.T, tm TokenManager, accessToken, refreshToken string) {
				// Refresh token should fail when validated as access token
				_, err := tm.ValidateAccessToken(refreshToken)
				require.Error(t, err)
				assert.ErrorIs(t, err, jwt.ErrSignatureInvalid)
			},
		},
		{
			name: "access_token_cannot_be_validated_as_refresh",
			setup: func(t *testing.T, tm TokenManager) (string, string) {
				access, err := tm.GenerateAccessToken("user-123", "emp-456", time.Now())
				require.NoError(t, err)
				refresh, _, err := tm.GenerateRefreshToken("user-123", time.Now())
				require.NoError(t, err)
				return access, refresh
			},
			test: func(t *testing.T, tm TokenManager, accessToken, refreshToken string) {
				// Access token should fail when validated as refresh token
				_, err := tm.ValidateRefreshToken(accessToken)
				require.Error(t, err)
				assert.ErrorIs(t, err, jwt.ErrSignatureInvalid)
			},
		},
		{
			name: "both_tokens_valid_with_correct_validators",
			setup: func(t *testing.T, tm TokenManager) (string, string) {
				access, err := tm.GenerateAccessToken("user-123", "emp-456", time.Now())
				require.NoError(t, err)
				refresh, _, err := tm.GenerateRefreshToken("user-123", time.Now())
				require.NoError(t, err)
				return access, refresh
			},
			test: func(t *testing.T, tm TokenManager, accessToken, refreshToken string) {
				// Both tokens should validate correctly with their respective validators
				accessClaims, err := tm.ValidateAccessToken(accessToken)
				require.NoError(t, err)
				assert.Equal(t, "user-123", accessClaims.Subject)

				refreshClaims, err := tm.ValidateRefreshToken(refreshToken)
				require.NoError(t, err)
				assert.Equal(t, "user-123", refreshClaims.Subject)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := newTestTokenManager()
			accessToken, refreshToken := tt.setup(t, tm)
			tt.test(t, tm, accessToken, refreshToken)
		})
	}
}
