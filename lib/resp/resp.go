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

type PaginationResponse[T any] struct {
	Data       []T `json:"data"`
	TotalCount int `json:"totalCount"`
	TotalPages int `json:"totalPages"`
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
}

func PagResp[T any](data []T, totalCount int, page int, pageSize int) PaginationResponse[T] {
	totalPages := (totalCount + pageSize - 1) / pageSize
	if totalPages < 0 {
		totalPages = 0
	}
	return PaginationResponse[T]{
		Data:       data,
		TotalCount: totalCount,
		TotalPages: totalPages,
		Page:       page,
		PageSize:   pageSize,
	}
}

// PagRespWithParams builds a pagination response using the provided pagination parameters
// This is a convenience function to ensure correct page/pageSize values are used
func PagRespWithParams[T any](data []T, totalCount int, page int32, pageSize int32) PaginationResponse[T] {
	return PagResp(data, totalCount, int(page), int(pageSize))
}
