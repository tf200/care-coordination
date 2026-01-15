package audit

import "time"

// ListAuditLogsRequest represents the request for listing audit logs
type ListAuditLogsRequest struct {
	UserID       *string    `form:"user_id"`
	ClientID     *string    `form:"client_id"`
	ResourceType *string    `form:"resource_type"`
	ResourceID   *string    `form:"resource_id"`
	Action       *string    `form:"action"`
	StartDate    *time.Time `form:"start_date"`
	EndDate      *time.Time `form:"end_date"`
}

// AuditLogResponse represents a single audit log entry in API responses
type AuditLogResponse struct {
	ID             string                 `json:"id"`
	SequenceNumber int64                  `json:"sequence_number"`
	UserID         *string                `json:"user_id,omitempty"`
	UserEmail      string                 `json:"user_email,omitempty"`
	EmployeeID     *string                `json:"employee_id,omitempty"`
	EmployeeName   string                 `json:"employee_name,omitempty"`
	ClientID       *string                `json:"client_id,omitempty"`
	ClientName     string                 `json:"client_name,omitempty"`
	Action         string                 `json:"action"`
	ResourceType   string                 `json:"resource_type"`
	ResourceID     *string                `json:"resource_id,omitempty"`
	OldValue       map[string]interface{} `json:"old_value,omitempty"`
	NewValue       map[string]interface{} `json:"new_value,omitempty"`
	IPAddress      *string                `json:"ip_address,omitempty"`
	UserAgent      *string                `json:"user_agent,omitempty"`
	RequestID      *string                `json:"request_id,omitempty"`
	Status         string                 `json:"status"`
	FailureReason  *string                `json:"failure_reason,omitempty"`
	PrevHash       string                 `json:"prev_hash"`
	CurrentHash    string                 `json:"current_hash"`
	CreatedAt      time.Time              `json:"created_at"`
}

// AuditLogStatsResponse represents audit log statistics
type AuditLogStatsResponse struct {
	TotalLogs    int64 `json:"total_logs"`
	ReadCount    int64 `json:"read_count"`
	CreateCount  int64 `json:"create_count"`
	UpdateCount  int64 `json:"update_count"`
	DeleteCount  int64 `json:"delete_count"`
	FailureCount int64 `json:"failure_count"`
}

// VerifyChainRequest represents a request to verify a range of audit logs
type VerifyChainRequest struct {
	StartSequence int64 `form:"start_sequence" binding:"required,min=1"`
	EndSequence   int64 `form:"end_sequence" binding:"required,gtefield=StartSequence"`
}
