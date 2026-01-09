package strategy

import (
	"sort"

	leaderboardmodels "leaderboard-service/internal/leaderboard/models"
)

// Ranking Strategies - различные стратегии ранжирования

// StandardRankingStrategy - стандартная стратегия ранжирования (1, 2, 3, 4, ...)
// Каждый игрок получает уникальный ранг
type StandardRankingStrategy struct{}

func NewStandardRankingStrategy() *StandardRankingStrategy {
	return &StandardRankingStrategy{}
}

func (s *StandardRankingStrategy) CalculateRanks(scores []*leaderboardmodels.Score) []*RankedScore {
	if len(scores) == 0 {
		return []*RankedScore{}
	}

	// Сортируем по убыванию счета
	sorted := make([]*leaderboardmodels.Score, len(scores))
	copy(sorted, scores)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Score > sorted[j].Score
	})

	// Присваиваем ранги
	ranked := make([]*RankedScore, len(sorted))
	for i, score := range sorted {
		ranked[i] = &RankedScore{
			Score:        score,
			Rank:         i + 1,
			TiedWithPrev: false,
		}
	}

	return ranked
}

func (s *StandardRankingStrategy) Name() string {
	return "Standard"
}

// DenseRankingStrategy - плотная стратегия ранжирования (1, 2, 2, 3, ...)
// Игроки с одинаковым счетом получают одинаковый ранг,
// следующий ранг = предыдущий + 1 (без пропусков)
type DenseRankingStrategy struct{}

func NewDenseRankingStrategy() *DenseRankingStrategy {
	return &DenseRankingStrategy{}
}

func (s *DenseRankingStrategy) CalculateRanks(scores []*leaderboardmodels.Score) []*RankedScore {
	if len(scores) == 0 {
		return []*RankedScore{}
	}

	// Сортируем по убыванию счета
	sorted := make([]*leaderboardmodels.Score, len(scores))
	copy(sorted, scores)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Score > sorted[j].Score
	})

	// Присваиваем ранги
	ranked := make([]*RankedScore, len(sorted))
	currentRank := 1

	for i, score := range sorted {
		if i > 0 && sorted[i-1].Score != score.Score {
			currentRank++
		}

		ranked[i] = &RankedScore{
			Score:        score,
			Rank:         currentRank,
			TiedWithPrev: i > 0 && sorted[i-1].Score == score.Score,
		}
	}

	return ranked
}

func (s *DenseRankingStrategy) Name() string {
	return "Dense"
}

// CompetitionRankingStrategy - соревновательная стратегия (1, 2, 2, 4, ...)
// Игроки с одинаковым счетом получают одинаковый ранг,
// следующий ранг пропускает номера (как в спорте)
type CompetitionRankingStrategy struct{}

func NewCompetitionRankingStrategy() *CompetitionRankingStrategy {
	return &CompetitionRankingStrategy{}
}

func (s *CompetitionRankingStrategy) CalculateRanks(scores []*leaderboardmodels.Score) []*RankedScore {
	if len(scores) == 0 {
		return []*RankedScore{}
	}

	// Сортируем по убыванию счета
	sorted := make([]*leaderboardmodels.Score, len(scores))
	copy(sorted, scores)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Score > sorted[j].Score
	})

	// Присваиваем ранги
	ranked := make([]*RankedScore, len(sorted))
	currentRank := 1

	for i, score := range sorted {
		if i > 0 && sorted[i-1].Score != score.Score {
			currentRank = i + 1
		}

		ranked[i] = &RankedScore{
			Score:        score,
			Rank:         currentRank,
			TiedWithPrev: i > 0 && sorted[i-1].Score == score.Score,
		}
	}

	return ranked
}

func (s *CompetitionRankingStrategy) Name() string {
	return "Competition"
}

// ModifiedCompetitionRankingStrategy - модифицированная соревновательная стратегия
// Игроки с одинаковым счетом получают средний ранг из диапазона
// Например, если 2-е и 3-е место равны, оба получат ранг 2.5
type ModifiedCompetitionRankingStrategy struct{}

func NewModifiedCompetitionRankingStrategy() *ModifiedCompetitionRankingStrategy {
	return &ModifiedCompetitionRankingStrategy{}
}

func (s *ModifiedCompetitionRankingStrategy) CalculateRanks(scores []*leaderboardmodels.Score) []*RankedScore {
	if len(scores) == 0 {
		return []*RankedScore{}
	}

	// Сортируем по убыванию счета
	sorted := make([]*leaderboardmodels.Score, len(scores))
	copy(sorted, scores)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Score > sorted[j].Score
	})

	// Присваиваем ранги
	ranked := make([]*RankedScore, len(sorted))

	i := 0
	for i < len(sorted) {
		// Находим всех с таким же счетом
		tieStart := i
		tieEnd := i
		for tieEnd < len(sorted) && sorted[tieEnd].Score == sorted[tieStart].Score {
			tieEnd++
		}

		// Вычисляем средний ранг
		avgRank := (tieStart + 1 + tieEnd) / 2

		// Присваиваем всем игрокам в группе
		for j := tieStart; j < tieEnd; j++ {
			ranked[j] = &RankedScore{
				Score:        sorted[j],
				Rank:         avgRank,
				TiedWithPrev: j > tieStart,
			}
		}

		i = tieEnd
	}

	return ranked
}

func (s *ModifiedCompetitionRankingStrategy) Name() string {
	return "ModifiedCompetition"
}

// OrdinalRankingStrategy - порядковая стратегия
// Определяет ранг на основе порядка добавления счета (по времени)
type OrdinalRankingStrategy struct{}

func NewOrdinalRankingStrategy() *OrdinalRankingStrategy {
	return &OrdinalRankingStrategy{}
}

func (s *OrdinalRankingStrategy) CalculateRanks(scores []*leaderboardmodels.Score) []*RankedScore {
	if len(scores) == 0 {
		return []*RankedScore{}
	}

	// Сортируем по счету (убывание), затем по времени (возрастание)
	sorted := make([]*leaderboardmodels.Score, len(scores))
	copy(sorted, scores)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Score != sorted[j].Score {
			return sorted[i].Score > sorted[j].Score
		}
		return sorted[i].Timestamp.Before(sorted[j].Timestamp)
	})

	// Присваиваем уникальные ранги
	ranked := make([]*RankedScore, len(sorted))
	for i, score := range sorted {
		ranked[i] = &RankedScore{
			Score:        score,
			Rank:         i + 1,
			TiedWithPrev: false,
		}
	}

	return ranked
}

func (s *OrdinalRankingStrategy) Name() string {
	return "Ordinal"
}

// PercentileRankingStrategy - процентильная стратегия
// Ранг определяется как процентиль в общей популяции
type PercentileRankingStrategy struct{}

func NewPercentileRankingStrategy() *PercentileRankingStrategy {
	return &PercentileRankingStrategy{}
}

func (s *PercentileRankingStrategy) CalculateRanks(scores []*leaderboardmodels.Score) []*RankedScore {
	if len(scores) == 0 {
		return []*RankedScore{}
	}

	// Сортируем по убыванию счета
	sorted := make([]*leaderboardmodels.Score, len(scores))
	copy(sorted, scores)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Score > sorted[j].Score
	})

	total := len(sorted)

	// Вычисляем процентильный ранг
	ranked := make([]*RankedScore, len(sorted))
	for i, score := range sorted {
		// Процентиль: (total - i) / total * 100
		percentileRank := ((total - i) * 100) / total

		ranked[i] = &RankedScore{
			Score:        score,
			Rank:         percentileRank,
			TiedWithPrev: i > 0 && sorted[i-1].Score == score.Score,
		}
	}

	return ranked
}

func (s *PercentileRankingStrategy) Name() string {
	return "Percentile"
}

// FractionalRankingStrategy - дробная стратегия
// Похожа на модифицированную соревновательную, но использует дробные ранги
type FractionalRankingStrategy struct{}

func NewFractionalRankingStrategy() *FractionalRankingStrategy {
	return &FractionalRankingStrategy{}
}

func (s *FractionalRankingStrategy) CalculateRanks(scores []*leaderboardmodels.Score) []*RankedScore {
	if len(scores) == 0 {
		return []*RankedScore{}
	}

	// Сортируем по убыванию счета
	sorted := make([]*leaderboardmodels.Score, len(scores))
	copy(sorted, scores)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Score > sorted[j].Score
	})

	// Присваиваем ранги
	ranked := make([]*RankedScore, len(sorted))

	i := 0
	for i < len(sorted) {
		// Находим всех с таким же счетом
		tieStart := i
		tieEnd := i
		for tieEnd < len(sorted) && sorted[tieEnd].Score == sorted[tieStart].Score {
			tieEnd++
		}

		// Вычисляем средний ранг (с дробями)
		sumRanks := 0
		for j := tieStart; j < tieEnd; j++ {
			sumRanks += j + 1
		}
		avgRank := sumRanks / (tieEnd - tieStart)

		// Присваиваем всем игрокам в группе
		for j := tieStart; j < tieEnd; j++ {
			ranked[j] = &RankedScore{
				Score:        sorted[j],
				Rank:         avgRank,
				TiedWithPrev: j > tieStart,
			}
		}

		i = tieEnd
	}

	return ranked
}

func (s *FractionalRankingStrategy) Name() string {
	return "Fractional"
}
