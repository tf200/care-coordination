package auth

import (
	"care-cordination/lib/middleware"
	"care-cordination/lib/resp"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService AuthService
	middleware  *middleware.Middleware
}

func NewAuthHandler(authService AuthService, middleware *middleware.Middleware) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		middleware:  middleware,
	}
}

func (h *AuthHandler) SetupAuthRoutes(router *gin.Engine) {
	auth := router.Group("/auth")
	auth.POST("/login", h.Login)
	auth.POST("/refresh", h.middleware.AuthMiddleware(), h.RefreshTokens)
	auth.POST("/logout", h.middleware.AuthMiddleware(), h.Logout)

}

// @Summary Login a user
// @Description Authenticate user and return access and refresh tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Param loginRequest body LoginRequest true "Login Request"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.ErrorResponse(ErrInvalidRequest))
		return
	}

	user, err := h.authService.Login(ctx, &req, ctx.Request.UserAgent(), ctx.ClientIP())
	if err != nil {
		switch err {
		case ErrInvalidCredentials:
			ctx.JSON(http.StatusUnauthorized, resp.ErrorResponse(err))
		case ErrInternal:
			ctx.JSON(http.StatusInternalServerError, resp.ErrorResponse(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.ErrorResponse(err))
		}
		return
	}
	ctx.JSON(http.StatusOK, user)
}

// @Summary Refresh tokens
// @Description Refresh access and refresh tokens using a valid refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param refreshTokensRequest body RefreshTokensRequest true "Refresh Tokens Request"
// @Success 200 {object} RefreshTokensResponse
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshTokens(ctx *gin.Context) {
	var req RefreshTokensRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.ErrorResponse(ErrInvalidRequest))
		return
	}
	tokens, err := h.authService.RefreshTokens(ctx, &req, ctx.Request.UserAgent(), ctx.ClientIP())
	if err != nil {
		switch err {
		case ErrInvalidToken:
			ctx.JSON(http.StatusUnauthorized, resp.ErrorResponse(err))
		case ErrInternal:
			ctx.JSON(http.StatusInternalServerError, resp.ErrorResponse(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.ErrorResponse(err))
		}
		return
	}
	ctx.JSON(http.StatusOK, tokens)
}

// @Summary Logout a user
// @Description Logout user by invalidating the refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param logoutRequest body LogoutRequest true "Logout Request"
// @Success 200 {object} resp.MessageResonse
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(ctx *gin.Context) {
	var req LogoutRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.ErrorResponse(ErrInvalidRequest))
		return
	}
	err := h.authService.Logout(ctx, &req)
	if err != nil {
		switch err {
		case ErrInvalidToken:
			ctx.JSON(http.StatusUnauthorized, resp.ErrorResponse(err))
		case ErrInternal:
			ctx.JSON(http.StatusInternalServerError, resp.ErrorResponse(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.ErrorResponse(err))
		}
		return
	}
	ctx.JSON(http.StatusOK, resp.MessageResonse("Successfully logged out"))
}
