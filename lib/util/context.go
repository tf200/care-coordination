package util

import (
	"context"

	"github.com/gin-gonic/gin"
)

const (
	UserIDKey     = "user_id"
	EmployeeIDKey = "employee_id"
	ClientIDKey   = "audit_client_id" // NEN7510: Track which client's data was accessed
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

func GetRequestID(ctx context.Context) string {
	if ginCtx, ok := ctx.(*gin.Context); ok {
		if v, exists := ginCtx.Get("X-Request-Id"); exists {
			if s, ok := v.(string); ok {
				return s
			}
		}
	}
	if v, ok := ctx.Value("X-Request-Id").(string); ok {
		return v
	}
	return ""
}

func GetIPAddress(ctx context.Context) string {
	if ginCtx, ok := ctx.(*gin.Context); ok {
		return ginCtx.ClientIP()
	}
	return ""
}

func GetUserAgent(ctx context.Context) string {
	if ginCtx, ok := ctx.(*gin.Context); ok {
		return ginCtx.Request.UserAgent()
	}
	return ""
}

// SetClientID sets the client ID in context for audit logging
// Call this from handlers after fetching client-related resources
func SetClientID(ctx *gin.Context, clientID string) {
	ctx.Set(ClientIDKey, clientID)
}

// GetClientID retrieves the client ID from context for audit logging
func GetClientID(ctx context.Context) string {
	if ginCtx, ok := ctx.(*gin.Context); ok {
		return ginCtx.GetString(ClientIDKey)
	}
	if clientID, ok := ctx.Value(ClientIDKey).(string); ok {
		return clientID
	}
	return ""
}
