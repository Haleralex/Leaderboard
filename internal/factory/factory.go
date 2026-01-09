package factory

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	"leaderboard-service/internal/shared/database"
	"leaderboard-service/internal/shared/repository"
)

// RepositoryFactory создает репозитории с автоматической конфигурацией
type RepositoryFactory interface {
	// CreateUserRepository создает репозиторий пользователей
	CreateUserRepository() repository.UserRepository

	// CreateScoreRepository создает репозиторий счетов
	CreateScoreRepository() repository.ScoreRepository

	// CreateUnitOfWork создает Unit of Work с репозиториями
	CreateUnitOfWork() repository.UnitOfWork
}

// ServiceFactory создает сервисы с инжекцией зависимостей
type ServiceFactory interface {
	// CreateAuthService создает сервис аутентификации
	CreateAuthService() interface{}

	// CreateLeaderboardService создает сервис лидерборда
	CreateLeaderboardService() interface{}

	// CreateUserManagementService создает сервис управления пользователями
	CreateUserManagementService() interface{}

	// CreateQueryService создает сервис для запросов
	CreateQueryService() interface{}
}

// RepositoryConfig конфигурация для создания репозиториев
type RepositoryConfig struct {
	// DB подключение к базе данных
	DB *database.PostgresDB

	// Redis подключение к Redis
	Redis *redis.Client

	// EnableCache включить кэширование
	EnableCache bool

	// CacheTTL время жизни кэша
	CacheTTL time.Duration

	// EnableLogging включить логирование
	EnableLogging bool
}

// DefaultRepositoryConfig возвращает конфигурацию по умолчанию
func DefaultRepositoryConfig(db *database.PostgresDB, redis *redis.Client) *RepositoryConfig {
	return &RepositoryConfig{
		DB:            db,
		Redis:         redis,
		EnableCache:   true,
		CacheTTL:      5 * time.Minute,
		EnableLogging: true,
	}
}

// RepositoryConfigOption функция для конфигурирования репозиториев
type RepositoryConfigOption func(*RepositoryConfig)

// WithCache включает/выключает кэширование
func WithCache(enable bool) RepositoryConfigOption {
	return func(cfg *RepositoryConfig) {
		cfg.EnableCache = enable
	}
}

// WithCacheTTL устанавливает время жизни кэша
func WithCacheTTL(ttl time.Duration) RepositoryConfigOption {
	return func(cfg *RepositoryConfig) {
		cfg.CacheTTL = ttl
	}
}

// WithLogging включает/выключает логирование
func WithLogging(enable bool) RepositoryConfigOption {
	return func(cfg *RepositoryConfig) {
		cfg.EnableLogging = enable
	}
}

// NewRepositoryConfig создает конфигурацию с опциями
func NewRepositoryConfig(db *database.PostgresDB, redis *redis.Client, opts ...RepositoryConfigOption) *RepositoryConfig {
	cfg := DefaultRepositoryConfig(db, redis)
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

// Cache интерфейс для кэширования
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context, pattern string) error
}

// RedisCache реализация кэша через Redis
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache создает новый Redis кэш
func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

// Get получает значение из кэша
func (c *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return val, err
}

// Set сохраняет значение в кэш
func (c *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

// Delete удаляет значение из кэша
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// Clear очищает кэш по паттерну
func (c *RedisCache) Clear(ctx context.Context, pattern string) error {
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := c.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}
