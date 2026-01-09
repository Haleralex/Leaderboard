package repository

import (
	leaderboardmodels "leaderboard-service/internal/leaderboard/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Score Specifications

// ScoreByUserIDSpec filters scores by user ID
type ScoreByUserIDSpec struct {
	BaseSpecification[leaderboardmodels.Score]
	UserID uuid.UUID
}

func NewScoreByUserIDSpec(userID uuid.UUID) Specification[leaderboardmodels.Score] {
	return &ScoreByUserIDSpec{UserID: userID}
}

func (s *ScoreByUserIDSpec) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("user_id = ?", s.UserID)
}

func (s *ScoreByUserIDSpec) IsSatisfiedBy(score leaderboardmodels.Score) bool {
	return score.UserID == s.UserID
}

// ScoreBySeasonSpec filters scores by season
type ScoreBySeasonSpec struct {
	BaseSpecification[leaderboardmodels.Score]
	Season string
}

func NewScoreBySeasonSpec(season string) Specification[leaderboardmodels.Score] {
	return &ScoreBySeasonSpec{Season: season}
}

func (s *ScoreBySeasonSpec) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("season = ?", s.Season)
}

func (s *ScoreBySeasonSpec) IsSatisfiedBy(score leaderboardmodels.Score) bool {
	return score.Season == s.Season
}

// ScoreMinValueSpec filters scores with minimum value
type ScoreMinValueSpec struct {
	BaseSpecification[leaderboardmodels.Score]
	MinScore int64
}

func NewScoreMinValueSpec(minScore int64) Specification[leaderboardmodels.Score] {
	return &ScoreMinValueSpec{MinScore: minScore}
}

func (s *ScoreMinValueSpec) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("score >= ?", s.MinScore)
}

func (s *ScoreMinValueSpec) IsSatisfiedBy(score leaderboardmodels.Score) bool {
	return score.Score >= s.MinScore
}

// ScoreMaxValueSpec filters scores with maximum value
type ScoreMaxValueSpec struct {
	BaseSpecification[leaderboardmodels.Score]
	MaxScore int64
}

func NewScoreMaxValueSpec(maxScore int64) Specification[leaderboardmodels.Score] {
	return &ScoreMaxValueSpec{MaxScore: maxScore}
}

func (s *ScoreMaxValueSpec) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("score <= ?", s.MaxScore)
}

func (s *ScoreMaxValueSpec) IsSatisfiedBy(score leaderboardmodels.Score) bool {
	return score.Score <= s.MaxScore
}

// ScoreRangeSpec filters scores within a range
type ScoreRangeSpec struct {
	BaseSpecification[leaderboardmodels.Score]
	MinScore int64
	MaxScore int64
}

func NewScoreRangeSpec(minScore, maxScore int64) Specification[leaderboardmodels.Score] {
	return &ScoreRangeSpec{MinScore: minScore, MaxScore: maxScore}
}

func (s *ScoreRangeSpec) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("score BETWEEN ? AND ?", s.MinScore, s.MaxScore)
}

func (s *ScoreRangeSpec) IsSatisfiedBy(score leaderboardmodels.Score) bool {
	return score.Score >= s.MinScore && score.Score <= s.MaxScore
}

// ScoreTopNSpec gets top N scores
type ScoreTopNSpec struct {
	BaseSpecification[leaderboardmodels.Score]
	N int
}

func NewScoreTopNSpec(n int) Specification[leaderboardmodels.Score] {
	return &ScoreTopNSpec{N: n}
}

func (s *ScoreTopNSpec) Apply(db *gorm.DB) *gorm.DB {
	return db.Order("score DESC").Limit(s.N)
}

func (s *ScoreTopNSpec) IsSatisfiedBy(score leaderboardmodels.Score) bool {
	return true // TopN doesn't affect individual entities
}

// ScoreOrderBySpec orders scores
type ScoreOrderBySpec struct {
	BaseSpecification[leaderboardmodels.Score]
	Field string
	Desc  bool
}

func NewScoreOrderBySpec(field string, desc bool) Specification[leaderboardmodels.Score] {
	return &ScoreOrderBySpec{Field: field, Desc: desc}
}

func (s *ScoreOrderBySpec) Apply(db *gorm.DB) *gorm.DB {
	order := s.Field
	if s.Desc {
		order += " DESC"
	}
	return db.Order(order)
}

func (s *ScoreOrderBySpec) IsSatisfiedBy(score leaderboardmodels.Score) bool {
	return true // Order doesn't affect individual entities
}

// ScoreLimitSpec limits results
type ScoreLimitSpec struct {
	BaseSpecification[leaderboardmodels.Score]
	Limit int
}

func NewScoreLimitSpec(limit int) Specification[leaderboardmodels.Score] {
	return &ScoreLimitSpec{Limit: limit}
}

func (s *ScoreLimitSpec) Apply(db *gorm.DB) *gorm.DB {
	return db.Limit(s.Limit)
}

func (s *ScoreLimitSpec) IsSatisfiedBy(score leaderboardmodels.Score) bool {
	return true
}

// ScoreOffsetSpec adds offset
type ScoreOffsetSpec struct {
	BaseSpecification[leaderboardmodels.Score]
	Offset int
}

func NewScoreOffsetSpec(offset int) Specification[leaderboardmodels.Score] {
	return &ScoreOffsetSpec{Offset: offset}
}

func (s *ScoreOffsetSpec) Apply(db *gorm.DB) *gorm.DB {
	return db.Offset(s.Offset)
}

func (s *ScoreOffsetSpec) IsSatisfiedBy(score leaderboardmodels.Score) bool {
	return true
}

// Convenience builders for common queries

// LeaderboardSpec - top scores for a season
func LeaderboardSpec(season string, limit int) Specification[leaderboardmodels.Score] {
	return And(
		NewScoreBySeasonSpec(season),
		NewScoreOrderBySpec("score", true),
		NewScoreLimitSpec(limit),
	)
}

// HighScoresSpec - scores above threshold
func HighScoresSpec(season string, minScore int64, limit int) Specification[leaderboardmodels.Score] {
	return And(
		NewScoreBySeasonSpec(season),
		NewScoreMinValueSpec(minScore),
		NewScoreOrderBySpec("score", true),
		NewScoreLimitSpec(limit),
	)
}

// UserScoresInSeasonSpec - all scores for user in season
func UserScoresInSeasonSpec(userID uuid.UUID, season string) Specification[leaderboardmodels.Score] {
	return And(
		NewScoreByUserIDSpec(userID),
		NewScoreBySeasonSpec(season),
		NewScoreOrderBySpec("timestamp", true),
	)
}

// PaginatedLeaderboardSpec - leaderboard with pagination
func PaginatedLeaderboardSpec(season string, page, pageSize int) Specification[leaderboardmodels.Score] {
	offset := (page - 1) * pageSize
	return And(
		NewScoreBySeasonSpec(season),
		NewScoreOrderBySpec("score", true),
		NewScoreLimitSpec(pageSize),
		NewScoreOffsetSpec(offset),
	)
}

// MidRangeScoresSpec - scores in specific range
func MidRangeScoresSpec(season string, minScore, maxScore int64) Specification[leaderboardmodels.Score] {
	return And(
		NewScoreBySeasonSpec(season),
		NewScoreRangeSpec(minScore, maxScore),
		NewScoreOrderBySpec("score", true),
	)
}
