package strategy

import (
	"context"
	"time"

	leaderboardmodels "leaderboard-service/internal/leaderboard/models"

	"github.com/google/uuid"
)

// Context - объекты, которые используют стратегии

// ScoreCalculator - контекст для вычисления счетов
type ScoreCalculator struct {
	strategy ScoringStrategy
}

// NewScoreCalculator создает новый калькулятор счетов
func NewScoreCalculator(strategy ScoringStrategy) *ScoreCalculator {
	return &ScoreCalculator{
		strategy: strategy,
	}
}

// SetStrategy устанавливает новую стратегию
func (c *ScoreCalculator) SetStrategy(strategy ScoringStrategy) {
	c.strategy = strategy
}

// Calculate вычисляет счет
func (c *ScoreCalculator) Calculate(baseScore int64, context *ScoringContext) int64 {
	if c.strategy == nil {
		return baseScore
	}
	return c.strategy.Calculate(baseScore, context)
}

// GetStrategyName возвращает название текущей стратегии
func (c *ScoreCalculator) GetStrategyName() string {
	if c.strategy == nil {
		return "None"
	}
	return c.strategy.Name()
}

// LeaderboardRanker - контекст для ранжирования лидерборда
type LeaderboardRanker struct {
	strategy RankingStrategy
}

// NewLeaderboardRanker создает новый ранкер лидерборда
func NewLeaderboardRanker(strategy RankingStrategy) *LeaderboardRanker {
	return &LeaderboardRanker{
		strategy: strategy,
	}
}

// SetStrategy устанавливает новую стратегию
func (r *LeaderboardRanker) SetStrategy(strategy RankingStrategy) {
	r.strategy = strategy
}

// Rank ранжирует список счетов
func (r *LeaderboardRanker) Rank(scores []*leaderboardmodels.Score) []*RankedScore {
	if r.strategy == nil {
		// По умолчанию стандартная стратегия
		r.strategy = NewStandardRankingStrategy()
	}
	return r.strategy.CalculateRanks(scores)
}

// GetStrategyName возвращает название текущей стратегии
func (r *LeaderboardRanker) GetStrategyName() string {
	if r.strategy == nil {
		return "None"
	}
	return r.strategy.Name()
}

// CachedRepository - контекст для кэширования с выбираемой стратегией
type CachedRepository struct {
	strategy CacheStrategy
}

// NewCachedRepository создает новый репозиторий с кэшированием
func NewCachedRepository(strategy CacheStrategy) *CachedRepository {
	return &CachedRepository{
		strategy: strategy,
	}
}

// SetStrategy устанавливает новую стратегию кэширования
func (r *CachedRepository) SetStrategy(strategy CacheStrategy) {
	r.strategy = strategy
}

// Get получает значение из кэша
func (r *CachedRepository) Get(ctx context.Context, key string) ([]byte, error) {
	if r.strategy == nil {
		return nil, nil
	}
	return r.strategy.Get(ctx, key)
}

// Set сохраняет значение в кэш
func (r *CachedRepository) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if r.strategy == nil {
		return nil
	}
	return r.strategy.Set(ctx, key, value, ttl)
}

// Delete удаляет значение из кэша
func (r *CachedRepository) Delete(ctx context.Context, key string) error {
	if r.strategy == nil {
		return nil
	}
	return r.strategy.Delete(ctx, key)
}

// ScoreProcessor - процессор счетов с множественными стратегиями
type ScoreProcessor struct {
	scoringStrategy    ScoringStrategy
	validationStrategy ValidationStrategy
	rankingStrategy    RankingStrategy
}

// NewScoreProcessor создает новый процессор
func NewScoreProcessor(
	scoring ScoringStrategy,
	validation ValidationStrategy,
	ranking RankingStrategy,
) *ScoreProcessor {
	return &ScoreProcessor{
		scoringStrategy:    scoring,
		validationStrategy: validation,
		rankingStrategy:    ranking,
	}
}

// ProcessScore обрабатывает счет: валидирует, вычисляет, сохраняет
func (p *ScoreProcessor) ProcessScore(
	ctx context.Context,
	baseScore int64,
	scoringCtx *ScoringContext,
	validationCtx interface{},
) (int64, error) {
	// 1. Валидация
	if p.validationStrategy != nil {
		if err := p.validationStrategy.Validate(validationCtx); err != nil {
			return 0, err
		}
	}

	// 2. Вычисление
	finalScore := baseScore
	if p.scoringStrategy != nil {
		finalScore = p.scoringStrategy.Calculate(baseScore, scoringCtx)
	}

	return finalScore, nil
}

// SetScoringStrategy устанавливает стратегию подсчета
func (p *ScoreProcessor) SetScoringStrategy(strategy ScoringStrategy) {
	p.scoringStrategy = strategy
}

// SetValidationStrategy устанавливает стратегию валидации
func (p *ScoreProcessor) SetValidationStrategy(strategy ValidationStrategy) {
	p.validationStrategy = strategy
}

// SetRankingStrategy устанавливает стратегию ранжирования
func (p *ScoreProcessor) SetRankingStrategy(strategy RankingStrategy) {
	p.rankingStrategy = strategy
}

// LeaderboardManager - менеджер лидерборда с выбираемыми стратегиями
type LeaderboardManager struct {
	rankingStrategy RankingStrategy
	sortStrategy    SortStrategy
	filterStrategy  FilterStrategy
}

// NewLeaderboardManager создает новый менеджер лидерборда
func NewLeaderboardManager(
	ranking RankingStrategy,
	sort SortStrategy,
	filter FilterStrategy,
) *LeaderboardManager {
	return &LeaderboardManager{
		rankingStrategy: ranking,
		sortStrategy:    sort,
		filterStrategy:  filter,
	}
}

// GetLeaderboard получает лидерборд с применением всех стратегий
func (m *LeaderboardManager) GetLeaderboard(scores []*leaderboardmodels.Score) []*RankedScore {
	// 1. Фильтрация
	filtered := scores
	if m.filterStrategy != nil {
		filtered = m.filterStrategy.Filter(scores)
	}

	// 2. Сортировка
	sorted := filtered
	if m.sortStrategy != nil {
		sorted = m.sortStrategy.Sort(filtered)
	}

	// 3. Ранжирование
	ranked := []*RankedScore{}
	if m.rankingStrategy != nil {
		ranked = m.rankingStrategy.CalculateRanks(sorted)
	}

	return ranked
}

// SetRankingStrategy устанавливает стратегию ранжирования
func (m *LeaderboardManager) SetRankingStrategy(strategy RankingStrategy) {
	m.rankingStrategy = strategy
}

// SetSortStrategy устанавливает стратегию сортировки
func (m *LeaderboardManager) SetSortStrategy(strategy SortStrategy) {
	m.sortStrategy = strategy
}

// SetFilterStrategy устанавливает стратегию фильтрации
func (m *LeaderboardManager) SetFilterStrategy(strategy FilterStrategy) {
	m.filterStrategy = strategy
}

// StrategyFactory - фабрика для создания стратегий
type StrategyFactory struct{}

// NewStrategyFactory создает новую фабрику стратегий
func NewStrategyFactory() *StrategyFactory {
	return &StrategyFactory{}
}

// CreateScoringStrategy создает стратегию подсчета по имени
func (f *StrategyFactory) CreateScoringStrategy(name string) ScoringStrategy {
	switch name {
	case "simple":
		return NewSimpleScoringStrategy()
	case "weighted":
		return NewWeightedScoringStrategy(1.5, 0.1)
	case "bonus":
		return NewBonusScoringStrategy(true, 100, 1000)
	case "multiplayer":
		return NewMultiplayerScoringStrategy(0.2, 100, 50, 50)
	case "percentage":
		return NewPercentageScoringStrategy(0.5)
	default:
		return NewSimpleScoringStrategy()
	}
}

// CreateRankingStrategy создает стратегию ранжирования по имени
func (f *StrategyFactory) CreateRankingStrategy(name string) RankingStrategy {
	switch name {
	case "standard":
		return NewStandardRankingStrategy()
	case "dense":
		return NewDenseRankingStrategy()
	case "competition":
		return NewCompetitionRankingStrategy()
	case "modified":
		return NewModifiedCompetitionRankingStrategy()
	case "ordinal":
		return NewOrdinalRankingStrategy()
	case "percentile":
		return NewPercentileRankingStrategy()
	case "fractional":
		return NewFractionalRankingStrategy()
	default:
		return NewStandardRankingStrategy()
	}
}

// StrategyRegistry - реестр стратегий (для динамического выбора)
type StrategyRegistry struct {
	scoringStrategies map[string]ScoringStrategy
	rankingStrategies map[string]RankingStrategy
}

// NewStrategyRegistry создает новый реестр
func NewStrategyRegistry() *StrategyRegistry {
	return &StrategyRegistry{
		scoringStrategies: make(map[string]ScoringStrategy),
		rankingStrategies: make(map[string]RankingStrategy),
	}
}

// RegisterScoringStrategy регистрирует стратегию подсчета
func (r *StrategyRegistry) RegisterScoringStrategy(name string, strategy ScoringStrategy) {
	r.scoringStrategies[name] = strategy
}

// RegisterRankingStrategy регистрирует стратегию ранжирования
func (r *StrategyRegistry) RegisterRankingStrategy(name string, strategy RankingStrategy) {
	r.rankingStrategies[name] = strategy
}

// GetScoringStrategy получает стратегию подсчета по имени
func (r *StrategyRegistry) GetScoringStrategy(name string) (ScoringStrategy, bool) {
	strategy, ok := r.scoringStrategies[name]
	return strategy, ok
}

// GetRankingStrategy получает стратегию ранжирования по имени
func (r *StrategyRegistry) GetRankingStrategy(name string) (RankingStrategy, bool) {
	strategy, ok := r.rankingStrategies[name]
	return strategy, ok
}

// GameSession - игровая сессия с динамическим выбором стратегий
type GameSession struct {
	ID                  uuid.UUID
	GameMode            string
	ScoringStrategyName string
	RankingStrategyName string
	registry            *StrategyRegistry
}

// NewGameSession создает новую игровую сессию
func NewGameSession(id uuid.UUID, gameMode string, registry *StrategyRegistry) *GameSession {
	return &GameSession{
		ID:                  id,
		GameMode:            gameMode,
		ScoringStrategyName: "simple",
		RankingStrategyName: "standard",
		registry:            registry,
	}
}

// CalculateScore вычисляет счет используя текущую стратегию
func (s *GameSession) CalculateScore(baseScore int64, ctx *ScoringContext) (int64, error) {
	strategy, ok := s.registry.GetScoringStrategy(s.ScoringStrategyName)
	if !ok {
		strategy = NewSimpleScoringStrategy()
	}

	return strategy.Calculate(baseScore, ctx), nil
}

// RankPlayers ранжирует игроков используя текущую стратегию
func (s *GameSession) RankPlayers(scores []*leaderboardmodels.Score) ([]*RankedScore, error) {
	strategy, ok := s.registry.GetRankingStrategy(s.RankingStrategyName)
	if !ok {
		strategy = NewStandardRankingStrategy()
	}

	return strategy.CalculateRanks(scores), nil
}

// SetScoringStrategy устанавливает стратегию подсчета
func (s *GameSession) SetScoringStrategy(name string) {
	s.ScoringStrategyName = name
}

// SetRankingStrategy устанавливает стратегию ранжирования
func (s *GameSession) SetRankingStrategy(name string) {
	s.RankingStrategyName = name
}
