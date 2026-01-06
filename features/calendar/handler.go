package calendar

import (
	"care-cordination/features/middleware"
	"care-cordination/lib/resp"
	"care-cordination/lib/util"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type CalendarHandler struct {
	service CalendarService
	mdw     *middleware.Middleware
}

func NewCalendarHandler(service CalendarService, mdw *middleware.Middleware) *CalendarHandler {
	return &CalendarHandler{
		service: service,
		mdw:     mdw,
	}
}

func (h *CalendarHandler) SetupRoutes(router *gin.RouterGroup) {
	calendar := router.Group("/calendar")
	calendar.Use(h.mdw.AuthMdw())
	{
		calendar.POST("/appointments", h.CreateAppointment)
		calendar.GET("/appointments", h.ListAppointments)
		calendar.GET("/appointments/:id", h.GetAppointment)
		calendar.PATCH("/appointments/:id", h.UpdateAppointment)
		calendar.DELETE("/appointments/:id", h.DeleteAppointment)

		calendar.POST("/reminders", h.CreateReminder)
		calendar.GET("/reminders", h.ListReminders)
		calendar.GET("/reminders/:id", h.GetReminder)
		calendar.PATCH("/reminders/:id", h.UpdateReminder)
		calendar.DELETE("/reminders/:id", h.DeleteReminder)

		calendar.GET("/view", h.GetCalendarView)
	}
}

// @Summary Get calendar view
// @Description Get a unified list of appointments and reminders for a date range
// @Tags Calendar
// @Accept json
// @Produce json
// @Param start query string true "Start time (ISO8601)"
// @Param end query string true "End time (ISO8601)"
// @Success 200 {object} resp.SuccessResponse[[]CalendarEventDTO]
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /api/v1/calendar/view [get]
func (h *CalendarHandler) GetCalendarView(ctx *gin.Context) {
	startStr := ctx.Query("start")
	endStr := ctx.Query("end")

	if startStr == "" || endStr == "" {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}

	start, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(err))
		return
	}

	end, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(err))
		return
	}

	targetEmployeeID := ctx.Query("employee_id")
	if targetEmployeeID == "" {
		targetEmployeeID = util.GetEmployeeID(ctx)
	}

	res, err := h.service.GetCalendarView(ctx, targetEmployeeID, start, end)
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp.Success(res, "Calendar view retrieved successfully"))
}

// Appointment handlers

func (h *CalendarHandler) CreateAppointment(ctx *gin.Context) {
	var req CreateAppointmentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(err))
		return
	}

	userID := util.GetUserID(ctx)
	res, err := h.service.CreateAppointment(ctx, userID, req)
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, resp.Success(res, "Appointment created successfully"))
}

func (h *CalendarHandler) GetAppointment(ctx *gin.Context) {
	id := ctx.Param("id")
	res, err := h.service.GetAppointment(ctx, id)
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp.Success(res, "Appointment retrieved successfully"))
}

func (h *CalendarHandler) ListAppointments(ctx *gin.Context) {
	targetEmployeeID := ctx.Query("employee_id")
	if targetEmployeeID == "" {
		targetEmployeeID = util.GetEmployeeID(ctx)
	}
	res, err := h.service.ListAppointments(ctx, targetEmployeeID)
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp.Success(res, "Appointments retrieved successfully"))
}

func (h *CalendarHandler) UpdateAppointment(ctx *gin.Context) {
	id := ctx.Param("id")
	var req UpdateAppointmentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(err))
		return
	}

	res, err := h.service.UpdateAppointment(ctx, id, req)
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp.Success(res, "Appointment updated successfully"))
}

func (h *CalendarHandler) DeleteAppointment(ctx *gin.Context) {
	id := ctx.Param("id")
	err := h.service.DeleteAppointment(ctx, id)
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp.MessageResonse("Appointment deleted successfully"))
}

// Reminder handlers

func (h *CalendarHandler) CreateReminder(ctx *gin.Context) {
	var req CreateReminderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(err))
		return
	}

	userID := util.GetUserID(ctx)
	res, err := h.service.CreateReminder(ctx, userID, req)
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, resp.Success(res, "Reminder created successfully"))
}

func (h *CalendarHandler) GetReminder(ctx *gin.Context) {
	id := ctx.Param("id")
	res, err := h.service.GetReminder(ctx, id)
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp.Success(res, "Reminder retrieved successfully"))
}

func (h *CalendarHandler) ListReminders(ctx *gin.Context) {
	targetEmployeeID := ctx.Query("employee_id")
	if targetEmployeeID == "" {
		targetEmployeeID = util.GetEmployeeID(ctx)
	}
	res, err := h.service.ListReminders(ctx, targetEmployeeID)
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp.Success(res, "Reminders retrieved successfully"))
}

func (h *CalendarHandler) UpdateReminder(ctx *gin.Context) {
	id := ctx.Param("id")
	var req struct {
		IsCompleted bool `json:"is_completed"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(err))
		return
	}

	res, err := h.service.UpdateReminder(ctx, id, req.IsCompleted)
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp.Success(res, "Reminder updated successfully"))
}

func (h *CalendarHandler) DeleteReminder(ctx *gin.Context) {
	id := ctx.Param("id")
	err := h.service.DeleteReminder(ctx, id)
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp.MessageResonse("Reminder deleted successfully"))
}

func (h *CalendarHandler) handleError(ctx *gin.Context, err error) {
	switch err {
	case ErrAppointmentNotFound, ErrReminderNotFound:
		ctx.JSON(http.StatusNotFound, resp.Error(err))
	case ErrUnauthorized:
		ctx.JSON(http.StatusUnauthorized, resp.Error(err))
	case ErrInternal:
		ctx.JSON(http.StatusInternalServerError, resp.Error(err))
	default:
		ctx.JSON(http.StatusInternalServerError, resp.Error(err))
	}
}
