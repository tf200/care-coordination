package middleware

import (
	"care-cordination/lib/resp"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (m *Middleware) AuthMdw() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationHeader := ctx.GetHeader(authorizationHeaderKey)
		if len(authorizationHeader) == 0 {
			err := ErrInvalidRequest
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, resp.Error(err))
			return
		}

		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			err := ErrInvalidRequest
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, resp.Error(err))
			return
		}

		authorizationType := strings.ToLower(fields[0])
		if authorizationType != authorizationTypeBearer {
			err := ErrInvalidRequest
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, resp.Error(err))
			return
		}

		accessToken := fields[1]
		payload, err := m.tokenMaker.ValidateAccessToken(accessToken)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, resp.Error(err))
			return
		}

		ctx.Set(UserIDKey, payload.Subject)
		ctx.Next()
	}
}
