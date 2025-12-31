package client

import (
	"context"
	"errors"
	"testing"
	"time"

	db "care-cordination/lib/db/sqlc"
	dbmocks "care-cordination/lib/db/sqlc/mocks"
	loggermocks "care-cordination/lib/logger/mocks"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestMoveClientToWaitingList(t *testing.T) {
	tests := []struct {
		name        string
		req         *MoveClientToWaitingListRequest
		setup       func(mockStore *dbmocks.MockStoreInterface)
		wantErr     bool
		expectedErr error
		validate    func(t *testing.T, resp *MoveClientToWaitingListResponse)
	}{
		{
			name: "success",
			req: &MoveClientToWaitingListRequest{
				IntakeFormID:        "intake-123",
				WaitingListPriority: "high",
			},
			setup: func(mockStore *dbmocks.MockStoreInterface) {
				mockStore.EXPECT().
					GetIntakeForm(gomock.Any(), "intake-123").
					Return(db.IntakeForm{
						ID:                 "intake-123",
						RegistrationFormID: "reg-123",
						LocationID:         "loc-123",
						CoordinatorID:      "coord-123",
					}, nil)

				mockStore.EXPECT().
					GetRegistrationForm(gomock.Any(), "reg-123").
					Return(db.RegistrationForm{
						ID:          "reg-123",
						FirstName:   "John",
						LastName:    "Doe",
						Bsn:         "123456789",
						DateOfBirth: pgtype.Date{Time: time.Now(), Valid: true},
					}, nil)

				mockStore.EXPECT().
					MoveClientToWaitingListTx(gomock.Any(), gomock.Any()).
					Return(db.MoveClientToWaitingListTxResult{ClientID: "client-123"}, nil)
			},
			wantErr: false,
			validate: func(t *testing.T, resp *MoveClientToWaitingListResponse) {
				assert.Equal(t, "client-123", resp.ClientID)
			},
		},
		{
			name: "missing_intake_form_id",
			req: &MoveClientToWaitingListRequest{
				IntakeFormID: "",
			},
			setup:       func(mockStore *dbmocks.MockStoreInterface) {},
			wantErr:     true,
			expectedErr: ErrInvalidRequest,
		},
		{
			name: "intake_form_not_found",
			req: &MoveClientToWaitingListRequest{
				IntakeFormID: "notfound",
			},
			setup: func(mockStore *dbmocks.MockStoreInterface) {
				mockStore.EXPECT().
					GetIntakeForm(gomock.Any(), "notfound").
					Return(db.IntakeForm{}, pgx.ErrNoRows)
			},
			wantErr:     true,
			expectedErr: ErrIntakeFormNotFound,
		},
		{
			name: "registration_form_not_found",
			req: &MoveClientToWaitingListRequest{
				IntakeFormID: "intake-123",
			},
			setup: func(mockStore *dbmocks.MockStoreInterface) {
				mockStore.EXPECT().
					GetIntakeForm(gomock.Any(), "intake-123").
					Return(db.IntakeForm{ID: "intake-123", RegistrationFormID: "reg-123"}, nil)

				mockStore.EXPECT().
					GetRegistrationForm(gomock.Any(), "reg-123").
					Return(db.RegistrationForm{}, pgx.ErrNoRows)
			},
			wantErr:     true,
			expectedErr: ErrRegistrationFormNotFound,
		},
		{
			name: "tx_error",
			req: &MoveClientToWaitingListRequest{
				IntakeFormID: "intake-123",
			},
			setup: func(mockStore *dbmocks.MockStoreInterface) {
				mockStore.EXPECT().
					GetIntakeForm(gomock.Any(), "intake-123").
					Return(db.IntakeForm{ID: "intake-123", RegistrationFormID: "reg-123"}, nil)

				mockStore.EXPECT().
					GetRegistrationForm(gomock.Any(), "reg-123").
					Return(db.RegistrationForm{ID: "reg-123"}, nil)

				mockStore.EXPECT().
					MoveClientToWaitingListTx(gomock.Any(), gomock.Any()).
					Return(db.MoveClientToWaitingListTxResult{}, errors.New("db error"))
			},
			wantErr:     true,
			expectedErr: ErrFailedToCreateClient,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := dbmocks.NewMockStoreInterface(ctrl)
			mockLogger := loggermocks.NewMockLogger(ctrl)

			mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
			mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			tt.setup(mockStore)

			service := NewClientService(mockStore, mockLogger)

			resp, err := service.MoveClientToWaitingList(context.Background(), tt.req)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
				return
			}

			require.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, resp)
			}
		})
	}
}

func TestMoveClientInCare(t *testing.T) {
	hours := int32(20)
	tests := []struct {
		name        string
		clientID    string
		req         *MoveClientInCareRequest
		setup       func(mockStore *dbmocks.MockStoreInterface)
		wantErr     bool
		expectedErr error
		validate    func(t *testing.T, resp *MoveClientInCareResponse)
	}{
		{
			name:     "success_ambulatory",
			clientID: "client-123",
			req: &MoveClientInCareRequest{
				CareStartDate:         "2023-01-01",
				CareEndDate:           "2023-12-31",
				AmbulatoryWeeklyHours: &hours,
			},
			setup: func(mockStore *dbmocks.MockStoreInterface) {
				mockStore.EXPECT().
					GetClientByID(gomock.Any(), "client-123").
					Return(db.Client{
						ID:       "client-123",
						Status:   db.ClientStatusEnumWaitingList,
						CareType: db.CareTypeEnumAmbulatoryCare,
					}, nil)

				mockStore.EXPECT().
					UpdateClient(gomock.Any(), gomock.Any()).
					Return("client-123", nil)
			},
			wantErr: false,
			validate: func(t *testing.T, resp *MoveClientInCareResponse) {
				assert.Equal(t, "client-123", resp.ClientID)
			},
		},
		{
			name:     "success_protected_living",
			clientID: "client-123",
			req: &MoveClientInCareRequest{
				CareStartDate: "2023-01-01",
				CareEndDate:   "2023-12-31",
			},
			setup: func(mockStore *dbmocks.MockStoreInterface) {
				mockStore.EXPECT().
					GetClientByID(gomock.Any(), "client-123").
					Return(db.Client{
						ID:       "client-123",
						Status:   db.ClientStatusEnumWaitingList,
						CareType: db.CareTypeEnumProtectedLiving,
					}, nil)

				mockStore.EXPECT().
					UpdateClient(gomock.Any(), gomock.Any()).
					Return("client-123", nil)
			},
			wantErr: false,
		},
		{
			name:     "client_not_found",
			clientID: "notfound",
			req:      &MoveClientInCareRequest{},
			setup: func(mockStore *dbmocks.MockStoreInterface) {
				mockStore.EXPECT().
					GetClientByID(gomock.Any(), "notfound").
					Return(db.Client{}, pgx.ErrNoRows)
			},
			wantErr:     true,
			expectedErr: ErrClientNotFound,
		},
		{
			name:     "invalid_status",
			clientID: "client-123",
			req:      &MoveClientInCareRequest{},
			setup: func(mockStore *dbmocks.MockStoreInterface) {
				mockStore.EXPECT().
					GetClientByID(gomock.Any(), "client-123").
					Return(db.Client{
						ID:     "client-123",
						Status: db.ClientStatusEnumInCare,
					}, nil)
			},
			wantErr:     true,
			expectedErr: ErrInvalidClientStatus,
		},
		{
			name:     "ambulatory_hours_required",
			clientID: "client-123",
			req: &MoveClientInCareRequest{
				CareStartDate: "2023-01-01",
				CareEndDate:   "2023-12-31",
			},
			setup: func(mockStore *dbmocks.MockStoreInterface) {
				mockStore.EXPECT().
					GetClientByID(gomock.Any(), "client-123").
					Return(db.Client{
						ID:       "client-123",
						Status:   db.ClientStatusEnumWaitingList,
						CareType: db.CareTypeEnumAmbulatoryCare,
					}, nil)
			},
			wantErr:     true,
			expectedErr: ErrAmbulatoryHoursRequired,
		},
		{
			name:     "ambulatory_hours_not_allowed",
			clientID: "client-123",
			req: &MoveClientInCareRequest{
				CareStartDate:         "2023-01-01",
				CareEndDate:           "2023-12-31",
				AmbulatoryWeeklyHours: &hours,
			},
			setup: func(mockStore *dbmocks.MockStoreInterface) {
				mockStore.EXPECT().
					GetClientByID(gomock.Any(), "client-123").
					Return(db.Client{
						ID:       "client-123",
						Status:   db.ClientStatusEnumWaitingList,
						CareType: db.CareTypeEnumProtectedLiving,
					}, nil)
			},
			wantErr:     true,
			expectedErr: ErrAmbulatoryHoursNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := dbmocks.NewMockStoreInterface(ctrl)
			mockLogger := loggermocks.NewMockLogger(ctrl)

			mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
			mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			tt.setup(mockStore)

			service := NewClientService(mockStore, mockLogger)

			resp, err := service.MoveClientInCare(context.Background(), tt.clientID, tt.req)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
				return
			}

			require.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, resp)
			}
		})
	}
}

func TestStartDischarge(t *testing.T) {
	tests := []struct {
		name        string
		clientID    string
		req         *StartDischargeRequest
		setup       func(mockStore *dbmocks.MockStoreInterface)
		wantErr     bool
		expectedErr error
	}{
		{
			name:     "success",
			clientID: "client-123",
			req: &StartDischargeRequest{
				DischargeDate:      "2023-12-31",
				ReasonForDischarge: "completed_treatment",
			},
			setup: func(mockStore *dbmocks.MockStoreInterface) {
				mockStore.EXPECT().
					GetClientByID(gomock.Any(), "client-123").
					Return(db.Client{
						ID:     "client-123",
						Status: db.ClientStatusEnumInCare,
					}, nil)

				mockStore.EXPECT().
					UpdateClient(gomock.Any(), gomock.Any()).
					Return("client-123", nil)
			},
			wantErr: false,
		},
		{
			name:     "client_not_found",
			clientID: "notfound",
			req:      &StartDischargeRequest{},
			setup: func(mockStore *dbmocks.MockStoreInterface) {
				mockStore.EXPECT().
					GetClientByID(gomock.Any(), "notfound").
					Return(db.Client{}, pgx.ErrNoRows)
			},
			wantErr:     true,
			expectedErr: ErrClientNotFound,
		},
		{
			name:     "client_not_in_care",
			clientID: "client-123",
			req:      &StartDischargeRequest{},
			setup: func(mockStore *dbmocks.MockStoreInterface) {
				mockStore.EXPECT().
					GetClientByID(gomock.Any(), "client-123").
					Return(db.Client{
						ID:     "client-123",
						Status: db.ClientStatusEnumWaitingList,
					}, nil)
			},
			wantErr:     true,
			expectedErr: ErrClientNotInCare,
		},
		{
			name:     "discharge_already_started",
			clientID: "client-123",
			req:      &StartDischargeRequest{},
			setup: func(mockStore *dbmocks.MockStoreInterface) {
				mockStore.EXPECT().
					GetClientByID(gomock.Any(), "client-123").
					Return(db.Client{
						ID:     "client-123",
						Status: db.ClientStatusEnumInCare,
						DischargeStatus: db.NullDischargeStatusEnum{
							DischargeStatusEnum: db.DischargeStatusEnumInProgress,
							Valid:               true,
						},
					}, nil)
			},
			wantErr:     true,
			expectedErr: ErrDischargeAlreadyStarted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := dbmocks.NewMockStoreInterface(ctrl)
			mockLogger := loggermocks.NewMockLogger(ctrl)

			mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
			mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			tt.setup(mockStore)

			service := NewClientService(mockStore, mockLogger)

			resp, err := service.StartDischarge(context.Background(), tt.clientID, tt.req)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, resp)
		})
	}
}

func TestCompleteDischarge(t *testing.T) {
	tests := []struct {
		name        string
		clientID    string
		req         *CompleteDischargeRequest
		setup       func(mockStore *dbmocks.MockStoreInterface)
		wantErr     bool
		expectedErr error
	}{
		{
			name:     "success",
			clientID: "client-123",
			req: &CompleteDischargeRequest{
				ClosingReport:    "Report",
				EvaluationReport: "Evaluation",
			},
			setup: func(mockStore *dbmocks.MockStoreInterface) {
				mockStore.EXPECT().
					GetClientByID(gomock.Any(), "client-123").
					Return(db.Client{
						ID:     "client-123",
						Status: db.ClientStatusEnumInCare,
						DischargeStatus: db.NullDischargeStatusEnum{
							DischargeStatusEnum: db.DischargeStatusEnumInProgress,
							Valid:               true,
						},
					}, nil)

				mockStore.EXPECT().
					UpdateClient(gomock.Any(), gomock.Any()).
					Return("client-123", nil)
			},
			wantErr: false,
		},
		{
			name:     "client_not_found",
			clientID: "notfound",
			req:      &CompleteDischargeRequest{},
			setup: func(mockStore *dbmocks.MockStoreInterface) {
				mockStore.EXPECT().
					GetClientByID(gomock.Any(), "notfound").
					Return(db.Client{}, pgx.ErrNoRows)
			},
			wantErr:     true,
			expectedErr: ErrClientNotFound,
		},
		{
			name:     "client_not_in_care",
			clientID: "client-123",
			req:      &CompleteDischargeRequest{},
			setup: func(mockStore *dbmocks.MockStoreInterface) {
				mockStore.EXPECT().
					GetClientByID(gomock.Any(), "client-123").
					Return(db.Client{
						ID:     "client-123",
						Status: db.ClientStatusEnumWaitingList,
					}, nil)
			},
			wantErr:     true,
			expectedErr: ErrClientNotInCare,
		},
		{
			name:     "discharge_not_started",
			clientID: "client-123",
			req:      &CompleteDischargeRequest{},
			setup: func(mockStore *dbmocks.MockStoreInterface) {
				mockStore.EXPECT().
					GetClientByID(gomock.Any(), "client-123").
					Return(db.Client{
						ID:     "client-123",
						Status: db.ClientStatusEnumInCare,
						DischargeStatus: db.NullDischargeStatusEnum{
							Valid: false,
						},
					}, nil)
			},
			wantErr:     true,
			expectedErr: ErrDischargeNotStarted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := dbmocks.NewMockStoreInterface(ctrl)
			mockLogger := loggermocks.NewMockLogger(ctrl)

			mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
			mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			tt.setup(mockStore)

			service := NewClientService(mockStore, mockLogger)

			resp, err := service.CompleteDischarge(context.Background(), tt.clientID, tt.req)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, resp)
		})
	}
}

func TestListWaitingListClients(t *testing.T) {
	tests := []struct {
		name    string
		req     *ListWaitingListClientsRequest
		setup   func(mockStore *dbmocks.MockStoreInterface)
		wantErr bool
	}{
		{
			name: "success",
			req:  &ListWaitingListClientsRequest{},
			setup: func(mockStore *dbmocks.MockStoreInterface) {
				mockStore.EXPECT().
					ExecTx(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := dbmocks.NewMockStoreInterface(ctrl)
			mockLogger := loggermocks.NewMockLogger(ctrl)

			mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			tt.setup(mockStore)

			service := NewClientService(mockStore, mockLogger)

			// Add pagination params to context
			ctx := context.WithValue(context.Background(), "limit", int32(10))
			ctx = context.WithValue(ctx, "offset", int32(0))
			ctx = context.WithValue(ctx, "page", 1)
			ctx = context.WithValue(ctx, "pageSize", 10)

			_, err := service.ListWaitingListClients(ctx, tt.req)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestGetWaitlistStats(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(mockStore *dbmocks.MockStoreInterface)
		wantErr bool
	}{
		{
			name: "success",
			setup: func(mockStore *dbmocks.MockStoreInterface) {
				mockStore.EXPECT().
					ExecTx(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := dbmocks.NewMockStoreInterface(ctrl)
			mockLogger := loggermocks.NewMockLogger(ctrl)

			mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			tt.setup(mockStore)

			service := NewClientService(mockStore, mockLogger)

			_, err := service.GetWaitlistStats(context.Background())

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestListClientGoals(t *testing.T) {
	tests := []struct {
		name     string
		clientID string
		setup    func(mockStore *dbmocks.MockStoreInterface)
		wantErr  bool
	}{
		{
			name:     "success",
			clientID: "client-123",
			setup: func(mockStore *dbmocks.MockStoreInterface) {
				mockStore.EXPECT().
					ListGoalsByClientID(gomock.Any(), gomock.Any()).
					Return([]db.ClientGoal{}, nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := dbmocks.NewMockStoreInterface(ctrl)
			mockLogger := loggermocks.NewMockLogger(ctrl)

			mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			tt.setup(mockStore)

			service := NewClientService(mockStore, mockLogger)

			_, err := service.ListClientGoals(context.Background(), tt.clientID)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}
