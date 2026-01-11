package handlers

import (
	"net/http"

	"leaderboard-service/internal/shared/config"
	"leaderboard-service/internal/shared/middleware"
	ws "leaderboard-service/internal/websocket"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for now (configure in production!)
		return true
	},
}

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	hub     *ws.Hub
	jwt     *middleware.JWTMiddleware
	config  *config.Config
	service interface {
		SendInitialSnapshot(season string, requestedLimit int, clientSend chan []byte)
	}
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(
	hub *ws.Hub,
	jwt *middleware.JWTMiddleware,
	cfg *config.Config,
	service interface {
		SendInitialSnapshot(season string, requestedLimit int, clientSend chan []byte)
	},
) *WebSocketHandler {
	return &WebSocketHandler{
		hub:     hub,
		jwt:     jwt,
		config:  cfg,
		service: service,
	}
}

// HandleLeaderboard handles WebSocket connections for leaderboard updates
// ws://localhost:8080/api/v1/ws/leaderboard?season=global&token=JWT
func (h *WebSocketHandler) HandleLeaderboard(w http.ResponseWriter, r *http.Request) {
	// Try to get user ID from context (set by JWT middleware)
	userID, ok := middleware.GetUserIDFromContext(r.Context())

	if !ok {
		// Fallback: try token from query parameter (for browser WebSocket)
		tokenString := r.URL.Query().Get("token")
		if tokenString == "" {
			log.Warn().Msg("WebSocket connection attempt without valid JWT")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Validate token from query param
		log.Debug().Msg("🔑 Validating token from query parameter")
		claims, err := h.jwt.ValidateTokenString(tokenString)
		if err != nil {
			log.Warn().Err(err).Msg("❌ Invalid token from query parameter")
			http.Error(w, "Unauthorized - invalid token", http.StatusUnauthorized)
			return
		}

		userID = claims.UserID
		log.Info().Str("user_id", userID.String()).Msg("✅ Token validated from query parameter")
	}

	// Get season from query parameter
	season := r.URL.Query().Get("season")
	if season == "" {
		season = "global"
	}

	log.Info().
		Str("user_id", userID.String()).
		Str("season", season).
		Str("remote_addr", r.RemoteAddr).
		Msg("🔌 WebSocket connection request")

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to upgrade to WebSocket")
		return
	}

	// Create new client with configuration from config
	clientConfig := ws.ClientConfig{
		WriteWait:      h.config.GetWebSocketWriteWait(),
		PongWait:       h.config.GetWebSocketPongWait(),
		PingPeriod:     h.config.GetWebSocketPingPeriod(),
		MaxMessageSize: h.config.WebSocket.MaxMessageSize,
	}
	client := ws.NewClient(h.hub, conn, userID, season, clientConfig)

	// Register client with hub
	h.hub.Register <- client

	log.Info().
		Str("user_id", userID.String()).
		Str("season", season).
		Str("remote_addr", r.RemoteAddr).
		Msg("🔌 New WebSocket connection established")

	// Send initial leaderboard snapshot to client
	if h.service != nil {
		go h.service.SendInitialSnapshot(season, client.RequestedLimit, client.Send)
	}

	// Start client goroutines
	// Allow collection of memory referenced by the caller by doing all work in new goroutines
	go client.WritePump()
	go client.ReadPump()
}

// HandleStats returns WebSocket hub statistics
func (h *WebSocketHandler) HandleStats(w http.ResponseWriter, r *http.Request) {
	stats := h.hub.GetStats()
	respondJSON(w, stats, http.StatusOK)
}
