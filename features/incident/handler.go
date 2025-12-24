package incident

import (
	"care-cordination/features/middleware"
	"care-cordination/lib/resp"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type IncidentHandler struct {
	incidentService IncidentService
	mdw             *middleware.Middleware
}

func NewIncidentHandler(
	incidentService IncidentService,
	mdw *middleware.Middleware,
) *IncidentHandler {
	return &IncidentHandler{
		incidentService: incidentService,
		mdw:             mdw,
	}
}

func (h *IncidentHandler) SetupIncidentRoutes(router *gin.Engine) {
	incident := router.Group("/incidents")

	incident.POST("", h.mdw.AuthMdw(), h.CreateIncident)
	incident.GET("/stats", h.mdw.AuthMdw(), h.GetIncidentStats)
	incident.GET("", h.mdw.AuthMdw(), h.mdw.PaginationMdw(), h.ListIncidents)
	incident.GET("/:id", h.mdw.AuthMdw(), h.GetIncident)
	incident.PATCH("/:id", h.mdw.AuthMdw(), h.UpdateIncident)
	incident.DELETE("/:id", h.mdw.AuthMdw(), h.DeleteIncident)
}

// @Summary Create an incident
// @Description Create a new incident
// @Tags Incident
// @Accept json
// @Produce json
// @Param incident body CreateIncidentRequest true "Incident"
// @Success 200 {object} resp.SuccessResponse[CreateIncidentResponse]
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /incidents [post]
func (h *IncidentHandler) CreateIncident(ctx *gin.Context) {
	var req CreateIncidentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}
	result, err := h.incidentService.CreateIncident(ctx, &req)
	if err != nil {
		switch err {
		case ErrInvalidRequest:
			ctx.JSON(http.StatusBadRequest, resp.Error(err))
		case ErrInternal:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		}
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(result, "Incident created successfully"))
}

// @Summary Get an incident
// @Description Get a single incident by ID
// @Tags Incident
// @Produce json
// @Param id path string true "Incident ID"
// @Success 200 {object} resp.SuccessResponse[GetIncidentResponse]
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 404 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /incidents/{id} [get]
func (h *IncidentHandler) GetIncident(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}
	result, err := h.incidentService.GetIncident(ctx, id)
	if err != nil {
		switch err {
		case ErrNotFound:
			ctx.JSON(http.StatusNotFound, resp.Error(err))
		case ErrInternal:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		}
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(result, "Incident retrieved successfully"))
}

// @Summary Update an incident
// @Description Update an existing incident by ID
// @Tags Incident
// @Accept json
// @Produce json
// @Param id path string true "Incident ID"
// @Param incident body UpdateIncidentRequest true "Incident update data"
// @Success 200 {object} resp.SuccessResponse[UpdateIncidentResponse]
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /incidents/{id} [patch]
func (h *IncidentHandler) UpdateIncident(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}
	var req UpdateIncidentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}
	result, err := h.incidentService.UpdateIncident(ctx, id, &req)
	if err != nil {
		switch err {
		case ErrInvalidRequest:
			ctx.JSON(http.StatusBadRequest, resp.Error(err))
		case ErrInternal:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		}
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(result, "Incident updated successfully"))
}

// @Summary Delete an incident
// @Description Soft delete an incident by ID
// @Tags Incident
// @Produce json
// @Param id path string true "Incident ID"
// @Success 200 {object} resp.SuccessResponse[DeleteIncidentResponse]
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /incidents/{id} [delete]
func (h *IncidentHandler) DeleteIncident(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}
	result, err := h.incidentService.DeleteIncident(ctx, id)
	if err != nil {
		switch err {
		case ErrInternal:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		}
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(result, "Incident deleted successfully"))
}

// @Summary List incidents
// @Description List all incidents with pagination and search by client name
// @Tags Incident
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 10, max: 100)"
// @Param search query string false "Search by client first name, last name, or full name"
// @Success 200 {object} resp.SuccessResponse[resp.PaginationResponse[[]ListIncidentsResponse]]
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /incidents [get]
func (h *IncidentHandler) ListIncidents(ctx *gin.Context) {
	var req ListIncidentsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}
	result, err := h.incidentService.ListIncidents(ctx, &req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInternal):
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		}
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(result, "Incidents listed successfully"))
}

// @Summary Get incident statistics
// @Description Get comprehensive statistics for incidents including counts by severity, status, and type
// @Tags Incident
// @Produce json
// @Success 200 {object} resp.SuccessResponse[GetIncidentStatsResponse]
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /incidents/stats [get]
func (h *IncidentHandler) GetIncidentStats(ctx *gin.Context) {
	result, err := h.incidentService.GetIncidentStats(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, resp.Error(ErrInternal))
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(result, "Incident statistics retrieved successfully"))
}
