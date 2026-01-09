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

// ScoreSubmitted - событие отправки счета
type ScoreSubmitted struct {
	BaseEvent
	ScoreID  uuid.UUID
	UserID   uuid.UUID
	Score    int64
	Season   string
	Metadata map[string]interface{}
}

func NewScoreSubmitted(scoreID, userID uuid.UUID, score int64, season string, metadata map[string]interface{}) *ScoreSubmitted {
	return &ScoreSubmitted{
		BaseEvent: BaseEvent{
			occurredAt:  time.Now(),
			eventType:   "score.submitted",
			aggregateID: scoreID,
		},
		ScoreID:  scoreID,
		UserID:   userID,
		Score:    score,
		Season:   season,
		Metadata: metadata,
	}
}

// ScoreUpdated - событие обновления счета
type ScoreUpdated struct {
	BaseEvent
	ScoreID  uuid.UUID
	UserID   uuid.UUID
	OldScore int64
	NewScore int64
	Season   string
}

func NewScoreUpdated(scoreID, userID uuid.UUID, oldScore, newScore int64, season string) *ScoreUpdated {
	return &ScoreUpdated{
		BaseEvent: BaseEvent{
			occurredAt:  time.Now(),
			eventType:   "score.updated",
			aggregateID: scoreID,
		},
		ScoreID:  scoreID,
		UserID:   userID,
		OldScore: oldScore,
		NewScore: newScore,
		Season:   season,
	}
}

// LeaderboardPositionChanged - событие изменения позиции в лидерборде
type LeaderboardPositionChanged struct {
	BaseEvent
	UserID      uuid.UUID
	Season      string
	OldPosition int
	NewPosition int
	Score       int64
}

func NewLeaderboardPositionChanged(userID uuid.UUID, season string, oldPos, newPos int, score int64) *LeaderboardPositionChanged {
	return &LeaderboardPositionChanged{
		BaseEvent: BaseEvent{
			occurredAt:  time.Now(),
			eventType:   "leaderboard.position_changed",
			aggregateID: userID,
		},
		UserID:      userID,
		Season:      season,
		OldPosition: oldPos,
		NewPosition: newPos,
		Score:       score,
	}
}

// NewRecordSet - событие установки нового рекорда
type NewRecordSet struct {
	BaseEvent
	UserID       uuid.UUID
	Season       string
	Score        int64
	PreviousHigh int64
}

func NewNewRecordSet(userID uuid.UUID, season string, score, previousHigh int64) *NewRecordSet {
	return &NewRecordSet{
		BaseEvent: BaseEvent{
			occurredAt:  time.Now(),
			eventType:   "leaderboard.new_record",
			aggregateID: userID,
		},
		UserID:       userID,
		Season:       season,
		Score:        score,
		PreviousHigh: previousHigh,
	}
}
