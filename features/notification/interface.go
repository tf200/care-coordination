package notification

import (
	"care-cordination/lib/resp"
	"context"
)

// NotificationService defines the interface for notification operations
type NotificationService interface {
	// Create creates a new notification and broadcasts it via WebSocket (synchronous)
	Create(ctx context.Context, req *CreateNotificationRequest) (*NotificationResponse, error)

	// Enqueue adds a notification to the async queue (non-blocking)
	// This is the preferred method for service triggers
	Enqueue(req *CreateNotificationRequest)

	// EnqueueForRole creates notifications for all users with the specified role (async)
	EnqueueForRole(ctx context.Context, roleName string, req *CreateNotificationRequest)

	// EnqueueForUsers creates notifications for multiple users (async)
	EnqueueForUsers(userIDs []string, req *CreateNotificationRequest)

	// List returns paginated notifications for the current user
	List(ctx context.Context, req *ListNotificationsRequest) (*resp.PaginationResponse[NotificationResponse], error)

	// MarkAsRead marks a single notification as read
	MarkAsRead(ctx context.Context, notificationID string) error

	// MarkAllAsRead marks all notifications as read for the current user
	MarkAllAsRead(ctx context.Context) error

	// GetUnreadCount returns the count of unread notifications for the current user
	GetUnreadCount(ctx context.Context) (int64, error)

	// Delete deletes a notification
	Delete(ctx context.Context, notificationID string) error
}
