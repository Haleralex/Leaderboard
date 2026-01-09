package repository

import (
	"context"

	"leaderboard-service/internal/auth/domain"
	"leaderboard-service/internal/auth/infrastructure"
	"leaderboard-service/internal/auth/models"
	"leaderboard-service/internal/shared/database"
	"leaderboard-service/internal/shared/repository"

	"github.com/google/uuid"
)

// PostgresUserRepository is a PostgreSQL implementation of UserRepository
// Использует чистую domain модель и отдельные entities для персистентности
// Реализует Clean Architecture: domain не зависит от инфраструктуры
type PostgresUserRepository struct {
	*repository.BaseRepository[infrastructure.UserEntity]
	db *database.PostgresDB
}

// NewPostgresUserRepository creates a new PostgreSQL user repository
func NewPostgresUserRepository(db *database.PostgresDB) repository.UserRepository {
	return &PostgresUserRepository{
		BaseRepository: repository.NewBaseRepository[infrastructure.UserEntity](db),
		db:             db,
	}
}

// Create creates a new user in the database
func (r *PostgresUserRepository) Create(ctx context.Context, user *models.User) error {
	// Конвертируем domain -> entity для персистентности
	domainUser := &domain.User{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Password:  user.Password,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
	entity := infrastructure.FromDomainUser(domainUser)
	if err := r.BaseRepository.Create(ctx, entity); err != nil {
		return err
	}
	// Обновляем ID если он был сгенерирован БД
	user.ID = entity.ID
	return nil
}

// FindByID retrieves a user by their UUID
func (r *PostgresUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	entity, err := r.BaseRepository.FindOne(ctx, "id = ?", id)
	if err != nil {
		return nil, err
	}
	domainUser := entity.ToDomain()
	return &models.User{
		ID:        domainUser.ID,
		Name:      domainUser.Name,
		Email:     domainUser.Email,
		Password:  domainUser.Password,
		CreatedAt: domainUser.CreatedAt,
		UpdatedAt: domainUser.UpdatedAt,
	}, nil
}

// FindByEmail retrieves a user by their email address
func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	entity, err := r.BaseRepository.FindOne(ctx, "email = ?", email)
	if err != nil {
		return nil, err
	}
	domainUser := entity.ToDomain()
	return &models.User{
		ID:        domainUser.ID,
		Name:      domainUser.Name,
		Email:     domainUser.Email,
		Password:  domainUser.Password,
		CreatedAt: domainUser.CreatedAt,
		UpdatedAt: domainUser.UpdatedAt,
	}, nil
}

// Update updates an existing user's information
func (r *PostgresUserRepository) Update(ctx context.Context, user *models.User) error {
	domainUser := &domain.User{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Password:  user.Password,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
	entity := infrastructure.FromDomainUser(domainUser)
	return r.BaseRepository.Update(ctx, entity)
}

// Delete removes a user from the database
func (r *PostgresUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.BaseRepository.Delete(ctx, "id = ?", id)
}

// FindBySpec finds users matching a specification
func (r *PostgresUserRepository) FindBySpec(ctx context.Context, spec repository.Specification[models.User]) ([]*models.User, error) {
	// Временно используем старый подход до полной миграции спецификаций
	return nil, nil
}

// FindOneBySpec finds first user matching a specification
func (r *PostgresUserRepository) FindOneBySpec(ctx context.Context, spec repository.Specification[models.User]) (*models.User, error) {
	// Временно используем старый подход до полной миграции спецификаций
	return nil, nil
}

// CountBySpec counts users matching a specification
func (r *PostgresUserRepository) CountBySpec(ctx context.Context, spec repository.Specification[models.User]) (int64, error) {
	// Временно используем старый подход до полной миграции спецификаций
	return 0, nil
}
