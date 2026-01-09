package domain

import (
	"time"

	"github.com/google/uuid"
)

// User - чистая domain-модель пользователя без инфраструктурных зависимостей
// Представляет игрока в системе лидерборда
type User struct {
	ID        uuid.UUID
	Name      string
	Email     string
	Password  string // Хэшированный пароль
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewUser создает нового пользователя
func NewUser(name, email, hashedPassword string) *User {
	now := time.Now()
	return &User{
		ID:        uuid.New(),
		Name:      name,
		Email:     email,
		Password:  hashedPassword,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// UpdateName обновляет имя пользователя
func (u *User) UpdateName(name string) {
	u.Name = name
	u.UpdatedAt = time.Now()
}

// UpdateEmail обновляет email пользователя
func (u *User) UpdateEmail(email string) {
	u.Email = email
	u.UpdatedAt = time.Now()
}

// ChangePassword изменяет пароль пользователя
func (u *User) ChangePassword(hashedPassword string) {
	u.Password = hashedPassword
	u.UpdatedAt = time.Now()
}
