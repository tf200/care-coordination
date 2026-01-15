package audit

import "errors"

var (
	ErrInternalServer   = errors.New("internal server error")
	ErrAuditLogNotFound = errors.New("audit log not found")
	ErrInvalidRequest   = errors.New("invalid request")
	ErrUnauthorized     = errors.New("unauthorized access")
	ErrForbidden        = errors.New("forbidden: admin access required")
)
