package resp

import "github.com/gin-gonic/gin"

func MessageResonse(message string) gin.H {
	return gin.H{"message": message}
}

func ErrorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
