package db

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================
// Test: CreateAppointment
// ============================================================

func TestCreateAppointment(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) CreateAppointmentParams
		wantErr  bool
		checkErr func(t *testing.T, err error)
	}{
		{
			name: "success",
			setup: func(t *testing.T, q *Queries) CreateAppointmentParams {
				userID := CreateTestUser(t, q, CreateTestUserOptions{})
				employeeID := CreateTestEmployee(t, q, CreateTestEmployeeOptions{UserID: userID})
				return CreateAppointmentParams{
					ID:          generateTestID(),
					Title:       "Success Appointment",
					Description: strPtr("Some description"),
					StartTime:   pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true},
					EndTime:     pgtype.Timestamptz{Time: time.Now().Add(2 * time.Hour), Valid: true},
					OrganizerID: employeeID,
					Status:      NullAppointmentStatusEnum{AppointmentStatusEnum: AppointmentStatusEnumConfirmed, Valid: true},
					Type:        AppointmentTypeEnumGeneral,
				}
			},
			wantErr: false,
		},
		{
			name: "invalid_organizer",
			setup: func(t *testing.T, q *Queries) CreateAppointmentParams {
				return CreateAppointmentParams{
					ID:          generateTestID(),
					Title:       "Invalid Organizer",
					StartTime:   pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true},
					EndTime:     pgtype.Timestamptz{Time: time.Now().Add(2 * time.Hour), Valid: true},
					OrganizerID: "non-existent-employee",
					Type:        AppointmentTypeEnumGeneral,
				}
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.True(t, IsForeignKeyViolation(err))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				params := tt.setup(t, q)

				result, err := q.CreateAppointment(ctx, params)

				if tt.wantErr {
					require.Error(t, err)
					if tt.checkErr != nil {
						tt.checkErr(t, err)
					}
					return
				}

				require.NoError(t, err)
				assert.Equal(t, params.ID, result.ID)
				assert.Equal(t, params.Title, result.Title)
			})
		})
	}
}

// ============================================================
// Test: GetAppointment
// ============================================================

func TestGetAppointment(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) string // returns ID to query
		wantErr  bool
		checkErr func(t *testing.T, err error)
		validate func(t *testing.T, result Appointment)
	}{
		{
			name: "found",
			setup: func(t *testing.T, q *Queries) string {
				userID := CreateTestUser(t, q, CreateTestUserOptions{})
				employeeID := CreateTestEmployee(t, q, CreateTestEmployeeOptions{UserID: userID})
				return CreateTestAppointment(t, q, CreateTestAppointmentOptions{OrganizerID: employeeID})
			},
			wantErr: false,
			validate: func(t *testing.T, result Appointment) {
				assert.NotEmpty(t, result.ID)
				assert.NotEmpty(t, result.Title)
			},
		},
		{
			name: "not_found",
			setup: func(t *testing.T, q *Queries) string {
				return "non-existent-id"
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.True(t, errors.Is(err, pgx.ErrNoRows))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				id := tt.setup(t, q)

				result, err := q.GetAppointment(ctx, id)

				if tt.wantErr {
					require.Error(t, err)
					if tt.checkErr != nil {
						tt.checkErr(t, err)
					}
					return
				}

				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			})
		})
	}
}

// ============================================================
// Test: UpdateAppointment
// ============================================================

func TestUpdateAppointment(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) UpdateAppointmentParams
		wantErr  bool
		validate func(t *testing.T, q *Queries, res Appointment)
	}{
		{
			name: "full_update",
			setup: func(t *testing.T, q *Queries) UpdateAppointmentParams {
				userID := CreateTestUser(t, q, CreateTestUserOptions{})
				employeeID := CreateTestEmployee(t, q, CreateTestEmployeeOptions{UserID: userID})
				id := CreateTestAppointment(t, q, CreateTestAppointmentOptions{OrganizerID: employeeID})
				return UpdateAppointmentParams{
					ID:    id,
					Title: "Updated Title",
					Status: NullAppointmentStatusEnum{
						AppointmentStatusEnum: AppointmentStatusEnumCancelled,
						Valid:                 true,
					},
					Type: NullAppointmentTypeEnum{
						AppointmentTypeEnum: AppointmentTypeEnumAmbulatory,
						Valid:               true,
					},
				}
			},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, res Appointment) {
				assert.Equal(t, "Updated Title", res.Title)
				assert.Equal(t, AppointmentStatusEnumCancelled, res.Status.AppointmentStatusEnum)
				assert.Equal(t, AppointmentTypeEnumAmbulatory, res.Type)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				params := tt.setup(t, q)

				res, err := q.UpdateAppointment(ctx, params)

				if tt.wantErr {
					require.Error(t, err)
					return
				}

				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, q, res)
				}
			})
		})
	}
}

// ============================================================
// Test: DeleteAppointment
// ============================================================

func TestDeleteAppointment(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, q *Queries) string // returns ID to delete
	}{
		{
			name: "existing",
			setup: func(t *testing.T, q *Queries) string {
				userID := CreateTestUser(t, q, CreateTestUserOptions{})
				employeeID := CreateTestEmployee(t, q, CreateTestEmployeeOptions{UserID: userID})
				return CreateTestAppointment(t, q, CreateTestAppointmentOptions{OrganizerID: employeeID})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				id := tt.setup(t, q)

				err := q.DeleteAppointment(ctx, id)
				require.NoError(t, err)

				_, err = q.GetAppointment(ctx, id)
				assert.ErrorIs(t, err, pgx.ErrNoRows)
			})
		})
	}
}

// ============================================================
// Test: ListAppointmentsByOrganizer
// ============================================================

func TestListAppointmentsByOrganizer(t *testing.T) {
	runTestWithTx(t, func(t *testing.T, q *Queries) {
		ctx := context.Background()
		userID := CreateTestUser(t, q, CreateTestUserOptions{})
		employeeID := CreateTestEmployee(t, q, CreateTestEmployeeOptions{UserID: userID})

		// Create 3 appointments
		for i := 0; i < 3; i++ {
			CreateTestAppointment(t, q, CreateTestAppointmentOptions{OrganizerID: employeeID})
		}

		results, err := q.ListAppointmentsByOrganizer(ctx, employeeID)
		require.NoError(t, err)
		assert.Len(t, results, 3)
	})
}

// ============================================================
// Test: ListAppointmentsByParticipant
// ============================================================

func TestListAppointmentsByParticipant(t *testing.T) {
	runTestWithTx(t, func(t *testing.T, q *Queries) {
		ctx := context.Background()
		userID := CreateTestUser(t, q, CreateTestUserOptions{})
		employeeID := CreateTestEmployee(t, q, CreateTestEmployeeOptions{UserID: userID})

		clientID, _ := CreateTestClientWithDependencies(t, q)

		appID := CreateTestAppointment(t, q, CreateTestAppointmentOptions{OrganizerID: employeeID})
		CreateTestAppointmentParticipant(t, q, appID, clientID, ParticipantTypeEnumClient)

		results, err := q.ListAppointmentsByParticipant(ctx, ListAppointmentsByParticipantParams{
			ParticipantID:   clientID,
			ParticipantType: ParticipantTypeEnumClient,
		})
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, appID, results[0].ID)
	})
}

// ============================================================
// Test: ListAppointmentsByRange
// ============================================================

func TestListAppointmentsByRange(t *testing.T) {
	runTestWithTx(t, func(t *testing.T, q *Queries) {
		ctx := context.Background()
		userID := CreateTestUser(t, q, CreateTestUserOptions{})
		employeeID := CreateTestEmployee(t, q, CreateTestEmployeeOptions{UserID: userID})

		now := time.Now().Truncate(time.Second)

		// In range
		app1 := CreateTestAppointment(t, q, CreateTestAppointmentOptions{
			OrganizerID: employeeID,
			StartTime:   strPtrTime(now.Add(time.Hour)),
		})
		// Out of range (too early)
		CreateTestAppointment(t, q, CreateTestAppointmentOptions{
			OrganizerID: employeeID,
			StartTime:   strPtrTime(now.Add(-2 * time.Hour)),
		})
		// Out of range (too late)
		CreateTestAppointment(t, q, CreateTestAppointmentOptions{
			OrganizerID: employeeID,
			StartTime:   strPtrTime(now.Add(10 * time.Hour)),
		})

		results, err := q.ListAppointmentsByRange(ctx, ListAppointmentsByRangeParams{
			OrganizerID: employeeID,
			StartTime:   pgtype.Timestamptz{Time: now, Valid: true},
			EndTime:     pgtype.Timestamptz{Time: now.Add(5 * time.Hour), Valid: true},
		})
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, app1, results[0].ID)
	})
}

// ============================================================
// Test: Appointment Participants
// ============================================================

func TestAppointmentParticipants(t *testing.T) {
	runTestWithTx(t, func(t *testing.T, q *Queries) {
		ctx := context.Background()
		userID := CreateTestUser(t, q, CreateTestUserOptions{})
		employeeID := CreateTestEmployee(t, q, CreateTestEmployeeOptions{UserID: userID})
		appID := CreateTestAppointment(t, q, CreateTestAppointmentOptions{OrganizerID: employeeID})

		clientID1, _ := CreateTestClientWithDependencies(t, q)
		clientID2, _ := CreateTestClientWithDependencies(t, q)

		// Test: AddAppointmentParticipant
		err := q.AddAppointmentParticipant(ctx, AddAppointmentParticipantParams{
			AppointmentID:   appID,
			ParticipantID:   clientID1,
			ParticipantType: ParticipantTypeEnumClient,
		})
		require.NoError(t, err)

		err = q.AddAppointmentParticipant(ctx, AddAppointmentParticipantParams{
			AppointmentID:   appID,
			ParticipantID:   clientID2,
			ParticipantType: ParticipantTypeEnumClient,
		})
		require.NoError(t, err)

		// Test: ListAppointmentParticipants
		participants, err := q.ListAppointmentParticipants(ctx, appID)
		require.NoError(t, err)
		assert.Len(t, participants, 2)

		// Test: RemoveAppointmentParticipants
		err = q.RemoveAppointmentParticipants(ctx, appID)
		require.NoError(t, err)

		participants, err = q.ListAppointmentParticipants(ctx, appID)
		require.NoError(t, err)
		assert.Len(t, participants, 0)
	})
}

// ============================================================
// Test: CreateReminder
// ============================================================

func TestCreateReminder(t *testing.T) {
	runTestWithTx(t, func(t *testing.T, q *Queries) {
		ctx := context.Background()
		userID := CreateTestUser(t, q, CreateTestUserOptions{})
		employeeID := CreateTestEmployee(t, q, CreateTestEmployeeOptions{UserID: userID})

		params := CreateReminderParams{
			ID:          generateTestID(),
			UserID:      employeeID,
			Title:       "Test Reminder",
			Description: strPtr("Reminder description"),
			DueTime:     pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true},
		}

		result, err := q.CreateReminder(ctx, params)
		require.NoError(t, err)
		assert.Equal(t, params.ID, result.ID)
		assert.Equal(t, params.Title, result.Title)
	})
}

// ============================================================
// Test: GetReminder
// ============================================================

func TestGetReminder(t *testing.T) {
	runTestWithTx(t, func(t *testing.T, q *Queries) {
		ctx := context.Background()
		userID := CreateTestUser(t, q, CreateTestUserOptions{})
		employeeID := CreateTestEmployee(t, q, CreateTestEmployeeOptions{UserID: userID})
		id := CreateTestReminder(t, q, CreateTestReminderOptions{UserID: employeeID})

		result, err := q.GetReminder(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, id, result.ID)
	})
}

// ============================================================
// Test: UpdateReminder
// ============================================================

func TestUpdateReminder(t *testing.T) {
	runTestWithTx(t, func(t *testing.T, q *Queries) {
		ctx := context.Background()
		userID := CreateTestUser(t, q, CreateTestUserOptions{})
		employeeID := CreateTestEmployee(t, q, CreateTestEmployeeOptions{UserID: userID})
		id := CreateTestReminder(t, q, CreateTestReminderOptions{UserID: employeeID})

		isComp := true
		params := UpdateReminderParams{
			ID:          id,
			Title:       "Updated Reminder",
			IsCompleted: &isComp,
		}

		result, err := q.UpdateReminder(ctx, params)
		require.NoError(t, err)
		assert.Equal(t, "Updated Reminder", result.Title)
		assert.True(t, *result.IsCompleted)
	})
}

// ============================================================
// Test: DeleteReminder
// ============================================================

func TestDeleteReminder(t *testing.T) {
	runTestWithTx(t, func(t *testing.T, q *Queries) {
		ctx := context.Background()
		userID := CreateTestUser(t, q, CreateTestUserOptions{})
		employeeID := CreateTestEmployee(t, q, CreateTestEmployeeOptions{UserID: userID})
		id := CreateTestReminder(t, q, CreateTestReminderOptions{UserID: employeeID})

		err := q.DeleteReminder(ctx, id)
		require.NoError(t, err)

		_, err = q.GetReminder(ctx, id)
		assert.ErrorIs(t, err, pgx.ErrNoRows)
	})
}

// ============================================================
// Test: ListRemindersByUser
// ============================================================

func TestListRemindersByUser(t *testing.T) {
	runTestWithTx(t, func(t *testing.T, q *Queries) {
		ctx := context.Background()
		userID := CreateTestUser(t, q, CreateTestUserOptions{})
		employeeID := CreateTestEmployee(t, q, CreateTestEmployeeOptions{UserID: userID})

		for i := 0; i < 3; i++ {
			CreateTestReminder(t, q, CreateTestReminderOptions{UserID: employeeID})
		}

		results, err := q.ListRemindersByUser(ctx, employeeID)
		require.NoError(t, err)
		assert.Len(t, results, 3)
	})
}

// ============================================================
// Test: ListRemindersByRange
// ============================================================

func TestListRemindersByRange(t *testing.T) {
	runTestWithTx(t, func(t *testing.T, q *Queries) {
		ctx := context.Background()
		userID := CreateTestUser(t, q, CreateTestUserOptions{})
		employeeID := CreateTestEmployee(t, q, CreateTestEmployeeOptions{UserID: userID})

		now := time.Now().Truncate(time.Second)

		// In range
		rem1 := CreateTestReminder(t, q, CreateTestReminderOptions{
			UserID:  employeeID,
			DueTime: strPtrTime(now.Add(time.Hour)),
		})
		// Out of range
		CreateTestReminder(t, q, CreateTestReminderOptions{
			UserID:  employeeID,
			DueTime: strPtrTime(now.Add(10 * time.Hour)),
		})

		results, err := q.ListRemindersByRange(ctx, ListRemindersByRangeParams{
			UserID:    employeeID,
			StartTime: pgtype.Timestamptz{Time: now, Valid: true},
			EndTime:   pgtype.Timestamptz{Time: now.Add(5 * time.Hour), Valid: true},
		})
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, rem1, results[0].ID)
	})
}

// Helper
func strPtrTime(t time.Time) *time.Time {
	return &t
}
