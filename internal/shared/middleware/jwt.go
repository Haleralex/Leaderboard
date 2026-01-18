package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	authmodels "leaderboard-service/internal/auth/models"
	"leaderboard-service/internal/shared/config"
	sharedmodels "leaderboard-service/internal/shared/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type contextKey string

const (
	UserIDKey contextKey = "user_id"
	EmailKey  contextKey = "email"
	RoleKey   contextKey = "role"
)

// JWTMiddleware validates JWT tokens
type JWTMiddleware struct {
	secret string
}

// NewJWTMiddleware creates a new JWT middleware
func NewJWTMiddleware(cfg *config.Config) *JWTMiddleware {
	return &JWTMiddleware{
		secret: cfg.JWT.Secret,
	}
}

// Authenticate is the middleware handler
func (m *JWTMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondError(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			respondError(w, "invalid authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]
		claims, err := m.validateToken(tokenString)
		if err != nil {
			respondError(w, "invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Add claims to request context
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, EmailKey, claims.Email)
		ctx = context.WithValue(ctx, RoleKey, claims.Role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// validateToken parses and validates a JWT token
func (m *JWTMiddleware) validateToken(tokenString string) (*authmodels.AuthClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(m.secret), nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, jwt.ErrTokenInvalidClaims
	}

	userIDStr, _ := claims["user_id"].(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, err
	}

	return &authmodels.AuthClaims{
		UserID: userID,
		Email:  claims["email"].(string),
		Role:   claims["role"].(string),
	}, nil
}

// ValidateTokenString is a public method to validate JWT token string (for WebSocket)
func (m *JWTMiddleware) ValidateTokenString(tokenString string) (*authmodels.AuthClaims, error) {
	return m.validateToken(tokenString)
}

// GenerateToken creates a new JWT token
func (m *JWTMiddleware) GenerateToken(userID uuid.UUID, email, role string, expiry time.Duration) (string, int64, error) {
	expiresAt := time.Now().Add(expiry)

	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"email":   email,
		"role":    role,
		"exp":     expiresAt.Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(m.secret))
	if err != nil {
		return "", 0, err
	}

	return tokenString, expiresAt.Unix(), nil
}

// GetUserIDFromContext extracts user ID from request context
func GetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return userID, ok
}

// respondError sends a JSON error response
func respondError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(sharedmodels.ErrorResponse{
		Error:   http.StatusText(code),
		Message: message,
		Code:    code,
	})
}
