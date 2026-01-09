package domain

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent - базовый интерфейс для всех domain events
type DomainEvent interface {
	OccurredAt() time.Time
	EventType() string
	AggregateID() uuid.UUID
}

// BaseEvent - базовая реализация domain event
type BaseEvent struct {
	occurredAt  time.Time
	eventType   string
	aggregateID uuid.UUID
}

func (e BaseEvent) OccurredAt() time.Time {
	return e.occurredAt
}

func (e BaseEvent) EventType() string {
	return e.eventType
}

func (e BaseEvent) AggregateID() uuid.UUID {
	return e.aggregateID
}

// UserRegistered - событие регистрации пользователя
type UserRegistered struct {
	BaseEvent
	UserID uuid.UUID
	Email  string
	Name   string
}

func NewUserRegistered(userID uuid.UUID, email, name string) *UserRegistered {
	return &UserRegistered{
		BaseEvent: BaseEvent{
			occurredAt:  time.Now(),
			eventType:   "user.registered",
			aggregateID: userID,
		},
		UserID: userID,
		Email:  email,
		Name:   name,
	}
}

// UserUpdated - событие обновления пользователя
type UserUpdated struct {
	BaseEvent
	UserID uuid.UUID
	Field  string
	OldVal string
	NewVal string
}

func NewUserUpdated(userID uuid.UUID, field, oldVal, newVal string) *UserUpdated {
	return &UserUpdated{
		BaseEvent: BaseEvent{
			occurredAt:  time.Now(),
			eventType:   "user.updated",
			aggregateID: userID,
		},
		UserID: userID,
		Field:  field,
		OldVal: oldVal,
		NewVal: newVal,
	}
}

// UserDeleted - событие удаления пользователя
type UserDeleted struct {
	BaseEvent
	UserID uuid.UUID
}

func NewUserDeleted(userID uuid.UUID) *UserDeleted {
	return &UserDeleted{
		BaseEvent: BaseEvent{
			occurredAt:  time.Now(),
			eventType:   "user.deleted",
			aggregateID: userID,
		},
		UserID: userID,
	}
}
