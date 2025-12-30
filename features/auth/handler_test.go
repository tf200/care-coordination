package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"care-cordination/features/auth"
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

func setupHandlerTest(t *testing.T) (*gin.Engine, *mocks.MockAuthService, *gomock.Controller) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	mockService := mocks.NewMockAuthService(ctrl)

	handler := auth.NewAuthHandler(mockService, nil)

	router := gin.New()
	router.POST("/auth/login", handler.Login)
	router.POST("/auth/refresh", handler.RefreshTokens)
	router.POST("/auth/logout", func(c *gin.Context) {
		// Simulate authenticated user for logout tests
		c.Set("user_id", "test-user-id")
		handler.Logout(c)
	})

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
// Test: Login Handler
// ============================================================

func TestLoginHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setup          func(mockService *mocks.MockAuthService)
		expectedStatus int
		validateBody   func(t *testing.T, body []byte)
	}{
		{
			name: "success",
			requestBody: auth.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			setup: func(mockService *mocks.MockAuthService) {
				mockService.EXPECT().
					Login(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&auth.LoginResponse{
						AccessToken:  "access-token-123",
						RefreshToken: "refresh-token-123",
					}, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var response resp.SuccessResponse[auth.LoginResponse]
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)

				assert.True(t, response.Success)
				assert.Equal(t, "access-token-123", response.Data.AccessToken)
				assert.Equal(t, "refresh-token-123", response.Data.RefreshToken)
				assert.Equal(t, "Login successful", response.Message)
			},
		},
		{
			name: "invalid_credentials",
			requestBody: auth.LoginRequest{
				Email:    "wrong@example.com",
				Password: "wrongpass",
			},
			setup: func(mockService *mocks.MockAuthService) {
				mockService.EXPECT().
					Login(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, auth.ErrInvalidCredentials)
			},
			expectedStatus: http.StatusUnauthorized,
			validateBody: func(t *testing.T, body []byte) {
				var response resp.ErrorResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)

				assert.False(t, response.Success)
				assert.Equal(t, "invalid_credentials", response.Error)
			},
		},
		{
			name: "internal_error",
			requestBody: auth.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			setup: func(mockService *mocks.MockAuthService) {
				mockService.EXPECT().
					Login(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, auth.ErrInternal)
			},
			expectedStatus: http.StatusInternalServerError,
			validateBody: func(t *testing.T, body []byte) {
				var response resp.ErrorResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)

				assert.False(t, response.Success)
				assert.Equal(t, "internal", response.Error)
			},
		},
		{
			name:           "invalid_json",
			requestBody:    "not valid json",
			setup:          func(mockService *mocks.MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body []byte) {
				var response resp.ErrorResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.False(t, response.Success)
			},
		},
		{
			name:           "missing_required_fields",
			requestBody:    map[string]string{"email": "test@example.com"}, // missing password
			setup:          func(mockService *mocks.MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body []byte) {
				var response resp.ErrorResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.False(t, response.Success)
			},
		},
		{
			name:           "invalid_email_format",
			requestBody:    map[string]string{"email": "not-an-email", "password": "password123"},
			setup:          func(mockService *mocks.MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body []byte) {
				var response resp.ErrorResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.False(t, response.Success)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, mockService, ctrl := setupHandlerTest(t)
			defer ctrl.Finish()

			tt.setup(mockService)

			w := performRequest(router, "POST", "/auth/login", tt.requestBody)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.validateBody != nil {
				tt.validateBody(t, w.Body.Bytes())
			}
		})
	}
}

// ============================================================
// Test: RefreshTokens Handler
// ============================================================

func TestRefreshTokensHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setup          func(mockService *mocks.MockAuthService)
		expectedStatus int
		validateBody   func(t *testing.T, body []byte)
	}{
		{
			name: "success",
			requestBody: auth.RefreshTokensRequest{
				RefreshToken: "valid-refresh-token",
			},
			setup: func(mockService *mocks.MockAuthService) {
				mockService.EXPECT().
					RefreshTokens(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&auth.RefreshTokensResponse{
						AccessToken:  "new-access-token",
						RefreshToken: "new-refresh-token",
					}, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var response resp.SuccessResponse[auth.RefreshTokensResponse]
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)

				assert.True(t, response.Success)
				assert.Equal(t, "new-access-token", response.Data.AccessToken)
				assert.Equal(t, "new-refresh-token", response.Data.RefreshToken)
				assert.Equal(t, "Tokens refreshed successfully", response.Message)
			},
		},
		{
			name: "invalid_token",
			requestBody: auth.RefreshTokensRequest{
				RefreshToken: "invalid-token",
			},
			setup: func(mockService *mocks.MockAuthService) {
				mockService.EXPECT().
					RefreshTokens(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, auth.ErrInvalidToken)
			},
			expectedStatus: http.StatusUnauthorized,
			validateBody: func(t *testing.T, body []byte) {
				var response resp.ErrorResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)

				assert.False(t, response.Success)
				assert.Equal(t, "invalid_token", response.Error)
			},
		},
		{
			name:           "missing_refresh_token",
			requestBody:    map[string]string{},
			setup:          func(mockService *mocks.MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body []byte) {
				var response resp.ErrorResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.False(t, response.Success)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, mockService, ctrl := setupHandlerTest(t)
			defer ctrl.Finish()

			tt.setup(mockService)

			w := performRequest(router, "POST", "/auth/refresh", tt.requestBody)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.validateBody != nil {
				tt.validateBody(t, w.Body.Bytes())
			}
		})
	}
}

// ============================================================
// Test: Logout Handler
// ============================================================

func TestLogoutHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setup          func(mockService *mocks.MockAuthService)
		expectedStatus int
		validateBody   func(t *testing.T, body []byte)
	}{
		{
			name: "success",
			requestBody: auth.LogoutRequest{
				RefreshToken: "valid-refresh-token",
			},
			setup: func(mockService *mocks.MockAuthService) {
				mockService.EXPECT().
					Logout(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var response resp.MessageResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)

				assert.True(t, response.Success)
				assert.Equal(t, "Successfully logged out", response.Message)
			},
		},
		{
			name: "invalid_token",
			requestBody: auth.LogoutRequest{
				RefreshToken: "invalid-token",
			},
			setup: func(mockService *mocks.MockAuthService) {
				mockService.EXPECT().
					Logout(gomock.Any(), gomock.Any()).
					Return(auth.ErrInvalidToken)
			},
			expectedStatus: http.StatusUnauthorized,
			validateBody: func(t *testing.T, body []byte) {
				var response resp.ErrorResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)

				assert.False(t, response.Success)
				assert.Equal(t, "invalid_token", response.Error)
			},
		},
		{
			name: "internal_error",
			requestBody: auth.LogoutRequest{
				RefreshToken: "valid-token",
			},
			setup: func(mockService *mocks.MockAuthService) {
				mockService.EXPECT().
					Logout(gomock.Any(), gomock.Any()).
					Return(auth.ErrInternal)
			},
			expectedStatus: http.StatusInternalServerError,
			validateBody: func(t *testing.T, body []byte) {
				var response resp.ErrorResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)

				assert.False(t, response.Success)
				assert.Equal(t, "internal", response.Error)
			},
		},
		{
			name:           "missing_refresh_token",
			requestBody:    map[string]string{},
			setup:          func(mockService *mocks.MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body []byte) {
				var response resp.ErrorResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.False(t, response.Success)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, mockService, ctrl := setupHandlerTest(t)
			defer ctrl.Finish()

			tt.setup(mockService)

			w := performRequest(router, "POST", "/auth/logout", tt.requestBody)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.validateBody != nil {
				tt.validateBody(t, w.Body.Bytes())
			}
		})
	}
}
