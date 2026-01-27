package auth

import (
	"care-cordination/lib/middleware"
	"care-cordination/lib/ratelimit"
	"care-cordination/lib/resp"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type AuthHandler struct {
	authService AuthService
	mdw         *middleware.Middleware
}

func NewAuthHandler(authService AuthService, mdw *middleware.Middleware) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		mdw:         mdw,
	}
}

func (h *AuthHandler) SetupAuthRoutes(
	router *gin.Engine,
	rateLimiter ratelimit.RateLimiter,
) {
	auth := router.Group("/auth")

	// Apply rate limiting to login endpoint
	if rateLimiter != nil {
		auth.POST("/login", h.mdw.LoginRateLimitMiddleware(rateLimiter), h.Login)
	} else {
		auth.POST("/login", h.Login)
	}

	auth.POST("/refresh", h.RefreshTokens)
	auth.POST("/logout", h.mdw.AuthMdw(), h.Logout)
	auth.POST("/reset-password", h.mdw.AuthMdw(), h.ResetPassword)
	auth.POST("/mfa/setup", h.mdw.AuthMdw(), h.SetupMFA)
	auth.POST("/mfa/enable", h.mdw.AuthMdw(), h.EnableMFA)
	if rateLimiter != nil {
		auth.POST("/mfa/verify", h.mdw.RateLimitMiddleware(), h.VerifyMFA)
	} else {
		auth.POST("/mfa/verify", h.VerifyMFA)
	}
	auth.POST("/mfa/disable", h.mdw.AuthMdw(), h.DisableMFA)
}

// @Summary Login a user
// @Description Authenticate user and return access and refresh tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Param loginRequest body LoginRequest true "Login Request"
// @Success 200 {object} resp.SuccessResponse[LoginResponse]
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /auth/login [post]
// @Security -
func (h *AuthHandler) Login(ctx *gin.Context) {
	var req LoginRequest
	// Use ShouldBindBodyWith to read from cached body (if rate limiting middleware ran)
	if err := ctx.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(err))
		return
	}

	user, err := h.authService.Login(ctx, &req, ctx.Request.UserAgent(), ctx.ClientIP())
	if err != nil {
		switch err {
		case ErrInvalidCredentials:
			ctx.JSON(http.StatusUnauthorized, resp.Error(err))
		case ErrInternal:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		}
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(user, "Login successful"))
}

// @Summary Refresh tokens
// @Description Refresh access and refresh tokens using a valid refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param refreshTokensRequest body RefreshTokensRequest true "Refresh Tokens Request"
// @Success 200 {object} resp.SuccessResponse[RefreshTokensResponse]
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshTokens(ctx *gin.Context) {
	var req RefreshTokensRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}
	tokens, err := h.authService.RefreshTokens(ctx, &req, ctx.Request.UserAgent(), ctx.ClientIP())
	if err != nil {
		switch err {
		case ErrInvalidToken:
			ctx.JSON(http.StatusUnauthorized, resp.Error(err))
		case ErrInternal:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		}
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(tokens, "Tokens refreshed successfully"))
}

// @Summary Logout a user
// @Description Logout user by invalidating the refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param logoutRequest body LogoutRequest true "Logout Request"
// @Success 200 {object} resp.MessageResponse
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(ctx *gin.Context) {
	var req LogoutRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}
	err := h.authService.Logout(ctx, &req)
	if err != nil {
		switch err {
		case ErrInvalidToken:
			ctx.JSON(http.StatusUnauthorized, resp.Error(err))
		case ErrInternal:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		}
		return
	}
	ctx.JSON(http.StatusOK, resp.MessageResonse("Successfully logged out"))
}

// @Summary Reset password
// @Description Reset password for the authenticated user
// @Tags Auth
// @Accept json
// @Produce json
// @Param resetPasswordRequest body ResetPasswordRequest true "Reset Password Request"
// @Success 200 {object} resp.MessageResponse
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /auth/reset-password [post]
func (h *AuthHandler) ResetPassword(ctx *gin.Context) {
	var req ResetPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}

	if err := h.authService.ResetPassword(ctx, &req); err != nil {
		switch err {
		case ErrInvalidCredentials, ErrInvalidToken:
			ctx.JSON(http.StatusUnauthorized, resp.Error(err))
		case ErrInvalidRequest:
			ctx.JSON(http.StatusBadRequest, resp.Error(err))
		case ErrInternal:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		}
		return
	}

	ctx.JSON(http.StatusOK, resp.MessageResonse("Password reset successfully"))
}

// @Summary Setup MFA
// @Description Generate MFA secret and OTP auth URL for the authenticated user
// @Tags Auth
// @Produce json
// @Success 200 {object} resp.SuccessResponse[SetupMFAResponse]
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /auth/mfa/setup [post]
func (h *AuthHandler) SetupMFA(ctx *gin.Context) {
	result, err := h.authService.SetupMFA(ctx)
	if err != nil {
		switch err {
		case ErrInternal:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		}
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(result, "MFA setup generated successfully"))
}

// @Summary Enable MFA
// @Description Verify code and enable MFA for the authenticated user
// @Tags Auth
// @Accept json
// @Produce json
// @Param enableMFARequest body EnableMFARequest true "Enable MFA Request"
// @Success 200 {object} resp.SuccessResponse[EnableMFAResponse]
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /auth/mfa/enable [post]
func (h *AuthHandler) EnableMFA(ctx *gin.Context) {
	var req EnableMFARequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}

	result, err := h.authService.EnableMFA(ctx, &req)
	if err != nil {
		switch err {
		case ErrInvalidMFACode, ErrMFANotSetup:
			ctx.JSON(http.StatusUnauthorized, resp.Error(err))
		case ErrMFAAlreadyEnabled:
			ctx.JSON(http.StatusBadRequest, resp.Error(err))
		case ErrInternal:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		}
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(result, "MFA enabled successfully"))
}

// @Summary Verify MFA
// @Description Verify MFA code using pre-auth token
// @Tags Auth
// @Accept json
// @Produce json
// @Param verifyMFARequest body VerifyMFARequest true "Verify MFA Request"
// @Success 200 {object} resp.SuccessResponse[VerifyMFAResponse]
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /auth/mfa/verify [post]
func (h *AuthHandler) VerifyMFA(ctx *gin.Context) {
	var req VerifyMFARequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}

	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		ctx.JSON(http.StatusUnauthorized, resp.Error(ErrInvalidToken))
		return
	}
	fields := strings.Fields(authHeader)
	if len(fields) < 2 || strings.ToLower(fields[0]) != "bearer" {
		ctx.JSON(http.StatusUnauthorized, resp.Error(ErrInvalidToken))
		return
	}
	preAuthToken := fields[1]

	result, err := h.authService.VerifyMFA(ctx, &req, preAuthToken)
	if err != nil {
		switch err {
		case ErrInvalidMFACode, ErrInvalidToken, ErrMFANotSetup:
			ctx.JSON(http.StatusUnauthorized, resp.Error(err))
		case ErrInternal:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		}
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(result, "MFA verified successfully"))
}

// @Summary Disable MFA
// @Description Disable MFA for the authenticated user
// @Tags Auth
// @Accept json
// @Produce json
// @Param disableMFARequest body DisableMFARequest true "Disable MFA Request"
// @Success 200 {object} resp.MessageResponse
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /auth/mfa/disable [post]
func (h *AuthHandler) DisableMFA(ctx *gin.Context) {
	var req DisableMFARequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}

	if err := h.authService.DisableMFA(ctx, &req); err != nil {
		switch err {
		case ErrInvalidCredentials:
			ctx.JSON(http.StatusUnauthorized, resp.Error(err))
		case ErrInternal:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		}
		return
	}

	ctx.JSON(http.StatusOK, resp.MessageResonse("MFA disabled successfully"))
}
