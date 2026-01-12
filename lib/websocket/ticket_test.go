package websocket

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================
// Test Helpers
// ============================================================

func setupTestRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return mr, client
}

// ============================================================
// Test: TicketManager
// ============================================================

func TestCreateTicket(t *testing.T) {
	tests := []struct {
		name     string
		userID   string
		ttl      time.Duration
		wantErr  bool
		validate func(t *testing.T, ticket string, mr *miniredis.Miniredis)
	}{
		{
			name:    "success",
			userID:  "user-123",
			ttl:     30 * time.Second,
			wantErr: false,
			validate: func(t *testing.T, ticket string, mr *miniredis.Miniredis) {
				// Ticket should be 20 characters (nanoid default)
				assert.Len(t, ticket, 20)

				// Check ticket exists in Redis
				key := "ws:ticket:" + ticket
				val, err := mr.Get(key)
				require.NoError(t, err)
				assert.Equal(t, "user-123", val)
			},
		},
		{
			name:    "different_users_different_tickets",
			userID:  "user-456",
			ttl:     30 * time.Second,
			wantErr: false,
			validate: func(t *testing.T, ticket string, mr *miniredis.Miniredis) {
				assert.Len(t, ticket, 20)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr, client := setupTestRedis(t)
			defer mr.Close()
			defer client.Close()

			tm := NewTicketManager(client, tt.ttl)
			ctx := context.Background()

			ticket, err := tm.CreateTicket(ctx, tt.userID)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, ticket, mr)
			}
		})
	}
}

func TestValidateTicket(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(mr *miniredis.Miniredis) string // returns ticket
		wantUserID string
		wantErr    bool
	}{
		{
			name: "success_valid_ticket",
			setup: func(mr *miniredis.Miniredis) string {
				ticket := "test-ticket-123"
				mr.Set("ws:ticket:"+ticket, "user-123")
				return ticket
			},
			wantUserID: "user-123",
			wantErr:    false,
		},
		{
			name: "error_ticket_not_found",
			setup: func(mr *miniredis.Miniredis) string {
				return "nonexistent-ticket"
			},
			wantUserID: "",
			wantErr:    true,
		},
		{
			name: "ticket_consumed_on_first_use",
			setup: func(mr *miniredis.Miniredis) string {
				ticket := "one-time-ticket"
				mr.Set("ws:ticket:"+ticket, "user-456")
				return ticket
			},
			wantUserID: "user-456",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr, client := setupTestRedis(t)
			defer mr.Close()
			defer client.Close()

			tm := NewTicketManager(client, 30*time.Second)
			ctx := context.Background()

			ticket := tt.setup(mr)

			userID, err := tm.ValidateTicket(ctx, ticket)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantUserID, userID)

			// Verify ticket was consumed (deleted from Redis)
			exists := mr.Exists("ws:ticket:" + ticket)
			assert.False(t, exists, "ticket should be deleted after validation")
		})
	}
}

func TestTicketIsOneTimeUse(t *testing.T) {
	mr, client := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	tm := NewTicketManager(client, 30*time.Second)
	ctx := context.Background()

	// Create a ticket
	ticket, err := tm.CreateTicket(ctx, "user-123")
	require.NoError(t, err)

	// First validation should succeed
	userID, err := tm.ValidateTicket(ctx, ticket)
	require.NoError(t, err)
	assert.Equal(t, "user-123", userID)

	// Second validation should fail (ticket consumed)
	_, err = tm.ValidateTicket(ctx, ticket)
	require.Error(t, err)
}

func TestTicketExpiration(t *testing.T) {
	mr, client := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	// Use very short TTL
	tm := NewTicketManager(client, 100*time.Millisecond)
	ctx := context.Background()

	// Create a ticket
	ticket, err := tm.CreateTicket(ctx, "user-123")
	require.NoError(t, err)

	// Ticket should be valid immediately
	mr.FastForward(50 * time.Millisecond)
	exists := mr.Exists("ws:ticket:" + ticket)
	assert.True(t, exists)

	// Fast forward past TTL
	mr.FastForward(100 * time.Millisecond)
	exists = mr.Exists("ws:ticket:" + ticket)
	assert.False(t, exists, "ticket should have expired")
}
