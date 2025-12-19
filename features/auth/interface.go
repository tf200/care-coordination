package auth

import "context"

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
}
