package service

import (
	"context"
	"errors"
	"fmt"

	"leaderboard-service/internal/auth/models"
	"leaderboard-service/internal/shared/config"
	"leaderboard-service/internal/shared/middleware"
	"leaderboard-service/internal/shared/repository"

	"golang.org/x/crypto/bcrypt"
)

// AuthService handles authentication operations
type AuthService struct {
	userRepo repository.UserRepository
	jwt      *middleware.JWTMiddleware
	cfg      *config.Config
}

// NewAuthService creates a new auth service
func NewAuthService(userRepo repository.UserRepository, jwt *middleware.JWTMiddleware, cfg *config.Config) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		jwt:      jwt,
		cfg:      cfg,
	}
}

// Register creates a new user
func (s *AuthService) Register(ctx context.Context, req *models.RegisterRequest) (*models.User, error) {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	if err := s.userRepo.Create(ctx, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// Login authenticates a user and returns a JWT token
func (s *AuthService) Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
	// Fetch user by email using repository
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		// Check if error message contains "not found"
		if errors.Is(err, fmt.Errorf("user not found")) || err.Error() == "user not found" {
			return nil, fmt.Errorf("invalid credentials")
		}
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Generate JWT token
	token, expiresAt, err := s.jwt.GenerateToken(user.ID, user.Email, "user", s.cfg.GetJWTExpiry())
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &models.LoginResponse{
		Token:     token,
		UserID:    user.ID,
		ExpiresAt: expiresAt,
	}, nil
}
