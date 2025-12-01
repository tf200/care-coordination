package resp

import "github.com/gin-gonic/gin"

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error" example:"error message"`
}

// MessageResponse represents a success message response
type MessageResponse struct {
	Message string `json:"message" example:"success message"`
}

func MessageResonse(message string) gin.H {
	return gin.H{"message": message}
}

func Error(err error) gin.H {
	return gin.H{"error": err.Error()}
}
