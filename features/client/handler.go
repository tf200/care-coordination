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
	clients.POST("/:id/move-to-care", h.mdw.AuthMdw(), h.MoveClientInCare)
	clients.GET("/waiting-list", h.mdw.AuthMdw(), h.mdw.PaginationMdw(), h.ListWaitingListClients)
	clients.GET("/in-care", h.mdw.AuthMdw(), h.mdw.PaginationMdw(), h.ListInCareClients)
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

// @Summary Move client to in care
// @Description Move a client from waiting list to in care status
// @Tags Client
// @Accept json
// @Produce json
// @Param id path string true "Client ID"
// @Param request body MoveClientInCareRequest true "Care start date and optional ambulatory hours"
// @Success 200 {object} MoveClientInCareResponse
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 404 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /clients/{id}/move-to-care [post]
func (h *ClientHandler) MoveClientInCare(ctx *gin.Context) {
	clientID := ctx.Param("id")
	if clientID == "" {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}

	var req MoveClientInCareRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}

	result, err := h.clientService.MoveClientInCare(ctx, clientID, &req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidRequest):
			ctx.JSON(http.StatusBadRequest, resp.Error(err))
		case errors.Is(err, ErrClientNotFound):
			ctx.JSON(http.StatusNotFound, resp.Error(err))
		case errors.Is(err, ErrInvalidClientStatus):
			ctx.JSON(http.StatusBadRequest, resp.Error(err))
		case errors.Is(err, ErrAmbulatoryHoursRequired):
			ctx.JSON(http.StatusBadRequest, resp.Error(err))
		case errors.Is(err, ErrAmbulatoryHoursNotAllowed):
			ctx.JSON(http.StatusBadRequest, resp.Error(err))
		case errors.Is(err, ErrInternal):
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(ErrInternal))
		}
		return
	}

	ctx.JSON(http.StatusOK, result)
}

// @Summary List waiting list clients
// @Description List all clients on the waiting list with pagination and search
// @Tags Client
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 10, max: 100)"
// @Param search query string false "Search by client first name or last name"
// @Success 200 {object} resp.PaginationResponse[[]ListWaitingListClientsResponse]
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /clients/waiting-list [get]
func (h *ClientHandler) ListWaitingListClients(ctx *gin.Context) {
	var req ListWaitingListClientsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(err))
		return
	}

	result, err := h.clientService.ListWaitingListClients(ctx, &req)
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

// @Summary List in-care clients
// @Description List all clients currently in care with pagination and search. Returns weeks in accommodation for living care types or used ambulatory hours for ambulatory care.
// @Tags Client
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 10, max: 100)"
// @Param search query string false "Search by client first name or last name"
// @Success 200 {object} resp.PaginationResponse[[]ListInCareClientsResponse]
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /clients/in-care [get]
func (h *ClientHandler) ListInCareClients(ctx *gin.Context) {
	var req ListInCareClientsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(err))
		return
	}

	result, err := h.clientService.ListInCareClients(ctx, &req)
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
