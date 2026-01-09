package models

import "github.com/google/uuid"

// AuthClaims represents JWT token claims
type AuthClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
}

// LoginRequest is the payload for user login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// LoginResponse contains the JWT token
type LoginResponse struct {
	Token     string    `json:"token"`
	UserID    uuid.UUID `json:"user_id"`
	ExpiresAt int64     `json:"expires_at"`
}

// RegisterRequest is the payload for user registration
type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}
