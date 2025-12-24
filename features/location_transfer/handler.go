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

func NewLocTransferHandler(
	locTransferService LocationTransferService,
	mdw *middleware.Middleware,
) *LocTransferHandler {
	return &LocTransferHandler{
		locTransferService: locTransferService,
		mdw:                mdw,
	}
}

func (h *LocTransferHandler) SetupLocTransferRoutes(router *gin.Engine) {
	locTransfers := router.Group("/location-transfers")

	locTransfers.POST("", h.mdw.AuthMdw(), h.RegisterLocationTransfer)
	locTransfers.GET("/stats", h.mdw.AuthMdw(), h.GetLocationTransferStats)
	locTransfers.GET("", h.mdw.AuthMdw(), h.mdw.PaginationMdw(), h.ListLocationTransfers)
	locTransfers.GET("/:id", h.mdw.AuthMdw(), h.GetLocationTransferByID)
	locTransfers.POST("/:id/confirm", h.mdw.AuthMdw(), h.ConfirmLocationTransfer)
	locTransfers.POST("/:id/refuse", h.mdw.AuthMdw(), h.RefuseLocationTransfer)
	locTransfers.PUT("/:id", h.mdw.AuthMdw(), h.UpdateLocationTransfer)
}

// @Summary Register a location transfer
// @Description Register a new location transfer for a client
// @Tags LocationTransfer
// @Accept json
// @Produce json
// @Param request body RegisterLocationTransferRequest true "Location Transfer Request"
// @Success 200 {object} resp.SuccessResponse[RegisterLocationTransferResponse]
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

	ctx.JSON(http.StatusOK, resp.Success(result, "Location transfer registered successfully"))
}

// @Summary List location transfers
// @Description List all location transfers with pagination and search
// @Tags LocationTransfer
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 10, max: 100)"
// @Param search query string false "Search by client name"
// @Success 200 {object} resp.SuccessResponse[resp.PaginationResponse[ListLocationTransfersResponse]]
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

	ctx.JSON(http.StatusOK, resp.Success(result, "Location transfers listed successfully"))
}

// @Summary Get a location transfer by ID
// @Description Get a single location transfer with all details
// @Tags LocationTransfer
// @Produce json
// @Param id path string true "Transfer ID"
// @Success 200 {object} resp.SuccessResponse[ListLocationTransfersResponse]
// @Failure 404 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /location-transfers/{id} [get]
func (h *LocTransferHandler) GetLocationTransferByID(ctx *gin.Context) {
	transferID := ctx.Param("id")

	result, err := h.locTransferService.GetLocationTransferByID(ctx, transferID)
	if err != nil {
		switch {
		case errors.Is(err, ErrTransferNotFound):
			ctx.JSON(http.StatusNotFound, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(ErrInternal))
		}
		return
	}

	ctx.JSON(http.StatusOK, resp.Success(result, "Location transfer retrieved successfully"))
}

// @Summary Confirm a location transfer
// @Description Confirm a pending location transfer, updating the client's location and coordinator
// @Tags LocationTransfer
// @Produce json
// @Param id path string true "Transfer ID"
// @Success 200 {object} resp.SuccessResponse[any]
// @Failure 400 {object} resp.ErrorResponse
// @Failure 404 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /location-transfers/{id}/confirm [post]
func (h *LocTransferHandler) ConfirmLocationTransfer(ctx *gin.Context) {
	transferID := ctx.Param("id")

	err := h.locTransferService.ConfirmLocationTransfer(ctx, transferID)
	if err != nil {
		switch {
		case errors.Is(err, ErrTransferNotFound):
			ctx.JSON(http.StatusNotFound, resp.Error(err))
		case errors.Is(err, ErrTransferAlreadyProcessed):
			ctx.JSON(http.StatusBadRequest, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(ErrInternal))
		}
		return
	}

	ctx.JSON(http.StatusOK, resp.MessageResonse("Location transfer confirmed successfully"))
}

// @Summary Refuse a location transfer
// @Description Refuse a pending location transfer with a reason
// @Tags LocationTransfer
// @Accept json
// @Produce json
// @Param id path string true "Transfer ID"
// @Param request body RefuseLocationTransferRequest true "Refusal reason"
// @Success 200 {object} resp.SuccessResponse[any]
// @Failure 400 {object} resp.ErrorResponse
// @Failure 404 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /location-transfers/{id}/refuse [post]
func (h *LocTransferHandler) RefuseLocationTransfer(ctx *gin.Context) {
	transferID := ctx.Param("id")

	var req RefuseLocationTransferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}

	err := h.locTransferService.RefuseLocationTransfer(ctx, transferID, &req)
	if err != nil {
		switch {
		case errors.Is(err, ErrTransferNotFound):
			ctx.JSON(http.StatusNotFound, resp.Error(err))
		case errors.Is(err, ErrTransferAlreadyProcessed):
			ctx.JSON(http.StatusBadRequest, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(ErrInternal))
		}
		return
	}

	ctx.JSON(http.StatusOK, resp.MessageResonse("Location transfer refused successfully"))
}

// @Summary Update a location transfer
// @Description Update a pending location transfer (new location, coordinator, or reason)
// @Tags LocationTransfer
// @Accept json
// @Produce json
// @Param id path string true "Transfer ID"
// @Param request body UpdateLocationTransferRequest true "Update fields"
// @Success 200 {object} resp.SuccessResponse[any]
// @Failure 400 {object} resp.ErrorResponse
// @Failure 404 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /location-transfers/{id} [put]
func (h *LocTransferHandler) UpdateLocationTransfer(ctx *gin.Context) {
	transferID := ctx.Param("id")

	var req UpdateLocationTransferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}

	err := h.locTransferService.UpdateLocationTransfer(ctx, transferID, &req)
	if err != nil {
		switch {
		case errors.Is(err, ErrTransferNotFound):
			ctx.JSON(http.StatusNotFound, resp.Error(err))
		case errors.Is(err, ErrTransferAlreadyProcessed):
			ctx.JSON(http.StatusBadRequest, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(ErrInternal))
		}
		return
	}

	ctx.JSON(http.StatusOK, resp.MessageResonse("Location transfer updated successfully"))
}

// @Summary Get location transfer statistics
// @Description Get comprehensive statistics for location transfers including total count, pending count, approval rate, and status breakdowns
// @Tags LocationTransfer
// @Produce json
// @Success 200 {object} resp.SuccessResponse[GetLocationTransferStatsResponse]
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /location-transfers/stats [get]
func (h *LocTransferHandler) GetLocationTransferStats(ctx *gin.Context) {
	result, err := h.locTransferService.GetLocationTransferStats(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, resp.Error(ErrInternal))
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(result, "Location transfer statistics retrieved successfully"))
}
