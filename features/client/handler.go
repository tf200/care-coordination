package client

import (
	"care-cordination/lib/middleware"
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
	clients.POST("/:id/start-discharge", h.mdw.AuthMdw(), h.StartDischarge)
	clients.POST("/:id/complete-discharge", h.mdw.AuthMdw(), h.CompleteDischarge)
	clients.GET("/waiting-list/stats", h.mdw.AuthMdw(), h.GetWaitlistStats)
	clients.GET("/waiting-list", h.mdw.AuthMdw(), h.mdw.PaginationMdw(), h.ListWaitingListClients)
	clients.GET("/in-care/stats", h.mdw.AuthMdw(), h.GetInCareStats)
	clients.GET("/in-care", h.mdw.AuthMdw(), h.mdw.PaginationMdw(), h.ListInCareClients)
	clients.GET("/discharged/stats", h.mdw.AuthMdw(), h.GetDischargeStats)
	clients.GET("/discharged", h.mdw.AuthMdw(), h.mdw.PaginationMdw(), h.ListDischargedClients)
	clients.GET("/:id/goals", h.mdw.AuthMdw(), h.ListClientGoals)
}

// @Summary Move client to waiting list
// @Description Move a client from intake form to waiting list by creating a client record
// @Tags Client
// @Accept json
// @Produce json
// @Param request body MoveClientToWaitingListRequest true "Intake Form ID"
// @Success 200 {object} resp.SuccessResponse[MoveClientToWaitingListResponse]
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

	ctx.JSON(http.StatusOK, resp.Success(result, "Client moved to waiting list successfully"))
}

// @Summary Move client to in care
// @Description Move a client from waiting list to in care status
// @Tags Client
// @Accept json
// @Produce json
// @Param id path string true "Client ID"
// @Param request body MoveClientInCareRequest true "Care start date and optional ambulatory hours"
// @Success 200 {object} resp.SuccessResponse[MoveClientInCareResponse]
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

	ctx.JSON(http.StatusOK, resp.Success(result, "Client moved to in care successfully"))
}

// @Summary Start client discharge
// @Description Start the discharge process for a client. Client remains in care with discharge_status = in_progress
// @Tags Client
// @Accept json
// @Produce json
// @Param id path string true "Client ID"
// @Param request body StartDischargeRequest true "Discharge date and reason"
// @Success 200 {object} resp.SuccessResponse[StartDischargeResponse]
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 404 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /clients/{id}/start-discharge [post]
func (h *ClientHandler) StartDischarge(ctx *gin.Context) {
	clientID := ctx.Param("id")
	if clientID == "" {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}

	var req StartDischargeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}

	result, err := h.clientService.StartDischarge(ctx, clientID, &req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidRequest):
			ctx.JSON(http.StatusBadRequest, resp.Error(err))
		case errors.Is(err, ErrClientNotFound):
			ctx.JSON(http.StatusNotFound, resp.Error(err))
		case errors.Is(err, ErrClientNotInCare):
			ctx.JSON(http.StatusBadRequest, resp.Error(err))
		case errors.Is(err, ErrDischargeAlreadyStarted):
			ctx.JSON(http.StatusBadRequest, resp.Error(err))
		case errors.Is(err, ErrInternal):
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(ErrInternal))
		}
		return
	}

	ctx.JSON(http.StatusOK, resp.Success(result, "Client discharge started successfully"))
}

// @Summary Complete client discharge
// @Description Complete the discharge process for a client. Requires closing and evaluation reports. Client status changes to discharged.
// @Tags Client
// @Accept json
// @Produce json
// @Param id path string true "Client ID"
// @Param request body CompleteDischargeRequest true "Reports and optional attachments"
// @Success 200 {object} resp.SuccessResponse[CompleteDischargeResponse]
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 404 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /clients/{id}/complete-discharge [post]
func (h *ClientHandler) CompleteDischarge(ctx *gin.Context) {
	clientID := ctx.Param("id")
	if clientID == "" {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}

	var req CompleteDischargeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}

	result, err := h.clientService.CompleteDischarge(ctx, clientID, &req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidRequest):
			ctx.JSON(http.StatusBadRequest, resp.Error(err))
		case errors.Is(err, ErrClientNotFound):
			ctx.JSON(http.StatusNotFound, resp.Error(err))
		case errors.Is(err, ErrClientNotInCare):
			ctx.JSON(http.StatusBadRequest, resp.Error(err))
		case errors.Is(err, ErrDischargeNotStarted):
			ctx.JSON(http.StatusBadRequest, resp.Error(err))
		case errors.Is(err, ErrInternal):
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(ErrInternal))
		}
		return
	}

	ctx.JSON(http.StatusOK, resp.Success(result, "Client discharged successfully"))
}

// @Summary List waiting list clients
// @Description List all clients on the waiting list with pagination and search
// @Tags Client
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 10, max: 100)"
// @Param search query string false "Search by client first name or last name"
// @Success 200 {object} resp.SuccessResponse[resp.PaginationResponse[[]ListWaitingListClientsResponse]]
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

	ctx.JSON(http.StatusOK, resp.Success(result, "Clients listed successfully"))
}

// @Summary List in-care clients
// @Description List all clients currently in care with pagination, search, and care type filter. Returns weeks in accommodation for living care types or used ambulatory hours for ambulatory care.
// @Tags Client
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 10, max: 100)"
// @Param search query string false "Search by client first name or last name"
// @Param careType query string false "Filter by care type (protected_living, semi_independent_living, independent_assisted_living, ambulatory_care)"
// @Success 200 {object} resp.SuccessResponse[resp.PaginationResponse[[]ListInCareClientsResponse]]
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

	ctx.JSON(http.StatusOK, resp.Success(result, "Clients listed successfully"))
}

// @Summary List discharged clients
// @Description List all clients with discharge status (both in_progress and completed) with pagination, search, and discharge status filter
// @Tags Client
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 10, max: 100)"
// @Param search query string false "Search by client first name or last name"
// @Param dischargeStatus query string false "Filter by discharge status (in_progress or completed)"
// @Success 200 {object} resp.SuccessResponse[resp.PaginationResponse[[]ListDischargedClientsResponse]]
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /clients/discharged [get]
func (h *ClientHandler) ListDischargedClients(ctx *gin.Context) {
	var req ListDischargedClientsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(err))
		return
	}

	result, err := h.clientService.ListDischargedClients(ctx, &req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInternal):
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(ErrInternal))
		}
		return
	}

	ctx.JSON(http.StatusOK, resp.Success(result, "Clients listed successfully"))
}

// @Summary Get waitlist statistics
// @Description Get comprehensive statistics for clients on the waiting list including total count, average wait time, and priority breakdowns
// @Tags Client
// @Produce json
// @Success 200 {object} resp.SuccessResponse[GetWaitlistStatsResponse]
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /clients/waiting-list/stats [get]
func (h *ClientHandler) GetWaitlistStats(ctx *gin.Context) {
	result, err := h.clientService.GetWaitlistStats(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, resp.Error(ErrInternal))
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(result, "Waitlist statistics retrieved successfully"))
}

// @Summary Get in-care statistics
// @Description Get comprehensive statistics for clients currently in care including total count, average days in care, and care type breakdowns
// @Tags Client
// @Produce json
// @Success 200 {object} resp.SuccessResponse[GetInCareStatsResponse]
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /clients/in-care/stats [get]
func (h *ClientHandler) GetInCareStats(ctx *gin.Context) {
	result, err := h.clientService.GetInCareStats(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, resp.Error(ErrInternal))
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(result, "In-care statistics retrieved successfully"))
}

// @Summary Get discharge statistics
// @Description Get comprehensive statistics for discharged clients including completed/premature breakdown, completion rate, and average days in care
// @Tags Client
// @Produce json
// @Success 200 {object} resp.SuccessResponse[GetDischargeStatsResponse]
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /clients/discharged/stats [get]
func (h *ClientHandler) GetDischargeStats(ctx *gin.Context) {
	result, err := h.clientService.GetDischargeStats(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, resp.Error(ErrInternal))
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(result, "Discharge statistics retrieved successfully"))
}

// @Summary List client goals
// @Description Get all goals for a specific client
// @Tags Client
// @Produce json
// @Param id path string true "Client ID"
// @Success 200 {object} resp.SuccessResponse[[]ListClientGoalsResponse]
// @Failure 401 {object} resp.ErrorResponse
// @Failure 404 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /clients/{id}/goals [get]
func (h *ClientHandler) ListClientGoals(ctx *gin.Context) {
	clientID := ctx.Param("id")
	if clientID == "" {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}

	result, err := h.clientService.ListClientGoals(ctx, clientID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, resp.Error(ErrInternal))
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(result, "Client goals retrieved successfully"))
}
