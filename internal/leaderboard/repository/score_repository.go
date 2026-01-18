package repository

import (
	"context"
	"fmt"

	"leaderboard-service/internal/leaderboard/domain"
	"leaderboard-service/internal/leaderboard/infrastructure"
	"leaderboard-service/internal/leaderboard/models"
	"leaderboard-service/internal/shared/database"
	"leaderboard-service/internal/shared/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// PostgresScoreRepository is a PostgreSQL implementation of ScoreRepository
// Использует чистую domain модель и отдельные entities для персистентности
type PostgresScoreRepository struct {
	*repository.BaseRepository[infrastructure.ScoreEntity]
	db *database.PostgresDB
}

// NewPostgresScoreRepository creates a new PostgreSQL score repository
func NewPostgresScoreRepository(db *database.PostgresDB) repository.ScoreRepository {
	return &PostgresScoreRepository{
		BaseRepository: repository.NewBaseRepository[infrastructure.ScoreEntity](db),
		db:             db,
	}
}

// Upsert inserts a new score or updates if the user already has a score for the season
func (r *PostgresScoreRepository) Upsert(ctx context.Context, score *models.Score) error {
	domainScore := &domain.Score{
		ID:        score.ID,
		UserID:    score.UserID,
		Score:     score.Score,
		Season:    score.Season,
		Metadata:  score.Metadata,
		Timestamp: score.Timestamp,
	}
	entity := infrastructure.FromDomainScore(domainScore)

	result := r.db.DB.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "season"}},
		DoUpdates: clause.AssignmentColumns([]string{"score", "metadata", "timestamp"}),
	}).Create(entity)

	if result.Error != nil {
		return fmt.Errorf("failed to upsert score: %w", result.Error)
	}
	score.ID = entity.ID
	return nil
}

// FindByUserAndSeason retrieves a user's score for a specific season
func (r *PostgresScoreRepository) FindByUserAndSeason(ctx context.Context, userID uuid.UUID, season string) (*models.Score, error) {
	entity, err := r.BaseRepository.FindOne(ctx, "user_id = ? AND season = ?", userID, season)
	if err != nil {
		return nil, err
	}
	domainScore := entity.ToDomain()
	return &models.Score{
		ID:        domainScore.ID,
		UserID:    domainScore.UserID,
		Score:     domainScore.Score,
		Season:    domainScore.Season,
		Metadata:  domainScore.Metadata,
		Timestamp: domainScore.Timestamp,
	}, nil
}

// GetLeaderboard retrieves paginated leaderboard entries for a season with user details
func (r *PostgresScoreRepository) GetLeaderboard(ctx context.Context, season string, limit, offset int, sortOrder string) ([]models.LeaderboardEntry, int64, error) {
	// Sort by score directly, not by rank (which is computed)
	// desc = highest scores first (default leaderboard view)
	// asc = lowest scores first (rare case)
	var orderBy string
	if sortOrder == "asc" {
		orderBy = "s.score ASC, s.timestamp ASC"
	} else {
		orderBy = "s.score DESC, s.timestamp ASC"
	}

	var entries []models.LeaderboardEntry
	// Force fresh query without prepared statement cache
	// Use a new connection to avoid transaction isolation issues
	err := r.db.DB.WithContext(ctx).
		Session(&gorm.Session{
			PrepareStmt:            false,
			SkipDefaultTransaction: true,
		}).
		Raw(`
			SELECT 
				DENSE_RANK() OVER (ORDER BY s.score DESC, s.timestamp ASC) as rank,
				s.user_id,
				u.name as user_name,
				s.score,
				s.season,
				s.timestamp
			FROM scores s
			JOIN users u ON s.user_id = u.id
			WHERE s.season = ?
			ORDER BY `+orderBy+`
			LIMIT ? OFFSET ?
		`, season, limit, offset).Scan(&entries).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query leaderboard: %w", err)
	}

	// Log query results for debugging
	top3Debug := "["
	for i := 0; i < 3 && i < len(entries); i++ {
		if i > 0 {
			top3Debug += ", "
		}
		top3Debug += fmt.Sprintf("{rank:%d, score:%d, name:%s, ts:%v}",
			entries[i].Rank,
			entries[i].Score,
			entries[i].UserName,
			entries[i].Timestamp)
	}
	_ = top3Debug // unused for now, kept for debugging

	// Get total count for pagination
	totalCount, err := r.CountBySeason(ctx, season)
	if err != nil {
		// If count fails, use entries length as fallback
		totalCount = int64(len(entries))
	}

	return entries, totalCount, nil
}

// CountBySeason returns the total number of scores for a given season
// Использует переиспользуемый метод из BaseRepository
func (r *PostgresScoreRepository) CountBySeason(ctx context.Context, season string) (int64, error) {
	return r.BaseRepository.Count(ctx, "season = ?", season)
}

// DeleteByUserAndSeason removes a user's score for a specific season
func (r *PostgresScoreRepository) DeleteByUserAndSeason(ctx context.Context, userID uuid.UUID, season string) error {
	return r.BaseRepository.Delete(ctx, "user_id = ? AND season = ?", userID, season)
}

// FindBySpec finds scores matching a specification
func (r *PostgresScoreRepository) FindBySpec(ctx context.Context, spec repository.Specification[models.Score]) ([]*models.Score, error) {
	// Временно возвращаем пустой список до полной миграции спецификаций
	return nil, nil
}

// FindOneBySpec finds first score matching a specification
func (r *PostgresScoreRepository) FindOneBySpec(ctx context.Context, spec repository.Specification[models.Score]) (*models.Score, error) {
	// Временно возвращаем nil до полной миграции спецификаций
	return nil, nil
}

// CountBySpec counts scores matching a specification
func (r *PostgresScoreRepository) CountBySpec(ctx context.Context, spec repository.Specification[models.Score]) (int64, error) {
	// Временно возвращаем 0 до полной миграции спецификаций
	return 0, nil
}
