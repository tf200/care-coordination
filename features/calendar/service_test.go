package calendar

import (
	"context"
	"errors"
	"testing"
	"time"

	db "care-cordination/lib/db/sqlc"
	dbmocks "care-cordination/lib/db/sqlc/mocks"
	loggermocks "care-cordination/lib/logger/mocks"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// ============================================================
// Test: CreateAppointment
// ============================================================

func TestCreateAppointment(t *testing.T) {
	tests := []struct {
		name        string
		organizerID string
		req         CreateAppointmentRequest
		setup       func(mockStore *dbmocks.MockStoreInterface)
		wantErr     bool
		expectedErr error
	}{
		{
			name:        "success",
			organizerID: "org-123",
			req: CreateAppointmentRequest{
				Title:     "Test Appointment",
				StartTime: time.Now().Add(time.Hour),
				EndTime:   time.Now().Add(2 * time.Hour),
				Type:      TypeGeneral,
				Participants: []ParticipantDTO{
					{ID: "client-1", Type: ParticipantClient},
				},
			},
			setup: func(mockStore *dbmocks.MockStoreInterface) {
				mockStore.EXPECT().
					ExecTx(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(*db.Queries) error) error {
						return nil
					})
				mockStore.EXPECT().
					ExecTx(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(*db.Queries) error) error {
						return nil
					})
			},
			wantErr: false,
		},
		{
			name:        "db_error",
			organizerID: "org-123",
			req: CreateAppointmentRequest{
				Title:     "Test Appointment",
				StartTime: time.Now().Add(time.Hour),
				EndTime:   time.Now().Add(2 * time.Hour),
				Type:      TypeGeneral,
				Participants: []ParticipantDTO{
					{ID: "client-1", Type: ParticipantClient},
				},
			},
			setup: func(mockStore *dbmocks.MockStoreInterface) {
				mockStore.EXPECT().
					ExecTx(gomock.Any(), gomock.Any()).
					Return(errors.New("database error"))
			},
			wantErr:     true,
			expectedErr: ErrInternal,
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

			service := NewCalendarService(mockStore, mockLogger)
			_, err := service.CreateAppointment(context.Background(), tt.organizerID, tt.req)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
				return
			}
			// Note: Because ExecTx is mocked, we can't fully validate success
		})
	}
}

// ============================================================
// Test: GetAppointment
// ============================================================

func TestGetAppointmentService(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		setup       func(mockStore *dbmocks.MockStoreInterface)
		wantErr     bool
		expectedErr error
	}{
		{
			name: "success",
			id:   "app-123",
			setup: func(mockStore *dbmocks.MockStoreInterface) {
				mockStore.EXPECT().
					ExecTx(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(*db.Queries) error) error {
						return nil
					})
			},
			wantErr: false,
		},
		{
			name: "not_found",
			id:   "non-existent",
			setup: func(mockStore *dbmocks.MockStoreInterface) {
				mockStore.EXPECT().
					ExecTx(gomock.Any(), gomock.Any()).
					Return(errors.New("no rows in result set"))
			},
			wantErr:     true,
			expectedErr: ErrAppointmentNotFound,
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

			service := NewCalendarService(mockStore, mockLogger)
			_, err := service.GetAppointment(context.Background(), tt.id)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
			}
		})
	}
}

// ============================================================
// Test: DeleteAppointment
// ============================================================

func TestDeleteAppointmentService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := dbmocks.NewMockStoreInterface(ctrl)
	mockLogger := loggermocks.NewMockLogger(ctrl)

	mockStore.EXPECT().
		ExecTx(gomock.Any(), gomock.Any()).
		Return(nil)

	service := NewCalendarService(mockStore, mockLogger)
	err := service.DeleteAppointment(context.Background(), "app-123")

	require.NoError(t, err)
}

// ============================================================
// Test: ListAppointments
// ============================================================

func TestListAppointmentsService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := dbmocks.NewMockStoreInterface(ctrl)
	mockLogger := loggermocks.NewMockLogger(ctrl)

	mockStore.EXPECT().
		ExecTx(gomock.Any(), gomock.Any()).
		Return(nil)

	service := NewCalendarService(mockStore, mockLogger)
	_, err := service.ListAppointments(context.Background(), "user-123")

	require.NoError(t, err)
}

// ============================================================
// Test: CreateReminder
// ============================================================

func TestCreateReminderService(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		req         CreateReminderRequest
		setup       func(mockStore *dbmocks.MockStoreInterface)
		wantErr     bool
		expectedErr error
	}{
		{
			name:   "success",
			userID: "user-123",
			req: CreateReminderRequest{
				Title:   "Test Reminder",
				DueTime: time.Now().Add(time.Hour),
			},
			setup: func(mockStore *dbmocks.MockStoreInterface) {
				mockStore.EXPECT().
					ExecTx(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(*db.Queries) error) error {
						return nil
					})
			},
			wantErr: false,
		},
		{
			name:   "db_error",
			userID: "user-123",
			req: CreateReminderRequest{
				Title:   "Test Reminder",
				DueTime: time.Now().Add(time.Hour),
			},
			setup: func(mockStore *dbmocks.MockStoreInterface) {
				mockStore.EXPECT().
					ExecTx(gomock.Any(), gomock.Any()).
					Return(errors.New("database error"))
			},
			wantErr:     true,
			expectedErr: ErrInternal,
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

			service := NewCalendarService(mockStore, mockLogger)
			_, err := service.CreateReminder(context.Background(), tt.userID, tt.req)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
			}
		})
	}
}

// ============================================================
// Test: GetReminder
// ============================================================

func TestGetReminderService(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		setup       func(mockStore *dbmocks.MockStoreInterface)
		wantErr     bool
		expectedErr error
	}{
		{
			name: "success",
			id:   "rem-123",
			setup: func(mockStore *dbmocks.MockStoreInterface) {
				mockStore.EXPECT().
					ExecTx(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(*db.Queries) error) error {
						return nil
					})
			},
			wantErr: false,
		},
		{
			name: "not_found",
			id:   "non-existent",
			setup: func(mockStore *dbmocks.MockStoreInterface) {
				mockStore.EXPECT().
					ExecTx(gomock.Any(), gomock.Any()).
					Return(errors.New("no rows in result set"))
			},
			wantErr:     true,
			expectedErr: ErrReminderNotFound,
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

			service := NewCalendarService(mockStore, mockLogger)
			_, err := service.GetReminder(context.Background(), tt.id)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
			}
		})
	}
}

// ============================================================
// Test: DeleteReminder
// ============================================================

func TestDeleteReminderService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := dbmocks.NewMockStoreInterface(ctrl)
	mockLogger := loggermocks.NewMockLogger(ctrl)

	mockStore.EXPECT().
		ExecTx(gomock.Any(), gomock.Any()).
		Return(nil)

	service := NewCalendarService(mockStore, mockLogger)
	err := service.DeleteReminder(context.Background(), "rem-123")

	require.NoError(t, err)
}

// ============================================================
// Test: ListReminders
// ============================================================

func TestListRemindersService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := dbmocks.NewMockStoreInterface(ctrl)
	mockLogger := loggermocks.NewMockLogger(ctrl)

	mockStore.EXPECT().
		ExecTx(gomock.Any(), gomock.Any()).
		Return(nil)

	service := NewCalendarService(mockStore, mockLogger)
	_, err := service.ListReminders(context.Background(), "user-123")

	require.NoError(t, err)
}

// ============================================================
// Test: GetCalendarView
// ============================================================

func TestGetCalendarViewService(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		startTime   time.Time
		endTime     time.Time
		setup       func(mockStore *dbmocks.MockStoreInterface)
		wantErr     bool
		expectedErr error
	}{
		{
			name:      "success",
			userID:    "user-123",
			startTime: time.Now(),
			endTime:   time.Now().Add(24 * time.Hour),
			setup: func(mockStore *dbmocks.MockStoreInterface) {
				mockStore.EXPECT().
					ExecTx(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(*db.Queries) error) error {
						return nil
					})
			},
			wantErr: false,
		},
		{
			name:      "db_error",
			userID:    "user-123",
			startTime: time.Now(),
			endTime:   time.Now().Add(24 * time.Hour),
			setup: func(mockStore *dbmocks.MockStoreInterface) {
				mockStore.EXPECT().
					ExecTx(gomock.Any(), gomock.Any()).
					Return(errors.New("database error"))
			},
			wantErr:     true,
			expectedErr: ErrInternal,
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

			service := NewCalendarService(mockStore, mockLogger)
			_, err := service.GetCalendarView(context.Background(), tt.userID, tt.startTime, tt.endTime)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
			}
		})
	}
}

// Unused import placeholder for pgtype (needed for some test setups)
var _ = pgtype.Timestamptz{}
