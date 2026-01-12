package websocket

import (
	"sync"
	"testing"
	"time"

	loggermocks "care-cordination/lib/logger/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// ============================================================
// Test: Hub
// ============================================================

func TestRegisterUnregister(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := loggermocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	hub := NewHub(mockLogger)
	go hub.Run()
	defer hub.Stop()

	// Create a mock client (without actual WebSocket)
	client := &Client{
		hub:    hub,
		UserID: "user-123",
		send:   make(chan *Message, 256),
	}

	// Register the client
	hub.Register(client)

	// Give hub time to process
	time.Sleep(50 * time.Millisecond)

	// Verify client is registered
	hub.mu.RLock()
	clients, exists := hub.clients["user-123"]
	hub.mu.RUnlock()

	assert.True(t, exists)
	assert.Contains(t, clients, client)

	// Unregister the client
	hub.Unregister(client)

	// Give hub time to process
	time.Sleep(50 * time.Millisecond)

	// Verify client is unregistered
	hub.mu.RLock()
	_, exists = hub.clients["user-123"]
	hub.mu.RUnlock()

	assert.False(t, exists)
}

func TestMultipleClientsPerUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := loggermocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	hub := NewHub(mockLogger)
	go hub.Run()
	defer hub.Stop()

	// Create multiple clients for the same user (e.g., multiple browser tabs)
	client1 := &Client{
		hub:    hub,
		UserID: "user-123",
		send:   make(chan *Message, 256),
	}
	client2 := &Client{
		hub:    hub,
		UserID: "user-123",
		send:   make(chan *Message, 256),
	}

	hub.Register(client1)
	hub.Register(client2)

	time.Sleep(50 * time.Millisecond)

	// Both clients should be registered
	hub.mu.RLock()
	clients := hub.clients["user-123"]
	hub.mu.RUnlock()

	assert.Len(t, clients, 2)
	assert.Contains(t, clients, client1)
	assert.Contains(t, clients, client2)

	// Unregister one client
	hub.Unregister(client1)
	time.Sleep(50 * time.Millisecond)

	// Only client2 should remain
	hub.mu.RLock()
	clients = hub.clients["user-123"]
	hub.mu.RUnlock()

	assert.Len(t, clients, 1)
	assert.Contains(t, clients, client2)
}

func TestSendToUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := loggermocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	hub := NewHub(mockLogger)
	go hub.Run()
	defer hub.Stop()

	// Create a client
	client := &Client{
		hub:    hub,
		UserID: "user-123",
		send:   make(chan *Message, 256),
	}

	hub.Register(client)
	time.Sleep(50 * time.Millisecond)

	// Send a message to the user
	msg := &Message{
		Type: MessageTypeNotification,
		Payload: NotificationPayload{
			ID:      "notif-123",
			Title:   "Test Notification",
			Message: "This is a test",
		},
	}

	hub.SendToUser("user-123", msg)

	// Wait for message to be received
	select {
	case received := <-client.send:
		assert.Equal(t, MessageTypeNotification, received.Type)
		payload, ok := received.Payload.(NotificationPayload)
		require.True(t, ok)
		assert.Equal(t, "Test Notification", payload.Title)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for message")
	}
}

func TestSendToUserMultipleClients(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := loggermocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	hub := NewHub(mockLogger)
	go hub.Run()
	defer hub.Stop()

	// Create two clients for same user
	client1 := &Client{
		hub:    hub,
		UserID: "user-123",
		send:   make(chan *Message, 256),
	}
	client2 := &Client{
		hub:    hub,
		UserID: "user-123",
		send:   make(chan *Message, 256),
	}

	hub.Register(client1)
	hub.Register(client2)
	time.Sleep(50 * time.Millisecond)

	// Send message to user
	msg := &Message{
		Type:    MessageTypeNotification,
		Payload: "test",
	}

	hub.SendToUser("user-123", msg)

	// Both clients should receive the message
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		select {
		case <-client1.send:
			// Success
		case <-time.After(100 * time.Millisecond):
			t.Error("client1 didn't receive message")
		}
	}()

	go func() {
		defer wg.Done()
		select {
		case <-client2.send:
			// Success
		case <-time.After(100 * time.Millisecond):
			t.Error("client2 didn't receive message")
		}
	}()

	wg.Wait()
}

func TestSendToNonexistentUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := loggermocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	hub := NewHub(mockLogger)
	go hub.Run()
	defer hub.Stop()

	// Sending to nonexistent user should not panic
	msg := &Message{
		Type:    MessageTypeNotification,
		Payload: "test",
	}

	// This should not panic
	hub.SendToUser("nonexistent-user", msg)
}

func TestBroadcast(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := loggermocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	hub := NewHub(mockLogger)
	go hub.Run()
	defer hub.Stop()

	// Create clients for different users
	client1 := &Client{hub: hub, UserID: "user-1", send: make(chan *Message, 256)}
	client2 := &Client{hub: hub, UserID: "user-2", send: make(chan *Message, 256)}
	client3 := &Client{hub: hub, UserID: "user-3", send: make(chan *Message, 256)}

	hub.Register(client1)
	hub.Register(client2)
	hub.Register(client3)
	time.Sleep(50 * time.Millisecond)

	// Broadcast message
	msg := &Message{
		Type:    MessageTypePing,
		Payload: nil,
	}

	hub.Broadcast(msg)

	// All clients should receive the message
	clients := []*Client{client1, client2, client3}
	for i, client := range clients {
		select {
		case received := <-client.send:
			assert.Equal(t, MessageTypePing, received.Type)
		case <-time.After(100 * time.Millisecond):
			t.Errorf("client %d didn't receive broadcast message", i+1)
		}
	}
}

func TestCountConnections(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := loggermocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	hub := NewHub(mockLogger)
	go hub.Run()
	defer hub.Stop()

	// Initially no connections
	assert.Equal(t, 0, hub.CountConnections())

	// Add clients
	client1 := &Client{hub: hub, UserID: "user-1", send: make(chan *Message, 256)}
	client2 := &Client{hub: hub, UserID: "user-1", send: make(chan *Message, 256)}
	client3 := &Client{hub: hub, UserID: "user-2", send: make(chan *Message, 256)}

	hub.Register(client1)
	hub.Register(client2)
	hub.Register(client3)
	time.Sleep(50 * time.Millisecond)

	assert.Equal(t, 3, hub.CountConnections())

	// Remove one
	hub.Unregister(client1)
	time.Sleep(50 * time.Millisecond)

	assert.Equal(t, 2, hub.CountConnections())
}
