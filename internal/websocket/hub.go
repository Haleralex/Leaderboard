package websocket

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
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

	// Hash of last broadcast per season (для избежания дублирующих broadcasts)
	lastBroadcastHash map[string]string

	// Callback for periodic updates (called every N seconds with max requested limit per season)
	OnPeriodicUpdate func(seasonLimits map[string]int)

	// Configuration
	broadcastInterval time.Duration
	defaultLimit      int
}

// BroadcastMessage contains the season and leaderboard data to broadcast
type BroadcastMessage struct {
	Season      string
	Leaderboard *leaderboardmodels.LeaderboardResponse
}

// NewHub creates a new Hub instance
func NewHub(ctx context.Context, broadcastInterval time.Duration, defaultLimit int) *Hub {
	return &Hub{
		BroadcastChan:     make(chan *BroadcastMessage, 256),
		Register:          make(chan *Client),
		Unregister:        make(chan *Client),
		Clients:           make(map[string]map[*Client]bool),
		lastBroadcastHash: make(map[string]string),
		ctx:               ctx,
		broadcastInterval: broadcastInterval,
		defaultLimit:      defaultLimit,
	}
}

// Run starts the hub's main loop (must be run in a goroutine)
func (h *Hub) Run() {
	log.Info().Msg("🔌 WebSocket Hub started")

	// Ticker for periodic broadcasts
	ticker := time.NewTicker(h.broadcastInterval)
	defer ticker.Stop()

	log.Info().
		Dur("interval", h.broadcastInterval).
		Int("default_limit", h.defaultLimit).
		Msg("⚙️ Hub configuration loaded")

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

	// DISABLED: Hash check causes issues when data doesn't change but needs to be sent
	// (e.g., initial snapshots, periodic updates, client reconnects)
	// Always broadcast to ensure clients get updates
	/*
		// Calculate hash of the leaderboard data to avoid sending duplicate broadcasts
		hashData, err := json.Marshal(message.Leaderboard)
		if err != nil {
			log.Error().Err(err).Msg("Failed to marshal leaderboard for hash calculation")
			return
		}
		hash := sha256.Sum256(hashData)
		hashStr := hex.EncodeToString(hash[:])

		// Check if this is the same data as last broadcast
		h.mu.Lock()
		lastHash, exists := h.lastBroadcastHash[message.Season]
		if exists && lastHash == hashStr {
			log.Info().
				Str("season", message.Season).
				Str("hash", hashStr[:16]+"...").
				Msg("🔄 Skipping broadcast - data unchanged from last broadcast")
			h.mu.Unlock()
			return
		}
		h.lastBroadcastHash[message.Season] = hashStr
		h.mu.Unlock()

		log.Info().
			Str("season", message.Season).
			Str("hash", hashStr[:16]+"...").
			Bool("data_changed", !exists || lastHash != hashStr).
			Msg("✨ Data changed - proceeding with broadcast")
	*/

	log.Info().
		Str("season", message.Season).
		Msg("✨ Proceeding with broadcast (hash check disabled)")

	// Log first 3 entries from the leaderboard before filtering
	top3Log := "["
	for i := 0; i < 3 && i < len(message.Leaderboard.Entries); i++ {
		entry := message.Leaderboard.Entries[i]
		if i > 0 {
			top3Log += ", "
		}
		top3Log += fmt.Sprintf("{rank:%d, score:%d, name:%s, ts:%v}",
			entry.Rank, entry.Score, entry.UserName, entry.Timestamp)
	}
	top3Log += "]"
	log.Info().
		Str("season", message.Season).
		Int("total_entries", len(message.Leaderboard.Entries)).
		Str("top3_in_hub", top3Log).
		Msg("📊 Hub received leaderboard data")

	// Broadcast to each client with their requested limit
	sentCount := 0
	failedCount := 0

	for client := range clients {
		// Filter entries based on client's requested limit
		filteredEntries := message.Leaderboard.Entries
		if len(filteredEntries) > client.RequestedLimit {
			filteredEntries = filteredEntries[:client.RequestedLimit]
		}

		// Create custom leaderboard response for this client
		clientLeaderboard := *message.Leaderboard // Copy struct
		clientLeaderboard.Entries = filteredEntries

		// Marshal message for this specific client
		jsonData, err := json.Marshal(map[string]interface{}{
			"type":        "leaderboard_update",
			"season":      message.Season,
			"leaderboard": clientLeaderboard,
			"timestamp":   time.Now().Unix(),
		})
		if err != nil {
			log.Error().Err(err).Msg("Failed to marshal broadcast message")
			continue
		}

		log.Info().
			Str("season", message.Season).
			Str("user_id", client.UserID.String()).
			Int("requested_limit", client.RequestedLimit).
			Int("total_entries", len(message.Leaderboard.Entries)).
			Int("filtered_entries", len(filteredEntries)).
			Int("json_size", len(jsonData)).
			Msg("📡 Broadcasting leaderboard update to client")

		select {
		case client.Send <- jsonData:
			sentCount++
			log.Info().
				Str("user_id", client.UserID.String()).
				Str("season", client.Season).
				Int("entries_sent", len(filteredEntries)).
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

// triggerPeriodicUpdates calls callback with max requested limit per season
func (h *Hub) triggerPeriodicUpdates() {
	h.mu.RLock()
	seasonLimits := make(map[string]int)
	totalClients := 0

	// Find max requested limit per season
	for season, clients := range h.Clients {
		if len(clients) > 0 {
			maxLimit := h.defaultLimit // Use default from config
			for client := range clients {
				if client.RequestedLimit > maxLimit {
					maxLimit = client.RequestedLimit
				}
			}
			seasonLimits[season] = maxLimit
			totalClients += len(clients)
		}
	}
	h.mu.RUnlock()

	log.Info().
		Int("active_seasons", len(seasonLimits)).
		Int("total_clients", totalClients).
		Interface("season_limits", seasonLimits).
		Bool("callback_set", h.OnPeriodicUpdate != nil).
		Msg("🔄🔄🔄 triggerPeriodicUpdates CALLED")

	// Call callback if set
	if h.OnPeriodicUpdate != nil && len(seasonLimits) > 0 {
		log.Info().
			Int("seasons", len(seasonLimits)).
			Interface("limits", seasonLimits).
			Msg("✅ Calling OnPeriodicUpdate callback with dynamic limits")
		h.OnPeriodicUpdate(seasonLimits)
	} else {
		if h.OnPeriodicUpdate == nil {
			log.Error().Msg("❌❌❌ OnPeriodicUpdate callback is NIL - NO UPDATES WILL BE SENT!")
		}
		if len(seasonLimits) == 0 {
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

// computeLeaderboardHash вычисляет SHA256 hash от leaderboard entries
// Используется для определения, изменился ли leaderboard
func (h *Hub) computeLeaderboardHash(leaderboard *leaderboardmodels.LeaderboardResponse) string {
	if leaderboard == nil || len(leaderboard.Entries) == 0 {
		return "empty"
	}

	// Сериализуем только entries (без timestamp и прочей метаданных)
	data, err := json.Marshal(leaderboard.Entries)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to marshal leaderboard for hashing")
		return "error"
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
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
