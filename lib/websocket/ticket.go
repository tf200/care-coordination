package websocket

import (
	"context"
	"errors"
	"fmt"
	"time"

	"care-cordination/lib/nanoid"

	"github.com/redis/go-redis/v9"
)

var (
	// ErrInvalidTicket is returned when a ticket is invalid or expired
	ErrInvalidTicket = errors.New("invalid or expired ticket")
)

// TicketManager handles one-time WebSocket authentication tickets
// This implements Socket.IO-style auth where JWT is exchanged for a short-lived ticket
type TicketManager struct {
	redis *redis.Client
	ttl   time.Duration
}

// NewTicketManager creates a new TicketManager instance
func NewTicketManager(redisClient *redis.Client, ttl time.Duration) *TicketManager {
	if ttl == 0 {
		ttl = 30 * time.Second // Default TTL
	}
	return &TicketManager{
		redis: redisClient,
		ttl:   ttl,
	}
}

// CreateTicket creates a one-time ticket for the given user ID
// The ticket expires after the configured TTL
func (t *TicketManager) CreateTicket(ctx context.Context, userID string) (string, error) {
	ticket := nanoid.Generate()
	key := t.ticketKey(ticket)

	err := t.redis.Set(ctx, key, userID, t.ttl).Err()
	if err != nil {
		return "", fmt.Errorf("failed to create ticket: %w", err)
	}

	return ticket, nil
}

// ValidateTicket validates and consumes a ticket
// Returns the user ID if valid, or ErrInvalidTicket if invalid/expired
// The ticket is consumed (deleted) immediately after validation
func (t *TicketManager) ValidateTicket(ctx context.Context, ticket string) (string, error) {
	key := t.ticketKey(ticket)

	// GETDEL: Get and delete atomically (one-time use)
	userID, err := t.redis.GetDel(ctx, key).Result()
	if err == redis.Nil {
		return "", ErrInvalidTicket
	}
	if err != nil {
		return "", fmt.Errorf("failed to validate ticket: %w", err)
	}

	return userID, nil
}

// ticketKey returns the Redis key for a ticket
func (t *TicketManager) ticketKey(ticket string) string {
	return "ws:ticket:" + ticket
}
