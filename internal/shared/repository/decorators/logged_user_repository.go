package decorators

import (
	"context"
	"time"

	authmodels "leaderboard-service/internal/auth/models"
	"leaderboard-service/internal/shared/repository"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// LoggedUserRepository decorates UserRepository with logging
type LoggedUserRepository struct {
	inner repository.UserRepository
}

// NewLoggedUserRepository creates a logged user repository
func NewLoggedUserRepository(inner repository.UserRepository) repository.UserRepository {
	return &LoggedUserRepository{
		inner: inner,
	}
}

// Create creates a user with logging
func (r *LoggedUserRepository) Create(ctx context.Context, user *authmodels.User) error {
	start := time.Now()
	err := r.inner.Create(ctx, user)
	duration := time.Since(start)

	logEvent := log.Info()
	if err != nil {
		logEvent = log.Error().Err(err)
	}

	logEvent.
		Str("method", "UserRepository.Create").
		Str("user_id", user.ID.String()).
		Str("email", user.Email).
		Dur("duration", duration).
		Msg("User creation")

	return err
}

// FindByID retrieves a user by ID with logging
func (r *LoggedUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*authmodels.User, error) {
	start := time.Now()
	user, err := r.inner.FindByID(ctx, id)
	duration := time.Since(start)

	logEvent := log.Debug()
	if err != nil {
		logEvent = log.Warn().Err(err)
	}

	logEvent.
		Str("method", "UserRepository.FindByID").
		Str("user_id", id.String()).
		Dur("duration", duration).
		Bool("found", err == nil).
		Msg("User lookup by ID")

	return user, err
}

// FindByEmail retrieves a user by email with logging
func (r *LoggedUserRepository) FindByEmail(ctx context.Context, email string) (*authmodels.User, error) {
	start := time.Now()
	user, err := r.inner.FindByEmail(ctx, email)
	duration := time.Since(start)

	logEvent := log.Debug()
	if err != nil {
		logEvent = log.Warn().Err(err)
	}

	logEvent.
		Str("method", "UserRepository.FindByEmail").
		Str("email", email).
		Dur("duration", duration).
		Bool("found", err == nil).
		Msg("User lookup by email")

	return user, err
}

// Update updates a user with logging
func (r *LoggedUserRepository) Update(ctx context.Context, user *authmodels.User) error {
	start := time.Now()
	err := r.inner.Update(ctx, user)
	duration := time.Since(start)

	logEvent := log.Info()
	if err != nil {
		logEvent = log.Error().Err(err)
	}

	logEvent.
		Str("method", "UserRepository.Update").
		Str("user_id", user.ID.String()).
		Dur("duration", duration).
		Msg("User update")

	return err
}

// Delete deletes a user with logging
func (r *LoggedUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	start := time.Now()
	err := r.inner.Delete(ctx, id)
	duration := time.Since(start)

	logEvent := log.Info()
	if err != nil {
		logEvent = log.Error().Err(err)
	}

	logEvent.
		Str("method", "UserRepository.Delete").
		Str("user_id", id.String()).
		Dur("duration", duration).
		Msg("User deletion")

	return err
}

// FindBySpec finds users by specification with logging
func (r *LoggedUserRepository) FindBySpec(ctx context.Context, spec repository.Specification[authmodels.User]) ([]*authmodels.User, error) {
	start := time.Now()
	users, err := r.inner.FindBySpec(ctx, spec)
	duration := time.Since(start)

	logEvent := log.Debug()
	if err != nil {
		logEvent = log.Warn().Err(err)
	}

	logEvent.
		Str("method", "UserRepository.FindBySpec").
		Dur("duration", duration).
		Int("count", len(users)).
		Msg("User query by specification")

	return users, err
}

// FindOneBySpec finds one user by specification with logging
func (r *LoggedUserRepository) FindOneBySpec(ctx context.Context, spec repository.Specification[authmodels.User]) (*authmodels.User, error) {
	start := time.Now()
	user, err := r.inner.FindOneBySpec(ctx, spec)
	duration := time.Since(start)

	logEvent := log.Debug()
	if err != nil {
		logEvent = log.Warn().Err(err)
	}

	logEvent.
		Str("method", "UserRepository.FindOneBySpec").
		Dur("duration", duration).
		Bool("found", err == nil).
		Msg("User query by specification")

	return user, err
}

// CountBySpec counts users by specification with logging
func (r *LoggedUserRepository) CountBySpec(ctx context.Context, spec repository.Specification[authmodels.User]) (int64, error) {
	start := time.Now()
	count, err := r.inner.CountBySpec(ctx, spec)
	duration := time.Since(start)

	logEvent := log.Debug()
	if err != nil {
		logEvent = log.Warn().Err(err)
	}

	logEvent.
		Str("method", "UserRepository.CountBySpec").
		Dur("duration", duration).
		Int64("count", count).
		Msg("User count by specification")

	return count, err
}
