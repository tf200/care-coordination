package resp

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"   example:"error message"`
	Success bool   `json:"success" example:"true"`
}

// MessageResponse represents a success message response
type MessageResponse struct {
	Message string `json:"message" example:"success message"`
	Success bool   `json:"success" example:"true"`
}

type SuccessResponse[T any] struct {
	Data    T      `json:"data"`
	Message string `json:"message" example:"success message"`
	Success bool   `json:"success" example:"true"`
}

func MessageResonse(message string) MessageResponse {
	return MessageResponse{Message: message, Success: true}
}

func Error(err error) ErrorResponse {
	return ErrorResponse{Error: err.Error(), Success: false}
}

func Success[T any](data T, message string) SuccessResponse[T] {
	return SuccessResponse[T]{Data: data, Message: message, Success: true}
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
func PagRespWithParams[T any](
	data []T,
	totalCount int,
	page int32,
	pageSize int32,
) PaginationResponse[T] {
	return PagResp(data, totalCount, int(page), int(pageSize))
}
