package notification

// CreateNotificationRequest is used to create a new notification
type CreateNotificationRequest struct {
	UserID       string  `json:"user_id" binding:"required"`
	Type         string  `json:"type" binding:"required"`
	Priority     string  `json:"priority"`
	Title        string  `json:"title" binding:"required"`
	Message      string  `json:"message" binding:"required"`
	ResourceType *string `json:"resource_type,omitempty"`
	ResourceID   *string `json:"resource_id,omitempty"`
}

// NotificationResponse is the response for a single notification
type NotificationResponse struct {
	ID           string  `json:"id"`
	Type         string  `json:"type"`
	Priority     string  `json:"priority"`
	Title        string  `json:"title"`
	Message      string  `json:"message"`
	ResourceType *string `json:"resource_type,omitempty"`
	ResourceID   *string `json:"resource_id,omitempty"`
	IsRead       bool    `json:"is_read"`
	ReadAt       *string `json:"read_at,omitempty"`
	CreatedAt    string  `json:"created_at"`
}

// ListNotificationsRequest is the query params for listing notifications
type ListNotificationsRequest struct {
	IsRead *bool `form:"is_read"`
}

// UnreadCountResponse is the response for unread count
type UnreadCountResponse struct {
	Count int64 `json:"count"`
}

// MarkReadRequest is the request for marking notifications as read
type MarkReadRequest struct {
	NotificationID string `json:"notification_id"`
}

// WSAuthRequest is the request for WebSocket auth ticket exchange
type WSAuthRequest struct {
	Token string `json:"token" binding:"required"`
}

// WSAuthResponse is the response containing the one-time ticket
type WSAuthResponse struct {
	Ticket string `json:"ticket"`
}
