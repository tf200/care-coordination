package websocket

// Message types for WebSocket communication
const (
	// Server -> Client message types
	MessageTypeNotification = "notification"
	MessageTypePing         = "ping"
	MessageTypeConnected    = "connected"
	MessageTypeError        = "error"
	MessageTypeUnreadCount  = "unread_count"

	// Client -> Server message types
	MessageTypePong        = "pong"
	MessageTypeMarkRead    = "mark_read"
	MessageTypeMarkAllRead = "mark_all_read"
)

// Message represents a WebSocket message
type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
}

// NotificationPayload is the payload for notification messages
type NotificationPayload struct {
	ID           string  `json:"id"`
	Type         string  `json:"type"`
	Priority     string  `json:"priority"`
	Title        string  `json:"title"`
	Message      string  `json:"message"`
	ResourceType *string `json:"resource_type,omitempty"`
	ResourceID   *string `json:"resource_id,omitempty"`
	CreatedAt    string  `json:"created_at"`
}

// ErrorPayload is the payload for error messages
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// UnreadCountPayload is the payload for unread count messages
type UnreadCountPayload struct {
	Count int64 `json:"count"`
}

// ClientMessage represents a message from client to server
type ClientMessage struct {
	Type    string `json:"type"`
	Payload string `json:"payload,omitempty"` // notification ID for mark_read
}
