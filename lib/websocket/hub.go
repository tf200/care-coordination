package websocket

import (
	"context"
	"sync"

	"care-cordination/lib/logger"

	"go.uber.org/zap"
)

// Hub maintains the set of active clients and broadcasts messages to clients
type Hub struct {
	// Map of userID -> set of connections (user can have multiple tabs/devices)
	clients map[string]map[*Client]bool

	// Channel for broadcasting messages to specific users
	broadcast chan *BroadcastMessage

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Mutex for thread-safe access to clients map
	mu sync.RWMutex

	// Logger
	logger logger.Logger

	// Worker done channel for graceful shutdown (used in tests)
	workerDone chan struct{}
}

// BroadcastMessage contains the message and target user
type BroadcastMessage struct {
	UserID  string   // Target user ID (empty string = broadcast to all)
	Message *Message // The message to send
}

// NewHub creates a new Hub instance
func NewHub(logger logger.Logger) *Hub {
	return &Hub{
		clients:    make(map[string]map[*Client]bool),
		broadcast:  make(chan *BroadcastMessage, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		logger:     logger,
		workerDone: make(chan struct{}),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastMessage(message)

		case <-h.workerDone:
			return
		}
	}
}

// registerClient adds a client to the hub
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[client.UserID] == nil {
		h.clients[client.UserID] = make(map[*Client]bool)
	}
	h.clients[client.UserID][client] = true

	h.logger.Info(context.Background(), "WebSocket", "Client registered",
		zap.String("userID", client.UserID),
		zap.Int("userConnections", len(h.clients[client.UserID])),
	)
}

// unregisterClient removes a client from the hub
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.clients[client.UserID]; ok {
		if _, ok := clients[client]; ok {
			delete(clients, client)
			close(client.send)

			// Clean up empty user entry
			if len(clients) == 0 {
				delete(h.clients, client.UserID)
			}

			h.logger.Info(context.Background(), "WebSocket", "Client unregistered",
				zap.String("userID", client.UserID),
			)
		}
	}
}

// broadcastMessage sends a message to specific user or all users
func (h *Hub) broadcastMessage(msg *BroadcastMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if msg.UserID != "" {
		// Send to specific user (all their connections)
		if clients, ok := h.clients[msg.UserID]; ok {
			for client := range clients {
				select {
				case client.send <- msg.Message:
				default:
					// Client's send buffer is full, close connection
					close(client.send)
					delete(clients, client)
				}
			}
		}
	} else {
		// Broadcast to all connected clients
		for _, clients := range h.clients {
			for client := range clients {
				select {
				case client.send <- msg.Message:
				default:
					close(client.send)
					delete(clients, client)
				}
			}
		}
	}
}

// Register adds a client to the hub
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister removes a client from the hub
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// SendToUser sends a message to a specific user
func (h *Hub) SendToUser(userID string, message *Message) {
	h.broadcast <- &BroadcastMessage{
		UserID:  userID,
		Message: message,
	}
}

// Broadcast sends a message to all connected clients
func (h *Hub) Broadcast(message *Message) {
	h.broadcast <- &BroadcastMessage{
		UserID:  "",
		Message: message,
	}
}

// GetConnectedUserCount returns the number of connected users
func (h *Hub) GetConnectedUserCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// IsUserConnected checks if a user has any active connections
func (h *Hub) IsUserConnected(userID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.clients[userID]
	return ok
}

// CountConnections returns the total number of active connections
func (h *Hub) CountConnections() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	count := 0
	for _, clients := range h.clients {
		count += len(clients)
	}
	return count
}

// Stop stops the hub's main loop (for testing)
func (h *Hub) Stop() {
	close(h.workerDone)
}
