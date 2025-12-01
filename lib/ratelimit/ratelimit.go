package ratelimit

import (
	"context"
	"time"
)

// RateLimiter defines the interface for rate limiting operations
type RateLimiter interface {
	// CheckIPLimit checks if the IP has exceeded the rate limit
	CheckIPLimit(ctx context.Context, ip string) (*LimitResult, error)

	// CheckEmailLimit checks if the email has exceeded the rate limit
	CheckEmailLimit(ctx context.Context, email string) (*LimitResult, error)

	// ResetEmailLimit resets the rate limit for an email (on successful login)
	ResetEmailLimit(ctx context.Context, email string) error

	// Close closes the rate limiter and cleans up resources
	Close() error
}

// LimitResult contains information about a rate limit check
type LimitResult struct {
	Allowed    bool          // Whether the request is allowed
	Limit      int           // Total number of allowed requests
	Remaining  int           // Number of remaining requests
	RetryAfter time.Duration // Duration until the rate limit resets
	ResetAt    time.Time     // Absolute time when the rate limit resets
}

// Config holds configuration for the rate limiter
type Config struct {
	RedisURL       string
	IPLimit        int
	IPWindow       time.Duration
	EmailLimit     int
	EmailWindow    time.Duration
	EnableFallback bool // Use in-memory fallback if Redis fails
}

// NewRateLimiter creates a new rate limiter instance
// It will use Redis-backed rate limiting with optional in-memory fallback
func NewRateLimiter(config *Config) (RateLimiter, error) {
	// Try to create Redis-backed rate limiter
	redisLimiter, err := NewRedisLimiter(config)
	if err != nil {
		// If Redis connection fails and fallback is enabled, use memory limiter
		if config.EnableFallback {
			return NewMemoryLimiter(config), nil
		}
		return nil, err
	}

	return redisLimiter, nil
}
