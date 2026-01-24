package calendar

import (
	"care-cordination/lib/middleware"
	"care-cordination/lib/resp"
	"care-cordination/lib/util"
	"net/http"

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

func (h *CalendarHandler) SetupRoutes(router *gin.Engine) {
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
// @Security BearerAuth
// @Param start query string true "Start time (RFC3339 format)"
// @Param end query string true "End time (RFC3339 format)"
// @Param employee_id query string false "Employee ID (defaults to current user)"
// @Success 200 {object} resp.SuccessResponse[[]CalendarEventDTO]
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /calendar/view [get]
func (h *CalendarHandler) GetCalendarView(ctx *gin.Context) {
	var req GetCalendarViewRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(err))
		return
	}

	targetEmployeeID := util.GetEmployeeID(ctx)
	if req.EmployeeID != nil {
		targetEmployeeID = *req.EmployeeID
	}

	res, err := h.service.GetCalendarView(ctx, targetEmployeeID, req.Start, req.End)
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp.Success(res, "Calendar view retrieved successfully"))
}

// Appointment handlers

// @Summary Create appointment
// @Description Create a new calendar appointment
// @Tags Calendar - Appointments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateAppointmentRequest true "Appointment details"
// @Success 201 {object} resp.SuccessResponse[AppointmentResponse]
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /calendar/appointments [post]
func (h *CalendarHandler) CreateAppointment(ctx *gin.Context) {
	var req CreateAppointmentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(err))
		return
	}

	userID := util.GetEmployeeID(ctx)
	res, err := h.service.CreateAppointment(ctx, userID, req)
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, resp.Success(res, "Appointment created successfully"))
}

// @Summary Get appointment
// @Description Get a specific appointment by ID
// @Tags Calendar - Appointments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Appointment ID"
// @Success 200 {object} resp.SuccessResponse[AppointmentResponse]
// @Failure 401 {object} resp.ErrorResponse
// @Failure 404 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /calendar/appointments/{id} [get]
func (h *CalendarHandler) GetAppointment(ctx *gin.Context) {
	id := ctx.Param("id")
	res, err := h.service.GetAppointment(ctx, id)
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp.Success(res, "Appointment retrieved successfully"))
}

// @Summary List appointments
// @Description List all appointments for an employee
// @Tags Calendar - Appointments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param employee_id query string false "Employee ID (defaults to current user)"
// @Success 200 {object} resp.SuccessResponse[[]AppointmentResponse]
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /calendar/appointments [get]
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

// @Summary Update appointment
// @Description Update an existing appointment
// @Tags Calendar - Appointments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Appointment ID"
// @Param request body UpdateAppointmentRequest true "Updated appointment details"
// @Success 200 {object} resp.SuccessResponse[AppointmentResponse]
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 404 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /calendar/appointments/{id} [patch]
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

// @Summary Delete appointment
// @Description Delete an appointment by ID
// @Tags Calendar - Appointments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Appointment ID"
// @Success 200 {object} resp.MessageResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 404 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /calendar/appointments/{id} [delete]
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

// @Summary Create reminder
// @Description Create a new reminder
// @Tags Calendar - Reminders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateReminderRequest true "Reminder details"
// @Success 201 {object} resp.SuccessResponse[ReminderResponse]
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /calendar/reminders [post]
func (h *CalendarHandler) CreateReminder(ctx *gin.Context) {
	var req CreateReminderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(err))
		return
	}

	userID := util.GetEmployeeID(ctx)
	res, err := h.service.CreateReminder(ctx, userID, req)
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, resp.Success(res, "Reminder created successfully"))
}

// @Summary Get reminder
// @Description Get a specific reminder by ID
// @Tags Calendar - Reminders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Reminder ID"
// @Success 200 {object} resp.SuccessResponse[ReminderResponse]
// @Failure 401 {object} resp.ErrorResponse
// @Failure 404 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /calendar/reminders/{id} [get]
func (h *CalendarHandler) GetReminder(ctx *gin.Context) {
	id := ctx.Param("id")
	res, err := h.service.GetReminder(ctx, id)
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp.Success(res, "Reminder retrieved successfully"))
}

// @Summary List reminders
// @Description List all reminders for an employee
// @Tags Calendar - Reminders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param employee_id query string false "Employee ID (defaults to current user)"
// @Success 200 {object} resp.SuccessResponse[[]ReminderResponse]
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /calendar/reminders [get]
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

// @Summary Update reminder
// @Description Update a reminder's completion status
// @Tags Calendar - Reminders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Reminder ID"
// @Param request body UpdateReminderRequest true "Updated reminder details"
// @Success 200 {object} resp.SuccessResponse[ReminderResponse]
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 404 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /calendar/reminders/{id} [patch]
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

// @Summary Delete reminder
// @Description Delete a reminder by ID
// @Tags Calendar - Reminders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Reminder ID"
// @Success 200 {object} resp.MessageResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 404 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /calendar/reminders/{id} [delete]
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
