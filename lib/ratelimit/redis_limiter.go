package ratelimit

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
)

// RedisLimiter implements RateLimiter using Redis
type RedisLimiter struct {
	client      *redis.Client
	limiter     *redis_rate.Limiter
	ipLimit     int
	ipWindow    time.Duration
	emailLimit  int
	emailWindow time.Duration
}

// NewRedisLimiter creates a new Redis-backed rate limiter
func NewRedisLimiter(config *Config) (*RedisLimiter, error) {
	// Parse Redis URL and create client
	opts, err := redis.ParseURL(config.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid Redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisLimiter{
		client:      client,
		limiter:     redis_rate.NewLimiter(client),
		ipLimit:     config.IPLimit,
		ipWindow:    config.IPWindow,
		emailLimit:  config.EmailLimit,
		emailWindow: config.EmailWindow,
	}, nil
}

// CheckIPLimit checks if the IP has exceeded the rate limit
func (r *RedisLimiter) CheckIPLimit(ctx context.Context, ip string) (*LimitResult, error) {
	key := fmt.Sprintf("ratelimit:ip:%s", hashKey(ip))
	return r.checkLimit(ctx, key, r.ipLimit, r.ipWindow)
}

// CheckEmailLimit checks if the email has exceeded the rate limit
func (r *RedisLimiter) CheckEmailLimit(ctx context.Context, email string) (*LimitResult, error) {
	key := fmt.Sprintf("ratelimit:email:%s", hashKey(email))
	return r.checkLimit(ctx, key, r.emailLimit, r.emailWindow)
}

// ResetEmailLimit resets the rate limit for an email
func (r *RedisLimiter) ResetEmailLimit(ctx context.Context, email string) error {
	key := fmt.Sprintf("ratelimit:email:%s", hashKey(email))
	return r.client.Del(ctx, key).Err()
}

// Close closes the Redis client connection
func (r *RedisLimiter) Close() error {
	return r.client.Close()
}

// checkLimit performs the actual rate limit check
func (r *RedisLimiter) checkLimit(
	ctx context.Context,
	key string,
	limit int,
	window time.Duration,
) (*LimitResult, error) {
	// Use redis_rate sliding window algorithm
	result, err := r.limiter.Allow(ctx, key, redis_rate.Limit{
		Rate:   limit,
		Burst:  limit,
		Period: window,
	})
	if err != nil {
		return nil, fmt.Errorf("rate limit check failed: %w", err)
	}

	// Calculate retry after duration
	retryAfter := time.Duration(0)
	allowed := result.Allowed > 0
	if !allowed {
		retryAfter = result.RetryAfter
	}

	return &LimitResult{
		Allowed:    allowed,
		Limit:      limit,
		Remaining:  result.Remaining,
		RetryAfter: retryAfter,
		ResetAt:    time.Now().Add(retryAfter),
	}, nil
}

// hashKey hashes a key using SHA-256 for privacy (GDPR compliance)
func hashKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}
