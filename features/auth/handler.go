package auth

import (
	"care-cordination/features/middleware"
	"care-cordination/lib/ratelimit"
	"care-cordination/lib/resp"
	"net/http"

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
