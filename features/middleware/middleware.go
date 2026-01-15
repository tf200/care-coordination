package middleware

import (
	db "care-cordination/lib/db/sqlc"
	"care-cordination/lib/logger"
	"care-cordination/lib/ratelimit"
	"care-cordination/lib/token"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
	UserIDKey               = "user_id"
	EmployeeIDKey           = "employee_id"
)

type Middleware struct {
	tokenMaker  token.TokenManager
	rateLimiter ratelimit.RateLimiter
	logger      logger.Logger
	store       *db.Store
	auditLogger AuditLogger
}

func NewMiddleware(
	tokenMaker token.TokenManager,
	rateLimiter ratelimit.RateLimiter,
	logger logger.Logger,
	store *db.Store,
	auditLogger AuditLogger,
) *Middleware {
	return &Middleware{
		tokenMaker:  tokenMaker,
		rateLimiter: rateLimiter,
		logger:      logger,
		store:       store,
		auditLogger: auditLogger,
	}
}
