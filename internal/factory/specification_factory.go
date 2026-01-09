package factory

import (
	leaderboardmodels "leaderboard-service/internal/leaderboard/models"
	authmodels "leaderboard-service/internal/auth/models"
	"leaderboard-service/internal/shared/repository"

	"github.com/google/uuid"
)

// SpecificationFactory упрощает создание сложных спецификаций

// LeaderboardSpecBuilder fluent builder для спецификаций лидерборда
type LeaderboardSpecBuilder struct {
	season      string
	minScore    *int64
	maxScore    *int64
	userIDs     []uuid.UUID
	excludeIDs  []uuid.UUID
	topN        *int
	offset      *int
	limit       *int
	orderByDesc bool
}

// NewLeaderboardSpec создает новый builder
func NewLeaderboardSpec() *LeaderboardSpecBuilder {
	return &LeaderboardSpecBuilder{
		orderByDesc: true, // По умолчанию сортировка по убыванию
	}
}

// ForSeason устанавливает сезон
func (b *LeaderboardSpecBuilder) ForSeason(season string) *LeaderboardSpecBuilder {
	b.season = season
	return b
}

// WithMinScore устанавливает минимальный счет
func (b *LeaderboardSpecBuilder) WithMinScore(score int64) *LeaderboardSpecBuilder {
	b.minScore = &score
	return b
}

// WithMaxScore устанавливает максимальный счет
func (b *LeaderboardSpecBuilder) WithMaxScore(score int64) *LeaderboardSpecBuilder {
	b.maxScore = &score
	return b
}

// WithScoreRange устанавливает диапазон счетов
func (b *LeaderboardSpecBuilder) WithScoreRange(min, max int64) *LeaderboardSpecBuilder {
	b.minScore = &min
	b.maxScore = &max
	return b
}

// ForUsers фильтрует по определенным пользователям
func (b *LeaderboardSpecBuilder) ForUsers(userIDs ...uuid.UUID) *LeaderboardSpecBuilder {
	b.userIDs = userIDs
	return b
}

// ExcludeUsers исключает определенных пользователей
func (b *LeaderboardSpecBuilder) ExcludeUsers(userIDs ...uuid.UUID) *LeaderboardSpecBuilder {
	b.excludeIDs = userIDs
	return b
}

// TopN ограничивает количество топ записей
func (b *LeaderboardSpecBuilder) TopN(n int) *LeaderboardSpecBuilder {
	b.topN = &n
	return b
}

// WithPagination добавляет пагинацию
func (b *LeaderboardSpecBuilder) WithPagination(page, pageSize int) *LeaderboardSpecBuilder {
	offset := (page - 1) * pageSize
	b.offset = &offset
	b.limit = &pageSize
	return b
}

// WithLimit устанавливает лимит
func (b *LeaderboardSpecBuilder) WithLimit(limit int) *LeaderboardSpecBuilder {
	b.limit = &limit
	return b
}

// WithOffset устанавливает смещение
func (b *LeaderboardSpecBuilder) WithOffset(offset int) *LeaderboardSpecBuilder {
	b.offset = &offset
	return b
}

// OrderByAsc сортировка по возрастанию
func (b *LeaderboardSpecBuilder) OrderByAsc() *LeaderboardSpecBuilder {
	b.orderByDesc = false
	return b
}

// OrderByDesc сортировка по убыванию
func (b *LeaderboardSpecBuilder) OrderByDesc() *LeaderboardSpecBuilder {
	b.orderByDesc = true
	return b
}

// Build создает спецификацию
func (b *LeaderboardSpecBuilder) Build() repository.Specification[leaderboardmodels.Score] {
	specs := []repository.Specification[leaderboardmodels.Score]{}

	// Сезон (обязательно)
	if b.season != "" {
		specs = append(specs, repository.NewScoreBySeasonSpec(b.season))
	}

	// Диапазон счетов
	if b.minScore != nil && b.maxScore != nil {
		specs = append(specs, repository.NewScoreRangeSpec(*b.minScore, *b.maxScore))
	} else if b.minScore != nil {
		specs = append(specs, repository.NewScoreMinValueSpec(*b.minScore))
	} else if b.maxScore != nil {
		specs = append(specs, repository.NewScoreMaxValueSpec(*b.maxScore))
	}

	// Фильтр по пользователям
	if len(b.userIDs) > 0 {
		userSpecs := []repository.Specification[leaderboardmodels.Score]{}
		for _, userID := range b.userIDs {
			userSpecs = append(userSpecs, repository.NewScoreByUserIDSpec(userID))
		}
		specs = append(specs, repository.Or(userSpecs...))
	}

	// Исключение пользователей
	if len(b.excludeIDs) > 0 {
		for _, userID := range b.excludeIDs {
			specs = append(specs, repository.Not(repository.NewScoreByUserIDSpec(userID)))
		}
	}

	// Сортировка
	specs = append(specs, repository.NewScoreOrderBySpec("score", b.orderByDesc))

	// TopN или Limit
	if b.topN != nil {
		specs = append(specs, repository.NewScoreTopNSpec(*b.topN))
	} else {
		// Offset
		if b.offset != nil {
			specs = append(specs, repository.NewScoreOffsetSpec(*b.offset))
		}

		// Limit
		if b.limit != nil {
			specs = append(specs, repository.NewScoreLimitSpec(*b.limit))
		}
	}

	return repository.And(specs...)
}

// UserSearchSpecBuilder fluent builder для поиска пользователей
type UserSearchSpecBuilder struct {
	query        string
	domain       string
	createdAfter *string
	limit        *int
	orderBy      string
	orderDesc    bool
}

// NewUserSearchSpec создает новый builder
func NewUserSearchSpec() *UserSearchSpecBuilder {
	return &UserSearchSpecBuilder{}
}

// WithQuery устанавливает поисковый запрос
func (b *UserSearchSpecBuilder) WithQuery(query string) *UserSearchSpecBuilder {
	b.query = query
	return b
}

// InDomain фильтрует по домену email
func (b *UserSearchSpecBuilder) InDomain(domain string) *UserSearchSpecBuilder {
	b.domain = domain
	return b
}

// CreatedAfter фильтрует по дате создания
func (b *UserSearchSpecBuilder) CreatedAfter(date string) *UserSearchSpecBuilder {
	b.createdAfter = &date
	return b
}

// WithLimit устанавливает лимит
func (b *UserSearchSpecBuilder) WithLimit(limit int) *UserSearchSpecBuilder {
	b.limit = &limit
	return b
}

// OrderBy устанавливает сортировку
func (b *UserSearchSpecBuilder) OrderBy(field string, desc bool) *UserSearchSpecBuilder {
	b.orderBy = field
	b.orderDesc = desc
	return b
}

// Build создает спецификацию
func (b *UserSearchSpecBuilder) Build() repository.Specification[authmodels.User] {
	specs := []repository.Specification[authmodels.User]{}

	// Поиск по имени или email
	if b.query != "" {
		specs = append(specs, repository.Or(
			repository.NewUserByNameSpec(b.query),
			repository.NewUserByEmailSpec(b.query),
		))
	}

	// Фильтр по домену
	if b.domain != "" {
		specs = append(specs, repository.NewUserByEmailDomainSpec(b.domain))
	}

	// Фильтр по дате
	if b.createdAfter != nil {
		specs = append(specs, repository.NewUserCreatedAfterSpec(*b.createdAfter))
	}

	// Сортировка
	if b.orderBy != "" {
		specs = append(specs, repository.NewUserOrderBySpec(b.orderBy, b.orderDesc))
	}

	// Лимит
	if b.limit != nil {
		specs = append(specs, repository.NewUserLimitSpec(*b.limit))
	}

	if len(specs) == 0 {
		return nil
	}

	return repository.And(specs...)
}

// QuickSpecifications быстрые готовые спецификации

// GlobalTopN создает спецификацию для топ N глобального лидерборда
func GlobalTopN(n int) repository.Specification[leaderboardmodels.Score] {
	return NewLeaderboardSpec().
		ForSeason("global").
		TopN(n).
		Build()
}

// SeasonTopN создает спецификацию для топ N лидерборда сезона
func SeasonTopN(season string, n int) repository.Specification[leaderboardmodels.Score] {
	return NewLeaderboardSpec().
		ForSeason(season).
		TopN(n).
		Build()
}

// HighScores создает спецификацию для высоких счетов
func HighScores(season string, minScore int64, limit int) repository.Specification[leaderboardmodels.Score] {
	return NewLeaderboardSpec().
		ForSeason(season).
		WithMinScore(minScore).
		WithLimit(limit).
		Build()
}

// PaginatedLeaderboard создает спецификацию для пагинированного лидерборда
func PaginatedLeaderboard(season string, page, pageSize int) repository.Specification[leaderboardmodels.Score] {
	return NewLeaderboardSpec().
		ForSeason(season).
		WithPagination(page, pageSize).
		Build()
}

// MidRangeScores создает спецификацию для счетов в диапазоне
func MidRangeScores(season string, minScore, maxScore int64, limit int) repository.Specification[leaderboardmodels.Score] {
	return NewLeaderboardSpec().
		ForSeason(season).
		WithScoreRange(minScore, maxScore).
		WithLimit(limit).
		Build()
}

// UserScoresInSeason создает спецификацию для всех счетов пользователя в сезоне
func UserScoresInSeason(userID uuid.UUID, season string) repository.Specification[leaderboardmodels.Score] {
	return repository.And(
		repository.NewScoreByUserIDSpec(userID),
		repository.NewScoreBySeasonSpec(season),
		repository.NewScoreOrderBySpec("score", true),
	)
}

// SearchUsers создает спецификацию для поиска пользователей
func SearchUsers(query string, limit int) repository.Specification[authmodels.User] {
	return NewUserSearchSpec().
		WithQuery(query).
		WithLimit(limit).
		Build()
}

// UsersInDomain создает спецификацию для пользователей домена
func UsersInDomain(domain string, limit int) repository.Specification[authmodels.User] {
	return NewUserSearchSpec().
		InDomain(domain).
		WithLimit(limit).
		Build()
}

// RecentUsers создает спецификацию для недавних пользователей
func RecentUsers(afterDate string, limit int) repository.Specification[authmodels.User] {
	return NewUserSearchSpec().
		CreatedAfter(afterDate).
		OrderBy("created_at", true).
		WithLimit(limit).
		Build()
}
