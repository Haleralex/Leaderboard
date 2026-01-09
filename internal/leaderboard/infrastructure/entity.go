package infrastructure

import (
	"time"

	"leaderboard-service/internal/leaderboard/domain"

	"github.com/google/uuid"
)

// ScoreEntity - persistence модель с GORM тегами
// Отделена от domain для чистоты архитектуры
type ScoreEntity struct {
	ID        uuid.UUID              `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID              `gorm:"type:uuid;not null;index:idx_scores_season_score"`
	Score     int64                  `gorm:"type:bigint;not null;index:idx_scores_season_score"`
	Season    string                 `gorm:"type:varchar(50);not null;default:'global';index:idx_scores_season_score"`
	Metadata  map[string]interface{} `gorm:"type:jsonb"`
	Timestamp time.Time              `gorm:"autoCreateTime"`
}

// TableName для GORM
func (ScoreEntity) TableName() string {
	return "scores"
}

// ToDomain конвертирует entity в domain модель
func (e *ScoreEntity) ToDomain() *domain.Score {
	return &domain.Score{
		ID:        e.ID,
		UserID:    e.UserID,
		Score:     e.Score,
		Season:    e.Season,
		Metadata:  e.Metadata,
		Timestamp: e.Timestamp,
	}
}

// FromDomain создает entity из domain модели
func FromDomainScore(s *domain.Score) *ScoreEntity {
	return &ScoreEntity{
		ID:        s.ID,
		UserID:    s.UserID,
		Score:     s.Score,
		Season:    s.Season,
		Metadata:  s.Metadata,
		Timestamp: s.Timestamp,
	}
}

// ToDomainList конвертирует список entities в domain модели
func ToDomainScoreList(entities []*ScoreEntity) []*domain.Score {
	result := make([]*domain.Score, len(entities))
	for i, e := range entities {
		result[i] = e.ToDomain()
	}
	return result
}
