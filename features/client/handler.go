package client

import (
	"care-cordination/features/middleware"
	"care-cordination/lib/resp"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ClientHandler struct {
	clientService ClientService
	mdw           *middleware.Middleware
}

func NewClientHandler(clientService ClientService, mdw *middleware.Middleware) *ClientHandler {
	return &ClientHandler{
		clientService: clientService,
		mdw:           mdw,
	}
}

func (h *ClientHandler) SetupClientRoutes(router *gin.Engine) {
	clients := router.Group("/clients")

	clients.POST("/move-to-waiting-list", h.mdw.AuthMdw(), h.MoveClientToWaitingList)
}

// @Summary Move client to waiting list
// @Description Move a client from intake form to waiting list by creating a client record
// @Tags Client
// @Accept json
// @Produce json
// @Param request body MoveClientToWaitingListRequest true "Intake Form ID"
// @Success 200 {object} MoveClientToWaitingListResponse
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 404 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /clients/move-to-waiting-list [post]
func (h *ClientHandler) MoveClientToWaitingList(ctx *gin.Context) {
	var req MoveClientToWaitingListRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}

	result, err := h.clientService.MoveClientToWaitingList(ctx, &req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidRequest):
			ctx.JSON(http.StatusBadRequest, resp.Error(err))
		case errors.Is(err, ErrIntakeFormNotFound):
			ctx.JSON(http.StatusNotFound, resp.Error(err))
		case errors.Is(err, ErrRegistrationFormNotFound):
			ctx.JSON(http.StatusNotFound, resp.Error(err))
		case errors.Is(err, ErrFailedToCreateClient):
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		case errors.Is(err, ErrInternal):
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(ErrInternal))
		}
		return
	}

	ctx.JSON(http.StatusOK, result)
}
