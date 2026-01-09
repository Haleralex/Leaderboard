package repository

import (
	"context"
	"errors"
	"fmt"

	"leaderboard-service/internal/auth/models"
	"leaderboard-service/internal/shared/database"
	"leaderboard-service/internal/shared/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PostgresUserRepository is a PostgreSQL implementation of UserRepository
type PostgresUserRepository struct {
	db *database.PostgresDB
}

// NewPostgresUserRepository creates a new PostgreSQL user repository
func NewPostgresUserRepository(db *database.PostgresDB) repository.UserRepository {
	return &PostgresUserRepository{
		db: db,
	}
}

// Create creates a new user in the database
func (r *PostgresUserRepository) Create(ctx context.Context, user *models.User) error {
	if err := r.db.DB.WithContext(ctx).Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// FindByID retrieves a user by their UUID
func (r *PostgresUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.DB.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	return &user, nil
}

// FindByEmail retrieves a user by their email address
func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.DB.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}
	return &user, nil
}

// Update updates an existing user's information
func (r *PostgresUserRepository) Update(ctx context.Context, user *models.User) error {
	result := r.db.DB.WithContext(ctx).Save(user)
	if result.Error != nil {
		return fmt.Errorf("failed to update user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

// Delete removes a user from the database
func (r *PostgresUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.DB.WithContext(ctx).Delete(&models.User{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

// FindBySpec finds users matching a specification
func (r *PostgresUserRepository) FindBySpec(ctx context.Context, spec repository.Specification[models.User]) ([]*models.User, error) {
	var users []*models.User

	query := r.db.DB.WithContext(ctx)
	query = spec.Apply(query)

	err := query.Find(&users).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find users by spec: %w", err)
	}

	return users, nil
}

// FindOneBySpec finds first user matching a specification
func (r *PostgresUserRepository) FindOneBySpec(ctx context.Context, spec repository.Specification[models.User]) (*models.User, error) {
	var user models.User

	query := r.db.DB.WithContext(ctx)
	query = spec.Apply(query)

	err := query.First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to find user by spec: %w", err)
	}

	return &user, nil
}

// CountBySpec counts users matching a specification
func (r *PostgresUserRepository) CountBySpec(ctx context.Context, spec repository.Specification[models.User]) (int64, error) {
	var count int64

	query := r.db.DB.WithContext(ctx).Model(&models.User{})
	query = spec.Apply(query)

	err := query.Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count users by spec: %w", err)
	}

	return count, nil
}
