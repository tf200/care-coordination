package audit

import (
	db "care-cordination/lib/db/sqlc"
	"care-cordination/lib/logger"
	"care-cordination/lib/nanoid"
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

// ComputeHash generates a SHA-256 hash of audit entry data + previous hash
func ComputeHash(id string, userID, employeeID *string, action, resourceType string,
	resourceID *string, oldValue, newValue []byte, ipAddress, userAgent, requestID *string,
	status, failureReason *string, prevHash string, createdAt time.Time) string {

	data := fmt.Sprintf("%s|%v|%v|%s|%s|%v|%s|%s|%v|%v|%v|%v|%v|%s|%s",
		id,
		PtrToStr(userID),
		PtrToStr(employeeID),
		action,
		resourceType,
		PtrToStr(resourceID),
		string(oldValue),
		string(newValue),
		PtrToStr(ipAddress),
		PtrToStr(userAgent),
		PtrToStr(requestID),
		PtrToStr(status),
		PtrToStr(failureReason),
		prevHash,
		createdAt.UTC().Format(time.RFC3339Nano),
	)

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func PtrToStr(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}

func StrToPtr(s string) *string {
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
	currentHash := ComputeHash(
		id,
		StrToPtr(entry.UserID),
		StrToPtr(entry.EmployeeID),
		actionStr,
		entry.ResourceType,
		StrToPtr(entry.ResourceID),
		oldValueJSON,
		newValueJSON,
		StrToPtr(entry.IPAddress),
		StrToPtr(entry.UserAgent),
		StrToPtr(entry.RequestID),
		&statusStr,
		StrToPtr(entry.FailureReason),
		prevHash,
		createdAt,
	)

	err = s.store.CreateAuditLog(ctx, db.CreateAuditLogParams{
		ID:            id,
		UserID:        StrToPtr(entry.UserID),
		EmployeeID:    StrToPtr(entry.EmployeeID),
		ClientID:      StrToPtr(entry.ClientID),
		Action:        db.AuditActionEnum(entry.Action),
		ResourceType:  entry.ResourceType,
		ResourceID:    StrToPtr(entry.ResourceID),
		OldValue:      oldValueJSON,
		NewValue:      newValueJSON,
		IpAddress:     StrToPtr(entry.IPAddress),
		UserAgent:     StrToPtr(entry.UserAgent),
		RequestID:     StrToPtr(entry.RequestID),
		Status:        db.AuditStatusEnum(entry.Status),
		FailureReason: StrToPtr(entry.FailureReason),
		PrevHash:      prevHash,
		CurrentHash:   currentHash,
	})

	if err != nil {
		s.logger.Error(ctx, "AuditLoggerService.Log", "Failed to create audit log", zap.Error(err))
		return nil
	}

	return nil
}

// LogEntry implements the AuditLogger interface
func (s *AuditLoggerService) LogEntry(ctx context.Context, entry AuditEntry) error {
	return s.Log(ctx, entry)
}
