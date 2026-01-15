---
description: Implementation plan for NEN7510/ISO27001 compliant audit logging system
---

# Audit Logging Implementation Plan

This document outlines the complete implementation plan for adding a NEN7510/ISO27001 compliant audit logging system to the care-coordination application.

## Overview

Audit logging tracks **who** did **what** to **which resource** and **when**, along with contextual information like IP address and user agent. This is critical for healthcare applications to meet NEN7510 and GDPR requirements.

## Architecture Decisions

### Approach: Middleware + Service Layer Hybrid

We'll use a **two-pronged approach**:

1. **Middleware Layer** (for READ operations): Automatically log all GET requests to sensitive resources
2. **Service Layer** (for WRITE operations): Explicitly log CREATE/UPDATE/DELETE with before/after values

This approach is chosen because:
- READ operations don't need before/after values (simpler, can be automated)
- WRITE operations need explicit before/after tracking (requires service layer awareness)

### What to Log (NEN7510 Requirements)

| Field | Description | Required |
|-------|-------------|----------|
| `id` | Unique audit log ID | ✅ |
| `timestamp` | When the action occurred | ✅ |
| `user_id` | Who performed the action | ✅ |
| `employee_id` | Employee ID (for traceability) | ✅ |
| `action` | read/create/update/delete | ✅ |
| `resource_type` | client/evaluation/intake/etc | ✅ |
| `resource_id` | ID of the affected resource | ✅ |
| `old_value` | Previous state (for updates) | ⚠️ For writes |
| `new_value` | New state (for creates/updates) | ⚠️ For writes |
| `ip_address` | Client IP address | ✅ |
| `user_agent` | Browser/client info | ✅ |
| `request_id` | Correlation ID for tracing | ✅ |
| `status` | success/failure | ✅ |
| `failure_reason` | Why it failed (if applicable) | ⚠️ On failure |
| `prev_hash` | SHA-256 hash of previous log entry | ✅ Tamper-evidence |
| `current_hash` | SHA-256 hash of this entry | ✅ Tamper-evidence |

### Hash Chain (Tamper-Evidence)

The audit log uses a **blockchain-style hash chain** to ensure tamper-evidence:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Log Entry 1   │    │   Log Entry 2   │    │   Log Entry 3   │
├─────────────────┤    ├─────────────────┤    ├─────────────────┤
│ prev_hash: NULL │───▶│ prev_hash: H1   │───▶│ prev_hash: H2   │
│ current_hash:H1 │    │ current_hash:H2 │    │ current_hash:H3 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

**How it works:**
1. Each entry's `current_hash` = SHA256(all fields + prev_hash)
2. Each entry's `prev_hash` = previous entry's `current_hash`
3. First entry has `prev_hash` = "GENESIS"
4. If any entry is modified, all subsequent hashes become invalid

**Verification:**
- Periodic integrity checks recalculate hashes and compare
- Any mismatch indicates tampering
- Alerts are triggered on chain breaks

### Sensitive Resources to Audit

High priority (PHI - Protected Health Information):
- `clients` - All operations
- `evaluations` - All operations  
- `intake_forms` - All operations
- `registration_forms` - All operations
- `incidents` - All operations
- `client_goals` - All operations

Medium priority:
- `employees` - All operations
- `locations` - Create/Update/Delete
- `appointments` - All operations involving clients

Lower priority:
- `auth` - Login/logout events (already somewhat covered by sessions)
- `notifications` - Not PHI, but useful for debugging

---

## Implementation Steps

### Phase 1: Database Schema (Migration)

// turbo-all

#### Step 1.1: Add audit_logs table to migration

Add the following to `/home/taha/care-cordination/lib/db/migrations/000001_init.up.sql`:

```sql
-- ============================================================
-- Audit Logging (NEN7510 / ISO27001 Compliance)
-- ============================================================

CREATE TYPE audit_action_enum AS ENUM ('read', 'create', 'update', 'delete', 'login', 'logout', 'export');
CREATE TYPE audit_status_enum AS ENUM ('success', 'failure');

CREATE TABLE audit_logs (
    id TEXT PRIMARY KEY,
    sequence_number BIGSERIAL NOT NULL,  -- For ordering and chain verification
    user_id TEXT REFERENCES users(id),
    employee_id TEXT REFERENCES employees(id),
    action audit_action_enum NOT NULL,
    resource_type TEXT NOT NULL,
    resource_id TEXT,
    old_value JSONB,
    new_value JSONB,
    ip_address TEXT,
    user_agent TEXT,
    request_id TEXT,
    status audit_status_enum NOT NULL DEFAULT 'success',
    failure_reason TEXT,
    prev_hash TEXT NOT NULL,       -- SHA-256 hash of previous entry (or "GENESIS" for first)
    current_hash TEXT NOT NULL,    -- SHA-256 hash of this entry including prev_hash
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for efficient querying
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_sequence ON audit_logs(sequence_number DESC);  -- For chain verification

-- Composite index for common query patterns
CREATE INDEX idx_audit_logs_user_resource_time ON audit_logs(user_id, resource_type, created_at DESC);

-- Unique index on sequence_number for integrity
CREATE UNIQUE INDEX idx_audit_logs_sequence_unique ON audit_logs(sequence_number);
```

### Phase 2: SQLC Queries

#### Step 2.1: Create audit queries file

Create `/home/taha/care-cordination/lib/db/queries/audit_logs.sql`:

```sql
-- name: CreateAuditLog :exec
INSERT INTO audit_logs (
    id, user_id, employee_id, action, resource_type, resource_id,
    old_value, new_value, ip_address, user_agent, request_id, status, failure_reason,
    prev_hash, current_hash
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
);

-- name: GetLatestAuditLog :one
-- Get the most recent audit log entry to retrieve its hash for the chain
SELECT id, current_hash, sequence_number 
FROM audit_logs 
ORDER BY sequence_number DESC 
LIMIT 1;

-- name: ListAuditLogs :many
SELECT 
    al.*,
    u.email as user_email,
    COALESCE(e.first_name || ' ' || e.last_name, '') as employee_name,
    COUNT(*) OVER() as total_count
FROM audit_logs al
LEFT JOIN users u ON al.user_id = u.id
LEFT JOIN employees e ON al.employee_id = e.id
WHERE 
    (sqlc.narg(user_id)::TEXT IS NULL OR al.user_id = sqlc.narg(user_id))
    AND (sqlc.narg(resource_type)::TEXT IS NULL OR al.resource_type = sqlc.narg(resource_type))
    AND (sqlc.narg(resource_id)::TEXT IS NULL OR al.resource_id = sqlc.narg(resource_id))
    AND (sqlc.narg(action)::audit_action_enum IS NULL OR al.action = sqlc.narg(action))
    AND (sqlc.narg(start_date)::TIMESTAMP IS NULL OR al.created_at >= sqlc.narg(start_date))
    AND (sqlc.narg(end_date)::TIMESTAMP IS NULL OR al.created_at <= sqlc.narg(end_date))
ORDER BY al.sequence_number DESC
LIMIT $1 OFFSET $2;

-- name: GetAuditLogsByResource :many
SELECT * FROM audit_logs
WHERE resource_type = $1 AND resource_id = $2
ORDER BY sequence_number DESC
LIMIT $3;

-- name: GetAuditLogsByUser :many
SELECT * FROM audit_logs
WHERE user_id = $1
ORDER BY sequence_number DESC
LIMIT $2 OFFSET $3;

-- name: GetAuditLogStats :one
SELECT 
    COUNT(*) as total_logs,
    COUNT(*) FILTER (WHERE action = 'read') as read_count,
    COUNT(*) FILTER (WHERE action = 'create') as create_count,
    COUNT(*) FILTER (WHERE action = 'update') as update_count,
    COUNT(*) FILTER (WHERE action = 'delete') as delete_count,
    COUNT(*) FILTER (WHERE status = 'failure') as failure_count
FROM audit_logs
WHERE created_at >= NOW() - INTERVAL '24 hours';

-- name: GetAuditLogsForVerification :many
-- Get audit logs in sequence order for hash chain verification
SELECT id, sequence_number, user_id, employee_id, action, resource_type, resource_id,
       old_value, new_value, ip_address, user_agent, request_id, status, failure_reason,
       prev_hash, current_hash, created_at
FROM audit_logs
WHERE sequence_number >= $1 AND sequence_number <= $2
ORDER BY sequence_number ASC;

-- name: GetAuditLogBySequence :one
SELECT * FROM audit_logs WHERE sequence_number = $1;

-- name: CountAuditLogs :one
SELECT COUNT(*) as total FROM audit_logs;
```

#### Step 2.2: Regenerate SQLC

```bash
cd /home/taha/care-cordination && make sqlc
```

### Phase 3: Audit Service

#### Step 3.1: Create audit feature directory structure

```bash
mkdir -p /home/taha/care-cordination/features/audit
```

#### Step 3.2: Create audit service interface and implementation

Create `/home/taha/care-cordination/features/audit/service.go`:

```go
package audit

import (
    "care-cordination/features/middleware"
    db "care-cordination/lib/db/sqlc"
    "care-cordination/lib/logger"
    "care-cordination/lib/nanoid"
    "context"
    "crypto/sha256"
    "database/sql"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "sync"
    "time"

    "go.uber.org/zap"
)

const GenesisHash = "GENESIS"

type Action string

const (
    ActionRead   Action = "read"
    ActionCreate Action = "create"
    ActionUpdate Action = "update"
    ActionDelete Action = "delete"
    ActionLogin  Action = "login"
    ActionLogout Action = "logout"
    ActionExport Action = "export"
)

type Status string

const (
    StatusSuccess Status = "success"
    StatusFailure Status = "failure"
)

type AuditEntry struct {
    UserID        string
    EmployeeID    string
    Action        Action
    ResourceType  string
    ResourceID    string
    OldValue      interface{}
    NewValue      interface{}
    IPAddress     string
    UserAgent     string
    RequestID     string
    Status        Status
    FailureReason string
}

// ChainVerificationResult represents the result of verifying the audit log chain
type ChainVerificationResult struct {
    IsValid          bool
    TotalEntries     int64
    VerifiedEntries  int64
    FirstBrokenSeq   *int64  // Sequence number where chain first breaks (nil if valid)
    BrokenEntryID    *string // ID of the entry where chain breaks
    ExpectedHash     string  // What the hash should be
    ActualHash       string  // What the hash actually is
    VerifiedAt       time.Time
}

//go:generate mockgen -destination=mocks/mock_audit_service.go -package=mocks care-cordination/features/audit AuditService
type AuditService interface {
    Log(ctx context.Context, entry AuditEntry) error
    LogRead(ctx context.Context, resourceType, resourceID string) error
    LogCreate(ctx context.Context, resourceType, resourceID string, newValue interface{}) error
    LogUpdate(ctx context.Context, resourceType, resourceID string, oldValue, newValue interface{}) error
    LogDelete(ctx context.Context, resourceType, resourceID string, oldValue interface{}) error
    LogFailure(ctx context.Context, action Action, resourceType, resourceID, reason string) error
    VerifyChain(ctx context.Context, startSeq, endSeq int64) (*ChainVerificationResult, error)
    VerifyFullChain(ctx context.Context) (*ChainVerificationResult, error)
}

type auditService struct {
    store    *db.Store
    logger   logger.Logger
    hashLock sync.Mutex // Mutex to ensure sequential hash chain integrity
}

func NewAuditService(store *db.Store, logger logger.Logger) AuditService {
    return &auditService{
        store:  store,
        logger: logger,
    }
}

// computeHash generates a SHA-256 hash of audit entry data + previous hash
func computeHash(id string, userID, employeeID *string, action, resourceType string, 
    resourceID *string, oldValue, newValue []byte, ipAddress, userAgent, requestID *string,
    status, failureReason *string, prevHash string, createdAt time.Time) string {
    
    data := fmt.Sprintf("%s|%v|%v|%s|%s|%v|%s|%s|%v|%v|%v|%v|%v|%s|%s",
        id,
        ptrToStr(userID),
        ptrToStr(employeeID),
        action,
        resourceType,
        ptrToStr(resourceID),
        string(oldValue),
        string(newValue),
        ptrToStr(ipAddress),
        ptrToStr(userAgent),
        ptrToStr(requestID),
        status,
        ptrToStr(failureReason),
        prevHash,
        createdAt.UTC().Format(time.RFC3339Nano),
    )
    
    hash := sha256.Sum256([]byte(data))
    return hex.EncodeToString(hash[:])
}

func ptrToStr(s *string) string {
    if s == nil {
        return "<nil>"
    }
    return *s
}

func (s *auditService) Log(ctx context.Context, entry AuditEntry) error {
    // Lock to ensure hash chain integrity (entries must be sequential)
    s.hashLock.Lock()
    defer s.hashLock.Unlock()
    
    id := nanoid.Generate()
    createdAt := time.Now().UTC()
    
    oldValueJSON, _ := json.Marshal(entry.OldValue)
    newValueJSON, _ := json.Marshal(entry.NewValue)
    
    // Get the latest audit log to retrieve its hash
    var prevHash string
    latestLog, err := s.store.GetLatestAuditLog(ctx)
    if err != nil {
        if err == sql.ErrNoRows {
            // First entry in the chain
            prevHash = GenesisHash
        } else {
            s.logger.Error(ctx, "AuditService.Log", "Failed to get latest audit log", zap.Error(err))
            return nil // Don't break main flow
        }
    } else {
        prevHash = latestLog.CurrentHash
    }
    
    // Compute the hash for this entry
    statusStr := string(entry.Status)
    actionStr := string(entry.Action)
    currentHash := computeHash(
        id,
        toNullString(entry.UserID),
        toNullString(entry.EmployeeID),
        actionStr,
        entry.ResourceType,
        toNullString(entry.ResourceID),
        oldValueJSON,
        newValueJSON,
        toNullString(entry.IPAddress),
        toNullString(entry.UserAgent),
        toNullString(entry.RequestID),
        &statusStr,
        toNullString(entry.FailureReason),
        prevHash,
        createdAt,
    )

    err = s.store.CreateAuditLog(ctx, db.CreateAuditLogParams{
        ID:            id,
        UserID:        toNullString(entry.UserID),
        EmployeeID:    toNullString(entry.EmployeeID),
        Action:        db.AuditActionEnum(entry.Action),
        ResourceType:  entry.ResourceType,
        ResourceID:    toNullString(entry.ResourceID),
        OldValue:      oldValueJSON,
        NewValue:      newValueJSON,
        IpAddress:     toNullString(entry.IPAddress),
        UserAgent:     toNullString(entry.UserAgent),
        RequestID:     toNullString(entry.RequestID),
        Status:        db.AuditStatusEnum(entry.Status),
        FailureReason: toNullString(entry.FailureReason),
        PrevHash:      prevHash,
        CurrentHash:   currentHash,
    })

    if err != nil {
        s.logger.Error(ctx, "AuditService.Log", "Failed to create audit log", zap.Error(err))
        return nil
    }

    return nil
}

// VerifyChain verifies the hash chain integrity between two sequence numbers
func (s *auditService) VerifyChain(ctx context.Context, startSeq, endSeq int64) (*ChainVerificationResult, error) {
    logs, err := s.store.GetAuditLogsForVerification(ctx, db.GetAuditLogsForVerificationParams{
        SequenceNumber:   startSeq,
        SequenceNumber_2: endSeq,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to get audit logs: %w", err)
    }
    
    result := &ChainVerificationResult{
        IsValid:         true,
        TotalEntries:    int64(len(logs)),
        VerifiedEntries: 0,
        VerifiedAt:      time.Now(),
    }
    
    var prevHash string
    for i, log := range logs {
        // For the first entry in our range, get the previous hash
        if i == 0 {
            if log.SequenceNumber == 1 {
                prevHash = GenesisHash
            } else {
                // Get the previous entry's hash
                prevLog, err := s.store.GetAuditLogBySequence(ctx, log.SequenceNumber-1)
                if err != nil {
                    return nil, fmt.Errorf("failed to get previous log: %w", err)
                }
                prevHash = prevLog.CurrentHash
            }
        }
        
        // Verify prev_hash matches
        if log.PrevHash != prevHash {
            result.IsValid = false
            result.FirstBrokenSeq = &log.SequenceNumber
            result.BrokenEntryID = &log.ID
            result.ExpectedHash = prevHash
            result.ActualHash = log.PrevHash
            return result, nil
        }
        
        // Recompute the hash and verify current_hash
        statusStr := string(log.Status)
        actionStr := string(log.Action)
        computedHash := computeHash(
            log.ID,
            log.UserID,
            log.EmployeeID,
            actionStr,
            log.ResourceType,
            log.ResourceID,
            log.OldValue,
            log.NewValue,
            log.IpAddress,
            log.UserAgent,
            log.RequestID,
            &statusStr,
            log.FailureReason,
            log.PrevHash,
            log.CreatedAt.UTC(),
        )
        
        if computedHash != log.CurrentHash {
            result.IsValid = false
            result.FirstBrokenSeq = &log.SequenceNumber
            result.BrokenEntryID = &log.ID
            result.ExpectedHash = computedHash
            result.ActualHash = log.CurrentHash
            return result, nil
        }
        
        prevHash = log.CurrentHash
        result.VerifiedEntries++
    }
    
    return result, nil
}

// VerifyFullChain verifies the entire audit log hash chain
func (s *auditService) VerifyFullChain(ctx context.Context) (*ChainVerificationResult, error) {
    count, err := s.store.CountAuditLogs(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to count audit logs: %w", err)
    }
    
    if count == 0 {
        return &ChainVerificationResult{
            IsValid:         true,
            TotalEntries:    0,
            VerifiedEntries: 0,
            VerifiedAt:      time.Now(),
        }, nil
    }
    
    return s.VerifyChain(ctx, 1, count)
}

func (s *auditService) LogRead(ctx context.Context, resourceType, resourceID string) error {
    return s.Log(ctx, AuditEntry{
        UserID:       middleware.GetUserID(ctx),
        EmployeeID:   middleware.GetEmployeeID(ctx),
        Action:       ActionRead,
        ResourceType: resourceType,
        ResourceID:   resourceID,
        IPAddress:    middleware.GetIPAddress(ctx),
        UserAgent:    middleware.GetUserAgent(ctx),
        RequestID:    middleware.GetRequestID(ctx),
        Status:       StatusSuccess,
    })
}

func (s *auditService) LogCreate(ctx context.Context, resourceType, resourceID string, newValue interface{}) error {
    return s.Log(ctx, AuditEntry{
        UserID:       middleware.GetUserID(ctx),
        EmployeeID:   middleware.GetEmployeeID(ctx),
        Action:       ActionCreate,
        ResourceType: resourceType,
        ResourceID:   resourceID,
        NewValue:     newValue,
        IPAddress:    middleware.GetIPAddress(ctx),
        UserAgent:    middleware.GetUserAgent(ctx),
        RequestID:    middleware.GetRequestID(ctx),
        Status:       StatusSuccess,
    })
}

func (s *auditService) LogUpdate(ctx context.Context, resourceType, resourceID string, oldValue, newValue interface{}) error {
    return s.Log(ctx, AuditEntry{
        UserID:       middleware.GetUserID(ctx),
        EmployeeID:   middleware.GetEmployeeID(ctx),
        Action:       ActionUpdate,
        ResourceType: resourceType,
        ResourceID:   resourceID,
        OldValue:     oldValue,
        NewValue:     newValue,
        IPAddress:    middleware.GetIPAddress(ctx),
        UserAgent:    middleware.GetUserAgent(ctx),
        RequestID:    middleware.GetRequestID(ctx),
        Status:       StatusSuccess,
    })
}

func (s *auditService) LogDelete(ctx context.Context, resourceType, resourceID string, oldValue interface{}) error {
    return s.Log(ctx, AuditEntry{
        UserID:       middleware.GetUserID(ctx),
        EmployeeID:   middleware.GetEmployeeID(ctx),
        Action:       ActionDelete,
        ResourceType: resourceType,
        ResourceID:   resourceID,
        OldValue:     oldValue,
        IPAddress:    middleware.GetIPAddress(ctx),
        UserAgent:    middleware.GetUserAgent(ctx),
        RequestID:    middleware.GetRequestID(ctx),
        Status:       StatusSuccess,
    })
}

func (s *auditService) LogFailure(ctx context.Context, action Action, resourceType, resourceID, reason string) error {
    return s.Log(ctx, AuditEntry{
        UserID:        middleware.GetUserID(ctx),
        EmployeeID:    middleware.GetEmployeeID(ctx),
        Action:        action,
        ResourceType:  resourceType,
        ResourceID:    resourceID,
        IPAddress:     middleware.GetIPAddress(ctx),
        UserAgent:     middleware.GetUserAgent(ctx),
        RequestID:     middleware.GetRequestID(ctx),
        Status:        StatusFailure,
        FailureReason: reason,
    })
}

func toNullString(s string) *string {
    if s == "" {
        return nil
    }
    return &s
}
```

### Phase 4: Middleware Helpers

#### Step 4.1: Add context helpers to middleware

Add to `/home/taha/care-cordination/features/middleware/middleware.go`:

```go
const (
    IPAddressKey  = "ip_address"
    UserAgentKey  = "user_agent"
)

// Add these helper functions for audit logging
func GetUserID(ctx context.Context) string {
    if ginCtx, ok := ctx.(*gin.Context); ok {
        return ginCtx.GetString(UserIDKey)
    }
    if v, ok := ctx.Value(UserIDKey).(string); ok {
        return v
    }
    return ""
}

func GetEmployeeID(ctx context.Context) string {
    if ginCtx, ok := ctx.(*gin.Context); ok {
        return ginCtx.GetString(EmployeeIDKey)
    }
    if v, ok := ctx.Value(EmployeeIDKey).(string); ok {
        return v
    }
    return ""
}

func GetRequestID(ctx context.Context) string {
    if ginCtx, ok := ctx.(*gin.Context); ok {
        if v, exists := ginCtx.Get("X-Request-Id"); exists {
            return v.(string)
        }
    }
    if v, ok := ctx.Value("X-Request-Id").(string); ok {
        return v
    }
    return ""
}

func GetIPAddress(ctx context.Context) string {
    if ginCtx, ok := ctx.(*gin.Context); ok {
        return ginCtx.ClientIP()
    }
    return ""
}

func GetUserAgent(ctx context.Context) string {
    if ginCtx, ok := ctx.(*gin.Context); ok {
        return ginCtx.Request.UserAgent()
    }
    return ""
}
```

### Phase 5: Integrate Audit Logging in Services

#### Step 5.1: Update services to inject AuditService

Example for intake service - similar pattern for all services:

```go
type intakeService struct {
    db     *db.Store
    logger logger.Logger
    audit  audit.AuditService  // Add this
}

func NewIntakeService(db *db.Store, logger logger.Logger, audit audit.AuditService) IntakeService {
    return &intakeService{
        db:     db,
        logger: logger,
        audit:  audit,
    }
}
```

#### Step 5.2: Add audit logging to CREATE operations

Example for CreateIntakeForm:

```go
func (s *intakeService) CreateIntakeForm(ctx context.Context, req *CreateIntakeFormRequest) (*CreateIntakeFormResponse, error) {
    id := nanoid.Generate()
    
    // ... existing creation logic ...
    
    if err != nil {
        s.audit.LogFailure(ctx, audit.ActionCreate, "intake_form", "", err.Error())
        return nil, ErrInternal
    }
    
    // Log successful creation
    s.audit.LogCreate(ctx, "intake_form", id, req)
    
    return &CreateIntakeFormResponse{ID: id}, nil
}
```

#### Step 5.3: Add audit logging to UPDATE operations

Example for UpdateIntakeForm:

```go
func (s *intakeService) UpdateIntakeForm(ctx context.Context, id string, req *UpdateIntakeFormRequest) (*UpdateIntakeFormResponse, error) {
    // Fetch current state BEFORE update
    oldValue, err := s.db.GetIntakeFormWithDetails(ctx, id)
    if err != nil {
        s.audit.LogFailure(ctx, audit.ActionUpdate, "intake_form", id, "not found")
        return nil, ErrInternal
    }
    
    // ... existing update logic ...
    
    // Log successful update with before/after
    s.audit.LogUpdate(ctx, "intake_form", id, oldValue, req)
    
    return &UpdateIntakeFormResponse{ID: id}, nil
}
```

#### Step 5.4: Add audit logging to READ operations

Example for GetIntakeForm:

```go
func (s *intakeService) GetIntakeForm(ctx context.Context, id string) (*GetIntakeFormResponse, error) {
    result, err := s.db.GetIntakeFormWithDetails(ctx, id)
    if err != nil {
        return nil, ErrInternal
    }
    
    // Log the read access
    s.audit.LogRead(ctx, "intake_form", id)
    
    return result, nil
}
```

### Phase 6: Admin API for Viewing Audit Logs

#### Step 6.1: Create audit handler

Create `/home/taha/care-cordination/features/audit/handler.go` for admin access to audit logs.

### Phase 7: Testing

#### Step 7.1: Create audit service tests

Create `/home/taha/care-cordination/features/audit/service_test.go`

---

## Summary: Files to Create/Modify

### New Files:
1. `features/audit/service.go` - Audit service implementation
2. `features/audit/handler.go` - Admin API endpoints  
3. `features/audit/dto.go` - Request/response DTOs
4. `features/audit/errors.go` - Error definitions
5. `lib/db/queries/audit_logs.sql` - SQLC queries

### Modified Files:
1. `lib/db/migrations/000001_init.up.sql` - Add audit_logs table
2. `features/middleware/middleware.go` - Add context helper functions
3. `features/intake/service.go` - Add audit logging
4. `features/client/service.go` - Add audit logging
5. `features/evaluation/service.go` - Add audit logging
6. `features/registration/service.go` - Add audit logging
7. `features/incident/service.go` - Add audit logging
8. `internal/factory/factory.go` - Wire up audit service
9. `api/server.go` - Add audit handler routes

---

## Compliance Checklist

After implementation, you will satisfy these NEN7510/ISO27001 requirements:

- [x] All access to PHI is logged
- [x] Logs contain: who, what, when, where (IP)
- [x] Before/after values for modifications
- [x] Failed access attempts are logged
- [x] Logs are tamper-resistant (no delete API)
- [x] Logs are tamper-evident (hash chain verification)
- [x] Logs can be queried by resource, user, time
- [x] Request IDs for correlation
- [x] Chain integrity can be verified programmatically

---

## Next Steps

Would you like me to start implementing this plan? I suggest we start with:

1. **Phase 1**: Database schema update
2. **Phase 2**: SQLC queries
3. **Phase 3**: Audit service

Then we can integrate into each feature module incrementally.
