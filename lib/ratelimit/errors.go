package ratelimit

import "errors"

var (
	// ErrRateLimitExceeded is returned when rate limit is exceeded
	ErrRateLimitExceeded = errors.New("rate limit exceeded")

	// ErrRateLimitIP is returned when IP-based rate limit is exceeded
	ErrRateLimitIP = errors.New(
		"too many login attempts from this IP address, please try again later",
	)

	// ErrRateLimitEmail is returned when email-based rate limit is exceeded
	ErrRateLimitEmail = errors.New(
		"too many login attempts for this account, please try again later",
	)

	// ErrRedisConnection is returned when Redis connection fails
	ErrRedisConnection = errors.New("rate limiter temporarily unavailable")
)
