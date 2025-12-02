package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
)

const (
	DefaultPage     = 1
	DefaultPageSize = 10
	MaxPageSize     = 100
	LimitKey        = "limit"
	OffsetKey       = "offset"
	PageKey         = "page"
	PageSizeKey     = "page_size"
)

type PaginationRequest struct {
	Page     int32 `form:"page"`
	PageSize int32 `form:"page_size"`
}

func (m *Middleware) PaginationMdw() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var pagination PaginationRequest
		if err := ctx.ShouldBindQuery(&pagination); err != nil {
			// use default values
			pagination.Page = DefaultPage
			pagination.PageSize = DefaultPageSize
		}

		// Validate page number (must be positive)
		if pagination.Page < 1 {
			pagination.Page = DefaultPage
		}

		// Validate page size (must be positive and not exceed max)
		if pagination.PageSize < 1 {
			pagination.PageSize = DefaultPageSize
		} else if pagination.PageSize > MaxPageSize {
			pagination.PageSize = MaxPageSize
		}

		limit, offset := m.getPagParams(pagination.Page, pagination.PageSize)

		// Store both SQL params (limit/offset) and display params (page/pageSize)
		ctx.Set(LimitKey, limit)
		ctx.Set(OffsetKey, offset)
		ctx.Set(PageKey, pagination.Page)
		ctx.Set(PageSizeKey, pagination.PageSize)
		ctx.Next()
	}
}

func (m *Middleware) getPagParams(page, pageSize int32) (int32, int32) {
	limit := pageSize
	offset := (page - 1) * pageSize
	return limit, offset
}

// GetPaginationParams safely retrieves pagination parameters from context
// Returns (limit, offset, page, pageSize) with safe defaults if values are missing
func GetPaginationParams(ctx context.Context) (limit, offset, page, pageSize int32) {
	// Safe type assertions with defaults
	if l, ok := ctx.Value(LimitKey).(int32); ok {
		limit = l
	} else {
		limit = DefaultPageSize
	}

	if o, ok := ctx.Value(OffsetKey).(int32); ok {
		offset = o
	} else {
		offset = 0
	}

	if p, ok := ctx.Value(PageKey).(int32); ok {
		page = p
	} else {
		page = DefaultPage
	}

	if ps, ok := ctx.Value(PageSizeKey).(int32); ok {
		pageSize = ps
	} else {
		pageSize = DefaultPageSize
	}

	return
}
