package employee_test

import (
	"context"
	"testing"
	"time"

	"care-cordination/features/employee"
	db "care-cordination/lib/db/sqlc"
	dbmocks "care-cordination/lib/db/sqlc/mocks"
	loggermocks "care-cordination/lib/logger/mocks"
	"care-cordination/lib/util"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCreateEmployee(t *testing.T) {
	hours := int32(40)
	tests := []struct {
		name    string
		req     *employee.CreateEmployeeRequest
		setup   func(mockStore *dbmocks.MockStoreInterface)
		wantErr bool
	}{
		{
			name: "success",
			req: &employee.CreateEmployeeRequest{
				Email:         "test@example.com",
				Password:      "password123",
				FirstName:     "John",
				LastName:      "Doe",
				BSN:           "123456789",
				DateOfBirth:   "1990-01-01",
				PhoneNumber:   "0612345678",
				Gender:        "male",
				LocationID:    "loc-123",
				ContractHours: &hours,
			},
			setup: func(mockStore *dbmocks.MockStoreInterface) {
				mockStore.EXPECT().
					CreateEmployeeTx(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "db_error",
			req: &employee.CreateEmployeeRequest{
				Email:       "test@example.com",
				Password:    "password123",
				FirstName:   "John",
				LastName:    "Doe",
				BSN:         "123456789",
				DateOfBirth: "1990-01-01",
				PhoneNumber: "0612345678",
				Gender:      "male",
				LocationID:  "loc-123",
			},
			setup: func(mockStore *dbmocks.MockStoreInterface) {
				mockStore.EXPECT().
					CreateEmployeeTx(gomock.Any(), gomock.Any()).
					Return(assert.AnError)
			},
			wantErr: true,
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

			service := employee.NewEmployeeService(mockStore, mockLogger)

			resp, err := service.CreateEmployee(context.Background(), tt.req)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, resp.ID)
		})
	}
}

func TestListEmployees(t *testing.T) {
	tests := []struct {
		name    string
		req     *employee.ListEmployeesRequest
		setup   func(mockStore *dbmocks.MockStoreInterface)
		wantErr bool
	}{
		{
			name: "success",
			req:  &employee.ListEmployeesRequest{},
			setup: func(mockStore *dbmocks.MockStoreInterface) {
				mockStore.EXPECT().
					ListEmployees(gomock.Any(), gomock.Any()).
					Return([]db.ListEmployeesRow{
						{
							ID:          "emp-123",
							FirstName:   "John",
							LastName:    "Doe",
							Email:       "john@example.com",
							TotalCount:  1,
							ClientCount: int64(0),
							DateOfBirth: pgtype.Date{Time: time.Now(), Valid: true},
						},
					}, nil)
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

			service := employee.NewEmployeeService(mockStore, mockLogger)

			// Add pagination params to context
			ctx := context.WithValue(context.Background(), "limit", int32(10))
			ctx = context.WithValue(ctx, "offset", int32(0))
			ctx = context.WithValue(ctx, "page", 1)
			ctx = context.WithValue(ctx, "pageSize", 10)

			resp, err := service.ListEmployees(ctx, tt.req)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, resp.Data, 1)
		})
	}
}

func TestGetEmployeeByID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		setup   func(mockStore *dbmocks.MockStoreInterface)
		wantErr bool
	}{
		{
			name: "success",
			id:   "emp-123",
			setup: func(mockStore *dbmocks.MockStoreInterface) {
				mockStore.EXPECT().
					GetEmployeeByID(gomock.Any(), "emp-123").
					Return(db.GetEmployeeByIDRow{
						ID:          "emp-123",
						FirstName:   "John",
						LastName:    "Doe",
						ClientCount: int64(0),
						DateOfBirth: pgtype.Date{Time: time.Now(), Valid: true},
					}, nil)
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

			service := employee.NewEmployeeService(mockStore, mockLogger)

			resp, err := service.GetEmployeeByID(context.Background(), tt.id)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.id, resp.ID)
		})
	}
}

func TestGetMyProfile(t *testing.T) {
	tests := []struct {
		name    string
		userID  string
		setup   func(mockStore *dbmocks.MockStoreInterface)
		wantErr bool
	}{
		{
			name:   "success",
			userID: "user-123",
			setup: func(mockStore *dbmocks.MockStoreInterface) {
				roleID := "role-123"
				mockStore.EXPECT().
					GetEmployeeByUserID(gomock.Any(), "user-123").
					Return(db.GetEmployeeByUserIDRow{
						ID:          "emp-123",
						UserID:      "user-123",
						RoleID:      &roleID,
						DateOfBirth: pgtype.Date{Time: time.Now(), Valid: true},
					}, nil)

				mockStore.EXPECT().
					ListPermissionsForRole(gomock.Any(), gomock.Any()).
					Return([]db.Permission{}, nil)
			},
			wantErr: false,
		},
		{
			name:    "unauthorized",
			userID:  "",
			setup:   func(mockStore *dbmocks.MockStoreInterface) {},
			wantErr: true,
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

			service := employee.NewEmployeeService(mockStore, mockLogger)

			ctx := context.Background()
			if tt.userID != "" {
				ctx = context.WithValue(ctx, "user_id", tt.userID)
			}

			resp, err := service.GetMyProfile(ctx)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.userID, resp.UserID)
		})
	}
}

func TestUpdateEmployee(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		req     *employee.UpdateEmployeeRequest
		setup   func(mockStore *dbmocks.MockStoreInterface)
		wantErr bool
	}{
		{
			name: "success",
			id:   "emp-123",
			req: &employee.UpdateEmployeeRequest{
				FirstName: util.StrPtr("Jane"),
			},
			setup: func(mockStore *dbmocks.MockStoreInterface) {
				mockStore.EXPECT().
					GetEmployeeByID(gomock.Any(), "emp-123").
					Return(db.GetEmployeeByIDRow{UserID: "user-123"}, nil)

				mockStore.EXPECT().
					UpdateEmployee(gomock.Any(), gomock.Any()).
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

			service := employee.NewEmployeeService(mockStore, mockLogger)

			resp, err := service.UpdateEmployee(context.Background(), tt.id, tt.req)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.id, resp.ID)
		})
	}
}

func TestDeleteEmployee(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		setup   func(mockStore *dbmocks.MockStoreInterface)
		wantErr bool
	}{
		{
			name: "success",
			id:   "emp-123",
			setup: func(mockStore *dbmocks.MockStoreInterface) {
				mockStore.EXPECT().
					SoftDeleteEmployee(gomock.Any(), "emp-123").
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

			service := employee.NewEmployeeService(mockStore, mockLogger)

			err := service.DeleteEmployee(context.Background(), tt.id)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}
