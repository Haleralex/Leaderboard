package infrastructure

import (
	"time"

	"leaderboard-service/internal/auth/domain"

	"github.com/google/uuid"
)

// UserEntity - persistence модель с GORM тегами
// Отделена от domain для чистоты архитектуры
type UserEntity struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name      string    `gorm:"type:varchar(255);not null"`
	Email     string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	Password  string    `gorm:"column:password_hash;type:varchar(255);not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// TableName для GORM
func (UserEntity) TableName() string {
	return "users"
}

// ToDomain конвертирует entity в domain модель
func (e *UserEntity) ToDomain() *domain.User {
	return &domain.User{
		ID:        e.ID,
		Name:      e.Name,
		Email:     e.Email,
		Password:  e.Password,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}

// FromDomain создает entity из domain модели
func FromDomainUser(u *domain.User) *UserEntity {
	return &UserEntity{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Password:  u.Password,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// ToDomainList конвертирует список entities в domain модели
func ToDomainUserList(entities []*UserEntity) []*domain.User {
	result := make([]*domain.User, len(entities))
	for i, e := range entities {
		result[i] = e.ToDomain()
	}
	return result
}
