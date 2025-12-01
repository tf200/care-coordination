package util

import (
	"context"

	"github.com/gin-gonic/gin"
)

const UserIDKey = "user_id"

func GetUserID(ctx context.Context) string {
	if ginCtx, ok := ctx.(*gin.Context); ok {
		return ginCtx.GetString(UserIDKey)
	}
	return ""
}
