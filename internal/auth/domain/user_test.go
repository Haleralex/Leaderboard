package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewUser(t *testing.T) {
	name := "John Doe"
	email := "john@example.com"
	password := "hashedPassword123"

	user := NewUser(name, email, password)

	assert.NotEqual(t, uuid.Nil, user.ID, "User ID should be generated")
	assert.Equal(t, name, user.Name)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, password, user.Password)
	assert.False(t, user.CreatedAt.IsZero(), "CreatedAt should be set")
	assert.False(t, user.UpdatedAt.IsZero(), "UpdatedAt should be set")
	assert.True(t, user.CreatedAt.Equal(user.UpdatedAt), "CreatedAt and UpdatedAt should be equal for new user")
}

func TestUser_UpdateName(t *testing.T) {
	user := NewUser("John Doe", "john@example.com", "hashedPassword123")
	originalUpdateTime := user.UpdatedAt

	time.Sleep(10 * time.Millisecond) // Small delay to ensure time difference

	newName := "Jane Doe"
	user.UpdateName(newName)

	assert.Equal(t, newName, user.Name, "Name should be updated")
	assert.True(t, user.UpdatedAt.After(originalUpdateTime), "UpdatedAt should be updated")
}

func TestUser_UpdateEmail(t *testing.T) {
	user := NewUser("John Doe", "john@example.com", "hashedPassword123")
	originalUpdateTime := user.UpdatedAt

	time.Sleep(10 * time.Millisecond)

	newEmail := "jane@example.com"
	user.UpdateEmail(newEmail)

	assert.Equal(t, newEmail, user.Email, "Email should be updated")
	assert.True(t, user.UpdatedAt.After(originalUpdateTime), "UpdatedAt should be updated")
}

func TestUser_ChangePassword(t *testing.T) {
	user := NewUser("John Doe", "john@example.com", "hashedPassword123")
	originalUpdateTime := user.UpdatedAt

	time.Sleep(10 * time.Millisecond)

	newPassword := "newHashedPassword456"
	user.ChangePassword(newPassword)

	assert.Equal(t, newPassword, user.Password, "Password should be updated")
	assert.True(t, user.UpdatedAt.After(originalUpdateTime), "UpdatedAt should be updated")
}

func TestUser_MultiplUpdates(t *testing.T) {
	user := NewUser("John Doe", "john@example.com", "hashedPassword123")

	time.Sleep(10 * time.Millisecond)
	user.UpdateName("Jane Doe")
	firstUpdate := user.UpdatedAt

	time.Sleep(10 * time.Millisecond)
	user.UpdateEmail("jane@example.com")
	secondUpdate := user.UpdatedAt

	time.Sleep(10 * time.Millisecond)
	user.ChangePassword("newHashedPassword")
	thirdUpdate := user.UpdatedAt

	assert.True(t, secondUpdate.After(firstUpdate), "Second update should be after first")
	assert.True(t, thirdUpdate.After(secondUpdate), "Third update should be after second")
}
