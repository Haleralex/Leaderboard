package websocket

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

// Client configuration
type ClientConfig struct {
	WriteWait      time.Duration
	PongWait       time.Duration
	PingPeriod     time.Duration
	MaxMessageSize int64
}

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

	// Requested limit - how many entries client wants (updated dynamically)
	RequestedLimit int

	// Configuration
	config ClientConfig
}

// NewClient creates a new WebSocket client
func NewClient(hub *Hub, conn *websocket.Conn, userID uuid.UUID, season string, config ClientConfig) *Client {
	return &Client{
		Hub:            hub,
		Conn:           conn,
		Send:           make(chan []byte, 256),
		UserID:         userID,
		Season:         season,
		RequestedLimit: hub.defaultLimit,
		config:         config,
	}
}

// ReadPump pumps messages from the WebSocket connection to the hub
// The application runs ReadPump in a per-connection goroutine
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(c.config.MaxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(c.config.PongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(c.config.PongWait))
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

		// Parse client messages
		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err == nil {
			if msgType, ok := msg["type"].(string); ok && msgType == "update_limit" {
				if limit, ok := msg["limit"].(float64); ok {
					c.RequestedLimit = int(limit)
					log.Info().
						Str("user_id", c.UserID.String()).
						Int("new_limit", c.RequestedLimit).
						Msg("📊 Client updated requested limit")
				}
			}
		}
	}
}

// WritePump pumps messages from the hub to the WebSocket connection
// A goroutine running WritePump is started for each connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(c.config.PingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(c.config.WriteWait))
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
			c.Conn.SetWriteDeadline(time.Now().Add(c.config.WriteWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
