package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewScore(t *testing.T) {
	userID := uuid.New()
	score := int64(1000)
	season := "summer2024"
	metadata := map[string]interface{}{"level": "expert"}

	s := NewScore(userID, score, season, metadata)

	assert.NotEqual(t, uuid.Nil, s.ID, "Score ID should be generated")
	assert.Equal(t, userID, s.UserID)
	assert.Equal(t, score, s.Score)
	assert.Equal(t, season, s.Season)
	assert.Equal(t, metadata, s.Metadata)
	assert.False(t, s.Timestamp.IsZero(), "Timestamp should be set")
}

func TestNewScore_EmptySeason(t *testing.T) {
	userID := uuid.New()
	score := int64(500)

	s := NewScore(userID, score, "", nil)

	assert.Equal(t, "global", s.Season, "Should default to 'global' season")
}

func TestScore_UpdateScore(t *testing.T) {
	userID := uuid.New()
	s := NewScore(userID, 100, "winter2024", nil)
	originalTimestamp := s.Timestamp

	time.Sleep(10 * time.Millisecond)

	newScore := int64(200)
	s.UpdateScore(newScore)

	assert.Equal(t, newScore, s.Score, "Score should be updated")
	assert.True(t, s.Timestamp.After(originalTimestamp), "Timestamp should be updated")
}

func TestScore_IsBetterThan(t *testing.T) {
	userID := uuid.New()
	score1 := NewScore(userID, 100, "global", nil)
	score2 := NewScore(userID, 50, "global", nil)
	score3 := NewScore(userID, 100, "global", nil)

	assert.True(t, score1.IsBetterThan(score2), "100 should be better than 50")
	assert.False(t, score2.IsBetterThan(score1), "50 should not be better than 100")
	assert.False(t, score1.IsBetterThan(score3), "Equal scores should not be better")
}

func TestScore_IsInSeason(t *testing.T) {
	userID := uuid.New()
	score := NewScore(userID, 100, "summer2024", nil)

	assert.True(t, score.IsInSeason("summer2024"), "Should be in summer2024")
	assert.False(t, score.IsInSeason("winter2024"), "Should not be in winter2024")
	assert.False(t, score.IsInSeason(""), "Should not be in empty season")
}

func TestScore_WithMetadata(t *testing.T) {
	userID := uuid.New()
	metadata := map[string]interface{}{
		"level":      "expert",
		"difficulty": 5,
		"bonus":      true,
	}

	score := NewScore(userID, 1000, "global", metadata)

	assert.Equal(t, "expert", score.Metadata["level"])
	assert.Equal(t, 5, score.Metadata["difficulty"])
	assert.Equal(t, true, score.Metadata["bonus"])
}

func TestScore_WithNilMetadata(t *testing.T) {
	userID := uuid.New()
	score := NewScore(userID, 500, "global", nil)

	assert.Nil(t, score.Metadata, "Metadata should be nil when not provided")
}
