package locTransfer

import (
	"care-cordination/features/middleware"
	"care-cordination/lib/resp"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type LocTransferHandler struct {
	locTransferService LocationTransferService
	mdw                *middleware.Middleware
}

func NewLocTransferHandler(locTransferService LocationTransferService, mdw *middleware.Middleware) *LocTransferHandler {
	return &LocTransferHandler{
		locTransferService: locTransferService,
		mdw:                mdw,
	}
}

func (h *LocTransferHandler) SetupLocTransferRoutes(router *gin.Engine) {
	locTransfers := router.Group("/location-transfers")

	locTransfers.POST("", h.mdw.AuthMdw(), h.RegisterLocationTransfer)
	locTransfers.GET("", h.mdw.AuthMdw(), h.mdw.PaginationMdw(), h.ListLocationTransfers)
}

// @Summary Register a location transfer
// @Description Register a new location transfer for a client
// @Tags LocationTransfer
// @Accept json
// @Produce json
// @Param request body RegisterLocationTransferRequest true "Location Transfer Request"
// @Success 200 {object} RegisterLocationTransferResponse
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 404 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /location-transfers [post]
func (h *LocTransferHandler) RegisterLocationTransfer(ctx *gin.Context) {
	var req RegisterLocationTransferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}

	result, err := h.locTransferService.RegisterLocationTransfer(ctx, &req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidRequest):
			ctx.JSON(http.StatusBadRequest, resp.Error(err))
		case errors.Is(err, ErrClientNotFound):
			ctx.JSON(http.StatusNotFound, resp.Error(err))
		case errors.Is(err, ErrInternal):
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(ErrInternal))
		}
		return
	}

	ctx.JSON(http.StatusOK, result)
}

// @Summary List location transfers
// @Description List all location transfers with pagination and search
// @Tags LocationTransfer
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 10, max: 100)"
// @Param search query string false "Search by client name"
// @Success 200 {object} resp.PaginationResponse[[]ListLocationTransfersResponse]
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /location-transfers [get]
func (h *LocTransferHandler) ListLocationTransfers(ctx *gin.Context) {
	var req ListLocationTransfersRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}

	result, err := h.locTransferService.ListLocationTransfers(ctx, &req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInternal):
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(ErrInternal))
		}
		return
	}

	ctx.JSON(http.StatusOK, result)
}
