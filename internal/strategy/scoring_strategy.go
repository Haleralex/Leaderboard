package strategy

import (
	"math"
)

// Scoring Strategies - различные стратегии подсчета очков

// SimpleScoringStrategy - простая стратегия, возвращает базовый счет без изменений
type SimpleScoringStrategy struct{}

func NewSimpleScoringStrategy() *SimpleScoringStrategy {
	return &SimpleScoringStrategy{}
}

func (s *SimpleScoringStrategy) Calculate(baseScore int64, context *ScoringContext) int64 {
	return baseScore
}

func (s *SimpleScoringStrategy) Name() string {
	return "Simple"
}

// WeightedScoringStrategy - стратегия с весовыми коэффициентами
type WeightedScoringStrategy struct {
	DifficultyMultiplier float64
	ComboBonus           float64
}

func NewWeightedScoringStrategy(difficultyMultiplier, comboBonus float64) *WeightedScoringStrategy {
	return &WeightedScoringStrategy{
		DifficultyMultiplier: difficultyMultiplier,
		ComboBonus:           comboBonus,
	}
}

func (s *WeightedScoringStrategy) Calculate(baseScore int64, context *ScoringContext) int64 {
	score := float64(baseScore)

	// Умножаем на сложность
	if context.Difficulty > 0 {
		score *= float64(context.Difficulty) * s.DifficultyMultiplier
	}

	// Добавляем бонус за комбо
	if context.Combo > 0 {
		comboMultiplier := 1.0 + (float64(context.Combo) * s.ComboBonus)
		score *= comboMultiplier
	}

	// Применяем общий множитель
	if context.Multiplier > 0 {
		score *= context.Multiplier
	}

	return int64(math.Round(score))
}

func (s *WeightedScoringStrategy) Name() string {
	return "Weighted"
}

// BonusScoringStrategy - стратегия с бонусами за достижения
type BonusScoringStrategy struct {
	TimeBonusEnabled       bool
	AchievementBonusPoints int64
	MaxTimeBonus           int64
}

func NewBonusScoringStrategy(timeBonusEnabled bool, achievementBonus, maxTimeBonus int64) *BonusScoringStrategy {
	return &BonusScoringStrategy{
		TimeBonusEnabled:       timeBonusEnabled,
		AchievementBonusPoints: achievementBonus,
		MaxTimeBonus:           maxTimeBonus,
	}
}

func (s *BonusScoringStrategy) Calculate(baseScore int64, context *ScoringContext) int64 {
	score := baseScore

	// Добавляем временной бонус
	if s.TimeBonusEnabled && context.TimeBonus > 0 {
		timeBonus := context.TimeBonus
		if timeBonus > s.MaxTimeBonus {
			timeBonus = s.MaxTimeBonus
		}
		score += timeBonus
	}

	// Добавляем бонус за достижения
	if len(context.Achievements) > 0 {
		score += int64(len(context.Achievements)) * s.AchievementBonusPoints
	}

	return score
}

func (s *BonusScoringStrategy) Name() string {
	return "Bonus"
}

// MultiplayerScoringStrategy - стратегия для мультиплеерных игр
type MultiplayerScoringStrategy struct {
	TeamBonus     float64
	KillsWeight   float64
	DeathsPenalty float64
	AssistsWeight float64
}

func NewMultiplayerScoringStrategy(teamBonus, killsWeight, deathsPenalty, assistsWeight float64) *MultiplayerScoringStrategy {
	return &MultiplayerScoringStrategy{
		TeamBonus:     teamBonus,
		KillsWeight:   killsWeight,
		DeathsPenalty: deathsPenalty,
		AssistsWeight: assistsWeight,
	}
}

func (s *MultiplayerScoringStrategy) Calculate(baseScore int64, context *ScoringContext) int64 {
	score := float64(baseScore)

	// Получаем статистику из метаданных
	kills := s.getMetadataInt(context, "kills", 0)
	deaths := s.getMetadataInt(context, "deaths", 0)
	assists := s.getMetadataInt(context, "assists", 0)
	isTeamWin := s.getMetadataBool(context, "team_win", false)

	// Добавляем очки за убийства
	score += float64(kills) * s.KillsWeight

	// Вычитаем за смерти
	score -= float64(deaths) * s.DeathsPenalty

	// Добавляем за ассисты
	score += float64(assists) * s.AssistsWeight

	// Бонус за победу команды
	if isTeamWin {
		score *= (1.0 + s.TeamBonus)
	}

	// Не допускаем отрицательных значений
	if score < 0 {
		score = 0
	}

	return int64(math.Round(score))
}

func (s *MultiplayerScoringStrategy) getMetadataInt(context *ScoringContext, key string, defaultValue int) int {
	if context.Metadata == nil {
		return defaultValue
	}
	if val, ok := context.Metadata[key]; ok {
		if intVal, ok := val.(int); ok {
			return intVal
		}
		if floatVal, ok := val.(float64); ok {
			return int(floatVal)
		}
	}
	return defaultValue
}

func (s *MultiplayerScoringStrategy) getMetadataBool(context *ScoringContext, key string, defaultValue bool) bool {
	if context.Metadata == nil {
		return defaultValue
	}
	if val, ok := context.Metadata[key]; ok {
		if boolVal, ok := val.(bool); ok {
			return boolVal
		}
	}
	return defaultValue
}

func (s *MultiplayerScoringStrategy) Name() string {
	return "Multiplayer"
}

// CompositeScoringStrategy - композитная стратегия, объединяющая несколько стратегий
type CompositeScoringStrategy struct {
	strategies []ScoringStrategy
}

func NewCompositeScoringStrategy(strategies ...ScoringStrategy) *CompositeScoringStrategy {
	return &CompositeScoringStrategy{
		strategies: strategies,
	}
}

func (s *CompositeScoringStrategy) Calculate(baseScore int64, context *ScoringContext) int64 {
	score := baseScore
	for _, strategy := range s.strategies {
		score = strategy.Calculate(score, context)
	}
	return score
}

func (s *CompositeScoringStrategy) Name() string {
	return "Composite"
}

// PercentageScoringStrategy - стратегия с процентными бонусами
type PercentageScoringStrategy struct {
	BonusPercentage float64 // Процент бонуса (например, 0.5 = 50%)
}

func NewPercentageScoringStrategy(bonusPercentage float64) *PercentageScoringStrategy {
	return &PercentageScoringStrategy{
		BonusPercentage: bonusPercentage,
	}
}

func (s *PercentageScoringStrategy) Calculate(baseScore int64, context *ScoringContext) int64 {
	bonus := float64(baseScore) * s.BonusPercentage
	return baseScore + int64(math.Round(bonus))
}

func (s *PercentageScoringStrategy) Name() string {
	return "Percentage"
}

// ThresholdScoringStrategy - стратегия с порогами
type ThresholdScoringStrategy struct {
	Thresholds []ThresholdBonus
}

type ThresholdBonus struct {
	MinScore int64
	Bonus    int64
}

func NewThresholdScoringStrategy(thresholds []ThresholdBonus) *ThresholdScoringStrategy {
	return &ThresholdScoringStrategy{
		Thresholds: thresholds,
	}
}

func (s *ThresholdScoringStrategy) Calculate(baseScore int64, context *ScoringContext) int64 {
	score := baseScore

	for _, threshold := range s.Thresholds {
		if baseScore >= threshold.MinScore {
			score += threshold.Bonus
		}
	}

	return score
}

func (s *ThresholdScoringStrategy) Name() string {
	return "Threshold"
}

// SeasonalScoringStrategy - стратегия с сезонными множителями
type SeasonalScoringStrategy struct {
	SeasonMultipliers map[string]float64
	DefaultMultiplier float64
}

func NewSeasonalScoringStrategy(seasonMultipliers map[string]float64, defaultMultiplier float64) *SeasonalScoringStrategy {
	return &SeasonalScoringStrategy{
		SeasonMultipliers: seasonMultipliers,
		DefaultMultiplier: defaultMultiplier,
	}
}

func (s *SeasonalScoringStrategy) Calculate(baseScore int64, context *ScoringContext) int64 {
	multiplier := s.DefaultMultiplier

	if mult, ok := s.SeasonMultipliers[context.Season]; ok {
		multiplier = mult
	}

	return int64(math.Round(float64(baseScore) * multiplier))
}

func (s *SeasonalScoringStrategy) Name() string {
	return "Seasonal"
}
