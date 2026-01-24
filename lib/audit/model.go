package audit

import (
	"context"
	"time"
)

// AuditAction represents an audit action type
type AuditAction string

const (
	ActionRead   AuditAction = "read"
	ActionCreate AuditAction = "create"
	ActionUpdate AuditAction = "update"
	ActionDelete AuditAction = "delete"
)

// AuditStatus represents the status of an audit entry
type AuditStatus string

const (
	StatusSuccess AuditStatus = "success"
	StatusFailure AuditStatus = "failure"
)

// AuditEntry represents a single audit log entry
type AuditEntry struct {
	UserID        string
	EmployeeID    string
	ClientID      string // NEN7510: Track which client's data was accessed
	Action        AuditAction
	ResourceType  string
	ResourceID    string
	OldValue      any
	NewValue      any
	IPAddress     string
	UserAgent     string
	RequestID     string
	Status        AuditStatus
	FailureReason string
	CreatedAt     time.Time
}

// AuditLogger is an interface for logging audit entries
type AuditLogger interface {
	LogEntry(ctx context.Context, entry AuditEntry) error
}
