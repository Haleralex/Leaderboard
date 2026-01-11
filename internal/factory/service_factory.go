package factory

import (
	authservice "leaderboard-service/internal/auth/service"
	leaderboardservice "leaderboard-service/internal/leaderboard/service"
	userservice "leaderboard-service/internal/service"
	"leaderboard-service/internal/shared/config"
	"leaderboard-service/internal/shared/database"
	"leaderboard-service/internal/shared/middleware"
	"leaderboard-service/internal/shared/repository"
)

// DefaultServiceFactory стандартная фабрика сервисов
type DefaultServiceFactory struct {
	repoFactory RepositoryFactory
	config      *config.Config
	redis       *database.RedisClient
	jwt         *middleware.JWTMiddleware
}

// NewServiceFactory создает новую фабрику сервисов
func NewServiceFactory(
	repoFactory RepositoryFactory,
	config *config.Config,
	redis *database.RedisClient,
) ServiceFactory {
	return &DefaultServiceFactory{
		repoFactory: repoFactory,
		config:      config,
		redis:       redis,
		jwt:         middleware.NewJWTMiddleware(config),
	}
}

// CreateAuthService создает сервис аутентификации
func (f *DefaultServiceFactory) CreateAuthService() interface{} {
	userRepo := f.repoFactory.CreateUserRepository()
	return authservice.NewAuthService(userRepo, f.jwt, f.config)
}

// CreateLeaderboardService создает сервис лидерборда
func (f *DefaultServiceFactory) CreateLeaderboardService() interface{} {
	scoreRepo := f.repoFactory.CreateScoreRepository()
	userRepo := f.repoFactory.CreateUserRepository()
	return leaderboardservice.NewLeaderboardService(scoreRepo, userRepo, f.redis, f.config)
}

// CreateUserManagementService создает сервис управления пользователями
func (f *DefaultServiceFactory) CreateUserManagementService() interface{} {
	uow := f.repoFactory.CreateUnitOfWork()
	return userservice.NewUserManagementService(uow)
}

// CreateQueryService создает сервис для запросов
func (f *DefaultServiceFactory) CreateQueryService() interface{} {
	userRepo := f.repoFactory.CreateUserRepository()
	scoreRepo := f.repoFactory.CreateScoreRepository()
	return leaderboardservice.NewQueryService(userRepo, scoreRepo)
}

// TypedServiceFactory типизированная фабрика сервисов (без interface{})
type TypedServiceFactory struct {
	repoFactory RepositoryFactory
	config      *config.Config
	redis       *database.RedisClient
	jwt         *middleware.JWTMiddleware
}

// NewTypedServiceFactory создает типизированную фабрику
func NewTypedServiceFactory(
	repoFactory RepositoryFactory,
	config *config.Config,
	redis *database.RedisClient,
) *TypedServiceFactory {
	return &TypedServiceFactory{
		repoFactory: repoFactory,
		config:      config,
		redis:       redis,
		jwt:         middleware.NewJWTMiddleware(config),
	}
}

// CreateAuthService создает AuthService
func (f *TypedServiceFactory) CreateAuthService() *authservice.AuthService {
	userRepo := f.repoFactory.CreateUserRepository()
	return authservice.NewAuthService(userRepo, f.jwt, f.config)
}

// CreateLeaderboardService создает LeaderboardService
func (f *TypedServiceFactory) CreateLeaderboardService() *leaderboardservice.LeaderboardService {
	scoreRepo := f.repoFactory.CreateScoreRepository()
	userRepo := f.repoFactory.CreateUserRepository()
	return leaderboardservice.NewLeaderboardService(scoreRepo, userRepo, f.redis, f.config)
}

// CreateUserManagementService создает UserManagementService
func (f *TypedServiceFactory) CreateUserManagementService() *userservice.UserManagementService {
	uow := f.repoFactory.CreateUnitOfWork()
	return userservice.NewUserManagementService(uow)
}

// CreateQueryService создает QueryService
func (f *TypedServiceFactory) CreateQueryService() *leaderboardservice.QueryService {
	userRepo := f.repoFactory.CreateUserRepository()
	scoreRepo := f.repoFactory.CreateScoreRepository()
	return leaderboardservice.NewQueryService(userRepo, scoreRepo)
}

// ServiceFactoryBuilder builder для создания фабрики сервисов
type ServiceFactoryBuilder struct {
	repoFactory RepositoryFactory
	config      *config.Config
	redis       *database.RedisClient
	jwt         *middleware.JWTMiddleware
	customAuth  func(repository.UserRepository, *middleware.JWTMiddleware, *config.Config) *authservice.AuthService
}

// NewServiceFactoryBuilder создает builder
func NewServiceFactoryBuilder() *ServiceFactoryBuilder {
	return &ServiceFactoryBuilder{}
}

// WithRepositoryFactory устанавливает фабрику репозиториев
func (b *ServiceFactoryBuilder) WithRepositoryFactory(factory RepositoryFactory) *ServiceFactoryBuilder {
	b.repoFactory = factory
	return b
}

// WithConfig устанавливает конфигурацию
func (b *ServiceFactoryBuilder) WithConfig(config *config.Config) *ServiceFactoryBuilder {
	b.config = config
	return b
}

// WithRedis устанавливает Redis клиент
func (b *ServiceFactoryBuilder) WithRedis(redis *database.RedisClient) *ServiceFactoryBuilder {
	b.redis = redis
	return b
}

// WithJWT устанавливает JWT middleware
func (b *ServiceFactoryBuilder) WithJWT(jwt *middleware.JWTMiddleware) *ServiceFactoryBuilder {
	b.jwt = jwt
	return b
}

// WithCustomAuthService устанавливает кастомную фабрику для AuthService
func (b *ServiceFactoryBuilder) WithCustomAuthService(
	factory func(repository.UserRepository, *middleware.JWTMiddleware, *config.Config) *authservice.AuthService,
) *ServiceFactoryBuilder {
	b.customAuth = factory
	return b
}

// Build создает фабрику сервисов
func (b *ServiceFactoryBuilder) Build() *TypedServiceFactory {
	if b.jwt == nil && b.config != nil {
		b.jwt = middleware.NewJWTMiddleware(b.config)
	}

	return &TypedServiceFactory{
		repoFactory: b.repoFactory,
		config:      b.config,
		redis:       b.redis,
		jwt:         b.jwt,
	}
}

// AllServicesFactory комбинированная фабрика для создания всех сервисов сразу
type AllServicesFactory struct {
	Auth           *authservice.AuthService
	Leaderboard    *leaderboardservice.LeaderboardService
	UserManagement *userservice.UserManagementService
	Query          *leaderboardservice.QueryService
}

// CreateAllServices создает все сервисы сразу
func CreateAllServices(
	repoFactory RepositoryFactory,
	config *config.Config,
	redis *database.RedisClient,
) *AllServicesFactory {
	factory := NewTypedServiceFactory(repoFactory, config, redis)

	return &AllServicesFactory{
		Auth:           factory.CreateAuthService(),
		Leaderboard:    factory.CreateLeaderboardService(),
		UserManagement: factory.CreateUserManagementService(),
		Query:          factory.CreateQueryService(),
	}
}
