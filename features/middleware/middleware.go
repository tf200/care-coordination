package middleware

import (
	"care-cordination/lib/logger"
	"care-cordination/lib/ratelimit"
	"care-cordination/lib/token"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
	UserIDKey               = "user_id"
)

type Middleware struct {
	tokenMaker  *token.TokenManager
	rateLimiter ratelimit.RateLimiter
	logger      *logger.Logger
}

func NewMiddleware(tokenMaker *token.TokenManager, rateLimiter ratelimit.RateLimiter, logger *logger.Logger) *Middleware {
	return &Middleware{
		tokenMaker:  tokenMaker,
		rateLimiter: rateLimiter,
		logger:      logger,
	}
}
