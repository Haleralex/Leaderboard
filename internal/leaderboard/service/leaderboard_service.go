package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"leaderboard-service/internal/leaderboard/models"
	"leaderboard-service/internal/shared/config"
	"leaderboard-service/internal/shared/database"
	"leaderboard-service/internal/shared/repository"
	ws "leaderboard-service/internal/websocket"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

const (
	redisLeaderboardPrefix = "leaderboard:"
	redisUserScorePrefix   = "user_score:"
)

// LeaderboardService handles leaderboard operations
type LeaderboardService struct {
	scoreRepo repository.ScoreRepository
	userRepo  repository.UserRepository
	redis     *database.RedisClient
	hub       BroadcastHub // WebSocket hub for real-time updates
	config    *config.Config
}

// BroadcastHub interface for WebSocket broadcasting
type BroadcastHub interface {
	Broadcast(season string, leaderboard *models.LeaderboardResponse)
}

// NewLeaderboardService creates a new leaderboard service
func NewLeaderboardService(
	scoreRepo repository.ScoreRepository,
	userRepo repository.UserRepository,
	redis *database.RedisClient,
	cfg *config.Config,
) *LeaderboardService {
	return &LeaderboardService{
		scoreRepo: scoreRepo,
		userRepo:  userRepo,
		redis:     redis,
		hub:       nil, // Will be set later via SetHub
		config:    cfg,
	}
}

// SetHub sets the WebSocket hub for broadcasting
func (s *LeaderboardService) SetHub(hub BroadcastHub) {
	s.hub = hub
	if hub != nil {
		log.Info().Msg("✅ WebSocket Hub connected to LeaderboardService")

		// Cast to concrete Hub type to set callback
		if concreteHub, ok := hub.(*ws.Hub); ok {
			log.Info().Msg("✅ Successfully cast to *ws.Hub - setting callback")
			concreteHub.OnPeriodicUpdate = s.handlePeriodicUpdates
			log.Info().
				Bool("callback_set", concreteHub.OnPeriodicUpdate != nil).
				Msg("✅✅✅ Periodic updates CALLBACK SET (every 3 seconds)")
		} else {
			log.Error().Msg("❌❌❌ FAILED to cast hub to *ws.Hub - NO PERIODIC UPDATES!")
		}
	} else {
		log.Warn().Msg("⚠️ WebSocket Hub is nil!")
	}
}

// SubmitScore submits or updates a user's score using GORM
func (s *LeaderboardService) SubmitScore(ctx context.Context, userID uuid.UUID, req *models.SubmitScoreRequest) (*models.Score, error) {
	season := req.Season
	if season == "" {
		season = "global"
	}

	// 1. Базовая валидация (используем config)
	if req.Score < s.config.Validation.MinScore {
		return nil, fmt.Errorf("score cannot be less than %d", s.config.Validation.MinScore)
	}
	if req.Score > s.config.Validation.MaxScore {
		return nil, fmt.Errorf("score exceeds maximum allowed value of %d", s.config.Validation.MaxScore)
	}

	// DISABLED: Redis score check disabled - using PostgreSQL as single source of truth
	// Database handles score improvement check through unique constraint and timestamp
	/*
		// 2. Проверяем текущий score в Redis (если доступен)
		if s.redis != nil {
			key := redisLeaderboardPrefix + season
			currentScore, err := s.redis.Client.ZScore(ctx, key, userID.String()).Result()

			if err == nil {
				// Обновляем только если новый score ЛУЧШЕ
				if req.Score <= int64(currentScore) {
					log.Info().
						Str("user_id", userID.String()).
						Int64("current", int64(currentScore)).
						Int64("new", req.Score).
						Msg("⏭️ Score not improved, skipping update")
					return nil, fmt.Errorf("score not improved: current=%d, new=%d", int64(currentScore), req.Score)
				}
			}
		}
	*/

	// 3. Создаём score объект
	score := models.Score{
		UserID:   userID,
		Score:    req.Score,
		Season:   season,
		Metadata: req.Metadata,
	}

	// 4. Сохраняем в базу данных (синхронно для надежности)
	if err := s.scoreRepo.Upsert(ctx, &score); err != nil {
		return nil, err
	}

	log.Info().
		Str("source", "GORM").
		Str("user_id", userID.String()).
		Int64("score", req.Score).
		Str("season", season).
		Msg("✅ Score saved to database")

	// DISABLED: Redis cache sync disabled - using PostgreSQL as single source of truth
	// Redis caching causes stale data issues with real-time WebSocket updates
	/*
		// 5. Обновляем Redis cache (синхронно для консистентности)
		if s.redis != nil {
			log.Debug().Str("action", "cache_update").Str("season", season).Msg("Updating Redis cache")
			s.updateRedisCache(ctx, userID, season, req.Score)
		} else {
			log.Debug().Msg("Redis not available, skipping cache update")
		}
	*/

	// 6. Broadcast к WebSocket клиентам (async, не блокируем ответ)
	if s.hub != nil {
		log.Info().Str("season", season).Msg("📡 Triggering WebSocket broadcast...")
		go s.broadcastLeaderboardUpdate(context.Background(), season)
	} else {
		log.Debug().Msg("Hub not available, skipping broadcast")
	}

	return &score, nil
}

// GetLeaderboard retrieves the leaderboard with pagination using GORM
func (s *LeaderboardService) GetLeaderboard(ctx context.Context, query *models.LeaderboardQuery) (*models.LeaderboardResponse, error) {
	season := query.Season
	if season == "" {
		season = "global"
	}

	// DISABLED: Redis cache causes stale data issues with WebSocket real-time updates
	// Always fetch from PostgreSQL to ensure fresh data
	// Redis sorted sets don't preserve order for same scores, causing missing entries
	/* skipCache := false

	if s.redis != nil && !skipCache {
		log.Debug().Str("source", "Redis").Str("season", season).Msg("Attempting to fetch leaderboard from Redis cache")
		entries, err := s.getLeaderboardFromRedis(ctx, season, query)
		if err == nil && len(entries) > 0 {
			log.Info().Str("source", "Redis").Str("season", season).Int("entries", len(entries)).Msg("✓ Leaderboard served from Redis cache")
			return s.buildResponse(entries, query), nil
		}
		log.Debug().Str("season", season).Err(err).Msg("Redis cache miss or empty")
	} else if skipCache {
		log.Debug().Int("limit", query.Limit).Msg("Skipping Redis cache for large request (limit > 100)")
	} else {
		log.Debug().Msg("Redis not available, skipping cache lookup")
	} */

	// Fetch directly from PostgreSQL (single source of truth)
	log.Info().Str("source", "PostgreSQL").Str("season", season).Msg("Fetching leaderboard from database")
	entries, totalCount, err := s.getLeaderboardFromDB(ctx, season, query)
	if err != nil {
		log.Error().Err(err).Str("season", season).Msg("Failed to fetch leaderboard from database")
		return nil, fmt.Errorf("failed to get leaderboard: %w", err)
	}

	log.Info().
		Str("source", "PostgreSQL").
		Str("season", season).
		Int("entries", len(entries)).
		Int64("total", totalCount).
		Msg("✓ Leaderboard loaded from database")

	// DISABLED: Redis caching disabled to ensure fresh data from PostgreSQL
	// Redis cache causes stale data issues - always fetch from PostgreSQL instead
	/*
		// Cache in Redis for next request (if available)
		if s.redis != nil {
			log.Debug().Str("action", "cache_store").Str("season", season).Int("entries", len(entries)).Msg("Caching leaderboard in Redis asynchronously")
			go s.cacheLeaderboardInRedis(context.Background(), season, entries)
		} else {
			log.Debug().Msg("Redis not available, skipping cache storage")
		}
	*/

	response := s.buildResponse(entries, query)
	response.TotalCount = totalCount

	return response, nil
}

// getLeaderboardFromRedis fetches leaderboard from Redis using sorted sets
func (s *LeaderboardService) getLeaderboardFromRedis(ctx context.Context, season string, query *models.LeaderboardQuery) ([]models.LeaderboardEntry, error) {
	key := redisLeaderboardPrefix + season

	// Calculate offset and limit for pagination
	offset := int64(query.Page * query.Limit)
	limit := int64(query.Limit)

	// Redis sorted set: ZREVRANGE for descending, ZRANGE for ascending
	var members []redis.Z
	var err error

	if query.SortOrder == "asc" {
		members, err = s.redis.Client.ZRangeWithScores(ctx, key, offset, offset+limit-1).Result()
	} else {
		members, err = s.redis.Client.ZRevRangeWithScores(ctx, key, offset, offset+limit-1).Result()
	}

	if err != nil {
		return nil, err
	}

	if len(members) == 0 {
		return nil, fmt.Errorf("cache miss")
	}

	// Build entries from Redis data
	entries := make([]models.LeaderboardEntry, 0, len(members))
	for i, member := range members {
		userID, _ := uuid.Parse(member.Member.(string))

		// Fetch user name from cache or DB
		userName, _ := s.getUserName(ctx, userID)

		entries = append(entries, models.LeaderboardEntry{
			Rank:      int(offset) + i + 1,
			UserID:    userID,
			UserName:  userName,
			Score:     int64(member.Score),
			Season:    season,
			Timestamp: time.Now(), // Approximate; full data is in DB
		})
	}

	return entries, nil
}

// getLeaderboardFromDB fetches leaderboard from PostgreSQL using repository
func (s *LeaderboardService) getLeaderboardFromDB(ctx context.Context, season string, query *models.LeaderboardQuery) ([]models.LeaderboardEntry, int64, error) {
	offset := query.Page * query.Limit

	// Use repository to fetch leaderboard
	entries, totalCount, err := s.scoreRepo.GetLeaderboard(ctx, season, query.Limit, offset, query.SortOrder)
	if err != nil {
		return nil, 0, err
	}

	return entries, totalCount, nil
}

// updateRedisCache updates the Redis sorted set with a new score
func (s *LeaderboardService) updateRedisCache(ctx context.Context, userID uuid.UUID, season string, score int64) {
	key := redisLeaderboardPrefix + season

	// Add to sorted set (score is the sort value, userID is the member)
	err := s.redis.Client.ZAdd(ctx, key, redis.Z{
		Score:  float64(score),
		Member: userID.String(),
	}).Err()

	if err != nil {
		log.Warn().Err(err).Str("key", key).Msg("Failed to update Redis cache")
		return
	}

	// Set expiry on the key
	s.redis.Client.Expire(ctx, key, s.config.GetCacheLeaderboardTTL())
	log.Debug().Str("source", "Redis").Str("key", key).Str("user_id", userID.String()).Int64("score", score).Msg("✓ Redis cache updated")
}

// cacheLeaderboardInRedis caches the entire leaderboard in Redis
func (s *LeaderboardService) cacheLeaderboardInRedis(ctx context.Context, season string, entries []models.LeaderboardEntry) {
	if len(entries) == 0 {
		return
	}

	key := redisLeaderboardPrefix + season

	// Build sorted set members
	members := make([]redis.Z, len(entries))
	for i, entry := range entries {
		members[i] = redis.Z{
			Score:  float64(entry.Score),
			Member: entry.UserID.String(),
		}
	}

	// Add all members to sorted set
	err := s.redis.Client.ZAdd(ctx, key, members...).Err()
	if err != nil {
		log.Warn().Err(err).Str("key", key).Msg("Failed to cache leaderboard in Redis")
		return
	}

	// Set expiry
	s.redis.Client.Expire(ctx, key, s.config.GetCacheLeaderboardTTL())
	log.Debug().Str("source", "Redis").Str("key", key).Int("entries", len(entries)).Dur("ttl", s.config.GetCacheLeaderboardTTL()).Msg("✓ Leaderboard cached in Redis")
}

// getUserName fetches a user's name (with caching) using GORM
func (s *LeaderboardService) getUserName(ctx context.Context, userID uuid.UUID) (string, error) {
	cacheKey := redisUserScorePrefix + userID.String()

	// Try cache first (if Redis available)
	if s.redis != nil {
		cachedName, err := s.redis.Client.Get(ctx, cacheKey).Result()
		if err == nil {
			log.Debug().Str("source", "Redis").Str("user_id", userID.String()).Str("name", cachedName).Msg("User name from cache")
			return cachedName, nil
		}
	}

	// Fetch from DB using repository
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		// If user not found, return "Unknown" instead of error
		if err.Error() == "user not found" {
			return "Unknown", nil
		}
		return "", err
	}

	log.Debug().Str("source", "Repository").Str("user_id", userID.String()).Str("name", user.Name).Msg("User name from database")

	// Cache for 10 minutes (if Redis available)
	if s.redis != nil {
		s.redis.Client.Set(ctx, cacheKey, user.Name, 10*time.Minute)
		log.Debug().Str("action", "cache_store").Str("key", cacheKey).Msg("User name cached in Redis")
	}

	return user.Name, nil
}

// buildResponse constructs the leaderboard response
func (s *LeaderboardService) buildResponse(entries []models.LeaderboardEntry, query *models.LeaderboardQuery) *models.LeaderboardResponse {
	hasNext := len(entries) == query.Limit

	var nextCursor string
	if hasNext && len(entries) > 0 {
		// Cursor-based pagination: encode the last score and rank
		lastEntry := entries[len(entries)-1]
		nextCursor = fmt.Sprintf("%d:%d", lastEntry.Rank, lastEntry.Score)
	}

	return &models.LeaderboardResponse{
		Entries:    entries,
		TotalCount: 0, // Will be set by caller if from DB
		Page:       query.Page,
		Limit:      query.Limit,
		HasNext:    hasNext,
		NextCursor: nextCursor,
	}
}

// GetUserRank gets a specific user's rank and score
func (s *LeaderboardService) GetUserRank(ctx context.Context, userID uuid.UUID, season string) (*models.LeaderboardEntry, error) {
	if season == "" {
		season = "global"
	}

	// Fetch all leaderboard entries (we need to calculate rank)
	// For large leaderboards, consider implementing a dedicated repository method
	entries, _, err := s.scoreRepo.GetLeaderboard(ctx, season, 100000, 0, "desc")
	if err != nil {
		return nil, fmt.Errorf("failed to get leaderboard: %w", err)
	}

	// Find the user in the leaderboard
	for _, entry := range entries {
		if entry.UserID == userID {
			log.Info().
				Str("source", "Repository").
				Str("user_id", userID.String()).
				Int("rank", entry.Rank).
				Int64("score", entry.Score).
				Str("season", season).
				Msg("User rank retrieved from database")
			return &entry, nil
		}
	}

	return nil, fmt.Errorf("user not found in leaderboard")
}

// broadcastLeaderboardUpdate fetches and broadcasts the current leaderboard
func (s *LeaderboardService) broadcastLeaderboardUpdate(ctx context.Context, season string) {
	s.broadcastLeaderboardUpdateWithLimit(ctx, season, 10000)
}

// broadcastLeaderboardUpdateWithLimit fetches and broadcasts the current leaderboard with custom limit
func (s *LeaderboardService) broadcastLeaderboardUpdateWithLimit(ctx context.Context, season string, limit int) {
	log.Info().
		Str("season", season).
		Int("limit", limit).
		Msg("🔔 broadcastLeaderboardUpdateWithLimit called")

	if s.hub == nil {
		log.Error().Msg("❌ Hub is nil in broadcastLeaderboardUpdateWithLimit!")
		return
	}

	query := &models.LeaderboardQuery{
		Season:    season,
		Limit:     limit, // Use dynamic limit from clients
		Page:      0,
		SortOrder: "desc",
	}

	leaderboard, err := s.GetLeaderboard(ctx, query)
	if err != nil {
		log.Warn().Err(err).Str("season", season).Msg("Failed to fetch leaderboard for broadcast")
		return
	}

	log.Info().
		Str("season", season).
		Int("entries", len(leaderboard.Entries)).
		Int("limit", limit).
		Msg("📡 Broadcasting leaderboard to WebSocket Hub...")

	s.hub.Broadcast(season, leaderboard)

	log.Info().
		Str("season", season).
		Int("limit", limit).
		Msg("✅ Broadcast sent to Hub")
}

// handlePeriodicUpdates broadcasts leaderboard with dynamic limits per season
func (s *LeaderboardService) handlePeriodicUpdates(seasonLimits map[string]int) {
	log.Info().
		Int("season_count", len(seasonLimits)).
		Interface("season_limits", seasonLimits).
		Msg("🔔🔔🔔 handlePeriodicUpdates CALLED - about to broadcast with dynamic limits")

	// Use background context - periodic updates are not critical
	for season, limit := range seasonLimits {
		log.Info().
			Str("season", season).
			Int("limit", limit).
			Msg("🚀 Launching broadcast goroutine for season with dynamic limit")
		// Create separate context for each goroutine
		go func(seasonName string, requestedLimit int) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			s.broadcastLeaderboardUpdateWithLimit(ctx, seasonName, requestedLimit)
		}(season, limit)
	}

	log.Info().Msg("🔔 handlePeriodicUpdates FINISHED - all goroutines launched")
}

// SendInitialSnapshot sends the current leaderboard to a newly connected client
func (s *LeaderboardService) SendInitialSnapshot(season string, requestedLimit int, clientSend chan []byte) {
	log.Info().
		Str("season", season).
		Int("requested_limit", requestedLimit).
		Msg("📸 Sending initial snapshot to new client")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Fetch requested number of entries (default to 50 if not specified)
	if requestedLimit <= 0 {
		requestedLimit = 50
	}

	query := &models.LeaderboardQuery{
		Season:    season,
		Limit:     requestedLimit, // Send only requested entries
		Page:      0,
		SortOrder: "desc",
	}

	// Redis cache not used - PostgreSQL is the single source of truth
	// No need to clear cache as we always fetch fresh data from database

	leaderboard, err := s.GetLeaderboard(ctx, query)
	if err != nil {
		log.Warn().Err(err).Str("season", season).Msg("Failed to fetch leaderboard for initial snapshot")
		return
	}

	// Marshal message
	jsonData, err := json.Marshal(map[string]any{
		"type":        "leaderboard_update",
		"season":      season,
		"leaderboard": leaderboard,
		"timestamp":   time.Now().Unix(),
	})

	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal initial snapshot")
		return
	}

	log.Info().
		Str("season", season).
		Int("entries", len(leaderboard.Entries)).
		Int("json_size", len(jsonData)).
		Msg("📦📦📦 Initial snapshot marshaled, sending to client channel")

	// Send to client
	select {
	case clientSend <- jsonData:
		log.Info().
			Str("season", season).
			Int("entries", len(leaderboard.Entries)).
			Int("json_size", len(jsonData)).
			Msg("✅✅✅ Initial snapshot QUEUED in client Send channel")
	default:
		log.Warn().Str("season", season).Msg("⚠️ Failed to send initial snapshot: channel full")
	}
}

// BroadcastLeaderboard manually broadcasts leaderboard (for testing/admin)
func (s *LeaderboardService) BroadcastLeaderboard(ctx context.Context, season string) error {
	log.Info().Str("season", season).Msg("🔔 Manual broadcast triggered")

	if s.hub == nil {
		return fmt.Errorf("WebSocket hub not initialized")
	}

	query := &models.LeaderboardQuery{
		Season:    season,
		Limit:     10000,
		Page:      0,
		SortOrder: "desc",
	}

	leaderboard, err := s.GetLeaderboard(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to fetch leaderboard: %w", err)
	}

	s.hub.Broadcast(season, leaderboard)

	log.Info().
		Str("season", season).
		Int("entries", len(leaderboard.Entries)).
		Msg("✅ Manual broadcast completed")

	return nil
}
