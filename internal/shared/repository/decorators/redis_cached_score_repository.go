package decorators

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	leaderboardmodels "leaderboard-service/internal/leaderboard/models"
	"leaderboard-service/internal/shared/database"
	"leaderboard-service/internal/shared/repository"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// RedisCachedScoreRepository decorates ScoreRepository with Redis caching
type RedisCachedScoreRepository struct {
	inner repository.ScoreRepository
	redis *database.RedisClient
	ttl   time.Duration
}

// NewRedisCachedScoreRepository creates a Redis-cached score repository
func NewRedisCachedScoreRepository(inner repository.ScoreRepository, redis *database.RedisClient) repository.ScoreRepository {
	if redis == nil {
		log.Warn().Msg("Redis is nil, returning uncached repository")
		return inner
	}

	return &RedisCachedScoreRepository{
		inner: inner,
		redis: redis,
		ttl:   30 * time.Second, // Short TTL for frequently changing data
	}
}

// Upsert inserts/updates a score and invalidates Redis cache
func (r *RedisCachedScoreRepository) Upsert(ctx context.Context, score *leaderboardmodels.Score) error {
	err := r.inner.Upsert(ctx, score)
	if err != nil {
		return err
	}

	// Invalidate ALL leaderboard caches for this season using Redis SCAN
	r.invalidateLeaderboardCache(ctx, score.Season)
	r.redis.Client.Del(ctx, r.scoreKey(score.UserID, score.Season))
	r.redis.Client.Del(ctx, r.countKey(score.Season))

	return nil
}

// FindByUserAndSeason retrieves a score with Redis caching
func (r *RedisCachedScoreRepository) FindByUserAndSeason(ctx context.Context, userID uuid.UUID, season string) (*leaderboardmodels.Score, error) {
	key := r.scoreKey(userID, season)

	// Try cache first
	cached, err := r.redis.Client.Get(ctx, key).Result()
	if err == nil {
		var score leaderboardmodels.Score
		if err := json.Unmarshal([]byte(cached), &score); err == nil {
			return &score, nil
		}
	}

	// Cache miss - fetch from DB
	score, err := r.inner.FindByUserAndSeason(ctx, userID, season)
	if err != nil {
		return nil, err
	}

	// Store in Redis
	if data, err := json.Marshal(score); err == nil {
		r.redis.Client.Set(ctx, key, data, r.ttl)
	}

	return score, nil
}

// GetLeaderboard retrieves leaderboard with Redis caching
func (r *RedisCachedScoreRepository) GetLeaderboard(ctx context.Context, season string, limit, offset int, sortOrder string) ([]leaderboardmodels.LeaderboardEntry, int64, error) {
	key := r.leaderboardKeyWithParams(season, limit, offset, sortOrder)

	// Try cache first
	cached, err := r.redis.Client.Get(ctx, key).Result()
	if err == nil {
		var result struct {
			Entries    []leaderboardmodels.LeaderboardEntry `json:"entries"`
			TotalCount int64                                `json:"total_count"`
		}
		if err := json.Unmarshal([]byte(cached), &result); err == nil {
			log.Info().
				Str("key", key).
				Int("entries", len(result.Entries)).
				Msg("🎯 CACHE HIT: Leaderboard served from Redis")
			return result.Entries, result.TotalCount, nil
		}
	}

	// Cache miss - fetch from DB
	entries, totalCount, err := r.inner.GetLeaderboard(ctx, season, limit, offset, sortOrder)
	if err != nil {
		return nil, 0, err
	}

	// Store in Redis
	result := struct {
		Entries    []leaderboardmodels.LeaderboardEntry `json:"entries"`
		TotalCount int64                                `json:"total_count"`
	}{
		Entries:    entries,
		TotalCount: totalCount,
	}

	if data, err := json.Marshal(result); err == nil {
		r.redis.Client.Set(ctx, key, data, r.ttl)
		log.Info().
			Str("key", key).
			Int("entries", len(entries)).
			Dur("ttl", r.ttl).
			Msg("💾 CACHE MISS: Leaderboard cached in Redis")
	}

	return entries, totalCount, nil
}

// CountBySeason retrieves count with Redis caching
func (r *RedisCachedScoreRepository) CountBySeason(ctx context.Context, season string) (int64, error) {
	key := r.countKey(season)

	// Try cache first
	count, err := r.redis.Client.Get(ctx, key).Int64()
	if err == nil {
		return count, nil
	}

	// Cache miss - fetch from DB
	count, err = r.inner.CountBySeason(ctx, season)
	if err != nil {
		return 0, err
	}

	// Store in Redis
	r.redis.Client.Set(ctx, key, count, r.ttl)

	return count, nil
}

// DeleteByUserAndSeason deletes a score and invalidates Redis cache
func (r *RedisCachedScoreRepository) DeleteByUserAndSeason(ctx context.Context, userID uuid.UUID, season string) error {
	err := r.inner.DeleteByUserAndSeason(ctx, userID, season)
	if err != nil {
		return err
	}

	// Invalidate caches in Redis
	r.invalidateLeaderboardCache(ctx, season)
	r.redis.Client.Del(ctx, r.scoreKey(userID, season))
	r.redis.Client.Del(ctx, r.countKey(season))

	return nil
}

// invalidateLeaderboardCache removes all leaderboard keys for a season using SCAN
func (r *RedisCachedScoreRepository) invalidateLeaderboardCache(ctx context.Context, season string) {
	pattern := fmt.Sprintf("leaderboard:%s:*", season)

	// Use SCAN to find all matching keys (non-blocking)
	iter := r.redis.Client.Scan(ctx, 0, pattern, 100).Iterator()
	keys := []string{}

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		log.Warn().Err(err).Str("pattern", pattern).Msg("Failed to scan Redis keys")
		return
	}

	// Delete all found keys
	if len(keys) > 0 {
		deleted, err := r.redis.Client.Del(ctx, keys...).Result()
		if err != nil {
			log.Warn().Err(err).Int("keys", len(keys)).Msg("Failed to delete Redis keys")
		} else {
			log.Info().
				Int64("deleted", deleted).
				Str("pattern", pattern).
				Msg("🗑️ INVALIDATED: Redis cache cleared")
		}
	}
}

// Helper methods for cache keys

func (r *RedisCachedScoreRepository) scoreKey(userID uuid.UUID, season string) string {
	return fmt.Sprintf("score:%s:%s", userID.String(), season)
}

func (r *RedisCachedScoreRepository) leaderboardKeyWithParams(season string, limit, offset int, sortOrder string) string {
	return fmt.Sprintf("leaderboard:%s:%d:%d:%s", season, limit, offset, sortOrder)
}

func (r *RedisCachedScoreRepository) countKey(season string) string {
	return fmt.Sprintf("count:%s", season)
}

// FindBySpec finds scores matching a specification (no caching for complex queries)
func (r *RedisCachedScoreRepository) FindBySpec(ctx context.Context, spec repository.Specification[leaderboardmodels.Score]) ([]*leaderboardmodels.Score, error) {
	return r.inner.FindBySpec(ctx, spec)
}

// FindOneBySpec finds first score matching a specification (no caching for complex queries)
func (r *RedisCachedScoreRepository) FindOneBySpec(ctx context.Context, spec repository.Specification[leaderboardmodels.Score]) (*leaderboardmodels.Score, error) {
	return r.inner.FindOneBySpec(ctx, spec)
}

// CountBySpec counts scores matching a specification (no caching for counts)
func (r *RedisCachedScoreRepository) CountBySpec(ctx context.Context, spec repository.Specification[leaderboardmodels.Score]) (int64, error) {
	return r.inner.CountBySpec(ctx, spec)
}
