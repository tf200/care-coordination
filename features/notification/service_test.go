package notification

import (
	"context"
	"testing"
	"time"

	db "care-cordination/lib/db/sqlc"
	dbmocks "care-cordination/lib/db/sqlc/mocks"
	loggermocks "care-cordination/lib/logger/mocks"
	"care-cordination/lib/websocket"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// ============================================================
// Test Helpers
// ============================================================

func setupTestService(t *testing.T) (*notificationService, *dbmocks.MockStoreInterface, *loggermocks.MockLogger, *websocket.Hub, *gomock.Controller) {
	ctrl := gomock.NewController(t)

	mockStore := dbmocks.NewMockStoreInterface(ctrl)
	mockLogger := loggermocks.NewMockLogger(ctrl)

	// Allow all log calls
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	// Create a real hub for testing WebSocket delivery
	hub := websocket.NewHub(mockLogger)
	go hub.Run()

	service := NewNotificationService(mockStore, hub, mockLogger).(*notificationService)

	return service, mockStore, mockLogger, hub, ctrl
}

// ============================================================
// Test: Create (synchronous)
// ============================================================

func TestCreate(t *testing.T) {
	tests := []struct {
		name     string
		req      *CreateNotificationRequest
		setup    func(mockStore *dbmocks.MockStoreInterface)
		wantErr  bool
		validate func(t *testing.T, resp *NotificationResponse)
	}{
		{
			name: "success",
			req: &CreateNotificationRequest{
				UserID:   "user-123",
				Type:     TypeIncidentCreated,
				Priority: PriorityHigh,
				Title:    "Test Notification",
				Message:  "This is a test",
			},
			setup: func(mockStore *dbmocks.MockStoreInterface) {
				mockStore.EXPECT().
					CreateNotification(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, params db.CreateNotificationParams) (db.Notification, error) {
						return db.Notification{
							ID:        params.ID,
							UserID:    params.UserID,
							Type:      params.Type,
							Priority:  params.Priority,
							Title:     params.Title,
							Message:   params.Message,
							CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
						}, nil
					})
			},
			wantErr: false,
			validate: func(t *testing.T, resp *NotificationResponse) {
				assert.NotEmpty(t, resp.ID)
				assert.Equal(t, TypeIncidentCreated, resp.Type)
				assert.Equal(t, PriorityHigh, resp.Priority)
				assert.Equal(t, "Test Notification", resp.Title)
				assert.False(t, resp.IsRead)
			},
		},
		{
			name: "default_priority",
			req: &CreateNotificationRequest{
				UserID:  "user-123",
				Type:    TypeEvaluationDue,
				Title:   "Evaluation Due",
				Message: "Check your evaluation",
				// No priority specified - should default to normal
			},
			setup: func(mockStore *dbmocks.MockStoreInterface) {
				mockStore.EXPECT().
					CreateNotification(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, params db.CreateNotificationParams) (db.Notification, error) {
						// Verify default priority was applied
						assert.Equal(t, db.NotificationPriorityEnum(PriorityNormal), params.Priority)
						return db.Notification{
							ID:        params.ID,
							UserID:    params.UserID,
							Type:      params.Type,
							Priority:  params.Priority,
							Title:     params.Title,
							Message:   params.Message,
							CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
						}, nil
					})
			},
			wantErr: false,
			validate: func(t *testing.T, resp *NotificationResponse) {
				assert.Equal(t, PriorityNormal, resp.Priority)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockStore, _, hub, ctrl := setupTestService(t)
			defer ctrl.Finish()
			defer hub.Stop()

			tt.setup(mockStore)

			resp, err := service.Create(context.Background(), tt.req)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, resp)
			}
		})
	}
}

// ============================================================
// Test: Enqueue (async)
// ============================================================

func TestEnqueue(t *testing.T) {
	service, mockStore, _, hub, ctrl := setupTestService(t)
	defer ctrl.Finish()
	defer hub.Stop()

	// Set up expectation for the worker to process
	created := make(chan bool, 1)
	mockStore.EXPECT().
		CreateNotification(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, params db.CreateNotificationParams) (db.Notification, error) {
			created <- true
			return db.Notification{
				ID:        params.ID,
				UserID:    params.UserID,
				Type:      params.Type,
				Priority:  params.Priority,
				Title:     params.Title,
				Message:   params.Message,
				CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}, nil
		})

	// Enqueue should return immediately (non-blocking)
	start := time.Now()
	service.Enqueue(&CreateNotificationRequest{
		UserID:  "user-123",
		Type:    TypeIncidentCreated,
		Title:   "Test",
		Message: "Test message",
	})
	elapsed := time.Since(start)

	// Should be nearly instant (non-blocking)
	assert.Less(t, elapsed, 10*time.Millisecond)

	// Wait for worker to process
	select {
	case <-created:
		// Success - worker processed the notification
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for notification to be created")
	}
}

// ============================================================
// Test: EnqueueForRole
// ============================================================

func TestEnqueueForRole(t *testing.T) {
	service, mockStore, _, hub, ctrl := setupTestService(t)
	defer ctrl.Finish()
	defer hub.Stop()

	// Setup: return 3 admin users
	mockStore.EXPECT().
		GetUserIDsByRoleName(gomock.Any(), "admin").
		Return([]string{"admin-1", "admin-2", "admin-3"}, nil)

	// Expect 3 notifications to be created
	createdCount := 0
	mockStore.EXPECT().
		CreateNotification(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, params db.CreateNotificationParams) (db.Notification, error) {
			createdCount++
			return db.Notification{
				ID:        params.ID,
				UserID:    params.UserID,
				Type:      params.Type,
				Priority:  params.Priority,
				Title:     params.Title,
				Message:   params.Message,
				CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}, nil
		}).Times(3)

	service.EnqueueForRole(context.Background(), "admin", &CreateNotificationRequest{
		Type:    TypeIncidentCreated,
		Title:   "New Incident",
		Message: "An incident was reported",
	})

	// Wait for workers to process
	time.Sleep(200 * time.Millisecond)

	assert.Equal(t, 3, createdCount)
}

// ============================================================
// Test: EnqueueForUsers
// ============================================================

func TestEnqueueForUsers(t *testing.T) {
	service, mockStore, _, hub, ctrl := setupTestService(t)
	defer ctrl.Finish()
	defer hub.Stop()

	userIDs := []string{"user-1", "user-2"}

	// Expect 2 notifications to be created
	createdCount := 0
	mockStore.EXPECT().
		CreateNotification(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, params db.CreateNotificationParams) (db.Notification, error) {
			createdCount++
			return db.Notification{
				ID:        params.ID,
				UserID:    params.UserID,
				Type:      params.Type,
				Priority:  params.Priority,
				Title:     params.Title,
				Message:   params.Message,
				CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}, nil
		}).Times(2)

	service.EnqueueForUsers(userIDs, &CreateNotificationRequest{
		Type:    TypeLocationTransferApproved,
		Title:   "Transfer Approved",
		Message: "Your transfer was approved",
	})

	// Wait for workers to process
	time.Sleep(200 * time.Millisecond)

	assert.Equal(t, 2, createdCount)
}

// ============================================================
// Test: WebSocket Delivery
// ============================================================

func TestWebSocketDelivery(t *testing.T) {
	service, mockStore, _, hub, ctrl := setupTestService(t)
	defer ctrl.Finish()
	defer hub.Stop()

	// Create a mock client
	client := &websocket.Client{
		UserID: "user-123",
	}
	// We need to set up the client's send channel
	client.SetSendChannel(make(chan *websocket.Message, 256))

	hub.Register(client)
	time.Sleep(50 * time.Millisecond)

	// Setup mock
	mockStore.EXPECT().
		CreateNotification(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, params db.CreateNotificationParams) (db.Notification, error) {
			return db.Notification{
				ID:        params.ID,
				UserID:    params.UserID,
				Type:      params.Type,
				Priority:  params.Priority,
				Title:     params.Title,
				Message:   params.Message,
				CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}, nil
		})

	// Create notification
	_, err := service.Create(context.Background(), &CreateNotificationRequest{
		UserID:  "user-123",
		Type:    TypeIncidentCreated,
		Title:   "Real-time Test",
		Message: "Should arrive via WebSocket",
	})
	require.NoError(t, err)

	// Check WebSocket received the message
	select {
	case msg := <-client.GetSendChannel():
		assert.Equal(t, websocket.MessageTypeNotification, msg.Type)
		payload, ok := msg.Payload.(websocket.NotificationPayload)
		require.True(t, ok)
		assert.Equal(t, "Real-time Test", payload.Title)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for WebSocket message")
	}
}

// ============================================================
// Test: List
// ============================================================

func TestList(t *testing.T) {
	service, mockStore, _, hub, ctrl := setupTestService(t)
	defer ctrl.Finish()
	defer hub.Stop()

	// Set up context with user ID and pagination
	ctx := context.WithValue(context.Background(), "user_id", "user-123")
	ctx = context.WithValue(ctx, "limit", int32(10))
	ctx = context.WithValue(ctx, "offset", int32(0))
	ctx = context.WithValue(ctx, "page", int32(1))
	ctx = context.WithValue(ctx, "pageSize", int32(10))

	isRead := false
	mockStore.EXPECT().
		ListNotifications(gomock.Any(), gomock.Any()).
		Return([]db.ListNotificationsRow{
			{
				ID:         "notif-1",
				Type:       db.NotificationTypeEnumIncidentCreated,
				Priority:   db.NotificationPriorityEnumHigh,
				Title:      "Notification 1",
				Message:    "Message 1",
				IsRead:     &isRead,
				TotalCount: 2,
				CreatedAt:  pgtype.Timestamptz{Time: time.Now(), Valid: true},
			},
			{
				ID:         "notif-2",
				Type:       db.NotificationTypeEnumEvaluationDue,
				Priority:   db.NotificationPriorityEnumNormal,
				Title:      "Notification 2",
				Message:    "Message 2",
				IsRead:     &isRead,
				TotalCount: 2,
				CreatedAt:  pgtype.Timestamptz{Time: time.Now(), Valid: true},
			},
		}, nil)

	result, err := service.List(ctx, &ListNotificationsRequest{})
	require.NoError(t, err)

	assert.Len(t, result.Data, 2)
	assert.Equal(t, 2, result.TotalCount)
	assert.Equal(t, "Notification 1", result.Data[0].Title)
	assert.Equal(t, "Notification 2", result.Data[1].Title)
}

// ============================================================
// Test: MarkAsRead
// ============================================================

func TestMarkAsRead(t *testing.T) {
	service, mockStore, _, hub, ctrl := setupTestService(t)
	defer ctrl.Finish()
	defer hub.Stop()

	ctx := context.WithValue(context.Background(), "user_id", "user-123")

	mockStore.EXPECT().
		MarkNotificationAsRead(gomock.Any(), db.MarkNotificationAsReadParams{
			ID:     "notif-123",
			UserID: "user-123",
		}).
		Return(nil)

	err := service.MarkAsRead(ctx, "notif-123")
	require.NoError(t, err)
}

// ============================================================
// Test: GetUnreadCount
// ============================================================

func TestGetUnreadCount(t *testing.T) {
	service, mockStore, _, hub, ctrl := setupTestService(t)
	defer ctrl.Finish()
	defer hub.Stop()

	ctx := context.WithValue(context.Background(), "user_id", "user-123")

	mockStore.EXPECT().
		GetUnreadCount(gomock.Any(), "user-123").
		Return(int64(5), nil)

	count, err := service.GetUnreadCount(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(5), count)
}
