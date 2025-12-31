package client_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"care-cordination/features/client"
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

func setupHandlerTest(t *testing.T) (*gin.Engine, *mocks.MockClientService, *gomock.Controller) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	mockService := mocks.NewMockClientService(ctrl)

	handler := client.NewClientHandler(mockService, nil)

	router := gin.New()
	router.POST("/clients/move-to-waiting-list", handler.MoveClientToWaitingList)
	router.POST("/clients/:id/move-to-care", handler.MoveClientInCare)
	router.POST("/clients/:id/start-discharge", handler.StartDischarge)
	router.POST("/clients/:id/complete-discharge", handler.CompleteDischarge)
	router.GET("/clients/waiting-list/stats", handler.GetWaitlistStats)
	router.GET("/clients/waiting-list", handler.ListWaitingListClients)
	router.GET("/clients/in-care/stats", handler.GetInCareStats)
	router.GET("/clients/in-care", handler.ListInCareClients)
	router.GET("/clients/discharged/stats", handler.GetDischargeStats)
	router.GET("/clients/discharged", handler.ListDischargedClients)
	router.GET("/clients/:id/goals", handler.ListClientGoals)

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
// Test: MoveClientToWaitingList
// ============================================================

func TestMoveClientToWaitingListHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setup          func(mockService *mocks.MockClientService)
		expectedStatus int
		validateBody   func(t *testing.T, body []byte)
	}{
		{
			name: "success",
			requestBody: client.MoveClientToWaitingListRequest{
				IntakeFormID:        "intake-123",
				WaitingListPriority: "high",
			},
			setup: func(mockService *mocks.MockClientService) {
				mockService.EXPECT().
					MoveClientToWaitingList(gomock.Any(), gomock.Any()).
					Return(&client.MoveClientToWaitingListResponse{ClientID: "client-123"}, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var response resp.SuccessResponse[client.MoveClientToWaitingListResponse]
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.True(t, response.Success)
				assert.Equal(t, "client-123", response.Data.ClientID)
			},
		},
		{
			name: "service_error",
			requestBody: client.MoveClientToWaitingListRequest{
				IntakeFormID:        "intake-123",
				WaitingListPriority: "high",
			},
			setup: func(mockService *mocks.MockClientService) {
				mockService.EXPECT().
					MoveClientToWaitingList(gomock.Any(), gomock.Any()).
					Return(nil, client.ErrIntakeFormNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, mockService, ctrl := setupHandlerTest(t)
			defer ctrl.Finish()

			tt.setup(mockService)

			w := performRequest(router, "POST", "/clients/move-to-waiting-list", tt.requestBody)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.validateBody != nil {
				tt.validateBody(t, w.Body.Bytes())
			}
		})
	}
}

// ============================================================
// Test: MoveClientInCare
// ============================================================

func TestMoveClientInCareHandler(t *testing.T) {
	hours := int32(20)
	tests := []struct {
		name           string
		clientID       string
		requestBody    interface{}
		setup          func(mockService *mocks.MockClientService)
		expectedStatus int
	}{
		{
			name:     "success",
			clientID: "client-123",
			requestBody: client.MoveClientInCareRequest{
				CareStartDate:         "2023-01-01",
				CareEndDate:           "2023-12-31",
				AmbulatoryWeeklyHours: &hours,
			},
			setup: func(mockService *mocks.MockClientService) {
				mockService.EXPECT().
					MoveClientInCare(gomock.Any(), "client-123", gomock.Any()).
					Return(&client.MoveClientInCareResponse{ClientID: "client-123"}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "not_found",
			clientID: "notfound",
			requestBody: client.MoveClientInCareRequest{
				CareStartDate: "2023-01-01",
				CareEndDate:   "2023-12-31",
			},
			setup: func(mockService *mocks.MockClientService) {
				mockService.EXPECT().
					MoveClientInCare(gomock.Any(), "notfound", gomock.Any()).
					Return(nil, client.ErrClientNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, mockService, ctrl := setupHandlerTest(t)
			defer ctrl.Finish()

			tt.setup(mockService)

			w := performRequest(router, "POST", "/clients/"+tt.clientID+"/move-to-care", tt.requestBody)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// ============================================================
// Test: StartDischarge
// ============================================================

func TestStartDischargeHandler(t *testing.T) {
	tests := []struct {
		name           string
		clientID       string
		requestBody    interface{}
		setup          func(mockService *mocks.MockClientService)
		expectedStatus int
	}{
		{
			name:     "success",
			clientID: "client-123",
			requestBody: client.StartDischargeRequest{
				DischargeDate:      "2023-12-31",
				ReasonForDischarge: "treatment_completed",
			},
			setup: func(mockService *mocks.MockClientService) {
				mockService.EXPECT().
					StartDischarge(gomock.Any(), "client-123", gomock.Any()).
					Return(&client.StartDischargeResponse{ClientID: "client-123"}, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, mockService, ctrl := setupHandlerTest(t)
			defer ctrl.Finish()

			tt.setup(mockService)

			w := performRequest(router, "POST", "/clients/"+tt.clientID+"/start-discharge", tt.requestBody)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// ============================================================
// Test: CompleteDischarge
// ============================================================

func TestCompleteDischargeHandler(t *testing.T) {
	tests := []struct {
		name           string
		clientID       string
		requestBody    interface{}
		setup          func(mockService *mocks.MockClientService)
		expectedStatus int
	}{
		{
			name:     "success",
			clientID: "client-123",
			requestBody: client.CompleteDischargeRequest{
				ClosingReport:    "Report content",
				EvaluationReport: "Evaluation content",
			},
			setup: func(mockService *mocks.MockClientService) {
				mockService.EXPECT().
					CompleteDischarge(gomock.Any(), "client-123", gomock.Any()).
					Return(&client.CompleteDischargeResponse{ClientID: "client-123"}, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, mockService, ctrl := setupHandlerTest(t)
			defer ctrl.Finish()

			tt.setup(mockService)

			w := performRequest(router, "POST", "/clients/"+tt.clientID+"/complete-discharge", tt.requestBody)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// ============================================================
// Test: GetStats Handlers
// ============================================================

func TestGetStatsHandlers(t *testing.T) {
	t.Run("GetWaitlistStats", func(t *testing.T) {
		router, mockService, ctrl := setupHandlerTest(t)
		defer ctrl.Finish()

		mockService.EXPECT().
			GetWaitlistStats(gomock.Any()).
			Return(&client.GetWaitlistStatsResponse{TotalCount: 10}, nil)

		w := performRequest(router, "GET", "/clients/waiting-list/stats", nil)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("GetInCareStats", func(t *testing.T) {
		router, mockService, ctrl := setupHandlerTest(t)
		defer ctrl.Finish()

		mockService.EXPECT().
			GetInCareStats(gomock.Any()).
			Return(&client.GetInCareStatsResponse{TotalCount: 5}, nil)

		w := performRequest(router, "GET", "/clients/in-care/stats", nil)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("GetDischargeStats", func(t *testing.T) {
		router, mockService, ctrl := setupHandlerTest(t)
		defer ctrl.Finish()

		mockService.EXPECT().
			GetDischargeStats(gomock.Any()).
			Return(&client.GetDischargeStatsResponse{TotalCount: 8}, nil)

		w := performRequest(router, "GET", "/clients/discharged/stats", nil)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// ============================================================
// Test: ListClients Handlers
// ============================================================

func TestListClientsHandlers(t *testing.T) {
	t.Run("ListWaitingListClients", func(t *testing.T) {
		router, mockService, ctrl := setupHandlerTest(t)
		defer ctrl.Finish()

		mockService.EXPECT().
			ListWaitingListClients(gomock.Any(), gomock.Any()).
			Return(&resp.PaginationResponse[client.ListWaitingListClientsResponse]{}, nil)

		w := performRequest(router, "GET", "/clients/waiting-list", nil)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("ListInCareClients", func(t *testing.T) {
		router, mockService, ctrl := setupHandlerTest(t)
		defer ctrl.Finish()

		mockService.EXPECT().
			ListInCareClients(gomock.Any(), gomock.Any()).
			Return(&resp.PaginationResponse[client.ListInCareClientsResponse]{}, nil)

		w := performRequest(router, "GET", "/clients/in-care", nil)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("ListDischargedClients", func(t *testing.T) {
		router, mockService, ctrl := setupHandlerTest(t)
		defer ctrl.Finish()

		mockService.EXPECT().
			ListDischargedClients(gomock.Any(), gomock.Any()).
			Return(&resp.PaginationResponse[client.ListDischargedClientsResponse]{}, nil)

		w := performRequest(router, "GET", "/clients/discharged", nil)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// ============================================================
// Test: ListClientGoals
// ============================================================

func TestListClientGoalsHandler(t *testing.T) {
	router, mockService, ctrl := setupHandlerTest(t)
	defer ctrl.Finish()

	mockService.EXPECT().
		ListClientGoals(gomock.Any(), "client-123").
		Return([]client.ListClientGoalsResponse{}, nil)

	w := performRequest(router, "GET", "/clients/client-123/goals", nil)

	assert.Equal(t, http.StatusOK, w.Code)
}
