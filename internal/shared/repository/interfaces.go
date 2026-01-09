package repository

import (
	"context"

	authmodels "leaderboard-service/internal/auth/models"
	leaderboardmodels "leaderboard-service/internal/leaderboard/models"

	"github.com/google/uuid"
)

// UserRepository defines the interface for user data access operations
// This abstraction allows us to decouple business logic from data persistence details
type UserRepository interface {
	// Create creates a new user in the database
	Create(ctx context.Context, user *authmodels.User) error

	// FindByID retrieves a user by their UUID
	FindByID(ctx context.Context, id uuid.UUID) (*authmodels.User, error)

	// FindByEmail retrieves a user by their email address
	FindByEmail(ctx context.Context, email string) (*authmodels.User, error)

	// Update updates an existing user's information
	Update(ctx context.Context, user *authmodels.User) error

	// Delete removes a user from the database
	Delete(ctx context.Context, id uuid.UUID) error

	// FindBySpec finds users matching a specification
	FindBySpec(ctx context.Context, spec Specification[authmodels.User]) ([]*authmodels.User, error)

	// FindOneBySpec finds first user matching a specification
	FindOneBySpec(ctx context.Context, spec Specification[authmodels.User]) (*authmodels.User, error)

	// CountBySpec counts users matching a specification
	CountBySpec(ctx context.Context, spec Specification[authmodels.User]) (int64, error)
}

// ScoreRepository defines the interface for score data access operations
// This abstraction allows us to switch database implementations without changing business logic
type ScoreRepository interface {
	// Upsert inserts a new score or updates if the user already has a score for the season
	Upsert(ctx context.Context, score *leaderboardmodels.Score) error

	// FindByUserAndSeason retrieves a user's score for a specific season
	FindByUserAndSeason(ctx context.Context, userID uuid.UUID, season string) (*leaderboardmodels.Score, error)

	// GetLeaderboard retrieves paginated leaderboard entries for a season with user details
	// Returns entries and total count for pagination
	GetLeaderboard(ctx context.Context, season string, limit, offset int, sortOrder string) ([]leaderboardmodels.LeaderboardEntry, int64, error)

	// CountBySeason returns the total number of scores for a given season
	CountBySeason(ctx context.Context, season string) (int64, error)

	// DeleteByUserAndSeason removes a user's score for a specific season
	DeleteByUserAndSeason(ctx context.Context, userID uuid.UUID, season string) error

	// FindBySpec finds scores matching a specification
	FindBySpec(ctx context.Context, spec Specification[leaderboardmodels.Score]) ([]*leaderboardmodels.Score, error)

	// FindOneBySpec finds first score matching a specification
	FindOneBySpec(ctx context.Context, spec Specification[leaderboardmodels.Score]) (*leaderboardmodels.Score, error)

	// CountBySpec counts scores matching a specification
	CountBySpec(ctx context.Context, spec Specification[leaderboardmodels.Score]) (int64, error)
}
