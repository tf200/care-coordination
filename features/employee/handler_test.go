package employee_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"care-cordination/features/employee"
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

func setupHandlerTest(t *testing.T) (*gin.Engine, *mocks.MockEmployeeService, *gomock.Controller) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	mockService := mocks.NewMockEmployeeService(ctrl)

	handler := employee.NewEmployeeHandler(mockService, nil)

	router := gin.New()
	router.POST("/employees", handler.CreateEmployee)
	router.GET("/employees", handler.ListEmployees)
	router.GET("/employees/:id", handler.GetEmployeeByID)
	router.GET("/employees/me", handler.GetMyProfile)
	router.PUT("/employees/:id", handler.UpdateEmployee)
	router.DELETE("/employees/:id", handler.DeleteEmployee)

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
// Test: CreateEmployee
// ============================================================

func TestCreateEmployeeHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setup          func(mockService *mocks.MockEmployeeService)
		expectedStatus int
		validateBody   func(t *testing.T, body []byte)
	}{
		{
			name: "success",
			requestBody: employee.CreateEmployeeRequest{
				Email:       "test@example.com",
				Password:    "password123",
				FirstName:   "John",
				LastName:    "Doe",
				BSN:         "123456789",
				DateOfBirth: "1990-01-01",
				PhoneNumber: "0612345678",
				Gender:      "male",
				Role:        "admin",
				LocationID:  "loc-123",
			},
			setup: func(mockService *mocks.MockEmployeeService) {
				mockService.EXPECT().
					CreateEmployee(gomock.Any(), gomock.Any()).
					Return(employee.CreateEmployeeResponse{ID: "emp-123"}, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var response resp.SuccessResponse[employee.CreateEmployeeResponse]
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.True(t, response.Success)
				assert.Equal(t, "emp-123", response.Data.ID)
			},
		},
		{
			name: "invalid_request",
			requestBody: map[string]string{
				"email": "invalid-email",
			},
			setup:          func(mockService *mocks.MockEmployeeService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, mockService, ctrl := setupHandlerTest(t)
			defer ctrl.Finish()

			tt.setup(mockService)

			w := performRequest(router, "POST", "/employees", tt.requestBody)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.validateBody != nil {
				tt.validateBody(t, w.Body.Bytes())
			}
		})
	}
}

// ============================================================
// Test: ListEmployees
// ============================================================

func TestListEmployeesHandler(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(mockService *mocks.MockEmployeeService)
		expectedStatus int
	}{
		{
			name: "success",
			setup: func(mockService *mocks.MockEmployeeService) {
				mockService.EXPECT().
					ListEmployees(gomock.Any(), gomock.Any()).
					Return(&resp.PaginationResponse[employee.ListEmployeesResponse]{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, mockService, ctrl := setupHandlerTest(t)
			defer ctrl.Finish()

			tt.setup(mockService)

			w := performRequest(router, "GET", "/employees", nil)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// ============================================================
// Test: GetEmployeeByID
// ============================================================

func TestGetEmployeeByIDHandler(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		setup          func(mockService *mocks.MockEmployeeService)
		expectedStatus int
	}{
		{
			name: "success",
			id:   "emp-123",
			setup: func(mockService *mocks.MockEmployeeService) {
				mockService.EXPECT().
					GetEmployeeByID(gomock.Any(), "emp-123").
					Return(&employee.GetEmployeeByIDResponse{ID: "emp-123"}, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, mockService, ctrl := setupHandlerTest(t)
			defer ctrl.Finish()

			tt.setup(mockService)

			w := performRequest(router, "GET", "/employees/"+tt.id, nil)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// ============================================================
// Test: GetMyProfile
// ============================================================

func TestGetMyProfileHandler(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(mockService *mocks.MockEmployeeService)
		expectedStatus int
	}{
		{
			name: "success",
			setup: func(mockService *mocks.MockEmployeeService) {
				mockService.EXPECT().
					GetMyProfile(gomock.Any()).
					Return(&employee.GetMyProfileResponse{ID: "emp-123"}, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, mockService, ctrl := setupHandlerTest(t)
			defer ctrl.Finish()

			tt.setup(mockService)

			w := performRequest(router, "GET", "/employees/me", nil)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// ============================================================
// Test: UpdateEmployee
// ============================================================

func TestUpdateEmployeeHandler(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		requestBody    interface{}
		setup          func(mockService *mocks.MockEmployeeService)
		expectedStatus int
	}{
		{
			name: "success",
			id:   "emp-123",
			requestBody: employee.UpdateEmployeeRequest{
				FirstName: func() *string { s := "Jane"; return &s }(),
			},
			setup: func(mockService *mocks.MockEmployeeService) {
				mockService.EXPECT().
					UpdateEmployee(gomock.Any(), "emp-123", gomock.Any()).
					Return(&employee.UpdateEmployeeResponse{ID: "emp-123"}, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, mockService, ctrl := setupHandlerTest(t)
			defer ctrl.Finish()

			tt.setup(mockService)

			w := performRequest(router, "PUT", "/employees/"+tt.id, tt.requestBody)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// ============================================================
// Test: DeleteEmployee
// ============================================================

func TestDeleteEmployeeHandler(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		setup          func(mockService *mocks.MockEmployeeService)
		expectedStatus int
	}{
		{
			name: "success",
			id:   "emp-123",
			setup: func(mockService *mocks.MockEmployeeService) {
				mockService.EXPECT().
					DeleteEmployee(gomock.Any(), "emp-123").
					Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, mockService, ctrl := setupHandlerTest(t)
			defer ctrl.Finish()

			tt.setup(mockService)

			w := performRequest(router, "DELETE", "/employees/"+tt.id, nil)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
