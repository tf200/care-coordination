package util

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGetUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		setupCtx func() context.Context
		want     string
	}{
		{
			name: "Gin context with user_id",
			setupCtx: func() context.Context {
				c, _ := gin.CreateTestContext(httptest.NewRecorder())
				c.Set("user_id", "12345")
				return c
			},
			want: "12345",
		},
		{
			name: "Gin context without user_id",
			setupCtx: func() context.Context {
				c, _ := gin.CreateTestContext(httptest.NewRecorder())
				return c
			},
			want: "",
		},
		{
			name: "Standard context",
			setupCtx: func() context.Context {
				return context.Background()
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetUserID(tt.setupCtx())
			if got != tt.want {
				t.Errorf("GetUserID() = %v, want %v", got, tt.want)
			}
		})
	}
}
