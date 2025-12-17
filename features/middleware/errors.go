package middleware

import "errors"

var (
	ErrInvalidRequest = errors.New("invalid request")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrForbidden      = errors.New("forbidden")
	ErrInternal       = errors.New("internal server error")

	// Rate limiting errors
	ErrRateLimitExceeded = errors.New("rate limit exceeded, please try again later")
	ErrRateLimitIP       = errors.New("too many login attempts from this IP address, please try again later")
	ErrRateLimitEmail    = errors.New("too many login attempts for this account, please try again later")
)
