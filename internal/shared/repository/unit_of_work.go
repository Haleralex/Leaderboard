package repository

import (
	"context"
	"fmt"

	"leaderboard-service/internal/shared/database"

	"gorm.io/gorm"
)

// UnitOfWork interface for managing transactions
type UnitOfWork interface {
	// Begin starts a new transaction
	Begin(ctx context.Context) error

	// Commit commits the transaction
	Commit() error

	// Rollback rolls back the transaction
	Rollback() error

	// GetUserRepository returns a user repository within the transaction context
	GetUserRepository() UserRepository

	// GetScoreRepository returns a score repository within the transaction context
	GetScoreRepository() ScoreRepository

	// Do executes a function within a transaction, automatically committing or rolling back
	Do(ctx context.Context, fn func(uow UnitOfWork) error) error
}

// GormUnitOfWork implements UnitOfWork using GORM
type GormUnitOfWork struct {
	db            *database.PostgresDB
	tx            *gorm.DB
	userRepo      UserRepository
	scoreRepo     ScoreRepository
	userFactory   func(*database.PostgresDB) UserRepository
	scoreFactory  func(*database.PostgresDB) ScoreRepository
	inTransaction bool
}

// NewUnitOfWork creates a new unit of work
func NewUnitOfWork(
	db *database.PostgresDB,
	userFactory func(*database.PostgresDB) UserRepository,
	scoreFactory func(*database.PostgresDB) ScoreRepository,
) UnitOfWork {
	return &GormUnitOfWork{
		db:           db,
		userFactory:  userFactory,
		scoreFactory: scoreFactory,
	}
}

// Begin starts a new transaction
func (uow *GormUnitOfWork) Begin(ctx context.Context) error {
	if uow.inTransaction {
		return fmt.Errorf("transaction already started")
	}

	tx := uow.db.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	uow.tx = tx
	uow.inTransaction = true

	// Create repositories with transaction context using factories
	txDB := &database.PostgresDB{DB: tx}
	uow.userRepo = uow.userFactory(txDB)
	uow.scoreRepo = uow.scoreFactory(txDB)

	return nil
}

// Commit commits the transaction
func (uow *GormUnitOfWork) Commit() error {
	if !uow.inTransaction {
		return fmt.Errorf("no active transaction")
	}

	err := uow.tx.Commit().Error
	uow.inTransaction = false
	uow.tx = nil
	uow.userRepo = nil
	uow.scoreRepo = nil

	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

// Rollback rolls back the transaction
func (uow *GormUnitOfWork) Rollback() error {
	if !uow.inTransaction {
		return fmt.Errorf("no active transaction")
	}

	err := uow.tx.Rollback().Error
	uow.inTransaction = false
	uow.tx = nil
	uow.userRepo = nil
	uow.scoreRepo = nil

	if err != nil {
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}
	return nil
}

// GetUserRepository returns a user repository within the transaction context
func (uow *GormUnitOfWork) GetUserRepository() UserRepository {
	if !uow.inTransaction {
		panic("no active transaction: call Begin() first")
	}
	return uow.userRepo
}

// GetScoreRepository returns a score repository within the transaction context
func (uow *GormUnitOfWork) GetScoreRepository() ScoreRepository {
	if !uow.inTransaction {
		panic("no active transaction: call Begin() first")
	}
	return uow.scoreRepo
}

// Do executes a function within a transaction
// Automatically commits on success or rolls back on error/panic
func (uow *GormUnitOfWork) Do(ctx context.Context, fn func(uow UnitOfWork) error) (err error) {
	// Start transaction
	if err := uow.Begin(ctx); err != nil {
		return err
	}

	// Ensure rollback on panic
	defer func() {
		if r := recover(); r != nil {
			_ = uow.Rollback()
			panic(r) // re-panic after rollback
		}
	}()

	// Execute function
	err = fn(uow)
	if err != nil {
		// Rollback on error
		if rbErr := uow.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction error: %v, rollback error: %w", err, rbErr)
		}
		return err
	}

	// Commit on success
	if err := uow.Commit(); err != nil {
		return err
	}

	return nil
}
