package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"leaderboard-service/internal/shared/config"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestJWTMiddleware_GenerateAndValidateToken(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:      "test-secret-key",
			ExpiryHours: 24,
		},
	}

	jwtMiddleware := NewJWTMiddleware(cfg)

	userID := uuid.New()
	email := "test@example.com"
	role := "user"

	// Generate token
	token, expiresAt, err := jwtMiddleware.GenerateToken(userID, email, role, 24*time.Hour)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Greater(t, expiresAt, time.Now().Unix())

	// Validate token
	claims, err := jwtMiddleware.validateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, role, claims.Role)
}

func TestJWTMiddleware_Authenticate_Success(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:      "test-secret-key",
			ExpiryHours: 24,
		},
	}

	jwtMiddleware := NewJWTMiddleware(cfg)

	userID := uuid.New()
	token, _, _ := jwtMiddleware.GenerateToken(userID, "test@example.com", "user", 24*time.Hour)

	// Create a test handler that checks if user ID is in context
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxUserID, ok := GetUserIDFromContext(r.Context())
		assert.True(t, ok)
		assert.Equal(t, userID, ctxUserID)
		w.WriteHeader(http.StatusOK)
	})

	handler := jwtMiddleware.Authenticate(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestJWTMiddleware_Authenticate_MissingToken(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:      "test-secret-key",
			ExpiryHours: 24,
		},
	}

	jwtMiddleware := NewJWTMiddleware(cfg)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Handler should not be called")
	})

	handler := jwtMiddleware.Authenticate(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestJWTMiddleware_Authenticate_InvalidToken(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:      "test-secret-key",
			ExpiryHours: 24,
		},
	}

	jwtMiddleware := NewJWTMiddleware(cfg)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Handler should not be called")
	})

	handler := jwtMiddleware.Authenticate(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}
