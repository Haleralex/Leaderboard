package strategy

import (
	"testing"

	leaderboardmodels "leaderboard-service/internal/leaderboard/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestStandardRankingStrategy(t *testing.T) {
	strategy := NewStandardRankingStrategy()

	scores := []*leaderboardmodels.Score{
		{ID: uuid.New(), UserID: uuid.New(), Score: 1000},
		{ID: uuid.New(), UserID: uuid.New(), Score: 800},
		{ID: uuid.New(), UserID: uuid.New(), Score: 900},
		{ID: uuid.New(), UserID: uuid.New(), Score: 700},
	}

	ranked := strategy.CalculateRanks(scores)

	assert.Equal(t, 4, len(ranked))
	assert.Equal(t, 1, ranked[0].Rank)
	assert.Equal(t, int64(1000), ranked[0].Score.Score)
	assert.Equal(t, 2, ranked[1].Rank)
	assert.Equal(t, int64(900), ranked[1].Score.Score)
	assert.Equal(t, 3, ranked[2].Rank)
	assert.Equal(t, int64(800), ranked[2].Score.Score)
	assert.Equal(t, 4, ranked[3].Rank)
	assert.Equal(t, int64(700), ranked[3].Score.Score)
}

func TestStandardRankingStrategy_EmptyList(t *testing.T) {
	strategy := NewStandardRankingStrategy()
	ranked := strategy.CalculateRanks([]*leaderboardmodels.Score{})

	assert.Equal(t, 0, len(ranked))
}

func TestStandardRankingStrategy_Name(t *testing.T) {
	strategy := NewStandardRankingStrategy()
	assert.Equal(t, "Standard", strategy.Name())
}

func TestDenseRankingStrategy(t *testing.T) {
	strategy := NewDenseRankingStrategy()

	scores := []*leaderboardmodels.Score{
		{ID: uuid.New(), UserID: uuid.New(), Score: 1000},
		{ID: uuid.New(), UserID: uuid.New(), Score: 900},
		{ID: uuid.New(), UserID: uuid.New(), Score: 900}, // Same score
		{ID: uuid.New(), UserID: uuid.New(), Score: 800},
	}

	ranked := strategy.CalculateRanks(scores)

	assert.Equal(t, 4, len(ranked))
	assert.Equal(t, 1, ranked[0].Rank)
	assert.Equal(t, int64(1000), ranked[0].Score.Score)
	assert.False(t, ranked[0].TiedWithPrev)

	assert.Equal(t, 2, ranked[1].Rank)
	assert.Equal(t, int64(900), ranked[1].Score.Score)
	assert.False(t, ranked[1].TiedWithPrev)

	assert.Equal(t, 2, ranked[2].Rank) // Same rank as previous
	assert.Equal(t, int64(900), ranked[2].Score.Score)
	assert.True(t, ranked[2].TiedWithPrev)

	assert.Equal(t, 3, ranked[3].Rank) // Rank 3, not 4 (dense ranking)
	assert.Equal(t, int64(800), ranked[3].Score.Score)
	assert.False(t, ranked[3].TiedWithPrev)
}

func TestDenseRankingStrategy_EmptyList(t *testing.T) {
	strategy := NewDenseRankingStrategy()
	ranked := strategy.CalculateRanks([]*leaderboardmodels.Score{})

	assert.Equal(t, 0, len(ranked))
}

func TestDenseRankingStrategy_Name(t *testing.T) {
	strategy := NewDenseRankingStrategy()
	assert.Equal(t, "Dense", strategy.Name())
}

func TestCompetitionRankingStrategy(t *testing.T) {
	strategy := NewCompetitionRankingStrategy()

	scores := []*leaderboardmodels.Score{
		{ID: uuid.New(), UserID: uuid.New(), Score: 1000},
		{ID: uuid.New(), UserID: uuid.New(), Score: 900},
		{ID: uuid.New(), UserID: uuid.New(), Score: 900}, // Same score
		{ID: uuid.New(), UserID: uuid.New(), Score: 800},
	}

	ranked := strategy.CalculateRanks(scores)

	assert.Equal(t, 4, len(ranked))
	assert.Equal(t, 1, ranked[0].Rank)
	assert.Equal(t, int64(1000), ranked[0].Score.Score)

	assert.Equal(t, 2, ranked[1].Rank)
	assert.Equal(t, int64(900), ranked[1].Score.Score)
	assert.False(t, ranked[1].TiedWithPrev)

	assert.Equal(t, 2, ranked[2].Rank) // Same rank as previous
	assert.Equal(t, int64(900), ranked[2].Score.Score)
	assert.True(t, ranked[2].TiedWithPrev)

	assert.Equal(t, 4, ranked[3].Rank) // Rank 4, skipping 3 (competition ranking)
	assert.Equal(t, int64(800), ranked[3].Score.Score)
	assert.False(t, ranked[3].TiedWithPrev)
}

func TestCompetitionRankingStrategy_Name(t *testing.T) {
	strategy := NewCompetitionRankingStrategy()
	assert.Equal(t, "Competition", strategy.Name())
}

func TestCompetitionRankingStrategy_EmptyList(t *testing.T) {
	strategy := NewCompetitionRankingStrategy()
	ranked := strategy.CalculateRanks([]*leaderboardmodels.Score{})

	assert.Equal(t, 0, len(ranked))
}

func TestCompetitionRankingStrategy_AllSameScores(t *testing.T) {
	strategy := NewCompetitionRankingStrategy()

	scores := []*leaderboardmodels.Score{
		{ID: uuid.New(), UserID: uuid.New(), Score: 1000},
		{ID: uuid.New(), UserID: uuid.New(), Score: 1000},
		{ID: uuid.New(), UserID: uuid.New(), Score: 1000},
	}

	ranked := strategy.CalculateRanks(scores)

	assert.Equal(t, 3, len(ranked))
	assert.Equal(t, 1, ranked[0].Rank)
	assert.Equal(t, 1, ranked[1].Rank)
	assert.Equal(t, 1, ranked[2].Rank)
	assert.False(t, ranked[0].TiedWithPrev)
	assert.True(t, ranked[1].TiedWithPrev)
	assert.True(t, ranked[2].TiedWithPrev)
}
