package util

import (
	"context"

	"github.com/gin-gonic/gin"
)

const (
	UserIDKey     = "user_id"
	EmployeeIDKey = "employee_id"
)

func GetUserID(ctx context.Context) string {
	if ginCtx, ok := ctx.(*gin.Context); ok {
		return ginCtx.GetString(UserIDKey)
	}
	return ""
}

func GetEmployeeID(ctx context.Context) string {
	if ginCtx, ok := ctx.(*gin.Context); ok {
		return ginCtx.GetString(EmployeeIDKey)
	}
	return ""
}
