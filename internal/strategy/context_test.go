package strategy

import (
	"testing"

	leaderboardmodels "leaderboard-service/internal/leaderboard/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScoreCalculator(t *testing.T) {
	simple := NewSimpleScoringStrategy()
	calculator := NewScoreCalculator(simple)

	t.Run("calculate with strategy", func(t *testing.T) {
		context := &ScoringContext{Difficulty: 5}
		result := calculator.Calculate(1000, context)
		assert.Equal(t, int64(1000), result)
	})

	t.Run("get strategy name", func(t *testing.T) {
		assert.Equal(t, "Simple", calculator.GetStrategyName())
	})

	t.Run("set new strategy", func(t *testing.T) {
		weighted := NewWeightedScoringStrategy(2.0, 0.1)
		calculator.SetStrategy(weighted)
		assert.Equal(t, "Weighted", calculator.GetStrategyName())

		context := &ScoringContext{Difficulty: 2}
		result := calculator.Calculate(1000, context)
		assert.Equal(t, int64(4000), result)
	})

	t.Run("nil strategy", func(t *testing.T) {
		nilCalc := NewScoreCalculator(nil)
		context := &ScoringContext{}
		result := nilCalc.Calculate(1000, context)
		assert.Equal(t, int64(1000), result)
		assert.Equal(t, "None", nilCalc.GetStrategyName())
	})
}

func TestLeaderboardRanker(t *testing.T) {
	standard := NewStandardRankingStrategy()
	ranker := NewLeaderboardRanker(standard)

	scores := []*leaderboardmodels.Score{
		{UserID: uuid.New(), Score: 1000},
		{UserID: uuid.New(), Score: 900},
		{UserID: uuid.New(), Score: 800},
	}

	t.Run("rank scores", func(t *testing.T) {
		rankedScores := ranker.Rank(scores)
		require.Len(t, rankedScores, 3)
		assert.Equal(t, 1, rankedScores[0].Rank)
		assert.Equal(t, 2, rankedScores[1].Rank)
		assert.Equal(t, 3, rankedScores[2].Rank)
	})

	t.Run("get strategy name", func(t *testing.T) {
		assert.Equal(t, "Standard", ranker.GetStrategyName())
	})

	t.Run("set new strategy", func(t *testing.T) {
		dense := NewDenseRankingStrategy()
		ranker.SetStrategy(dense)
		assert.Equal(t, "Dense", ranker.GetStrategyName())
	})

	t.Run("nil strategy", func(t *testing.T) {
		nilRanker := NewLeaderboardRanker(nil)
		result := nilRanker.Rank(scores)
		require.Len(t, result, 3)
	})

	t.Run("empty scores", func(t *testing.T) {
		result := ranker.Rank([]*leaderboardmodels.Score{})
		assert.Empty(t, result)
	})
}
