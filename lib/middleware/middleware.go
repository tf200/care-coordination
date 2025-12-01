package middleware

import (
	"care-cordination/lib/resp"
	"care-cordination/lib/token"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
	UserIDKey               = "user_id"
)

type Middleware struct {
	tokenMaker *token.TokenManager
}

func NewMiddleware(tokenMaker *token.TokenManager) *Middleware {
	return &Middleware{
		tokenMaker: tokenMaker,
	}
}

func (m *Middleware) AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationHeader := ctx.GetHeader(authorizationHeaderKey)
		if len(authorizationHeader) == 0 {
			err := ErrInvalidRequest
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, resp.ErrorResponse(err))
			return
		}

		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			err := ErrInvalidRequest
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, resp.ErrorResponse(err))
			return
		}

		authorizationType := strings.ToLower(fields[0])
		if authorizationType != authorizationTypeBearer {
			err := ErrInvalidRequest
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, resp.ErrorResponse(err))
			return
		}

		accessToken := fields[1]
		payload, err := m.tokenMaker.ValidateAccessToken(accessToken)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, resp.ErrorResponse(err))
			return
		}

		ctx.Set(UserIDKey, payload.Subject)
		ctx.Next()
	}
}
