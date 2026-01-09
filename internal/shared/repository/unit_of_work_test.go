package repository

import (
	"context"
	"testing"

	leaderboardmodels "leaderboard-service/internal/leaderboard/models"
	authmodels "leaderboard-service/internal/auth/models"
	"leaderboard-service/internal/shared/config"
	"leaderboard-service/internal/shared/database"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestUoW(t *testing.T) (*database.PostgresDB, UnitOfWork) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			URL:      "postgresql://postgres:postgres@localhost:5432/leaderboard?sslmode=disable",
			MaxConns: 25,
			MinConns: 5,
		},
		Server: config.ServerConfig{
			Env: "test",
		},
	}

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	require.NotNil(t, db)

	// Create simple factories for tests
	userFactory := func(db *database.PostgresDB) UserRepository {
		return nil // mock or real implementation
	}
	scoreFactory := func(db *database.PostgresDB) ScoreRepository {
		return nil // mock or real implementation
	}

	uow := NewUnitOfWork(db, userFactory, scoreFactory)
	return db, uow
}

func TestUnitOfWork_BeginCommit(t *testing.T) {
	_, uow := setupTestUoW(t)
	ctx := context.Background()

	// Begin
	err := uow.Begin(ctx)
	assert.NoError(t, err)

	// Get repositories
	userRepo := uow.GetUserRepository()
	assert.NotNil(t, userRepo)

	scoreRepo := uow.GetScoreRepository()
	assert.NotNil(t, scoreRepo)

	// Commit
	err = uow.Commit()
	assert.NoError(t, err)
}

func TestUnitOfWork_BeginRollback(t *testing.T) {
	_, uow := setupTestUoW(t)
	ctx := context.Background()

	// Begin
	err := uow.Begin(ctx)
	assert.NoError(t, err)

	// Rollback
	err = uow.Rollback()
	assert.NoError(t, err)
}

func TestUnitOfWork_DoSuccess(t *testing.T) {
	_, uow := setupTestUoW(t)
	ctx := context.Background()

	var userCreated *authmodels.User
	var scoreCreated *leaderboardmodels.Score

	err := uow.Do(ctx, func(uow UnitOfWork) error {
		userRepo := uow.GetUserRepository()
		scoreRepo := uow.GetScoreRepository()

		// Create user
		user := &authmodels.User{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: "hashed_password",
		}
		err := userRepo.Create(ctx, user)
		if err != nil {
			return err
		}
		userCreated = user

		// Create score
		score := &leaderboardmodels.Score{
			UserID: user.ID,
			Score:  1000,
			Season: "test-season",
		}
		err = scoreRepo.Upsert(ctx, score)
		if err != nil {
			return err
		}
		scoreCreated = score

		return nil
	})

	assert.NoError(t, err)
	assert.NotNil(t, userCreated)
	assert.NotNil(t, scoreCreated)
	assert.NotEqual(t, uuid.Nil, userCreated.ID)

	// Cleanup
	cleanupUser(t, userCreated.ID)
}

func TestUnitOfWork_DoRollback(t *testing.T) {
	db, uow := setupTestUoW(t)
	ctx := context.Background()

	var userID uuid.UUID

	// This transaction should rollback due to error
	err := uow.Do(ctx, func(uow UnitOfWork) error {
		userRepo := uow.GetUserRepository()

		// Create user
		user := &authmodels.User{
			Name:     "Rollback User",
			Email:    "rollback@example.com",
			Password: "hashed_password",
		}
		err := userRepo.Create(ctx, user)
		if err != nil {
			return err
		}
		userID = user.ID

		// Simulate error - this should trigger rollback
		return assert.AnError
	})

	assert.Error(t, err)
	assert.NotEqual(t, uuid.Nil, userID)

	// Verify user was NOT created (rolled back)
	var count int64
	db.DB.Model(&authmodels.User{}).Where("id = ?", userID).Count(&count)
	assert.Equal(t, int64(0), count, "User should not exist after rollback")
}

func TestUnitOfWork_MultipleRepositories(t *testing.T) {
	_, uow := setupTestUoW(t)
	ctx := context.Background()

	var user1, user2 *authmodels.User

	err := uow.Do(ctx, func(uow UnitOfWork) error {
		userRepo := uow.GetUserRepository()
		scoreRepo := uow.GetScoreRepository()

		// Create first user with score
		user1 = &authmodels.User{
			Name:     "User 1",
			Email:    "user1@example.com",
			Password: "password1",
		}
		if err := userRepo.Create(ctx, user1); err != nil {
			return err
		}

		score1 := &leaderboardmodels.Score{
			UserID: user1.ID,
			Score:  500,
			Season: "global",
		}
		if err := scoreRepo.Upsert(ctx, score1); err != nil {
			return err
		}

		// Create second user with score
		user2 = &authmodels.User{
			Name:     "User 2",
			Email:    "user2@example.com",
			Password: "password2",
		}
		if err := userRepo.Create(ctx, user2); err != nil {
			return err
		}

		score2 := &leaderboardmodels.Score{
			UserID: user2.ID,
			Score:  1000,
			Season: "global",
		}
		if err := scoreRepo.Upsert(ctx, score2); err != nil {
			return err
		}

		return nil
	})

	assert.NoError(t, err)
	assert.NotNil(t, user1)
	assert.NotNil(t, user2)

	// Cleanup
	cleanupUser(t, user1.ID)
	cleanupUser(t, user2.ID)
}

func TestUnitOfWork_PanicRecovery(t *testing.T) {
	db, uow := setupTestUoW(t)
	ctx := context.Background()

	var userID uuid.UUID

	// This should panic and rollback
	assert.Panics(t, func() {
		_ = uow.Do(ctx, func(uow UnitOfWork) error {
			userRepo := uow.GetUserRepository()

			// Create user
			user := &authmodels.User{
				Name:     "Panic User",
				Email:    "panic@example.com",
				Password: "password",
			}
			_ = userRepo.Create(ctx, user)
			userID = user.ID

			// Simulate panic
			panic("test panic")
		})
	})

	// Verify user was NOT created (rolled back after panic)
	var count int64
	db.DB.Model(&authmodels.User{}).Where("id = ?", userID).Count(&count)
	assert.Equal(t, int64(0), count, "User should not exist after panic rollback")
}

func cleanupUser(t *testing.T, userID uuid.UUID) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			URL:      "postgresql://postgres:postgres@localhost:5432/leaderboard?sslmode=disable",
			MaxConns: 25,
			MinConns: 5,
		},
		Server: config.ServerConfig{
			Env: "test",
		},
	}

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)

	// Delete scores first (foreign key)
	db.DB.Where("user_id = ?", userID).Delete(&leaderboardmodels.Score{})
	// Delete user
	db.DB.Where("id = ?", userID).Delete(&authmodels.User{})
}
