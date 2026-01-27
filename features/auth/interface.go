package auth

import "context"

//go:generate mockgen -destination=../../internal/mocks/mock_auth_service.go -package=mocks care-cordination/features/auth AuthService
type AuthService interface {
	Login(
		ctx context.Context,
		req *LoginRequest,
		userAgent string,
		ipAddress string,
	) (*LoginResponse, error)
	RefreshTokens(
		ctx context.Context,
		req *RefreshTokensRequest,
		userAgent string,
		ipAddress string,
	) (*RefreshTokensResponse, error)
	Logout(ctx context.Context, req *LogoutRequest) error
	ResetPassword(ctx context.Context, req *ResetPasswordRequest) error
	SetupMFA(ctx context.Context) (*SetupMFAResponse, error)
	EnableMFA(ctx context.Context, req *EnableMFARequest) (*EnableMFAResponse, error)
	VerifyMFA(ctx context.Context, req *VerifyMFARequest, preAuthToken string) (*VerifyMFAResponse, error)
	DisableMFA(ctx context.Context, req *DisableMFARequest) error
}
