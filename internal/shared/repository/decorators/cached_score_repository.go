package decorators

import (
	"context"
	"fmt"
	"time"

	leaderboardmodels "leaderboard-service/internal/leaderboard/models"
	"leaderboard-service/internal/shared/repository"

	"github.com/google/uuid"
)

// CachedScoreRepository decorates ScoreRepository with caching
type CachedScoreRepository struct {
	inner repository.ScoreRepository
	cache *SimpleCache
	ttl   time.Duration
}

// NewCachedScoreRepository creates a cached score repository
func NewCachedScoreRepository(inner repository.ScoreRepository, cache *SimpleCache) repository.ScoreRepository {
	return &CachedScoreRepository{
		inner: inner,
		cache: cache,
		ttl:   2 * time.Minute, // Shorter TTL for scores (more dynamic data)
	}
}

// Upsert inserts/updates a score and invalidates cache
func (r *CachedScoreRepository) Upsert(ctx context.Context, score *leaderboardmodels.Score) error {
	err := r.inner.Upsert(ctx, score)
	if err != nil {
		return err
	}

	// Invalidate ALL leaderboard caches for this season (including all pagination/sort combinations)
	r.cache.Delete(r.scoreKey(score.UserID, score.Season))
	r.cache.DeleteByPrefix(fmt.Sprintf("leaderboard:%s:", score.Season))
	r.cache.Delete(r.countKey(score.Season))

	return nil
}

// FindByUserAndSeason retrieves a score with caching
func (r *CachedScoreRepository) FindByUserAndSeason(ctx context.Context, userID uuid.UUID, season string) (*leaderboardmodels.Score, error) {
	key := r.scoreKey(userID, season)

	// Check cache
	if cached, ok := r.cache.Get(key); ok {
		return cached.(*leaderboardmodels.Score), nil
	}

	// Cache miss
	score, err := r.inner.FindByUserAndSeason(ctx, userID, season)
	if err != nil {
		return nil, err
	}

	// Store in cache
	r.cache.Set(key, score, r.ttl)

	return score, nil
}

// GetLeaderboard retrieves leaderboard WITHOUT caching (dynamic data)
func (r *CachedScoreRepository) GetLeaderboard(ctx context.Context, season string, limit, offset int, sortOrder string) ([]leaderboardmodels.LeaderboardEntry, int64, error) {
	// Leaderboard changes frequently - always fetch fresh data from DB
	// Caching leaderboard causes stale data issues with real-time updates
	return r.inner.GetLeaderboard(ctx, season, limit, offset, sortOrder)
}

// CountBySeason retrieves count with caching
func (r *CachedScoreRepository) CountBySeason(ctx context.Context, season string) (int64, error) {
	key := r.countKey(season)

	// Check cache
	if cached, ok := r.cache.Get(key); ok {
		return cached.(int64), nil
	}

	// Cache miss
	count, err := r.inner.CountBySeason(ctx, season)
	if err != nil {
		return 0, err
	}

	// Store in cache
	r.cache.Set(key, count, r.ttl)

	return count, nil
}

// DeleteByUserAndSeason deletes a score and invalidates cache
func (r *CachedScoreRepository) DeleteByUserAndSeason(ctx context.Context, userID uuid.UUID, season string) error {
	err := r.inner.DeleteByUserAndSeason(ctx, userID, season)
	if err != nil {
		return err
	}

	// Invalidate ALL leaderboard caches for this season
	r.cache.Delete(r.scoreKey(userID, season))
	r.cache.DeleteByPrefix(fmt.Sprintf("leaderboard:%s:", season))
	r.cache.Delete(r.countKey(season))

	return nil
}

// Helper types and methods

type leaderboardCacheEntry struct {
	Entries    []leaderboardmodels.LeaderboardEntry
	TotalCount int64
}

func (r *CachedScoreRepository) scoreKey(userID uuid.UUID, season string) string {
	return fmt.Sprintf("score:%s:%s", userID.String(), season)
}

func (r *CachedScoreRepository) leaderboardKey(season string) string {
	return fmt.Sprintf("leaderboard:%s", season)
}

func (r *CachedScoreRepository) leaderboardKeyWithParams(season string, limit, offset int, sortOrder string) string {
	return fmt.Sprintf("leaderboard:%s:%d:%d:%s", season, limit, offset, sortOrder)
}

func (r *CachedScoreRepository) countKey(season string) string {
	return fmt.Sprintf("count:%s", season)
}

// FindBySpec finds scores matching a specification (no caching for complex queries)
func (r *CachedScoreRepository) FindBySpec(ctx context.Context, spec repository.Specification[leaderboardmodels.Score]) ([]*leaderboardmodels.Score, error) {
	// Specifications are too complex to cache efficiently, delegate to inner repository
	return r.inner.FindBySpec(ctx, spec)
}

// FindOneBySpec finds first score matching a specification (no caching for complex queries)
func (r *CachedScoreRepository) FindOneBySpec(ctx context.Context, spec repository.Specification[leaderboardmodels.Score]) (*leaderboardmodels.Score, error) {
	// Specifications are too complex to cache efficiently, delegate to inner repository
	return r.inner.FindOneBySpec(ctx, spec)
}

// CountBySpec counts scores matching a specification (no caching for counts)
func (r *CachedScoreRepository) CountBySpec(ctx context.Context, spec repository.Specification[leaderboardmodels.Score]) (int64, error) {
	// Counts are typically fast and don't benefit much from caching
	return r.inner.CountBySpec(ctx, spec)
}
