package middleware

import (
	"care-cordination/lib/nanoid"

	"github.com/gin-gonic/gin"
)

const (
	RequestIDKey    = "X-Request-Id"
	RequestIDHeader = "X-Request-Id"
)

// RequestIDMiddleware generates a unique request ID for each request.
// If the client sends an X-Request-Id header, it will be used instead.
// The request ID is set in the context and also returned in the response headers.
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(RequestIDHeader)
		if requestID == "" {
			requestID = nanoid.Generate()
		}

		c.Set(RequestIDKey, requestID)
		c.Header(RequestIDHeader, requestID)
		c.Next()
	}
}
