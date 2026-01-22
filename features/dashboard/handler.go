package dashboard

import (
	"care-cordination/features/middleware"
	"care-cordination/lib/resp"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	dashboardService DashboardService
	mdw              *middleware.Middleware
}

func NewDashboardHandler(
	dashboardService DashboardService,
	mdw *middleware.Middleware,
) *DashboardHandler {
	return &DashboardHandler{
		dashboardService: dashboardService,
		mdw:              mdw,
	}
}

func (h *DashboardHandler) SetupDashboardRoutes(router *gin.Engine) {
	dashboard := router.Group("/dashboard")

	// Admin Dashboard
	dashboard.GET("/overview-stats", h.mdw.AuthMdw(), h.GetOverviewStats)
	dashboard.GET("/critical-alerts", h.mdw.AuthMdw(), h.GetCriticalAlerts)
	dashboard.GET("/pipeline-stats", h.mdw.AuthMdw(), h.GetPipelineStats)
	dashboard.GET("/care-type-distribution", h.mdw.AuthMdw(), h.GetCareTypeDistribution)
	dashboard.GET("/location-capacity", h.mdw.AuthMdw(), h.GetLocationCapacity)
	dashboard.GET("/today-appointments", h.mdw.AuthMdw(), h.GetTodayAppointments)
	dashboard.GET("/evaluation-stats", h.mdw.AuthMdw(), h.GetEvaluationStats)
	dashboard.GET("/discharge-stats", h.mdw.AuthMdw(), h.GetDischargeStats)

	// Coordinator Dashboard
	coordinator := dashboard.Group("/coordinator")
	coordinator.GET("/urgent-alerts", h.mdw.AuthMdw(), h.GetCoordinatorUrgentAlerts)
	coordinator.GET("/today-schedule", h.mdw.AuthMdw(), h.GetCoordinatorTodaySchedule)
}

// @Summary Get dashboard overview stats
// @Description Get overview statistics for the admin dashboard
// @Tags Dashboard
// @Produce json
// @Success 200 {object} resp.SuccessResponse[OverviewDTO]
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /dashboard/overview-stats [get]
func (h *DashboardHandler) GetOverviewStats(ctx *gin.Context) {
	stats, err := h.dashboardService.GetOverviewStats(ctx)
	if err != nil {
		switch err {
		case ErrInternal:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		}
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(stats, "Dashboard overview stats retrieved successfully"))
}

// @Summary Get critical alerts
// @Description Get critical alerts for the admin dashboard
// @Tags Dashboard
// @Produce json
// @Success 200 {object} resp.SuccessResponse[CriticalAlertsDTO]
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /dashboard/critical-alerts [get]
func (h *DashboardHandler) GetCriticalAlerts(ctx *gin.Context) {
	alerts, err := h.dashboardService.GetCriticalAlerts(ctx)
	if err != nil {
		switch err {
		case ErrInternal:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		}
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(alerts, "Critical alerts retrieved successfully"))
}

// @Summary Get pipeline stats
// @Description Get pipeline statistics showing client journey through the care system
// @Tags Dashboard
// @Produce json
// @Success 200 {object} resp.SuccessResponse[PipelineStatsDTO]
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /dashboard/pipeline-stats [get]
func (h *DashboardHandler) GetPipelineStats(ctx *gin.Context) {
	stats, err := h.dashboardService.GetPipelineStats(ctx)
	if err != nil {
		switch err {
		case ErrInternal:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		}
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(stats, "Pipeline stats retrieved successfully"))
}

// @Summary Get care type distribution
// @Description Get distribution of in-care clients by care type
// @Tags Dashboard
// @Produce json
// @Success 200 {object} resp.SuccessResponse[CareTypeDistributionDTO]
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /dashboard/care-type-distribution [get]
func (h *DashboardHandler) GetCareTypeDistribution(ctx *gin.Context) {
	distribution, err := h.dashboardService.GetCareTypeDistribution(ctx)
	if err != nil {
		switch err {
		case ErrInternal:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		}
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(distribution, "Care type distribution retrieved successfully"))
}

// @Summary Get location capacity
// @Description Get location capacity statistics with optional limit and sorting
// @Tags Dashboard
// @Produce json
// @Param limit query int false "Number of locations to return (default: 4, max: 100)"
// @Param sort query string false "Sort order: occupancy_desc, occupancy_asc, name (default: occupancy_desc)"
// @Success 200 {object} resp.SuccessResponse[LocationCapacityDTO]
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /dashboard/location-capacity [get]
func (h *DashboardHandler) GetLocationCapacity(ctx *gin.Context) {
	var req LocationCapacityRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(err))
		return
	}

	capacity, err := h.dashboardService.GetLocationCapacity(ctx, &req)
	if err != nil {
		switch err {
		case ErrInternal:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		}
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(capacity, "Location capacity retrieved successfully"))
}

// @Summary Get today's appointments
// @Description Get today's appointments for the logged-in user
// @Tags Dashboard
// @Produce json
// @Success 200 {object} resp.SuccessResponse[TodayAppointmentsDTO]
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /dashboard/today-appointments [get]
func (h *DashboardHandler) GetTodayAppointments(ctx *gin.Context) {
	employeeID, exists := ctx.Get(middleware.EmployeeIDKey)
	if !exists {
		ctx.JSON(http.StatusUnauthorized, resp.Error(ErrInternal))
		return
	}

	appointments, err := h.dashboardService.GetTodayAppointments(ctx, employeeID.(string))
	if err != nil {
		switch err {
		case ErrInternal:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		}
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(appointments, "Today's appointments retrieved successfully"))
}

// @Summary Get evaluation stats
// @Description Get evaluation statistics for all coordinators
// @Tags Dashboard
// @Produce json
// @Success 200 {object} resp.SuccessResponse[EvaluationStatsDTO]
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /dashboard/evaluation-stats [get]
func (h *DashboardHandler) GetEvaluationStats(ctx *gin.Context) {
	stats, err := h.dashboardService.GetEvaluationStats(ctx)
	if err != nil {
		switch err {
		case ErrInternal:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		}
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(stats, "Evaluation stats retrieved successfully"))
}

// @Summary Get discharge stats
// @Description Get discharge statistics
// @Tags Dashboard
// @Produce json
// @Success 200 {object} resp.SuccessResponse[DischargeStatsDTO]
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /dashboard/discharge-stats [get]
func (h *DashboardHandler) GetDischargeStats(ctx *gin.Context) {
	stats, err := h.dashboardService.GetDischargeStats(ctx)
	if err != nil {
		switch err {
		case ErrInternal:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		}
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(stats, "Discharge stats retrieved successfully"))
}

// Coordinator Dashboard Handlers

// @Summary Get coordinator urgent alerts
// @Description Get urgent alerts for the logged-in coordinator's clients
// @Tags Dashboard - Coordinator
// @Produce json
// @Success 200 {object} resp.SuccessResponse[CoordinatorUrgentAlertsDTO]
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /dashboard/coordinator/urgent-alerts [get]
func (h *DashboardHandler) GetCoordinatorUrgentAlerts(ctx *gin.Context) {
	employeeID, exists := ctx.Get(middleware.EmployeeIDKey)
	if !exists {
		ctx.JSON(http.StatusUnauthorized, resp.Error(ErrInternal))
		return
	}

	alerts, err := h.dashboardService.GetCoordinatorUrgentAlerts(ctx, employeeID.(string))
	if err != nil {
		switch err {
		case ErrInternal:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		}
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(alerts, "Coordinator urgent alerts retrieved successfully"))
}

// @Summary Get coordinator today schedule
// @Description Get today's schedule for the logged-in coordinator
// @Tags Dashboard - Coordinator
// @Produce json
// @Success 200 {object} resp.SuccessResponse[CoordinatorTodayScheduleDTO]
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /dashboard/coordinator/today-schedule [get]
func (h *DashboardHandler) GetCoordinatorTodaySchedule(ctx *gin.Context) {
	employeeID, exists := ctx.Get(middleware.EmployeeIDKey)
	if !exists {
		ctx.JSON(http.StatusUnauthorized, resp.Error(ErrInternal))
		return
	}

	schedule, err := h.dashboardService.GetCoordinatorTodaySchedule(ctx, employeeID.(string))
	if err != nil {
		switch err {
		case ErrInternal:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		}
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(schedule, "Coordinator today schedule retrieved successfully"))
}
