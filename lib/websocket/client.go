package websocket

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

// Client represents a single WebSocket connection
type Client struct {
	// The user ID this client belongs to
	UserID string

	// The hub this client is registered with
	hub *Hub

	// The WebSocket connection
	conn *websocket.Conn

	// Buffered channel of outbound messages
	send chan *Message

	// Handler for client messages (optional)
	messageHandler func(client *Client, msg *ClientMessage)
}

// NewClient creates a new Client instance
func NewClient(hub *Hub, conn *websocket.Conn, userID string) *Client {
	return &Client{
		UserID: userID,
		hub:    hub,
		conn:   conn,
		send:   make(chan *Message, 256),
	}
}

// SetMessageHandler sets a handler for incoming client messages
func (c *Client) SetMessageHandler(handler func(client *Client, msg *ClientMessage)) {
	c.messageHandler = handler
}

// ReadPump pumps messages from the WebSocket connection to the hub
// This should be run in a goroutine
func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// Log unexpected close errors
			}
			break
		}

		// Parse and handle client message
		if c.messageHandler != nil {
			var clientMsg ClientMessage
			if err := json.Unmarshal(message, &clientMsg); err == nil {
				c.messageHandler(c, &clientMsg)
			}
		}
	}
}

// WritePump pumps messages from the hub to the WebSocket connection
// This should be run in a goroutine
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Write JSON message
			if err := c.conn.WriteJSON(message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// SendMessage sends a message to this client
func (c *Client) SendMessage(msg *Message) {
	select {
	case c.send <- msg:
	default:
		// Buffer full, message dropped
	}
}

// SetSendChannel sets the send channel (for testing)
func (c *Client) SetSendChannel(ch chan *Message) {
	c.send = ch
}

// GetSendChannel returns the send channel (for testing)
func (c *Client) GetSendChannel() chan *Message {
	return c.send
}
