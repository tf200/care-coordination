package middleware

import (
	db "care-cordination/lib/db/sqlc"
	"care-cordination/lib/resp"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (m *Middleware) RequirePermission(resource, action string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID := ctx.GetString(UserIDKey)
		if userID == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, resp.Error(ErrUnauthorized))
			return
		}

		hasPermission, err := m.store.HasPermission(ctx, db.HasPermissionParams{
			UserID:   userID,
			Resource: resource,
			Action:   action,
		})

		if err != nil {
			m.logger.Error(
				ctx,
				"Middleware.RequirePermission",
				"failed to check permission",
				zap.Error(err),
			)
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, resp.Error(ErrInternal))
			return
		}

		if !hasPermission {
			ctx.AbortWithStatusJSON(http.StatusForbidden, resp.Error(ErrForbidden))
			return
		}

		ctx.Next()
	}
}
