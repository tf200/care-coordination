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

func NewIncidentHandler(incidentService IncidentService, mdw *middleware.Middleware) *IncidentHandler {
	return &IncidentHandler{
		incidentService: incidentService,
		mdw:             mdw,
	}
}

func (h *IncidentHandler) SetupIncidentRoutes(router *gin.Engine) {
	incident := router.Group("/incidents")

	incident.POST("", h.mdw.AuthMdw(), h.CreateIncident)
	incident.GET("", h.mdw.AuthMdw(), h.mdw.PaginationMdw(), h.ListIncidents)
}

// @Summary Create an incident
// @Description Create a new incident
// @Tags Incident
// @Accept json
// @Produce json
// @Param incident body CreateIncidentRequest true "Incident"
// @Success 200 {object} CreateIncidentResponse
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
	ctx.JSON(http.StatusOK, result)
}

// @Summary List incidents
// @Description List all incidents with pagination and search by client name
// @Tags Incident
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 10, max: 100)"
// @Param search query string false "Search by client first name, last name, or full name"
// @Success 200 {object} resp.PaginationResponse[[]ListIncidentsResponse]
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
	ctx.JSON(http.StatusOK, result)
}
