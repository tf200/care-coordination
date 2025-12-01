package auth

import (
	db "care-cordination/lib/db/sqlc"
	"care-cordination/lib/logger"
	"care-cordination/lib/nanoid"
	"care-cordination/lib/token"
	"care-cordination/lib/util"
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type authService struct {
	store        *db.Store
	tokenManager *token.TokenManager
	logger       *logger.Logger
}

func NewAuthService(store *db.Store, tokenManager *token.TokenManager, logger *logger.Logger) AuthService {
	return &authService{
		store:        store,
		tokenManager: tokenManager,
		logger:       logger,
	}
}

func (s *authService) Login(ctx context.Context, req *LoginRequest, userAgent string, ipAddress string) (*LoginResponse, error) {
	user, err := s.store.GetUserByEmail(ctx, req.Email)
	if err != nil {
		s.logger.Error(ctx, "Login", "User not found", zap.String("email", req.Email))
		return nil, ErrInvalidCredentials
	}

	if err := s.comparePassword(user.PasswordHash, req.Password); err != nil {
		s.logger.Error(ctx, "Login", "Invalid password", zap.String("email", req.Email))
		return nil, ErrInvalidCredentials
	}
	accessToken, err := s.tokenManager.GenerateAccessToken(user.ID, time.Now())
	if err != nil {
		s.logger.Error(ctx, "Login", "Failed to generate access token", zap.String("email", req.Email))
		return nil, ErrInternal
	}
	refreshToken, refreshClaims, err := s.tokenManager.GenerateRefreshToken(user.ID, time.Now())
	if err != nil {
		s.logger.Error(ctx, "Login", "Failed to generate refresh token", zap.String("email", req.Email))
		return nil, ErrInternal
	}
	if err := s.store.CreateUserSession(ctx,
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

func (s *authService) RefreshTokens(ctx context.Context, req *RefreshTokensRequest, userAgent string, ipAddress string) (*RefreshTokensResponse, error) {
	refreshClaims, err := s.tokenManager.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		s.logger.Error(ctx, "RefreshTokens", "Invalid refresh token", zap.String("refreshToken", req.RefreshToken))
		return nil, ErrInvalidToken
	}
	userSession, err := s.store.GetUserSession(ctx, refreshClaims.TokenHash)
	if err != nil {
		s.logger.Error(ctx, "RefreshTokens", "User session not found", zap.String("refreshToken", req.RefreshToken))
		return nil, ErrInvalidToken
	}
	if userSession.TokenFamily != refreshClaims.Tokenfamily {
		s.logger.Error(ctx, "RefreshTokens", "Invalid refresh token", zap.String("refreshToken", req.RefreshToken))
		return nil, ErrInvalidToken
	}
	accessToken, err := s.tokenManager.GenerateAccessToken(userSession.UserID, time.Now())
	if err != nil {
		s.logger.Error(ctx, "RefreshTokens", "Failed to generate access token", zap.String("refreshToken", req.RefreshToken))
		return nil, ErrInternal
	}
	refreshToken, refreshClaims, err := s.tokenManager.GenerateRefreshToken(userSession.UserID, time.Now())
	if err != nil {
		s.logger.Error(ctx, "RefreshTokens", "Failed to generate refresh token", zap.String("refreshToken", req.RefreshToken))
		return nil, ErrInternal
	}
	if err := s.store.UpdateUserSession(ctx,
		db.UpdateUserSessionParams{
			ID:          userSession.ID,
			TokenHash:   refreshClaims.TokenHash,
			TokenFamily: refreshClaims.Tokenfamily,
			ExpiresAt:   pgtype.Timestamptz{Time: refreshClaims.ExpiresAt.Time, Valid: true},
			UserAgent:   &userAgent,
			IpAddress:   &ipAddress,
		}); err != nil {
		s.logger.Error(ctx, "RefreshTokens", "Failed to store user session", zap.String("refreshToken", req.RefreshToken))
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
	if err := s.store.DeleteUserSession(ctx, req.RefreshToken); err != nil {
		s.logger.Error(ctx, "Logout", "Failed to delete user session", zap.String("refreshToken", req.RefreshToken))
		return ErrInternal
	}
	return nil
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
