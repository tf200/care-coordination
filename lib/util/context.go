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
	// Fallback for regular context (e.g., in tests)
	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		return userID
	}
	return ""
}

func GetEmployeeID(ctx context.Context) string {
	if ginCtx, ok := ctx.(*gin.Context); ok {
		return ginCtx.GetString(EmployeeIDKey)
	}
	// Fallback for regular context (e.g., in tests)
	if employeeID, ok := ctx.Value(EmployeeIDKey).(string); ok {
		return employeeID
	}
	return ""
}
