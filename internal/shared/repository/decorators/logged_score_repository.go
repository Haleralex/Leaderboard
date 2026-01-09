package decorators

import (
	"context"
	"time"

	leaderboardmodels "leaderboard-service/internal/leaderboard/models"
	"leaderboard-service/internal/shared/repository"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// LoggedScoreRepository decorates ScoreRepository with logging
type LoggedScoreRepository struct {
	inner repository.ScoreRepository
}

// NewLoggedScoreRepository creates a logged score repository
func NewLoggedScoreRepository(inner repository.ScoreRepository) repository.ScoreRepository {
	return &LoggedScoreRepository{
		inner: inner,
	}
}

// Upsert inserts/updates a score with logging
func (r *LoggedScoreRepository) Upsert(ctx context.Context, score *leaderboardmodels.Score) error {
	start := time.Now()
	err := r.inner.Upsert(ctx, score)
	duration := time.Since(start)

	logEvent := log.Info()
	if err != nil {
		logEvent = log.Error().Err(err)
	}

	logEvent.
		Str("method", "ScoreRepository.Upsert").
		Str("user_id", score.UserID.String()).
		Int64("score", score.Score).
		Str("season", score.Season).
		Dur("duration", duration).
		Msg("Score upsert")

	return err
}

// FindByUserAndSeason retrieves a score with logging
func (r *LoggedScoreRepository) FindByUserAndSeason(ctx context.Context, userID uuid.UUID, season string) (*leaderboardmodels.Score, error) {
	start := time.Now()
	score, err := r.inner.FindByUserAndSeason(ctx, userID, season)
	duration := time.Since(start)

	logEvent := log.Debug()
	if err != nil {
		logEvent = log.Warn().Err(err)
	}

	logEvent.
		Str("method", "ScoreRepository.FindByUserAndSeason").
		Str("user_id", userID.String()).
		Str("season", season).
		Dur("duration", duration).
		Bool("found", err == nil).
		Msg("Score lookup")

	return score, err
}

// GetLeaderboard retrieves leaderboard with logging
func (r *LoggedScoreRepository) GetLeaderboard(ctx context.Context, season string, limit, offset int, sortOrder string) ([]leaderboardmodels.LeaderboardEntry, int64, error) {
	start := time.Now()
	entries, totalCount, err := r.inner.GetLeaderboard(ctx, season, limit, offset, sortOrder)
	duration := time.Since(start)

	logEvent := log.Debug()
	if err != nil {
		logEvent = log.Error().Err(err)
	}

	logEvent.
		Str("method", "ScoreRepository.GetLeaderboard").
		Str("season", season).
		Int("limit", limit).
		Int("offset", offset).
		Str("sort_order", sortOrder).
		Int("entries_count", len(entries)).
		Int64("total_count", totalCount).
		Dur("duration", duration).
		Msg("Leaderboard query")

	return entries, totalCount, err
}

// CountBySeason retrieves count with logging
func (r *LoggedScoreRepository) CountBySeason(ctx context.Context, season string) (int64, error) {
	start := time.Now()
	count, err := r.inner.CountBySeason(ctx, season)
	duration := time.Since(start)

	logEvent := log.Debug()
	if err != nil {
		logEvent = log.Error().Err(err)
	}

	logEvent.
		Str("method", "ScoreRepository.CountBySeason").
		Str("season", season).
		Int64("count", count).
		Dur("duration", duration).
		Msg("Score count")

	return count, err
}

// DeleteByUserAndSeason deletes a score with logging
func (r *LoggedScoreRepository) DeleteByUserAndSeason(ctx context.Context, userID uuid.UUID, season string) error {
	start := time.Now()
	err := r.inner.DeleteByUserAndSeason(ctx, userID, season)
	duration := time.Since(start)

	logEvent := log.Info()
	if err != nil {
		logEvent = log.Error().Err(err)
	}

	logEvent.
		Str("method", "ScoreRepository.DeleteByUserAndSeason").
		Str("user_id", userID.String()).
		Str("season", season).
		Dur("duration", duration).
		Msg("Score deletion")

	return err
}

// FindBySpec finds scores by specification with logging
func (r *LoggedScoreRepository) FindBySpec(ctx context.Context, spec repository.Specification[leaderboardmodels.Score]) ([]*leaderboardmodels.Score, error) {
	start := time.Now()
	scores, err := r.inner.FindBySpec(ctx, spec)
	duration := time.Since(start)

	logEvent := log.Debug()
	if err != nil {
		logEvent = log.Warn().Err(err)
	}

	logEvent.
		Str("method", "ScoreRepository.FindBySpec").
		Dur("duration", duration).
		Int("count", len(scores)).
		Msg("Score query by specification")

	return scores, err
}

// FindOneBySpec finds one score by specification with logging
func (r *LoggedScoreRepository) FindOneBySpec(ctx context.Context, spec repository.Specification[leaderboardmodels.Score]) (*leaderboardmodels.Score, error) {
	start := time.Now()
	score, err := r.inner.FindOneBySpec(ctx, spec)
	duration := time.Since(start)

	logEvent := log.Debug()
	if err != nil {
		logEvent = log.Warn().Err(err)
	}

	logEvent.
		Str("method", "ScoreRepository.FindOneBySpec").
		Dur("duration", duration).
		Bool("found", err == nil).
		Msg("Score query by specification")

	return score, err
}

// CountBySpec counts scores by specification with logging
func (r *LoggedScoreRepository) CountBySpec(ctx context.Context, spec repository.Specification[leaderboardmodels.Score]) (int64, error) {
	start := time.Now()
	count, err := r.inner.CountBySpec(ctx, spec)
	duration := time.Since(start)

	logEvent := log.Debug()
	if err != nil {
		logEvent = log.Warn().Err(err)
	}

	logEvent.
		Str("method", "ScoreRepository.CountBySpec").
		Dur("duration", duration).
		Int64("count", count).
		Msg("Score count by specification")

	return count, err
}
