package service

import (
	"context"
	"errors"
	"testing"

	"leaderboard-service/internal/auth/models"
	"leaderboard-service/internal/shared/config"
	"leaderboard-service/internal/shared/middleware"
	"leaderboard-service/internal/shared/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// Mock UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) FindAll(ctx context.Context, limit, offset int) ([]*models.User, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) FindBySpec(ctx context.Context, spec repository.Specification[models.User]) ([]*models.User, error) {
	args := m.Called(ctx, spec)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepository) FindOneBySpec(ctx context.Context, spec repository.Specification[models.User]) (*models.User, error) {
	args := m.Called(ctx, spec)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) CountBySpec(ctx context.Context, spec repository.Specification[models.User]) (int64, error) {
	args := m.Called(ctx, spec)
	return args.Get(0).(int64), args.Error(1)
}

func TestAuthService_Register_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:      "test-secret",
			ExpiryHours: 24,
		},
	}
	jwtMiddleware := middleware.NewJWTMiddleware(cfg)
	service := NewAuthService(mockRepo, jwtMiddleware, cfg)

	req := &models.RegisterRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	}

	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(u *models.User) bool {
		return u.Name == req.Name && u.Email == req.Email
	})).Return(nil)

	user, err := service.Register(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, req.Name, user.Name)
	assert.Equal(t, req.Email, user.Email)
	assert.NotEmpty(t, user.Password)
	assert.NotEqual(t, req.Password, user.Password) // Should be hashed
	mockRepo.AssertExpectations(t)
}

func TestAuthService_Register_RepositoryError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:      "test-secret",
			ExpiryHours: 24,
		},
	}
	jwtMiddleware := middleware.NewJWTMiddleware(cfg)
	service := NewAuthService(mockRepo, jwtMiddleware, cfg)

	req := &models.RegisterRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	}

	mockRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("database error"))

	user, err := service.Register(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_Login_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:      "test-secret",
			ExpiryHours: 24,
		},
	}
	jwtMiddleware := middleware.NewJWTMiddleware(cfg)
	service := NewAuthService(mockRepo, jwtMiddleware, cfg)

	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	existingUser := &models.User{
		ID:       uuid.New(),
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: string(hashedPassword),
	}

	req := &models.LoginRequest{
		Email:    "john@example.com",
		Password: password,
	}

	mockRepo.On("FindByEmail", mock.Anything, req.Email).Return(existingUser, nil)

	response, err := service.Login(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.Token)
	assert.Equal(t, existingUser.ID, response.UserID)
	assert.Greater(t, response.ExpiresAt, int64(0))
	mockRepo.AssertExpectations(t)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:      "test-secret",
			ExpiryHours: 24,
		},
	}
	jwtMiddleware := middleware.NewJWTMiddleware(cfg)
	service := NewAuthService(mockRepo, jwtMiddleware, cfg)

	req := &models.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}

	mockRepo.On("FindByEmail", mock.Anything, req.Email).Return(nil, errors.New("user not found"))

	response, err := service.Login(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "invalid credentials")
	mockRepo.AssertExpectations(t)
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	mockRepo := new(MockUserRepository)
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:      "test-secret",
			ExpiryHours: 24,
		},
	}
	jwtMiddleware := middleware.NewJWTMiddleware(cfg)
	service := NewAuthService(mockRepo, jwtMiddleware, cfg)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)

	existingUser := &models.User{
		ID:       uuid.New(),
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: string(hashedPassword),
	}

	req := &models.LoginRequest{
		Email:    "john@example.com",
		Password: "wrongpassword",
	}

	mockRepo.On("FindByEmail", mock.Anything, req.Email).Return(existingUser, nil)

	response, err := service.Login(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "invalid credentials")
	mockRepo.AssertExpectations(t)
}
