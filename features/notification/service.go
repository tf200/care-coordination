package notification

import (
	"care-cordination/lib/middleware"
	db "care-cordination/lib/db/sqlc"
	"care-cordination/lib/logger"
	"care-cordination/lib/nanoid"
	"care-cordination/lib/resp"
	"care-cordination/lib/util"
	"care-cordination/lib/websocket"
	"context"
	"sync"

	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

const (
	// Queue capacity for async notifications
	defaultQueueCapacity = 1000
	// Number of worker goroutines
	defaultWorkerCount = 3
)

type notificationService struct {
	store  db.StoreInterface
	hub    *websocket.Hub
	logger logger.Logger

	// Async queue
	queue      chan *CreateNotificationRequest
	workerWg   sync.WaitGroup
	workerDone chan struct{}
}

// NewNotificationService creates a new notification service
func NewNotificationService(
	store db.StoreInterface,
	hub *websocket.Hub,
	logger logger.Logger,
) NotificationService {
	s := &notificationService{
		store:      store,
		hub:        hub,
		logger:     logger,
		queue:      make(chan *CreateNotificationRequest, defaultQueueCapacity),
		workerDone: make(chan struct{}),
	}

	// Start background workers
	s.startWorkers(defaultWorkerCount)

	return s
}

// startWorkers starts N worker goroutines to process the notification queue
func (s *notificationService) startWorkers(count int) {
	for i := 0; i < count; i++ {
		s.workerWg.Add(1)
		go s.worker(i)
	}
	s.logger.Info(context.Background(), "NotificationService", "Workers started",
		zap.Int("count", count),
		zap.Int("queueCapacity", defaultQueueCapacity),
	)
}

// worker processes notifications from the queue
func (s *notificationService) worker(id int) {
	defer s.workerWg.Done()
	ctx := context.Background()

	for {
		select {
		case req, ok := <-s.queue:
			if !ok {
				// Channel closed, exit
				return
			}
			// Process the notification
			_, err := s.createInternal(ctx, req)
			if err != nil {
				s.logger.Error(ctx, "NotificationWorker", "Failed to create notification",
					zap.Int("workerID", id),
					zap.Error(err),
				)
			}
		case <-s.workerDone:
			return
		}
	}
}

// Enqueue adds a notification request to the async queue (non-blocking)
// This is the preferred method for service triggers
func (s *notificationService) Enqueue(req *CreateNotificationRequest) {
	select {
	case s.queue <- req:
		// Successfully queued
	default:
		// Queue is full, log warning and drop
		s.logger.Warn(context.Background(), "NotificationService", "Queue full, notification dropped",
			zap.String("userID", req.UserID),
			zap.String("type", req.Type),
		)
	}
}

// EnqueueForRole creates notifications for all users with the specified role (async)
func (s *notificationService) EnqueueForRole(ctx context.Context, roleName string, req *CreateNotificationRequest) {
	// Get all user IDs with the role
	userIDs, err := s.store.GetUserIDsByRoleName(ctx, roleName)
	if err != nil {
		s.logger.Error(ctx, "EnqueueForRole", "Failed to get users by role",
			zap.String("role", roleName),
			zap.Error(err),
		)
		return
	}

	// Enqueue notification for each user
	for _, userID := range userIDs {
		s.Enqueue(&CreateNotificationRequest{
			UserID:       userID,
			Type:         req.Type,
			Priority:     req.Priority,
			Title:        req.Title,
			Message:      req.Message,
			ResourceType: req.ResourceType,
			ResourceID:   req.ResourceID,
		})
	}

	s.logger.Info(ctx, "EnqueueForRole", "Notifications queued for role",
		zap.String("role", roleName),
		zap.Int("userCount", len(userIDs)),
	)
}

// EnqueueForUsers creates notifications for multiple users (async)
func (s *notificationService) EnqueueForUsers(userIDs []string, req *CreateNotificationRequest) {
	for _, userID := range userIDs {
		s.Enqueue(&CreateNotificationRequest{
			UserID:       userID,
			Type:         req.Type,
			Priority:     req.Priority,
			Title:        req.Title,
			Message:      req.Message,
			ResourceType: req.ResourceType,
			ResourceID:   req.ResourceID,
		})
	}
}

// createInternal is the internal create method used by workers
func (s *notificationService) createInternal(ctx context.Context, req *CreateNotificationRequest) (*NotificationResponse, error) {
	// Set default priority if not provided
	priority := req.Priority
	if priority == "" {
		priority = PriorityNormal
	}

	// Create the notification in the database
	notification, err := s.store.CreateNotification(ctx, db.CreateNotificationParams{
		ID:           nanoid.Generate(),
		UserID:       req.UserID,
		Type:         db.NotificationTypeEnum(req.Type),
		Priority:     db.NotificationPriorityEnum(priority),
		Title:        req.Title,
		Message:      req.Message,
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
		ExpiresAt:    pgtype.Timestamptz{}, // No expiration by default
	})
	if err != nil {
		return nil, err
	}

	// Build response
	response := s.mapToResponse(notification)

	// Broadcast via WebSocket if hub is available
	if s.hub != nil {
		s.hub.SendToUser(req.UserID, &websocket.Message{
			Type: websocket.MessageTypeNotification,
			Payload: websocket.NotificationPayload{
				ID:           response.ID,
				Type:         response.Type,
				Priority:     response.Priority,
				Title:        response.Title,
				Message:      response.Message,
				ResourceType: response.ResourceType,
				ResourceID:   response.ResourceID,
				CreatedAt:    response.CreatedAt,
			},
		})
	}

	return response, nil
}

// Create creates a new notification and broadcasts it via WebSocket (synchronous)
// Use Enqueue for async non-blocking creation
func (s *notificationService) Create(ctx context.Context, req *CreateNotificationRequest) (*NotificationResponse, error) {
	response, err := s.createInternal(ctx, req)
	if err != nil {
		s.logger.Error(ctx, "CreateNotification", "Failed to create notification", zap.Error(err))
		return nil, ErrInternal
	}

	s.logger.Info(ctx, "CreateNotification", "Notification created",
		zap.String("notificationID", response.ID),
		zap.String("userID", req.UserID),
		zap.String("type", req.Type),
	)

	return response, nil
}

// List returns paginated notifications for the current user
func (s *notificationService) List(ctx context.Context, req *ListNotificationsRequest) (*resp.PaginationResponse[NotificationResponse], error) {
	userID := util.GetUserID(ctx)
	limit, offset, page, pageSize := middleware.GetPaginationParams(ctx)

	notifications, err := s.store.ListNotifications(ctx, db.ListNotificationsParams{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
		IsRead: req.IsRead,
	})
	if err != nil {
		s.logger.Error(ctx, "ListNotifications", "Failed to list notifications", zap.Error(err))
		return nil, ErrInternal
	}

	// Map to response
	items := make([]NotificationResponse, 0, len(notifications))
	var totalCount int64

	for _, n := range notifications {
		items = append(items, *s.mapRowToResponse(n))
		if totalCount == 0 {
			totalCount = n.TotalCount
		}
	}

	result := resp.PagRespWithParams(items, int(totalCount), page, pageSize)
	return &result, nil
}

// MarkAsRead marks a single notification as read
func (s *notificationService) MarkAsRead(ctx context.Context, notificationID string) error {
	userID := util.GetUserID(ctx)

	err := s.store.MarkNotificationAsRead(ctx, db.MarkNotificationAsReadParams{
		ID:     notificationID,
		UserID: userID,
	})
	if err != nil {
		s.logger.Error(ctx, "MarkAsRead", "Failed to mark notification as read", zap.Error(err))
		return ErrInternal
	}

	return nil
}

// MarkAllAsRead marks all notifications as read for the current user
func (s *notificationService) MarkAllAsRead(ctx context.Context) error {
	userID := util.GetUserID(ctx)

	err := s.store.MarkAllNotificationsAsRead(ctx, userID)
	if err != nil {
		s.logger.Error(ctx, "MarkAllAsRead", "Failed to mark all notifications as read", zap.Error(err))
		return ErrInternal
	}

	return nil
}

// GetUnreadCount returns the count of unread notifications for the current user
func (s *notificationService) GetUnreadCount(ctx context.Context) (int64, error) {
	userID := util.GetUserID(ctx)

	count, err := s.store.GetUnreadCount(ctx, userID)
	if err != nil {
		s.logger.Error(ctx, "GetUnreadCount", "Failed to get unread count", zap.Error(err))
		return 0, ErrInternal
	}

	return count, nil
}

// Delete deletes a notification
func (s *notificationService) Delete(ctx context.Context, notificationID string) error {
	userID := util.GetUserID(ctx)

	err := s.store.DeleteNotification(ctx, db.DeleteNotificationParams{
		ID:     notificationID,
		UserID: userID,
	})
	if err != nil {
		s.logger.Error(ctx, "DeleteNotification", "Failed to delete notification", zap.Error(err))
		return ErrInternal
	}

	return nil
}

// mapToResponse maps a database notification to response DTO
func (s *notificationService) mapToResponse(n db.Notification) *NotificationResponse {
	resp := &NotificationResponse{
		ID:           n.ID,
		Type:         string(n.Type),
		Priority:     string(n.Priority),
		Title:        n.Title,
		Message:      n.Message,
		ResourceType: n.ResourceType,
		ResourceID:   n.ResourceID,
		IsRead:       false,
		CreatedAt:    util.PgtypeTimestamptzToStr(n.CreatedAt),
	}

	if n.IsRead != nil {
		resp.IsRead = *n.IsRead
	}

	if n.ReadAt.Valid {
		readAt := util.PgtypeTimestamptzToStr(n.ReadAt)
		resp.ReadAt = &readAt
	}

	return resp
}

// mapRowToResponse maps a list row to response DTO
func (s *notificationService) mapRowToResponse(n db.ListNotificationsRow) *NotificationResponse {
	resp := &NotificationResponse{
		ID:           n.ID,
		Type:         string(n.Type),
		Priority:     string(n.Priority),
		Title:        n.Title,
		Message:      n.Message,
		ResourceType: n.ResourceType,
		ResourceID:   n.ResourceID,
		IsRead:       false,
		CreatedAt:    util.PgtypeTimestamptzToStr(n.CreatedAt),
	}

	if n.IsRead != nil {
		resp.IsRead = *n.IsRead
	}

	if n.ReadAt.Valid {
		readAt := util.PgtypeTimestamptzToStr(n.ReadAt)
		resp.ReadAt = &readAt
	}

	return resp
}
