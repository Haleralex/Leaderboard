package repository

import (
	"context"
	"errors"
	"fmt"

	"leaderboard-service/internal/leaderboard/models"
	"leaderboard-service/internal/shared/database"
	"leaderboard-service/internal/shared/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// PostgresScoreRepository is a PostgreSQL implementation of ScoreRepository
type PostgresScoreRepository struct {
	db *database.PostgresDB
}

// NewPostgresScoreRepository creates a new PostgreSQL score repository
func NewPostgresScoreRepository(db *database.PostgresDB) repository.ScoreRepository {
	return &PostgresScoreRepository{
		db: db,
	}
}

// Upsert inserts a new score or updates if the user already has a score for the season
func (r *PostgresScoreRepository) Upsert(ctx context.Context, score *models.Score) error {
	result := r.db.DB.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "season"}},
		DoUpdates: clause.AssignmentColumns([]string{"score", "metadata", "timestamp"}),
	}).Create(score)

	if result.Error != nil {
		return fmt.Errorf("failed to upsert score: %w", result.Error)
	}
	return nil
}

// FindByUserAndSeason retrieves a user's score for a specific season
func (r *PostgresScoreRepository) FindByUserAndSeason(ctx context.Context, userID uuid.UUID, season string) (*models.Score, error) {
	var score models.Score
	err := r.db.DB.WithContext(ctx).
		Where("user_id = ? AND season = ?", userID, season).
		First(&score).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("score not found")
		}
		return nil, fmt.Errorf("failed to find score: %w", err)
	}
	return &score, nil
}

// GetLeaderboard retrieves paginated leaderboard entries for a season with user details
func (r *PostgresScoreRepository) GetLeaderboard(ctx context.Context, season string, limit, offset int, sortOrder string) ([]models.LeaderboardEntry, int64, error) {
	orderBy := "scores.score DESC"
	if sortOrder == "asc" {
		orderBy = "score ASC"
	}

	var entries []models.LeaderboardEntry
	err := r.db.DB.WithContext(ctx).Raw(`
		SELECT rank, user_id, user_name, score, season, timestamp
		FROM leaderboard_view
		WHERE season = ?
		ORDER BY `+orderBy+`
		LIMIT ? OFFSET ?
	`, season, limit, offset).Scan(&entries).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to query leaderboard: %w", err)
	}

	// Get total count for pagination
	totalCount, err := r.CountBySeason(ctx, season)
	if err != nil {
		// If count fails, use entries length as fallback
		totalCount = int64(len(entries))
	}

	return entries, totalCount, nil
}

// CountBySeason returns the total number of scores for a given season
func (r *PostgresScoreRepository) CountBySeason(ctx context.Context, season string) (int64, error) {
	var count int64
	err := r.db.DB.WithContext(ctx).
		Model(&models.Score{}).
		Where("season = ?", season).
		Count(&count).Error

	if err != nil {
		return 0, fmt.Errorf("failed to count scores: %w", err)
	}
	return count, nil
}

// DeleteByUserAndSeason removes a user's score for a specific season
func (r *PostgresScoreRepository) DeleteByUserAndSeason(ctx context.Context, userID uuid.UUID, season string) error {
	result := r.db.DB.WithContext(ctx).
		Where("user_id = ? AND season = ?", userID, season).
		Delete(&models.Score{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete score: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("score not found")
	}
	return nil
}

// FindBySpec finds scores matching a specification
func (r *PostgresScoreRepository) FindBySpec(ctx context.Context, spec repository.Specification[models.Score]) ([]*models.Score, error) {
	var scores []*models.Score

	query := r.db.DB.WithContext(ctx)
	query = spec.Apply(query)

	err := query.Find(&scores).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find scores by spec: %w", err)
	}

	return scores, nil
}

// FindOneBySpec finds first score matching a specification
func (r *PostgresScoreRepository) FindOneBySpec(ctx context.Context, spec repository.Specification[models.Score]) (*models.Score, error) {
	var score models.Score

	query := r.db.DB.WithContext(ctx)
	query = spec.Apply(query)

	err := query.First(&score).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("score not found")
		}
		return nil, fmt.Errorf("failed to find score by spec: %w", err)
	}

	return &score, nil
}

// CountBySpec counts scores matching a specification
func (r *PostgresScoreRepository) CountBySpec(ctx context.Context, spec repository.Specification[models.Score]) (int64, error) {
	var count int64

	query := r.db.DB.WithContext(ctx).Model(&models.Score{})
	query = spec.Apply(query)

	err := query.Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count scores by spec: %w", err)
	}

	return count, nil
}
