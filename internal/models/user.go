package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a game player
type User struct {
	ID        uuid.UUID `json:"id" db:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name      string    `json:"name" db:"name" gorm:"type:varchar(255);not null"`
	Email     string    `json:"email,omitempty" db:"email" gorm:"type:varchar(255);uniqueIndex;not null"`
	Password  string    `json:"-" db:"password_hash" gorm:"column:password_hash;type:varchar(255);not null"`
	CreatedAt time.Time `json:"created_at" db:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at" gorm:"autoUpdateTime"`
}

// TableName specifies the table name for GORM
func (User) TableName() string {
	return "users"
}

// CreateUserRequest is the payload for creating a new user
type CreateUserRequest struct {
	Name  string `json:"name" validate:"required,min=3,max=50"`
	Email string `json:"email" validate:"required,email"`
}
