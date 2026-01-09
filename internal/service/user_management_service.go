package service

import (
	"context"
	"fmt"

	authmodels "leaderboard-service/internal/auth/models"
	leaderboardmodels "leaderboard-service/internal/leaderboard/models"
	"leaderboard-service/internal/shared/repository"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UserManagementService demonstrates Unit of Work usage
type UserManagementService struct {
	uow repository.UnitOfWork
}

// NewUserManagementService creates a new user management service
func NewUserManagementService(uow repository.UnitOfWork) *UserManagementService {
	return &UserManagementService{
		uow: uow,
	}
}

// RegisterUserWithInitialScore creates a user and gives them an initial score
// This operation must be atomic - both or neither should succeed
func (s *UserManagementService) RegisterUserWithInitialScore(
	ctx context.Context,
	name, email, password string,
	initialScore int64,
	season string,
) (*authmodels.User, *leaderboardmodels.Score, error) {
	var user *authmodels.User
	var score *leaderboardmodels.Score

	// Execute in transaction
	err := s.uow.Do(ctx, func(uow repository.UnitOfWork) error {
		// Get repositories from UoW (these are in transaction context)
		userRepo := uow.GetUserRepository()
		scoreRepo := uow.GetScoreRepository()

		// 1. Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		// 2. Create user
		user = &authmodels.User{
			Name:     name,
			Email:    email,
			Password: string(hashedPassword),
		}
		if err := userRepo.Create(ctx, user); err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		// 3. Create initial score
		if season == "" {
			season = "global"
		}
		score = &leaderboardmodels.Score{
			UserID: user.ID,
			Score:  initialScore,
			Season: season,
		}
		if err := scoreRepo.Upsert(ctx, score); err != nil {
			return fmt.Errorf("failed to create initial score: %w", err)
		}

		// Both operations succeeded - will auto-commit
		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	return user, score, nil
}

// TransferScore transfers score from one user to another
// This must be atomic - both updates or neither
func (s *UserManagementService) TransferScore(
	ctx context.Context,
	fromUserID, toUserID uuid.UUID,
	amount int64,
	season string,
) error {
	if season == "" {
		season = "global"
	}

	return s.uow.Do(ctx, func(uow repository.UnitOfWork) error {
		scoreRepo := uow.GetScoreRepository()

		// 1. Get source user's score
		fromScore, err := scoreRepo.FindByUserAndSeason(ctx, fromUserID, season)
		if err != nil {
			return fmt.Errorf("failed to find source score: %w", err)
		}

		// 2. Check if enough score
		if fromScore.Score < amount {
			return fmt.Errorf("insufficient score: has %d, needs %d", fromScore.Score, amount)
		}

		// 3. Deduct from source
		fromScore.Score -= amount
		if err := scoreRepo.Upsert(ctx, fromScore); err != nil {
			return fmt.Errorf("failed to update source score: %w", err)
		}

		// 4. Get or create destination score
		toScore, err := scoreRepo.FindByUserAndSeason(ctx, toUserID, season)
		if err != nil {
			// Create new score if not found
			toScore = &leaderboardmodels.Score{
				UserID: toUserID,
				Score:  0,
				Season: season,
			}
		}

		// 5. Add to destination
		toScore.Score += amount
		if err := scoreRepo.Upsert(ctx, toScore); err != nil {
			return fmt.Errorf("failed to update destination score: %w", err)
		}

		return nil
	})
}

// DeleteUserWithScores deletes a user and all their scores atomically
func (s *UserManagementService) DeleteUserWithScores(
	ctx context.Context,
	userID uuid.UUID,
) error {
	return s.uow.Do(ctx, func(uow repository.UnitOfWork) error {
		userRepo := uow.GetUserRepository()
		scoreRepo := uow.GetScoreRepository()

		// 1. Delete all scores for this user
		// Note: This would need a DeleteByUserID method in ScoreRepository
		// For now, we can at least delete the user

		// 2. Delete user
		if err := userRepo.Delete(ctx, userID); err != nil {
			return fmt.Errorf("failed to delete user: %w", err)
		}

		// Note: In production, you'd want CASCADE DELETE in DB schema
		// or implement DeleteByUserID in ScoreRepository
		_ = scoreRepo // TODO: add DeleteByUserID method

		return nil
	})
}

// BatchUpdateScores updates multiple user scores atomically
func (s *UserManagementService) BatchUpdateScores(
	ctx context.Context,
	updates []struct {
		UserID uuid.UUID
		Score  int64
		Season string
	},
) error {
	return s.uow.Do(ctx, func(uow repository.UnitOfWork) error {
		scoreRepo := uow.GetScoreRepository()

		for _, update := range updates {
			season := update.Season
			if season == "" {
				season = "global"
			}

			score := &leaderboardmodels.Score{
				UserID: update.UserID,
				Score:  update.Score,
				Season: season,
			}

			if err := scoreRepo.Upsert(ctx, score); err != nil {
				return fmt.Errorf("failed to update score for user %s: %w", update.UserID, err)
			}
		}

		return nil
	})
}
