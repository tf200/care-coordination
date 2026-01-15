package middleware

import (
	db "care-cordination/lib/db/sqlc"
	"care-cordination/lib/logger"
	"care-cordination/lib/nanoid"
	"care-cordination/lib/util"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

const GenesisHash = "GENESIS"

// AuditLoggerService handles tamper-evident audit logging
type AuditLoggerService struct {
	store    db.Store
	logger   logger.Logger
	hashLock sync.Mutex // Mutex to ensure sequential hash chain integrity
}

// NewAuditLoggerService creates a new audit logger service
func NewAuditLoggerService(store db.Store, logger logger.Logger) *AuditLoggerService {
	return &AuditLoggerService{
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
		ptrToStr(status),
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

func strToPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// Log creates a tamper-evident audit log entry
func (s *AuditLoggerService) Log(ctx context.Context, entry AuditEntry) error {
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
		if errors.Is(err, pgx.ErrNoRows) {
			// First entry in the chain
			prevHash = GenesisHash
		} else {
			s.logger.Error(ctx, "AuditLoggerService.Log", "Failed to get latest audit log", zap.Error(err))
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
		strToPtr(entry.UserID),
		strToPtr(entry.EmployeeID),
		actionStr,
		entry.ResourceType,
		strToPtr(entry.ResourceID),
		oldValueJSON,
		newValueJSON,
		strToPtr(entry.IPAddress),
		strToPtr(entry.UserAgent),
		strToPtr(entry.RequestID),
		&statusStr,
		strToPtr(entry.FailureReason),
		prevHash,
		createdAt,
	)

	err = s.store.CreateAuditLog(ctx, db.CreateAuditLogParams{
		ID:            id,
		UserID:        strToPtr(entry.UserID),
		EmployeeID:    strToPtr(entry.EmployeeID),
		ClientID:      strToPtr(entry.ClientID),
		Action:        db.AuditActionEnum(entry.Action),
		ResourceType:  entry.ResourceType,
		ResourceID:    strToPtr(entry.ResourceID),
		OldValue:      oldValueJSON,
		NewValue:      newValueJSON,
		IpAddress:     strToPtr(entry.IPAddress),
		UserAgent:     strToPtr(entry.UserAgent),
		RequestID:     strToPtr(entry.RequestID),
		Status:        db.AuditStatusEnum(entry.Status),
		FailureReason: strToPtr(entry.FailureReason),
		PrevHash:      prevHash,
		CurrentHash:   currentHash,
	})

	if err != nil {
		s.logger.Error(ctx, "AuditLoggerService.Log", "Failed to create audit log", zap.Error(err))
		return nil
	}

	return nil
}

// LogEntry implements the AuditLogger interface for the middleware
func (s *AuditLoggerService) LogEntry(ctx context.Context, entry AuditEntry) error {
	return s.Log(ctx, entry)
}

// LogRead logs a read operation
func (s *AuditLoggerService) LogRead(ctx context.Context, resourceType, resourceID string) error {
	return s.Log(ctx, AuditEntry{
		UserID:       util.GetUserID(ctx),
		EmployeeID:   util.GetEmployeeID(ctx),
		Action:       ActionRead,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		IPAddress:    util.GetIPAddress(ctx),
		UserAgent:    util.GetUserAgent(ctx),
		RequestID:    util.GetRequestID(ctx),
		Status:       StatusSuccess,
	})
}

// LogCreate logs a create operation
func (s *AuditLoggerService) LogCreate(ctx context.Context, resourceType, resourceID string, newValue interface{}) error {
	return s.Log(ctx, AuditEntry{
		UserID:       util.GetUserID(ctx),
		EmployeeID:   util.GetEmployeeID(ctx),
		Action:       ActionCreate,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		NewValue:     newValue,
		IPAddress:    util.GetIPAddress(ctx),
		UserAgent:    util.GetUserAgent(ctx),
		RequestID:    util.GetRequestID(ctx),
		Status:       StatusSuccess,
	})
}

// LogUpdate logs an update operation
func (s *AuditLoggerService) LogUpdate(ctx context.Context, resourceType, resourceID string, oldValue, newValue interface{}) error {
	return s.Log(ctx, AuditEntry{
		UserID:       util.GetUserID(ctx),
		EmployeeID:   util.GetEmployeeID(ctx),
		Action:       ActionUpdate,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		OldValue:     oldValue,
		NewValue:     newValue,
		IPAddress:    util.GetIPAddress(ctx),
		UserAgent:    util.GetUserAgent(ctx),
		RequestID:    util.GetRequestID(ctx),
		Status:       StatusSuccess,
	})
}

// LogDelete logs a delete operation
func (s *AuditLoggerService) LogDelete(ctx context.Context, resourceType, resourceID string, oldValue interface{}) error {
	return s.Log(ctx, AuditEntry{
		UserID:       util.GetUserID(ctx),
		EmployeeID:   util.GetEmployeeID(ctx),
		Action:       ActionDelete,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		OldValue:     oldValue,
		IPAddress:    util.GetIPAddress(ctx),
		UserAgent:    util.GetUserAgent(ctx),
		RequestID:    util.GetRequestID(ctx),
		Status:       StatusSuccess,
	})
}

// LogFailure logs a failed operation
func (s *AuditLoggerService) LogFailure(ctx context.Context, action AuditAction, resourceType, resourceID, reason string) error {
	return s.Log(ctx, AuditEntry{
		UserID:        util.GetUserID(ctx),
		EmployeeID:    util.GetEmployeeID(ctx),
		Action:        action,
		ResourceType:  resourceType,
		ResourceID:    resourceID,
		IPAddress:     util.GetIPAddress(ctx),
		UserAgent:     util.GetUserAgent(ctx),
		RequestID:     util.GetRequestID(ctx),
		Status:        StatusFailure,
		FailureReason: reason,
	})
}
