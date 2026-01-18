package strategy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleScoringStrategy(t *testing.T) {
	strategy := NewSimpleScoringStrategy()
	assert.Equal(t, "Simple", strategy.Name())

	context := &ScoringContext{
		Difficulty: 5,
		Combo:      10,
	}

	result := strategy.Calculate(1000, context)
	assert.Equal(t, int64(1000), result)
}

func TestWeightedScoringStrategy(t *testing.T) {
	strategy := NewWeightedScoringStrategy(1.5, 0.1)
	assert.Equal(t, "Weighted", strategy.Name())

	t.Run("with difficulty", func(t *testing.T) {
		context := &ScoringContext{
			Difficulty: 5,
			Combo:      0,
		}
		result := strategy.Calculate(1000, context)
		expected := int64(1000 * 5 * 1.5)
		assert.Equal(t, expected, result)
	})

	t.Run("with combo", func(t *testing.T) {
		context := &ScoringContext{
			Difficulty: 0,
			Combo:      10,
		}
		result := strategy.Calculate(1000, context)
		assert.Equal(t, int64(2000), result)
	})

	t.Run("with both", func(t *testing.T) {
		context := &ScoringContext{
			Difficulty: 2,
			Combo:      5,
		}
		result := strategy.Calculate(1000, context)
		assert.Equal(t, int64(4500), result)
	})
}

func TestCompositeScoringStrategy(t *testing.T) {
	simple := NewSimpleScoringStrategy()
	weighted := NewWeightedScoringStrategy(1.5, 0.1)
	composite := NewCompositeScoringStrategy(simple, weighted)

	assert.Equal(t, "Composite", composite.Name())

	context := &ScoringContext{
		Difficulty: 2,
		Combo:      0,
	}

	result := composite.Calculate(1000, context)
	assert.Equal(t, int64(3000), result)
}
