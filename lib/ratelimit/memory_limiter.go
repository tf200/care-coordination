package ratelimit

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// MemoryLimiter implements RateLimiter using in-memory storage
// This is a fallback implementation for development or when Redis is unavailable
type MemoryLimiter struct {
	mu          sync.RWMutex
	buckets     map[string]*bucket
	ipLimit     int
	ipWindow    time.Duration
	emailLimit  int
	emailWindow time.Duration
	cleanupDone chan struct{}
}

type bucket struct {
	tokens     int
	lastRefill time.Time
	window     time.Duration
	limit      int
}

// NewMemoryLimiter creates a new in-memory rate limiter
func NewMemoryLimiter(config *Config) *MemoryLimiter {
	ml := &MemoryLimiter{
		buckets:     make(map[string]*bucket),
		ipLimit:     config.IPLimit,
		ipWindow:    config.IPWindow,
		emailLimit:  config.EmailLimit,
		emailWindow: config.EmailWindow,
		cleanupDone: make(chan struct{}),
	}

	// Start cleanup goroutine
	go ml.cleanup()

	return ml
}

// CheckIPLimit checks if the IP has exceeded the rate limit
func (m *MemoryLimiter) CheckIPLimit(ctx context.Context, ip string) (*LimitResult, error) {
	key := fmt.Sprintf("ip:%s", hashKeyMem(ip))
	return m.checkLimit(key, m.ipLimit, m.ipWindow), nil
}

// CheckEmailLimit checks if the email has exceeded the rate limit
func (m *MemoryLimiter) CheckEmailLimit(ctx context.Context, email string) (*LimitResult, error) {
	key := fmt.Sprintf("email:%s", hashKeyMem(email))
	return m.checkLimit(key, m.emailLimit, m.emailWindow), nil
}

// ResetEmailLimit resets the rate limit for an email
func (m *MemoryLimiter) ResetEmailLimit(ctx context.Context, email string) error {
	key := fmt.Sprintf("email:%s", hashKeyMem(email))
	m.mu.Lock()
	delete(m.buckets, key)
	m.mu.Unlock()
	return nil
}

// Close stops the cleanup goroutine
func (m *MemoryLimiter) Close() error {
	close(m.cleanupDone)
	return nil
}

// checkLimit performs the actual rate limit check using token bucket algorithm
func (m *MemoryLimiter) checkLimit(key string, limit int, window time.Duration) *LimitResult {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	b, exists := m.buckets[key]

	if !exists {
		// Create new bucket
		b = &bucket{
			tokens:     limit - 1, // Take one token
			lastRefill: now,
			window:     window,
			limit:      limit,
		}
		m.buckets[key] = b

		return &LimitResult{
			Allowed:    true,
			Limit:      limit,
			Remaining:  b.tokens,
			RetryAfter: 0,
			ResetAt:    now.Add(window),
		}
	}

	// Refill tokens based on elapsed time
	elapsed := now.Sub(b.lastRefill)
	if elapsed >= window {
		// Full refill
		b.tokens = limit
		b.lastRefill = now
	} else {
		// Partial refill (linear refill)
		tokensToAdd := int(float64(limit) * (float64(elapsed) / float64(window)))
		b.tokens = min(b.tokens+tokensToAdd, limit)
		if tokensToAdd > 0 {
			b.lastRefill = now
		}
	}

	// Check if we can take a token
	if b.tokens > 0 {
		b.tokens--
		return &LimitResult{
			Allowed:    true,
			Limit:      limit,
			Remaining:  b.tokens,
			RetryAfter: 0,
			ResetAt:    b.lastRefill.Add(window),
		}
	}

	// Rate limit exceeded
	retryAfter := window - elapsed
	if retryAfter < 0 {
		retryAfter = 0
	}

	return &LimitResult{
		Allowed:    false,
		Limit:      limit,
		Remaining:  0,
		RetryAfter: retryAfter,
		ResetAt:    b.lastRefill.Add(window),
	}
}

// cleanup periodically removes expired buckets
func (m *MemoryLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.mu.Lock()
			now := time.Now()
			for key, b := range m.buckets {
				// Remove buckets that haven't been accessed in 2x their window
				if now.Sub(b.lastRefill) > b.window*2 {
					delete(m.buckets, key)
				}
			}
			m.mu.Unlock()
		case <-m.cleanupDone:
			return
		}
	}
}

// hashKeyMem hashes a key for privacy
func hashKeyMem(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
