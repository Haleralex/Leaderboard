package models

import (
	"time"

	"github.com/google/uuid"
)

// Score represents a player's score entry
type Score struct {
	ID        uuid.UUID              `json:"id" db:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID              `json:"user_id" db:"user_id" gorm:"type:uuid;not null;index:idx_scores_season_score"`
	Score     int64                  `json:"score" db:"score" gorm:"type:bigint;not null;index:idx_scores_season_score"`
	Season    string                 `json:"season" db:"season" gorm:"type:varchar(50);not null;default:'global';index:idx_scores_season_score"`
	Metadata  map[string]interface{} `json:"metadata,omitempty" db:"metadata" gorm:"type:jsonb"`
	Timestamp time.Time              `json:"timestamp" db:"timestamp" gorm:"autoCreateTime"`
}

// TableName specifies the table name for GORM
func (Score) TableName() string {
	return "scores"
}

// SubmitScoreRequest is the payload for submitting a score
type SubmitScoreRequest struct {
	Score    int64                  `json:"score" validate:"required,min=0"`
	Season   string                 `json:"season" validate:"omitempty,max=50"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// LeaderboardEntry represents a leaderboard row with user info
type LeaderboardEntry struct {
	Rank      int       `json:"rank"`
	UserID    uuid.UUID `json:"user_id"`
	UserName  string    `json:"user_name"`
	Score     int64     `json:"score"`
	Season    string    `json:"season"`
	Timestamp time.Time `json:"timestamp"`
}

// LeaderboardResponse is the paginated leaderboard response
type LeaderboardResponse struct {
	Entries    []LeaderboardEntry `json:"entries"`
	TotalCount int64              `json:"total_count"`
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
	HasNext    bool               `json:"has_next"`
	NextCursor string             `json:"next_cursor,omitempty"`
}

// LeaderboardQuery represents query parameters for fetching leaderboard
type LeaderboardQuery struct {
	Season    string
	UserID    *uuid.UUID
	SortOrder string // "asc" or "desc"
	Limit     int
	Page      int
	Cursor    string // For cursor-based pagination
}
