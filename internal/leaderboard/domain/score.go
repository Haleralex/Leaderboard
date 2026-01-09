package domain

import (
	"time"

	"github.com/google/uuid"
)

// Score - чистая domain-модель счета без инфраструктурных зависимостей
// Представляет результат игрока в определенном сезоне
type Score struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Score     int64
	Season    string
	Metadata  map[string]interface{}
	Timestamp time.Time
}

// NewScore создает новый счет
func NewScore(userID uuid.UUID, score int64, season string, metadata map[string]interface{}) *Score {
	if season == "" {
		season = "global"
	}
	return &Score{
		ID:        uuid.New(),
		UserID:    userID,
		Score:     score,
		Season:    season,
		Metadata:  metadata,
		Timestamp: time.Now(),
	}
}

// UpdateScore обновляет значение счета
func (s *Score) UpdateScore(newScore int64) {
	s.Score = newScore
	s.Timestamp = time.Now()
}

// IsBetterThan проверяет, лучше ли текущий счет чем другой
func (s *Score) IsBetterThan(other *Score) bool {
	return s.Score > other.Score
}

// IsInSeason проверяет, относится ли счет к указанному сезону
func (s *Score) IsInSeason(season string) bool {
	return s.Season == season
}
