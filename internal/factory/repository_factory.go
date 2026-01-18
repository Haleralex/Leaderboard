package factory

import (
	"time"

	"github.com/redis/go-redis/v9"

	authrepository "leaderboard-service/internal/auth/repository"
	leaderboardrepository "leaderboard-service/internal/leaderboard/repository"
	"leaderboard-service/internal/shared/database"
	"leaderboard-service/internal/shared/repository"
	"leaderboard-service/internal/shared/repository/decorators"
)

// DefaultRepositoryFactory стандартная фабрика репозиториев
type DefaultRepositoryFactory struct {
	config *RepositoryConfig
	cache  *decorators.SimpleCache
}

// NewRepositoryFactory создает новую фабрику репозиториев
func NewRepositoryFactory(config *RepositoryConfig) RepositoryFactory {
	return &DefaultRepositoryFactory{
		config: config,
		cache:  decorators.NewSimpleCache(),
	}
}

// CreateUserRepository создает репозиторий пользователей с декораторами
func (f *DefaultRepositoryFactory) CreateUserRepository() repository.UserRepository {
	// 1. Создаем базовый репозиторий
	baseRepo := authrepository.NewPostgresUserRepository(f.config.DB)

	// 2. Оборачиваем в декораторы в правильном порядке
	var repo repository.UserRepository = baseRepo

	// Сначала кэширование (ближе к базе данных)
	if f.config.EnableCache {
		repo = decorators.NewCachedUserRepository(repo, f.cache)
	}

	// Потом логирование (ближе к бизнес-логике)
	if f.config.EnableLogging {
		repo = decorators.NewLoggedUserRepository(repo)
	}

	return repo
}

// CreateScoreRepository создает репозиторий счетов с декораторами
func (f *DefaultRepositoryFactory) CreateScoreRepository() repository.ScoreRepository {
	// 1. Создаем базовый репозиторий
	baseRepo := leaderboardrepository.NewPostgresScoreRepository(f.config.DB)

	// 2. Оборачиваем в декораторы в правильном порядке
	var repo repository.ScoreRepository = baseRepo

	// Сначала кэширование (ближе к базе данных)
	if f.config.EnableCache {
		repo = decorators.NewCachedScoreRepository(repo, f.cache)
	}

	// Потом логирование (ближе к бизнес-логике)
	if f.config.EnableLogging {
		repo = decorators.NewLoggedScoreRepository(repo)
	}

	return repo
}

// CreateUnitOfWork создает Unit of Work с настроенными репозиториями
func (f *DefaultRepositoryFactory) CreateUnitOfWork() repository.UnitOfWork {
	// Передаем простые фабрики без декораторов для транзакционного контекста
	userFactory := func(db *database.PostgresDB) repository.UserRepository {
		authrepo := authrepository.NewPostgresUserRepository(db)
		return authrepo
	}
	scoreFactory := func(db *database.PostgresDB) repository.ScoreRepository {
		leaderboardrepo := leaderboardrepository.NewPostgresScoreRepository(db)
		return leaderboardrepo
	}
	return repository.NewUnitOfWork(f.config.DB, userFactory, scoreFactory)
}

// CustomRepositoryFactory позволяет создавать репозитории с кастомной логикой
type CustomRepositoryFactory struct {
	config            *RepositoryConfig
	cache             *decorators.SimpleCache
	userRepoBuilder   func(*database.PostgresDB) repository.UserRepository
	scoreRepoBuilder  func(*database.PostgresDB) repository.ScoreRepository
	decoratorBuilders []DecoratorBuilder
}

// DecoratorBuilder функция для построения декоратора
type DecoratorBuilder func(repo interface{}) interface{}

// NewCustomRepositoryFactory создает кастомную фабрику
func NewCustomRepositoryFactory(config *RepositoryConfig) *CustomRepositoryFactory {
	return &CustomRepositoryFactory{
		config:            config,
		cache:             decorators.NewSimpleCache(),
		decoratorBuilders: []DecoratorBuilder{},
	}
}

// WithUserRepositoryBuilder устанавливает кастомный builder для user репозитория
func (f *CustomRepositoryFactory) WithUserRepositoryBuilder(
	builder func(*database.PostgresDB) repository.UserRepository,
) *CustomRepositoryFactory {
	f.userRepoBuilder = builder
	return f
}

// WithScoreRepositoryBuilder устанавливает кастомный builder для score репозитория
func (f *CustomRepositoryFactory) WithScoreRepositoryBuilder(
	builder func(*database.PostgresDB) repository.ScoreRepository,
) *CustomRepositoryFactory {
	f.scoreRepoBuilder = builder
	return f
}

// WithDecorator добавляет кастомный декоратор
func (f *CustomRepositoryFactory) WithDecorator(builder DecoratorBuilder) *CustomRepositoryFactory {
	f.decoratorBuilders = append(f.decoratorBuilders, builder)
	return f
}

// CreateUserRepository создает user репозиторий
func (f *CustomRepositoryFactory) CreateUserRepository() repository.UserRepository {
	var repo repository.UserRepository

	// Используем кастомный builder или стандартный
	if f.userRepoBuilder != nil {
		repo = f.userRepoBuilder(f.config.DB)
	} else {
		repo = authrepository.NewPostgresUserRepository(f.config.DB)
	}

	// Применяем стандартные декораторы
	if f.config.EnableCache {
		repo = decorators.NewCachedUserRepository(repo, f.cache)
	}

	if f.config.EnableLogging {
		repo = decorators.NewLoggedUserRepository(repo)
	}

	// Применяем кастомные декораторы
	for _, builder := range f.decoratorBuilders {
		if decorated, ok := builder(repo).(repository.UserRepository); ok {
			repo = decorated
		}
	}

	return repo
}

// CreateScoreRepository создает score репозиторий
func (f *CustomRepositoryFactory) CreateScoreRepository() repository.ScoreRepository {
	var repo repository.ScoreRepository

	// Используем кастомный builder или стандартный
	if f.scoreRepoBuilder != nil {
		repo = f.scoreRepoBuilder(f.config.DB)
	} else {
		repo = leaderboardrepository.NewPostgresScoreRepository(f.config.DB)
	}

	// Применяем стандартные декораторы
	if f.config.EnableCache {
		repo = decorators.NewCachedScoreRepository(repo, f.cache)
	}

	if f.config.EnableLogging {
		repo = decorators.NewLoggedScoreRepository(repo)
	}

	// Применяем кастомные декораторы
	for _, builder := range f.decoratorBuilders {
		if decorated, ok := builder(repo).(repository.ScoreRepository); ok {
			repo = decorated
		}
	}

	return repo
}

// CreateUnitOfWork создает Unit of Work
func (f *CustomRepositoryFactory) CreateUnitOfWork() repository.UnitOfWork {
	return repository.NewUnitOfWork(
		f.config.DB,
		authrepository.NewPostgresUserRepository,
		leaderboardrepository.NewPostgresScoreRepository,
	)
}

// RepositoryFactoryBuilder builder для фабрики репозиториев (fluent interface)
type RepositoryFactoryBuilder struct {
	config           *RepositoryConfig
	enableCache      bool
	cacheTTL         time.Duration
	enableLogging    bool
	customDecorators []DecoratorBuilder
	userRepoBuilder  func(*database.PostgresDB) repository.UserRepository
	scoreRepoBuilder func(*database.PostgresDB) repository.ScoreRepository
}

// NewRepositoryFactoryBuilder создает новый builder
func NewRepositoryFactoryBuilder(db *database.PostgresDB, redis *redis.Client) *RepositoryFactoryBuilder {
	return &RepositoryFactoryBuilder{
		config: &RepositoryConfig{
			DB:    db,
			Redis: redis,
		},
		enableCache:   true,
		cacheTTL:      5 * time.Minute,
		enableLogging: true,
	}
}

// WithCache включает кэширование
func (b *RepositoryFactoryBuilder) WithCache(ttl time.Duration) *RepositoryFactoryBuilder {
	b.enableCache = true
	b.cacheTTL = ttl
	return b
}

// WithoutCache выключает кэширование
func (b *RepositoryFactoryBuilder) WithoutCache() *RepositoryFactoryBuilder {
	b.enableCache = false
	return b
}

// WithLogging включает логирование
func (b *RepositoryFactoryBuilder) WithLogging() *RepositoryFactoryBuilder {
	b.enableLogging = true
	return b
}

// WithoutLogging выключает логирование
func (b *RepositoryFactoryBuilder) WithoutLogging() *RepositoryFactoryBuilder {
	b.enableLogging = false
	return b
}

// WithCustomDecorator добавляет кастомный декоратор
func (b *RepositoryFactoryBuilder) WithCustomDecorator(decorator DecoratorBuilder) *RepositoryFactoryBuilder {
	b.customDecorators = append(b.customDecorators, decorator)
	return b
}

// WithUserRepositoryBuilder устанавливает кастомный user репозиторий
func (b *RepositoryFactoryBuilder) WithUserRepositoryBuilder(
	builder func(*database.PostgresDB) repository.UserRepository,
) *RepositoryFactoryBuilder {
	b.userRepoBuilder = builder
	return b
}

// WithScoreRepositoryBuilder устанавливает кастомный score репозиторий
func (b *RepositoryFactoryBuilder) WithScoreRepositoryBuilder(
	builder func(*database.PostgresDB) repository.ScoreRepository,
) *RepositoryFactoryBuilder {
	b.scoreRepoBuilder = builder
	return b
}

// Build создает фабрику
func (b *RepositoryFactoryBuilder) Build() RepositoryFactory {
	b.config.EnableCache = b.enableCache
	b.config.CacheTTL = b.cacheTTL
	b.config.EnableLogging = b.enableLogging

	if len(b.customDecorators) > 0 || b.userRepoBuilder != nil || b.scoreRepoBuilder != nil {
		factory := NewCustomRepositoryFactory(b.config)

		if b.userRepoBuilder != nil {
			factory.WithUserRepositoryBuilder(b.userRepoBuilder)
		}

		if b.scoreRepoBuilder != nil {
			factory.WithScoreRepositoryBuilder(b.scoreRepoBuilder)
		}

		for _, decorator := range b.customDecorators {
			factory.WithDecorator(decorator)
		}

		return factory
	}

	return NewRepositoryFactory(b.config)
}
