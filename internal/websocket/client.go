package websocket

import (
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
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

	// Buffer size for client send channel
	sendBufferSize = 256
)

// Client represents a single WebSocket connection
type Client struct {
	// The hub this client belongs to
	Hub *Hub

	// The WebSocket connection
	Conn *websocket.Conn

	// Buffered channel of outbound messages
	Send chan []byte

	// User ID of the connected client
	UserID uuid.UUID

	// Season this client is subscribed to
	Season string
}

// NewClient creates a new WebSocket client
func NewClient(hub *Hub, conn *websocket.Conn, userID uuid.UUID, season string) *Client {
	return &Client{
		Hub:    hub,
		Conn:   conn,
		Send:   make(chan []byte, sendBufferSize),
		UserID: userID,
		Season: season,
	}
}

// ReadPump pumps messages from the WebSocket connection to the hub
// The application runs ReadPump in a per-connection goroutine
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Warn().Err(err).Str("user_id", c.UserID.String()).Msg("WebSocket unexpected close")
			}
			break
		}

		// We don't expect messages from clients (read-only for now)
		log.Debug().
			Str("user_id", c.UserID.String()).
			Str("message", string(message)).
			Msg("Received message from client (ignoring)")
	}
}

// WritePump pumps messages from the hub to the WebSocket connection
// A goroutine running WritePump is started for each connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			log.Info().
				Str("user_id", c.UserID.String()).
				Str("season", c.Season).
				Int("message_size", len(message)).
				Msg("📤📤📤 WritePump: Sending message to WebSocket")

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Error().Err(err).Msg("❌ WritePump: Failed to get NextWriter")
				return
			}
			w.Write(message)

			// Add queued messages to the current WebSocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				log.Error().Err(err).Msg("❌ WritePump: Failed to close writer")
				return
			}

			log.Info().Msg("✅ WritePump: Message successfully written to WebSocket")

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
