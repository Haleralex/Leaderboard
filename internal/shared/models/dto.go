package models

import (
	"time"

	"github.com/google/uuid"
)

// UserDTO is a data transfer object for user information
// Used for cross-module communication without exposing internal domain models
type UserDTO struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// ScoreDTO is a data transfer object for score information
// Used for cross-module communication without exposing internal domain models
type ScoreDTO struct {
	ID        uuid.UUID              `json:"id"`
	UserID    uuid.UUID              `json:"user_id"`
	Score     int64                  `json:"score"`
	Season    string                 `json:"season"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// ErrorResponse represents an error API response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// SuccessResponse represents a generic success API response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}
