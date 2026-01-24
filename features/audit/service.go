package audit

import (
	"care-cordination/lib/audit"
	db "care-cordination/lib/db/sqlc"
	"care-cordination/lib/logger"
	"care-cordination/lib/util"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

// ChainVerificationResult represents the result of verifying the audit log chain
type ChainVerificationResult struct {
	IsValid         bool      `json:"is_valid"`
	TotalEntries    int64     `json:"total_entries"`
	VerifiedEntries int64     `json:"verified_entries"`
	FirstBrokenSeq  *int64    `json:"first_broken_seq,omitempty"`
	BrokenEntryID   *string   `json:"broken_entry_id,omitempty"`
	ExpectedHash    string    `json:"expected_hash,omitempty"`
	ActualHash      string    `json:"actual_hash,omitempty"`
	VerifiedAt      time.Time `json:"verified_at"`
}

type auditService struct {
	store  db.Store
	logger logger.Logger
}

func NewAuditService(store db.Store, logger logger.Logger) AuditService {
	return &auditService{
		store:  store,
		logger: logger,
	}
}

// ListAuditLogs retrieves audit logs with filters and pagination
func (s *auditService) ListAuditLogs(ctx context.Context, req *ListAuditLogsRequest, limit, offset int32) ([]AuditLogResponse, int64, error) {
	logs, err := s.store.ListAuditLogs(ctx, db.ListAuditLogsParams{
		Limit:        limit,
		Offset:       offset,
		UserID:       req.UserID,
		ClientID:     req.ClientID,
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
		Action:       toNullAuditActionEnum(req.Action),
		StartDate:    toPgtypeTimestamp(req.StartDate),
		EndDate:      toPgtypeTimestamp(req.EndDate),
	})
	if err != nil {
		s.logger.Error(ctx, "AuditService.ListAuditLogs", "Failed to list audit logs", zap.Error(err))
		return nil, 0, ErrInternalServer
	}

	response := make([]AuditLogResponse, 0, len(logs))
	for _, log := range logs {
		response = append(response, toAuditLogResponse(log))
	}

	totalCount := int64(0)
	if len(logs) > 0 {
		totalCount = logs[0].TotalCount
	}

	return response, totalCount, nil
}

// GetAuditLogByID retrieves a single audit log by ID
func (s *auditService) GetAuditLogByID(ctx context.Context, id string) (*AuditLogResponse, error) {
	log, err := s.store.GetAuditLogByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAuditLogNotFound
		}
		s.logger.Error(ctx, "AuditService.GetAuditLogByID", "Failed to get audit log", zap.Error(err))
		return nil, ErrInternalServer
	}

	response := toAuditLogResponseFromByID(log)
	return &response, nil
}

// GetAuditStats retrieves audit log statistics
func (s *auditService) GetAuditStats(ctx context.Context) (*AuditLogStatsResponse, error) {
	stats, err := s.store.GetAuditLogStats(ctx)
	if err != nil {
		s.logger.Error(ctx, "AuditService.GetAuditStats", "Failed to get audit stats", zap.Error(err))
		return nil, ErrInternalServer
	}

	return &AuditLogStatsResponse{
		TotalLogs:    stats.TotalLogs,
		ReadCount:    stats.ReadCount,
		CreateCount:  stats.CreateCount,
		UpdateCount:  stats.UpdateCount,
		DeleteCount:  stats.DeleteCount,
		FailureCount: stats.FailureCount,
	}, nil
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
				prevHash = audit.GenesisHash
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
		computedHash := audit.ComputeHash(
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
			log.CreatedAt.Time.UTC(),
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

// Helper functions for type conversions (internal to service)

func toNullAuditActionEnum(action *string) db.NullAuditActionEnum {
	if action == nil {
		return db.NullAuditActionEnum{Valid: false}
	}
	return db.NullAuditActionEnum{
		AuditActionEnum: db.AuditActionEnum(*action),
		Valid:           true,
	}
}

func toPgtypeTimestamp(t *time.Time) pgtype.Timestamp {
	if t == nil {
		return pgtype.Timestamp{Valid: false}
	}
	return pgtype.Timestamp{
		Time:  *t,
		Valid: true,
	}
}

func pgtypeTimestamptzToTime(ts pgtype.Timestamptz) time.Time {
	if !ts.Valid {
		return time.Time{}
	}
	return ts.Time
}

func toAuditLogResponse(log db.ListAuditLogsRow) AuditLogResponse {
	var userEmail string
	if log.UserEmail != nil {
		userEmail = *log.UserEmail
	}

	var employeeName string
	if log.EmployeeName != nil {
		if name, ok := log.EmployeeName.(string); ok {
			employeeName = name
		}
	}

	var clientName string
	if log.ClientName != nil {
		if name, ok := log.ClientName.(string); ok {
			clientName = name
		}
	}

	return AuditLogResponse{
		ID:             log.ID,
		SequenceNumber: log.SequenceNumber,
		UserID:         log.UserID,
		UserEmail:      userEmail,
		EmployeeID:     log.EmployeeID,
		EmployeeName:   employeeName,
		ClientID:       log.ClientID,
		ClientName:     clientName,
		Action:         string(log.Action),
		ResourceType:   log.ResourceType,
		ResourceID:     log.ResourceID,
		OldValue:       util.ParseJSONB(log.OldValue),
		NewValue:       util.ParseJSONB(log.NewValue),
		IPAddress:      log.IpAddress,
		UserAgent:      log.UserAgent,
		RequestID:      log.RequestID,
		Status:         string(log.Status),
		FailureReason:  log.FailureReason,
		PrevHash:       log.PrevHash,
		CurrentHash:    log.CurrentHash,
		CreatedAt:      pgtypeTimestamptzToTime(log.CreatedAt),
	}
}

func toAuditLogResponseFromByID(log db.GetAuditLogByIDRow) AuditLogResponse {
	var userEmail string
	if log.UserEmail != nil {
		userEmail = *log.UserEmail
	}

	var employeeName string
	if log.EmployeeName != nil {
		if name, ok := log.EmployeeName.(string); ok {
			employeeName = name
		}
	}

	var clientName string
	if log.ClientName != nil {
		if name, ok := log.ClientName.(string); ok {
			clientName = name
		}
	}

	return AuditLogResponse{
		ID:             log.ID,
		SequenceNumber: log.SequenceNumber,
		UserID:         log.UserID,
		UserEmail:      userEmail,
		EmployeeID:     log.EmployeeID,
		EmployeeName:   employeeName,
		ClientID:       log.ClientID,
		ClientName:     clientName,
		Action:         string(log.Action),
		ResourceType:   log.ResourceType,
		ResourceID:     log.ResourceID,
		OldValue:       util.ParseJSONB(log.OldValue),
		NewValue:       util.ParseJSONB(log.NewValue),
		IPAddress:      log.IpAddress,
		UserAgent:      log.UserAgent,
		RequestID:      log.RequestID,
		Status:         string(log.Status),
		FailureReason:  log.FailureReason,
		PrevHash:       log.PrevHash,
		CurrentHash:    log.CurrentHash,
		CreatedAt:      pgtypeTimestamptzToTime(log.CreatedAt),
	}
}

func toAuditLogResponseFromFull(log db.AuditLog) AuditLogResponse {
	return AuditLogResponse{
		ID:             log.ID,
		SequenceNumber: log.SequenceNumber,
		UserID:         log.UserID,
		EmployeeID:     log.EmployeeID,
		ClientID:       log.ClientID,
		Action:         string(log.Action),
		ResourceType:   log.ResourceType,
		ResourceID:     log.ResourceID,
		OldValue:       util.ParseJSONB(log.OldValue),
		NewValue:       util.ParseJSONB(log.NewValue),
		IPAddress:      log.IpAddress,
		UserAgent:      log.UserAgent,
		RequestID:      log.RequestID,
		Status:         string(log.Status),
		FailureReason:  log.FailureReason,
		PrevHash:       log.PrevHash,
		CurrentHash:    log.CurrentHash,
		CreatedAt:      pgtypeTimestamptzToTime(log.CreatedAt),
	}
}
