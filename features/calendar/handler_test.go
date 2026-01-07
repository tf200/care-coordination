package calendar_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"care-cordination/features/calendar"
	"care-cordination/internal/mocks"
	"care-cordination/lib/resp"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// ============================================================
// Test Helpers
// ============================================================

func setupHandlerTest(t *testing.T) (*gin.Engine, *mocks.MockCalendarService, *gomock.Controller) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	mockService := mocks.NewMockCalendarService(ctrl)

	handler := calendar.NewCalendarHandler(mockService, nil)

	router := gin.New()
	api := router.Group("/api/v1")

	// Simulate authenticated user
	api.Use(func(c *gin.Context) {
		c.Set("user_id", "test-user-id")
		c.Set("employee_id", "test-employee-id")
		c.Next()
	})

	api.POST("/calendar/appointments", handler.CreateAppointment)
	api.GET("/calendar/appointments", handler.ListAppointments)
	api.GET("/calendar/appointments/:id", handler.GetAppointment)
	api.PATCH("/calendar/appointments/:id", handler.UpdateAppointment)
	api.DELETE("/calendar/appointments/:id", handler.DeleteAppointment)

	api.POST("/calendar/reminders", handler.CreateReminder)
	api.GET("/calendar/reminders", handler.ListReminders)
	api.GET("/calendar/reminders/:id", handler.GetReminder)
	api.PATCH("/calendar/reminders/:id", handler.UpdateReminder)
	api.DELETE("/calendar/reminders/:id", handler.DeleteReminder)

	api.GET("/calendar/view", handler.GetCalendarView)

	return router, mockService, ctrl
}

func performRequest(router *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBytes, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonBytes)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, _ := http.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// ============================================================
// Test: Create Appointment Handler
// ============================================================

func TestCreateAppointmentHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setup          func(mockService *mocks.MockCalendarService)
		expectedStatus int
		validateBody   func(t *testing.T, body []byte)
	}{
		{
			name: "success",
			requestBody: calendar.CreateAppointmentRequest{
				Title:     "Test Appointment",
				StartTime: time.Now().Add(time.Hour),
				EndTime:   time.Now().Add(2 * time.Hour),
				Type:      calendar.TypeGeneral,
				Participants: []calendar.ParticipantDTO{
					{ID: "client-1", Type: calendar.ParticipantClient},
				},
			},
			setup: func(mockService *mocks.MockCalendarService) {
				mockService.EXPECT().
					CreateAppointment(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&calendar.AppointmentResponse{
						ID:    "app-123",
						Title: "Test Appointment",
					}, nil)
			},
			expectedStatus: http.StatusCreated,
			validateBody: func(t *testing.T, body []byte) {
				var response resp.SuccessResponse[calendar.AppointmentResponse]
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.True(t, response.Success)
				assert.Equal(t, "app-123", response.Data.ID)
			},
		},
		{
			name:           "invalid_json",
			requestBody:    "not valid json",
			setup:          func(mockService *mocks.MockCalendarService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "internal_error",
			requestBody: calendar.CreateAppointmentRequest{
				Title:     "Test Appointment",
				StartTime: time.Now().Add(time.Hour),
				EndTime:   time.Now().Add(2 * time.Hour),
				Type:      calendar.TypeGeneral,
				Participants: []calendar.ParticipantDTO{
					{ID: "client-1", Type: calendar.ParticipantClient},
				},
			},
			setup: func(mockService *mocks.MockCalendarService) {
				mockService.EXPECT().
					CreateAppointment(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, calendar.ErrInternal)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, mockService, ctrl := setupHandlerTest(t)
			defer ctrl.Finish()

			tt.setup(mockService)

			w := performRequest(router, "POST", "/api/v1/calendar/appointments", tt.requestBody)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.validateBody != nil {
				tt.validateBody(t, w.Body.Bytes())
			}
		})
	}
}

// ============================================================
// Test: Get Appointment Handler
// ============================================================

func TestGetAppointmentHandler(t *testing.T) {
	tests := []struct {
		name           string
		appointmentID  string
		setup          func(mockService *mocks.MockCalendarService)
		expectedStatus int
	}{
		{
			name:          "success",
			appointmentID: "app-123",
			setup: func(mockService *mocks.MockCalendarService) {
				mockService.EXPECT().
					GetAppointment(gomock.Any(), "app-123").
					Return(&calendar.AppointmentResponse{ID: "app-123"}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:          "not_found",
			appointmentID: "non-existent",
			setup: func(mockService *mocks.MockCalendarService) {
				mockService.EXPECT().
					GetAppointment(gomock.Any(), "non-existent").
					Return(nil, calendar.ErrAppointmentNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, mockService, ctrl := setupHandlerTest(t)
			defer ctrl.Finish()

			tt.setup(mockService)

			w := performRequest(router, "GET", "/api/v1/calendar/appointments/"+tt.appointmentID, nil)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// ============================================================
// Test: List Appointments Handler
// ============================================================

func TestListAppointmentsHandler(t *testing.T) {
	router, mockService, ctrl := setupHandlerTest(t)
	defer ctrl.Finish()

	mockService.EXPECT().
		ListAppointments(gomock.Any(), gomock.Any()).
		Return([]calendar.AppointmentResponse{{ID: "app-1"}, {ID: "app-2"}}, nil)

	w := performRequest(router, "GET", "/api/v1/calendar/appointments", nil)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ============================================================
// Test: Update Appointment Handler
// ============================================================

func TestUpdateAppointmentHandler(t *testing.T) {
	router, mockService, ctrl := setupHandlerTest(t)
	defer ctrl.Finish()

	title := "Updated Title"
	mockService.EXPECT().
		UpdateAppointment(gomock.Any(), "app-123", gomock.Any()).
		Return(&calendar.AppointmentResponse{ID: "app-123", Title: title}, nil)

	w := performRequest(router, "PATCH", "/api/v1/calendar/appointments/app-123", calendar.UpdateAppointmentRequest{
		Title: &title,
	})

	assert.Equal(t, http.StatusOK, w.Code)
}

// ============================================================
// Test: Delete Appointment Handler
// ============================================================

func TestDeleteAppointmentHandler(t *testing.T) {
	router, mockService, ctrl := setupHandlerTest(t)
	defer ctrl.Finish()

	mockService.EXPECT().
		DeleteAppointment(gomock.Any(), "app-123").
		Return(nil)

	w := performRequest(router, "DELETE", "/api/v1/calendar/appointments/app-123", nil)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ============================================================
// Test: Create Reminder Handler
// ============================================================

func TestCreateReminderHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setup          func(mockService *mocks.MockCalendarService)
		expectedStatus int
	}{
		{
			name: "success",
			requestBody: calendar.CreateReminderRequest{
				Title:   "Test Reminder",
				DueTime: time.Now().Add(time.Hour),
			},
			setup: func(mockService *mocks.MockCalendarService) {
				mockService.EXPECT().
					CreateReminder(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&calendar.ReminderResponse{ID: "rem-123"}, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "missing_title",
			requestBody:    map[string]interface{}{"due_time": time.Now().Add(time.Hour)},
			setup:          func(mockService *mocks.MockCalendarService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, mockService, ctrl := setupHandlerTest(t)
			defer ctrl.Finish()

			tt.setup(mockService)

			w := performRequest(router, "POST", "/api/v1/calendar/reminders", tt.requestBody)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// ============================================================
// Test: Get Reminder Handler
// ============================================================

func TestGetReminderHandler(t *testing.T) {
	router, mockService, ctrl := setupHandlerTest(t)
	defer ctrl.Finish()

	mockService.EXPECT().
		GetReminder(gomock.Any(), "rem-123").
		Return(&calendar.ReminderResponse{ID: "rem-123"}, nil)

	w := performRequest(router, "GET", "/api/v1/calendar/reminders/rem-123", nil)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ============================================================
// Test: List Reminders Handler
// ============================================================

func TestListRemindersHandler(t *testing.T) {
	router, mockService, ctrl := setupHandlerTest(t)
	defer ctrl.Finish()

	mockService.EXPECT().
		ListReminders(gomock.Any(), gomock.Any()).
		Return([]calendar.ReminderResponse{{ID: "rem-1"}}, nil)

	w := performRequest(router, "GET", "/api/v1/calendar/reminders", nil)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ============================================================
// Test: Update Reminder Handler
// ============================================================

func TestUpdateReminderHandler(t *testing.T) {
	router, mockService, ctrl := setupHandlerTest(t)
	defer ctrl.Finish()

	mockService.EXPECT().
		UpdateReminder(gomock.Any(), "rem-123", true).
		Return(&calendar.ReminderResponse{ID: "rem-123", IsCompleted: true}, nil)

	w := performRequest(router, "PATCH", "/api/v1/calendar/reminders/rem-123", map[string]bool{"is_completed": true})

	assert.Equal(t, http.StatusOK, w.Code)
}

// ============================================================
// Test: Delete Reminder Handler
// ============================================================

func TestDeleteReminderHandler(t *testing.T) {
	router, mockService, ctrl := setupHandlerTest(t)
	defer ctrl.Finish()

	mockService.EXPECT().
		DeleteReminder(gomock.Any(), "rem-123").
		Return(nil)

	w := performRequest(router, "DELETE", "/api/v1/calendar/reminders/rem-123", nil)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ============================================================
// Test: Get Calendar View Handler
// ============================================================

func TestGetCalendarViewHandler(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		setup          func(mockService *mocks.MockCalendarService)
		expectedStatus int
	}{
		{
			name:        "success",
			queryParams: "?start=2025-01-01T00:00:00Z&end=2025-01-31T23:59:59Z",
			setup: func(mockService *mocks.MockCalendarService) {
				mockService.EXPECT().
					GetCalendarView(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return([]calendar.CalendarEventDTO{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing_params",
			queryParams:    "",
			setup:          func(mockService *mocks.MockCalendarService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid_start_time",
			queryParams:    "?start=invalid&end=2025-01-31T23:59:59Z",
			setup:          func(mockService *mocks.MockCalendarService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, mockService, ctrl := setupHandlerTest(t)
			defer ctrl.Finish()

			tt.setup(mockService)

			w := performRequest(router, "GET", "/api/v1/calendar/view"+tt.queryParams, nil)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
