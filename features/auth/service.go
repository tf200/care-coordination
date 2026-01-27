package auth

import (
	db "care-cordination/lib/db/sqlc"
	"care-cordination/lib/logger"
	"care-cordination/lib/nanoid"
	"care-cordination/lib/token"
	"care-cordination/lib/util"
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pquerna/otp/totp"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type authService struct {
	db           db.StoreInterface
	tokenManager token.TokenManager
	logger       logger.Logger
	mfaSecretKey string
	mfaIssuer    string
}

func NewAuthService(
	db db.StoreInterface,
	tokenManager token.TokenManager,
	logger logger.Logger,
) AuthService {
	return &authService{
		db:           db,
		tokenManager: tokenManager,
		logger:       logger,
		mfaIssuer:    "care-coordination",
	}
}

func NewAuthServiceWithMFA(
	db db.StoreInterface,
	tokenManager token.TokenManager,
	logger logger.Logger,
	mfaSecretKey string,
	mfaIssuer string,
) AuthService {
	return &authService{
		db:           db,
		tokenManager: tokenManager,
		logger:       logger,
		mfaSecretKey: mfaSecretKey,
		mfaIssuer:    mfaIssuer,
	}
}

func (s *authService) Login(
	ctx context.Context,
	req *LoginRequest,
	userAgent string,
	ipAddress string,
) (*LoginResponse, error) {
	user, err := s.db.GetUserByEmail(ctx, req.Email)
	if err != nil {
		s.logger.Error(ctx, "Login", "User not found", zap.String("email", req.Email))
		return nil, ErrInvalidCredentials
	}

	if err := s.comparePassword(user.PasswordHash, req.Password); err != nil {
		s.logger.Error(ctx, "Login", "Invalid password", zap.String("email", req.Email))
		return nil, ErrInvalidCredentials
	}

	employee, err := s.db.GetEmployeeByUserID(ctx, user.ID)
	if err != nil {
		s.logger.Error(ctx, "Login", "Employee not found for user", zap.String("email", req.Email), zap.Error(err))
		return nil, ErrInvalidCredentials
	}

	if user.IsMfaEnabled {
		preAuthToken, err := s.tokenManager.GenerateMFAPendingToken(user.ID, time.Now())
		if err != nil {
			s.logger.Error(
				ctx,
				"Login",
				"Failed to generate MFA pre-auth token",
				zap.String("email", req.Email),
			)
			return nil, ErrInternal
		}
		return &LoginResponse{
			MFARequired:  true,
			PreAuthToken: preAuthToken,
		}, nil
	}

	accessToken, err := s.tokenManager.GenerateAccessToken(user.ID, employee.ID, time.Now())
	if err != nil {
		s.logger.Error(
			ctx,
			"Login",
			"Failed to generate access token",
			zap.String("email", req.Email),
		)
		return nil, ErrInternal
	}
	refreshToken, refreshClaims, err := s.tokenManager.GenerateRefreshToken(user.ID, time.Now())
	if err != nil {
		s.logger.Error(
			ctx,
			"Login",
			"Failed to generate refresh token",
			zap.String("email", req.Email),
		)
		return nil, ErrInternal
	}
	if err := s.db.CreateUserSession(ctx,
		db.CreateUserSessionParams{
			ID:          nanoid.Generate(),
			UserID:      user.ID,
			TokenHash:   refreshClaims.TokenHash,
			TokenFamily: refreshClaims.Tokenfamily,
			ExpiresAt:   pgtype.Timestamptz{Time: refreshClaims.ExpiresAt.Time, Valid: true},
			UserAgent:   &userAgent,
			IpAddress:   &ipAddress,
		}); err != nil {
		return nil, ErrInternal
	}
	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil

}

func (s *authService) RefreshTokens(
	ctx context.Context,
	req *RefreshTokensRequest,
	userAgent string,
	ipAddress string,
) (*RefreshTokensResponse, error) {
	refreshClaims, err := s.tokenManager.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		s.logger.Error(
			ctx,
			"RefreshTokens",
			"Invalid refresh token",
			zap.String("refreshToken", req.RefreshToken),
		)
		return nil, ErrInvalidToken
	}
	userSession, err := s.db.GetUserSession(ctx, refreshClaims.TokenHash)
	if err != nil {
		s.logger.Error(
			ctx,
			"RefreshTokens",
			"User session not found",
			zap.String("refreshToken", req.RefreshToken),
		)
		return nil, ErrInvalidToken
	}
	if userSession.TokenFamily != refreshClaims.Tokenfamily {
		s.logger.Error(
			ctx,
			"RefreshTokens",
			"Invalid refresh token",
			zap.String("refreshToken", req.RefreshToken),
		)
		return nil, ErrInvalidToken
	}
	// Try to get employee ID
	employeeID := ""
	employee, err := s.db.GetEmployeeByUserID(ctx, userSession.UserID)
	if err == nil {
		employeeID = employee.ID
	}

	accessToken, err := s.tokenManager.GenerateAccessToken(
		userSession.UserID,
		employeeID,
		time.Now(),
	)
	if err != nil {
		s.logger.Error(
			ctx,
			"RefreshTokens",
			"Failed to generate access token",
			zap.String("refreshToken", req.RefreshToken),
		)
		return nil, ErrInternal
	}
	refreshToken, refreshClaims, err := s.tokenManager.GenerateRefreshToken(
		userSession.UserID,
		time.Now(),
	)
	if err != nil {
		s.logger.Error(
			ctx,
			"RefreshTokens",
			"Failed to generate refresh token",
			zap.String("refreshToken", req.RefreshToken),
		)
		return nil, ErrInternal
	}
	if err := s.db.UpdateUserSession(ctx,
		db.UpdateUserSessionParams{
			ID:          userSession.ID,
			TokenHash:   refreshClaims.TokenHash,
			TokenFamily: refreshClaims.Tokenfamily,
			ExpiresAt:   pgtype.Timestamptz{Time: refreshClaims.ExpiresAt.Time, Valid: true},
			UserAgent:   &userAgent,
			IpAddress:   &ipAddress,
		}); err != nil {
		s.logger.Error(
			ctx,
			"RefreshTokens",
			"Failed to store user session",
			zap.String("refreshToken", req.RefreshToken),
		)
		return nil, ErrInternal
	}
	return &RefreshTokensResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *authService) Logout(ctx context.Context, req *LogoutRequest) error {
	userID := util.GetUserID(ctx)
	if userID == "" {
		return ErrInvalidToken
	}
	if err := s.db.DeleteUserSession(ctx, req.RefreshToken); err != nil {
		s.logger.Error(
			ctx,
			"Logout",
			"Failed to delete user session",
			zap.String("refreshToken", req.RefreshToken),
		)
		return ErrInternal
	}
	return nil
}

func (s *authService) ResetPassword(ctx context.Context, req *ResetPasswordRequest) error {
	userID := util.GetUserID(ctx)
	if userID == "" {
		return ErrInvalidToken
	}

	user, err := s.db.GetUserByID(ctx, userID)
	if err != nil {
		s.logger.Error(ctx, "ResetPassword", "Failed to get user", zap.String("userID", userID), zap.Error(err))
		return ErrInternal
	}

	if err := s.comparePassword(user.PasswordHash, req.CurrentPassword); err != nil {
		return ErrInvalidCredentials
	}

	passwordHash, err := s.hashPassword(req.NewPassword)
	if err != nil {
		s.logger.Error(ctx, "ResetPassword", "Failed to generate password hash", zap.Error(err))
		return ErrInternal
	}

	if err := s.db.UpdateUser(ctx, db.UpdateUserParams{
		ID:           userID,
		PasswordHash: &passwordHash,
	}); err != nil {
		s.logger.Error(ctx, "ResetPassword", "Failed to update user password", zap.String("userID", userID), zap.Error(err))
		return ErrInternal
	}

	return nil
}

func (s *authService) SetupMFA(ctx context.Context) (*SetupMFAResponse, error) {
	userID := util.GetUserID(ctx)
	if userID == "" {
		return nil, ErrInvalidToken
	}

	user, err := s.db.GetUserByID(ctx, userID)
	if err != nil {
		s.logger.Error(ctx, "SetupMFA", "Failed to get user", zap.String("userID", userID), zap.Error(err))
		return nil, ErrInternal
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      s.mfaIssuer,
		AccountName: user.Email,
	})
	if err != nil {
		s.logger.Error(ctx, "SetupMFA", "Failed to generate MFA secret", zap.String("userID", userID), zap.Error(err))
		return nil, ErrInternal
	}

	encryptedSecret, err := util.EncryptString(s.mfaSecretKey, key.Secret())
	if err != nil {
		s.logger.Error(ctx, "SetupMFA", "Failed to encrypt MFA secret", zap.String("userID", userID), zap.Error(err))
		return nil, ErrInternal
	}

	if err := s.db.UpdateUserMFASecret(ctx, db.UpdateUserMFASecretParams{
		ID:        userID,
		MfaSecret: &encryptedSecret,
	}); err != nil {
		s.logger.Error(ctx, "SetupMFA", "Failed to store MFA secret", zap.String("userID", userID), zap.Error(err))
		return nil, ErrInternal
	}

	return &SetupMFAResponse{
		Secret:     key.Secret(),
		OtpAuthURL: key.URL(),
	}, nil
}

func (s *authService) EnableMFA(ctx context.Context, req *EnableMFARequest) (*EnableMFAResponse, error) {
	userID := util.GetUserID(ctx)
	if userID == "" {
		return nil, ErrInvalidToken
	}

	mfaState, err := s.db.GetUserMFAState(ctx, userID)
	if err != nil {
		s.logger.Error(ctx, "EnableMFA", "Failed to get MFA state", zap.String("userID", userID), zap.Error(err))
		return nil, ErrInternal
	}
	if mfaState.IsMfaEnabled {
		return nil, ErrMFAAlreadyEnabled
	}
	if mfaState.MfaSecret == nil || *mfaState.MfaSecret == "" {
		return nil, ErrMFANotSetup
	}

	secret, err := util.DecryptString(s.mfaSecretKey, *mfaState.MfaSecret)
	if err != nil {
		s.logger.Error(ctx, "EnableMFA", "Failed to decrypt MFA secret", zap.String("userID", userID), zap.Error(err))
		return nil, ErrInternal
	}

	if !totp.Validate(req.Code, secret) {
		return nil, ErrInvalidMFACode
	}

	backupCodes := s.generateBackupCodes(10)
	backupJSON, err := json.Marshal(backupCodes)
	if err != nil {
		s.logger.Error(ctx, "EnableMFA", "Failed to marshal backup codes", zap.String("userID", userID), zap.Error(err))
		return nil, ErrInternal
	}

	encryptedCodes, err := util.EncryptString(s.mfaSecretKey, string(backupJSON))
	if err != nil {
		s.logger.Error(ctx, "EnableMFA", "Failed to encrypt backup codes", zap.String("userID", userID), zap.Error(err))
		return nil, ErrInternal
	}

	if err := s.db.EnableUserMFA(ctx, db.EnableUserMFAParams{
		ID:             userID,
		MfaBackupCodes: &encryptedCodes,
	}); err != nil {
		s.logger.Error(ctx, "EnableMFA", "Failed to enable MFA", zap.String("userID", userID), zap.Error(err))
		return nil, ErrInternal
	}

	return &EnableMFAResponse{BackupCodes: backupCodes}, nil
}

func (s *authService) VerifyMFA(ctx context.Context, req *VerifyMFARequest, preAuthToken string) (*VerifyMFAResponse, error) {
	claims, err := s.tokenManager.ValidateAccessToken(preAuthToken)
	if err != nil || claims.Scope != token.ScopeMFAPending {
		return nil, ErrInvalidToken
	}

	userID := claims.Subject
	mfaState, err := s.db.GetUserMFAState(ctx, userID)
	if err != nil {
		s.logger.Error(ctx, "VerifyMFA", "Failed to get MFA state", zap.String("userID", userID), zap.Error(err))
		return nil, ErrInternal
	}
	if !mfaState.IsMfaEnabled || mfaState.MfaSecret == nil || *mfaState.MfaSecret == "" {
		return nil, ErrMFANotSetup
	}

	secret, err := util.DecryptString(s.mfaSecretKey, *mfaState.MfaSecret)
	if err != nil {
		s.logger.Error(ctx, "VerifyMFA", "Failed to decrypt MFA secret", zap.String("userID", userID), zap.Error(err))
		return nil, ErrInternal
	}

	if !totp.Validate(req.Code, secret) {
		return nil, ErrInvalidMFACode
	}

	employeeID := ""
	employee, err := s.db.GetEmployeeByUserID(ctx, userID)
	if err == nil {
		employeeID = employee.ID
	}

	accessToken, err := s.tokenManager.GenerateAccessToken(userID, employeeID, time.Now())
	if err != nil {
		s.logger.Error(ctx, "VerifyMFA", "Failed to generate access token", zap.String("userID", userID))
		return nil, ErrInternal
	}

	refreshToken, refreshClaims, err := s.tokenManager.GenerateRefreshToken(userID, time.Now())
	if err != nil {
		s.logger.Error(ctx, "VerifyMFA", "Failed to generate refresh token", zap.String("userID", userID))
		return nil, ErrInternal
	}

	userAgent := util.GetUserAgent(ctx)
	ipAddress := util.GetIPAddress(ctx)
	if err := s.db.CreateUserSession(ctx, db.CreateUserSessionParams{
		ID:          nanoid.Generate(),
		UserID:      userID,
		TokenHash:   refreshClaims.TokenHash,
		TokenFamily: refreshClaims.Tokenfamily,
		ExpiresAt:   pgtype.Timestamptz{Time: refreshClaims.ExpiresAt.Time, Valid: true},
		UserAgent:   &userAgent,
		IpAddress:   &ipAddress,
	}); err != nil {
		s.logger.Error(ctx, "VerifyMFA", "Failed to store user session", zap.String("userID", userID), zap.Error(err))
		return nil, ErrInternal
	}

	return &VerifyMFAResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *authService) DisableMFA(ctx context.Context, req *DisableMFARequest) error {
	userID := util.GetUserID(ctx)
	if userID == "" {
		return ErrInvalidToken
	}

	user, err := s.db.GetUserByID(ctx, userID)
	if err != nil {
		s.logger.Error(ctx, "DisableMFA", "Failed to get user", zap.String("userID", userID), zap.Error(err))
		return ErrInternal
	}

	if err := s.comparePassword(user.PasswordHash, req.Password); err != nil {
		return ErrInvalidCredentials
	}

	if err := s.db.DisableUserMFA(ctx, userID); err != nil {
		s.logger.Error(ctx, "DisableMFA", "Failed to disable MFA", zap.String("userID", userID), zap.Error(err))
		return ErrInternal
	}

	return nil
}

func (s *authService) generateBackupCodes(count int) []string {
	codes := make([]string, 0, count)
	for i := 0; i < count; i++ {
		code := nanoid.Generate()
		if len(code) > 10 {
			code = code[:10]
		}
		codes = append(codes, code)
	}
	return codes
}

func (s *authService) hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func (s *authService) comparePassword(hashedPassword string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
