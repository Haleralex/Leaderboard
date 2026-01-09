package repository

import (
	"context"
	"errors"
	"fmt"

	"leaderboard-service/internal/shared/database"

	"gorm.io/gorm"
)

// BaseRepository - переиспользуемый базовый репозиторий с общими методами
// Реализует общие паттерны работы с БД для всех доменных репозиториев
type BaseRepository[T any] struct {
	db *database.PostgresDB
}

// NewBaseRepository создает новый базовый репозиторий
func NewBaseRepository[T any](db *database.PostgresDB) *BaseRepository[T] {
	return &BaseRepository[T]{
		db: db,
	}
}

// GetDB возвращает подключение к БД
func (r *BaseRepository[T]) GetDB() *gorm.DB {
	return r.db.DB
}

// FindBySpec находит записи по спецификации - переиспользуемый метод
func (r *BaseRepository[T]) FindBySpec(ctx context.Context, spec Specification[T]) ([]*T, error) {
	var results []*T

	query := r.db.DB.WithContext(ctx)
	query = spec.Apply(query)

	err := query.Find(&results).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find by spec: %w", err)
	}

	return results, nil
}

// FindOneBySpec находит первую запись по спецификации - переиспользуемый метод
func (r *BaseRepository[T]) FindOneBySpec(ctx context.Context, spec Specification[T]) (*T, error) {
	var result T

	query := r.db.DB.WithContext(ctx)
	query = spec.Apply(query)

	err := query.First(&result).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("record not found")
		}
		return nil, fmt.Errorf("failed to find one by spec: %w", err)
	}

	return &result, nil
}

// CountBySpec подсчитывает записи по спецификации - переиспользуемый метод
func (r *BaseRepository[T]) CountBySpec(ctx context.Context, spec Specification[T]) (int64, error) {
	var count int64
	var model T

	query := r.db.DB.WithContext(ctx).Model(&model)
	query = spec.Apply(query)

	err := query.Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count by spec: %w", err)
	}

	return count, nil
}

// Create создает новую запись - переиспользуемый метод
func (r *BaseRepository[T]) Create(ctx context.Context, entity *T) error {
	if err := r.db.DB.WithContext(ctx).Create(entity).Error; err != nil {
		return fmt.Errorf("failed to create: %w", err)
	}
	return nil
}

// Update обновляет существующую запись - переиспользуемый метод
func (r *BaseRepository[T]) Update(ctx context.Context, entity *T) error {
	result := r.db.DB.WithContext(ctx).Save(entity)
	if result.Error != nil {
		return fmt.Errorf("failed to update: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("record not found")
	}
	return nil
}

// Delete удаляет запись по условию - переиспользуемый метод
func (r *BaseRepository[T]) Delete(ctx context.Context, condition string, args ...interface{}) error {
	var model T
	result := r.db.DB.WithContext(ctx).Where(condition, args...).Delete(&model)
	if result.Error != nil {
		return fmt.Errorf("failed to delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("record not found")
	}
	return nil
}

// FindOne находит одну запись по условию - переиспользуемый метод
func (r *BaseRepository[T]) FindOne(ctx context.Context, condition string, args ...interface{}) (*T, error) {
	var result T
	err := r.db.DB.WithContext(ctx).Where(condition, args...).First(&result).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("record not found")
		}
		return nil, fmt.Errorf("failed to find one: %w", err)
	}
	return &result, nil
}

// FindAll находит все записи по условию - переиспользуемый метод
func (r *BaseRepository[T]) FindAll(ctx context.Context, condition string, args ...interface{}) ([]*T, error) {
	var results []*T
	err := r.db.DB.WithContext(ctx).Where(condition, args...).Find(&results).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find all: %w", err)
	}
	return results, nil
}

// Count подсчитывает записи по условию - переиспользуемый метод
func (r *BaseRepository[T]) Count(ctx context.Context, condition string, args ...interface{}) (int64, error) {
	var count int64
	var model T
	err := r.db.DB.WithContext(ctx).Model(&model).Where(condition, args...).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count: %w", err)
	}
	return count, nil
}
