package dashboard

import (
	"care-cordination/lib/middleware"
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
	dashboard.Use(h.mdw.AuthMdw())

	// Admin Dashboard
	admin := dashboard.Group("")
	admin.Use(h.mdw.RequirePermission("dashboard", "read"))
	admin.GET("/overview-stats", h.GetOverviewStats)
	admin.GET("/critical-alerts", h.GetCriticalAlerts)
	admin.GET("/pipeline-stats", h.GetPipelineStats)
	admin.GET("/care-type-distribution", h.GetCareTypeDistribution)
	admin.GET("/location-capacity", h.GetLocationCapacity)
	admin.GET("/today-appointments", h.GetTodayAppointments)
	admin.GET("/evaluation-stats", h.GetEvaluationStats)
	admin.GET("/discharge-stats", h.GetDischargeStats)

	// Coordinator Dashboard
	coordinator := dashboard.Group("/coordinator")
	coordinator.GET("/urgent-alerts", h.GetCoordinatorUrgentAlerts)
	coordinator.GET("/today-schedule", h.GetCoordinatorTodaySchedule)
	coordinator.GET("/stats", h.GetCoordinatorStats)
	coordinator.GET("/reminders", h.GetCoordinatorReminders)
	coordinator.GET("/clients", h.GetCoordinatorClients)
	coordinator.GET("/goals-progress", h.GetCoordinatorGoalsProgress)
	coordinator.GET("/incidents", h.GetCoordinatorIncidents)
}

// @Summary Get dashboard overview stats
// @Description Get overview statistics for the admin dashboard
// @Tags Dashboard
// @Produce json
// @Success 200 {object} resp.SuccessResponse[OverviewResponse]
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
// @Success 200 {object} resp.SuccessResponse[CriticalAlertsResponse]
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
// @Success 200 {object} resp.SuccessResponse[PipelineStatsResponse]
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
// @Success 200 {object} resp.SuccessResponse[CareTypeDistributionResponse]
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
// @Success 200 {object} resp.SuccessResponse[LocationCapacityResponse]
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
// @Success 200 {object} resp.SuccessResponse[TodayAppointmentsResponse]
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
// @Success 200 {object} resp.SuccessResponse[EvaluationStatsResponse]
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
// @Success 200 {object} resp.SuccessResponse[DischargeStatsResponse]
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
// @Success 200 {object} resp.SuccessResponse[CoordinatorUrgentAlertsResponse]
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
// @Success 200 {object} resp.SuccessResponse[CoordinatorTodayScheduleResponse]
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

// @Summary Get coordinator personal stats
// @Description Get personal statistics for the coordinator's dashboard summary
// @Tags Dashboard - Coordinator
// @Produce json
// @Success 200 {object} resp.SuccessResponse[CoordinatorStatsResponse]
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /dashboard/coordinator/stats [get]
func (h *DashboardHandler) GetCoordinatorStats(ctx *gin.Context) {
	employeeID, exists := ctx.Get(middleware.EmployeeIDKey)
	if !exists {
		ctx.JSON(http.StatusUnauthorized, resp.Error(ErrInternal))
		return
	}

	stats, err := h.dashboardService.GetCoordinatorStats(ctx, employeeID.(string))
	if err != nil {
		switch err {
		case ErrInternal:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		}
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(stats, "Coordinator stats retrieved successfully"))
}

// @Summary Get coordinator reminders
// @Description Get pending reminders and tasks for the coordinator
// @Tags Dashboard - Coordinator
// @Produce json
// @Success 200 {object} resp.SuccessResponse[CoordinatorRemindersResponse]
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /dashboard/coordinator/reminders [get]
func (h *DashboardHandler) GetCoordinatorReminders(ctx *gin.Context) {
	employeeID, exists := ctx.Get(middleware.EmployeeIDKey)
	if !exists {
		ctx.JSON(http.StatusUnauthorized, resp.Error(ErrInternal))
		return
	}

	reminders, err := h.dashboardService.GetCoordinatorReminders(ctx, employeeID.(string))
	if err != nil {
		switch err {
		case ErrInternal:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		}
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(reminders, "Coordinator reminders retrieved successfully"))
}

// @Summary Get coordinator clients
// @Description Get list of clients assigned to this coordinator
// @Tags Dashboard - Coordinator
// @Produce json
// @Success 200 {object} resp.SuccessResponse[CoordinatorClientsResponse]
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /dashboard/coordinator/clients [get]
func (h *DashboardHandler) GetCoordinatorClients(ctx *gin.Context) {
	employeeID, exists := ctx.Get(middleware.EmployeeIDKey)
	if !exists {
		ctx.JSON(http.StatusUnauthorized, resp.Error(ErrInternal))
		return
	}

	clients, err := h.dashboardService.GetCoordinatorClients(ctx, employeeID.(string))
	if err != nil {
		switch err {
		case ErrInternal:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		}
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(clients, "Coordinator clients retrieved successfully"))
}

// @Summary Get coordinator goals progress
// @Description Get aggregated goals progress for all coordinator's clients
// @Tags Dashboard - Coordinator
// @Produce json
// @Success 200 {object} resp.SuccessResponse[CoordinatorGoalsProgressResponse]
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /dashboard/coordinator/goals-progress [get]
func (h *DashboardHandler) GetCoordinatorGoalsProgress(ctx *gin.Context) {
	employeeID, exists := ctx.Get(middleware.EmployeeIDKey)
	if !exists {
		ctx.JSON(http.StatusUnauthorized, resp.Error(ErrInternal))
		return
	}

	progress, err := h.dashboardService.GetCoordinatorGoalsProgress(ctx, employeeID.(string))
	if err != nil {
		switch err {
		case ErrInternal:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		}
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(progress, "Coordinator goals progress retrieved successfully"))
}

// @Summary Get coordinator incidents
// @Description Get incidents for coordinator's assigned clients
// @Tags Dashboard - Coordinator
// @Produce json
// @Success 200 {object} resp.SuccessResponse[CoordinatorIncidentsResponse]
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /dashboard/coordinator/incidents [get]
func (h *DashboardHandler) GetCoordinatorIncidents(ctx *gin.Context) {
	employeeID, exists := ctx.Get(middleware.EmployeeIDKey)
	if !exists {
		ctx.JSON(http.StatusUnauthorized, resp.Error(ErrInternal))
		return
	}

	incidents, err := h.dashboardService.GetCoordinatorIncidents(ctx, employeeID.(string))
	if err != nil {
		switch err {
		case ErrInternal:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		}
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(incidents, "Coordinator incidents retrieved successfully"))
}
