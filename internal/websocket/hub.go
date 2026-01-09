package websocket

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	leaderboardmodels "leaderboard-service/internal/leaderboard/models"
	"github.com/rs/zerolog/log"
)

// Hub maintains the set of active clients and broadcasts messages to clients
type Hub struct {
	// Registered clients per season
	Clients map[string]map[*Client]bool

	// Broadcast message to all clients in a season
	BroadcastChan chan *BroadcastMessage

	// Register requests from clients
	Register chan *Client

	// Unregister requests from clients
	Unregister chan *Client

	// Mutex for thread-safe access
	mu sync.RWMutex

	// Context for graceful shutdown
	ctx context.Context

	// Callback for periodic updates (called every N seconds)
	OnPeriodicUpdate func(seasons []string)
}

// BroadcastMessage contains the season and leaderboard data to broadcast
type BroadcastMessage struct {
	Season      string
	Leaderboard *leaderboardmodels.LeaderboardResponse
}

// NewHub creates a new Hub instance
func NewHub(ctx context.Context) *Hub {
	return &Hub{
		BroadcastChan: make(chan *BroadcastMessage, 256),
		Register:      make(chan *Client),
		Unregister:    make(chan *Client),
		Clients:       make(map[string]map[*Client]bool),
		ctx:           ctx,
	}
}

// Run starts the hub's main loop (must be run in a goroutine)
func (h *Hub) Run() {
	log.Info().Msg("🔌 WebSocket Hub started")

	// Ticker for periodic broadcasts (real-time updates every 3 seconds)
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case client := <-h.Register:
			h.registerClient(client)

		case client := <-h.Unregister:
			h.unregisterClient(client)

		case message := <-h.BroadcastChan:
			h.broadcastToSeason(message)

		case <-ticker.C:
			h.triggerPeriodicUpdates()

		case <-h.ctx.Done():
			log.Info().Msg("🛑 WebSocket Hub shutting down")
			h.closeAllClients()
			return
		}
	}
}

// registerClient registers a new client
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.Clients[client.Season] == nil {
		h.Clients[client.Season] = make(map[*Client]bool)
	}
	h.Clients[client.Season][client] = true

	log.Info().
		Str("season", client.Season).
		Str("user_id", client.UserID.String()).
		Int("total_clients", h.getTotalClients()).
		Msg("✅ WebSocket client connected")
}

// unregisterClient unregisters a client
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.Clients[client.Season]; ok {
		if _, exists := clients[client]; exists {
			delete(clients, client)
			close(client.Send)

			// Clean up empty season maps
			if len(clients) == 0 {
				delete(h.Clients, client.Season)
			}

			log.Info().
				Str("season", client.Season).
				Str("user_id", client.UserID.String()).
				Int("total_clients", h.getTotalClients()).
				Msg("❌ WebSocket client disconnected")
		}
	}
}

// broadcastToSeason sends a message to all clients in a specific season
func (h *Hub) broadcastToSeason(message *BroadcastMessage) {
	h.mu.RLock()
	clients := h.Clients[message.Season]
	clientCount := len(clients)
	h.mu.RUnlock()

	log.Info().
		Str("season", message.Season).
		Int("clients", clientCount).
		Int("entries", len(message.Leaderboard.Entries)).
		Msg("📤 broadcastToSeason called")

	if clientCount == 0 {
		log.Warn().Str("season", message.Season).Msg("⚠️ No clients connected for this season")
		return
	}

	// Marshal message once
	jsonData, err := json.Marshal(map[string]interface{}{
		"type":        "leaderboard_update",
		"season":      message.Season,
		"leaderboard": message.Leaderboard,
		"timestamp":   time.Now().Unix(),
	})

	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal broadcast message")
		return
	}

	log.Info().
		Str("season", message.Season).
		Int("clients", clientCount).
		Int("entries", len(message.Leaderboard.Entries)).
		Int("json_size", len(jsonData)).
		Msg("📡 Broadcasting leaderboard update to clients")

	// Broadcast to all clients in parallel
	sentCount := 0
	failedCount := 0
	for client := range clients {
		select {
		case client.Send <- jsonData:
			sentCount++
			log.Info().
				Str("user_id", client.UserID.String()).
				Str("season", client.Season).
				Int("message_size", len(jsonData)).
				Msg("✅ Message queued to client send channel")
		default:
			// Client's send channel is full, close it
			failedCount++
			h.mu.Lock()
			close(client.Send)
			delete(clients, client)
			h.mu.Unlock()
			log.Warn().
				Str("season", client.Season).
				Str("user_id", client.UserID.String()).
				Msg("⚠️ Client send buffer full, disconnecting")
		}
	}

	log.Info().
		Str("season", message.Season).
		Int("sent", sentCount).
		Int("failed", failedCount).
		Int("total", clientCount).
		Msg("✅ Broadcast complete")
}

// Broadcast sends a leaderboard update to all clients in a season
func (h *Hub) Broadcast(season string, leaderboard *leaderboardmodels.LeaderboardResponse) {
	log.Info().
		Str("season", season).
		Int("entries", len(leaderboard.Entries)).
		Msg("🔔 Hub.Broadcast() called")

	select {
	case h.BroadcastChan <- &BroadcastMessage{
		Season:      season,
		Leaderboard: leaderboard,
	}:
		log.Info().Str("season", season).Msg("✅ Message queued to BroadcastChan")
	default:
		log.Warn().Str("season", season).Msg("⚠️ Broadcast channel full, dropping message")
	}
}

// getTotalClients returns the total number of connected clients
func (h *Hub) getTotalClients() int {
	count := 0
	for _, clients := range h.Clients {
		count += len(clients)
	}
	return count
}

// logStats logs connection statistics
func (h *Hub) logStats() {
	h.mu.RLock()
	defer h.mu.RUnlock()

	totalClients := h.getTotalClients()
	if totalClients > 0 {
		log.Debug().
			Int("total_clients", totalClients).
			Int("seasons", len(h.Clients)).
			Msg("📊 WebSocket stats")
	}
}

// triggerPeriodicUpdates calls callback for each active season
func (h *Hub) triggerPeriodicUpdates() {
	h.mu.RLock()
	activeSeasons := make([]string, 0, len(h.Clients))
	totalClients := 0
	for season, clients := range h.Clients {
		if len(clients) > 0 {
			activeSeasons = append(activeSeasons, season)
			totalClients += len(clients)
		}
	}
	h.mu.RUnlock()

	log.Info().
		Int("active_seasons", len(activeSeasons)).
		Int("total_clients", totalClients).
		Bool("callback_set", h.OnPeriodicUpdate != nil).
		Msg("🔄🔄🔄 triggerPeriodicUpdates CALLED")

	// Call callback if set
	if h.OnPeriodicUpdate != nil && len(activeSeasons) > 0 {
		log.Info().
			Int("seasons", len(activeSeasons)).
			Strs("season_list", activeSeasons).
			Msg("✅ Calling OnPeriodicUpdate callback")
		h.OnPeriodicUpdate(activeSeasons)
	} else {
		if h.OnPeriodicUpdate == nil {
			log.Error().Msg("❌❌❌ OnPeriodicUpdate callback is NIL - NO UPDATES WILL BE SENT!")
		}
		if len(activeSeasons) == 0 {
			log.Warn().Msg("⚠️ No active seasons with clients")
		}
	}
}

// closeAllClients closes all client connections
func (h *Hub) closeAllClients() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for season, clients := range h.Clients {
		for client := range clients {
			close(client.Send)
			delete(clients, client)
		}
		delete(h.Clients, season)
	}

	log.Info().Msg("All WebSocket clients closed")
}

// GetStats returns current hub statistics
func (h *Hub) GetStats() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	seasonStats := make(map[string]int)
	for season, clients := range h.Clients {
		seasonStats[season] = len(clients)
	}

	return map[string]interface{}{
		"total_clients": h.getTotalClients(),
		"seasons":       seasonStats,
	}
}
