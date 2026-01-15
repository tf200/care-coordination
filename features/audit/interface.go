package audit

import "context"

//go:generate mockgen -destination=mocks/mock_audit_service.go -package=mocks care-cordination/features/audit AuditService
type AuditService interface {
	// Query methods
	ListAuditLogs(ctx context.Context, req *ListAuditLogsRequest, limit, offset int32) ([]AuditLogResponse, int64, error)
	GetAuditLogByID(ctx context.Context, id string) (*AuditLogResponse, error)
	GetAuditStats(ctx context.Context) (*AuditLogStatsResponse, error)

	// Verification methods
	VerifyChain(ctx context.Context, startSeq, endSeq int64) (*ChainVerificationResult, error)
	VerifyFullChain(ctx context.Context) (*ChainVerificationResult, error)
}
