package eventbus

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// DomainEvent - интерфейс domain события
type DomainEvent interface {
	EventType() string
	AggregateID() uuid.UUID
}

// EventHandler - обработчик события
type EventHandler func(ctx context.Context, event DomainEvent) error

// EventBus - шина событий для domain events
type EventBus struct {
	mu       sync.RWMutex
	handlers map[string][]EventHandler
}

// NewEventBus создает новую шину событий
func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(map[string][]EventHandler),
	}
}

// Subscribe подписывается на события определенного типа
func (b *EventBus) Subscribe(eventType string, handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[eventType] = append(b.handlers[eventType], handler)
	log.Info().
		Str("event_type", eventType).
		Msg("Subscribed to domain event")
}

// Publish публикует событие синхронно
func (b *EventBus) Publish(ctx context.Context, event DomainEvent) error {
	b.mu.RLock()
	handlers, exists := b.handlers[event.EventType()]
	b.mu.RUnlock()

	if !exists || len(handlers) == 0 {
		log.Debug().
			Str("event_type", event.EventType()).
			Msg("No handlers for event")
		return nil
	}

	log.Info().
		Str("event_type", event.EventType()).
		Str("aggregate_id", event.AggregateID().String()).
		Int("handlers_count", len(handlers)).
		Msg("Publishing domain event")

	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			log.Error().
				Err(err).
				Str("event_type", event.EventType()).
				Msg("Error handling domain event")
			return err
		}
	}

	return nil
}

// PublishAsync публикует событие асинхронно
func (b *EventBus) PublishAsync(ctx context.Context, event DomainEvent) {
	go func() {
		if err := b.Publish(ctx, event); err != nil {
			log.Error().
				Err(err).
				Str("event_type", event.EventType()).
				Msg("Error in async event publishing")
		}
	}()
}

// Clear очищает все подписки (для тестов)
func (b *EventBus) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers = make(map[string][]EventHandler)
}
