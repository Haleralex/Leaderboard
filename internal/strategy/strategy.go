package strategy

import (
	"context"
	"time"

	leaderboardmodels "leaderboard-service/internal/leaderboard/models"

	"github.com/google/uuid"
)

// Strategy Pattern - поведенческий паттерн, который определяет семейство алгоритмов,
// инкапсулирует каждый из них и делает их взаимозаменяемыми.

// ScoringStrategy определяет стратегию подсчета очков
type ScoringStrategy interface {
	// Calculate вычисляет итоговый счет на основе базового значения и контекста
	Calculate(baseScore int64, context *ScoringContext) int64

	// Name возвращает название стратегии
	Name() string
}

// ScoringContext контекст для подсчета очков
type ScoringContext struct {
	UserID       uuid.UUID
	Season       string
	GameMode     string
	Difficulty   int
	Multiplier   float64
	Combo        int
	TimeBonus    int64
	Achievements []string
	Metadata     map[string]interface{}
}

// RankingStrategy определяет стратегию ранжирования игроков в лидерборде
type RankingStrategy interface {
	// CalculateRanks вычисляет ранги для списка счетов
	CalculateRanks(scores []*leaderboardmodels.Score) []*RankedScore

	// Name возвращает название стратегии
	Name() string
}

// RankedScore счет с рангом
type RankedScore struct {
	*leaderboardmodels.Score
	Rank         int
	TiedWithPrev bool // Есть ли ничья с предыдущим
}

// CacheStrategy определяет стратегию кэширования
type CacheStrategy interface {
	// Get получает значение из кэша
	Get(ctx context.Context, key string) ([]byte, error)

	// Set сохраняет значение в кэш
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

	// Delete удаляет значение из кэша
	Delete(ctx context.Context, key string) error

	// Name возвращает название стратегии
	Name() string
}

// ValidationStrategy определяет стратегию валидации данных
type ValidationStrategy interface {
	// Validate проверяет данные
	Validate(data interface{}) error

	// Name возвращает название стратегии
	Name() string
}

// ScoreValidationContext контекст валидации счета
type ScoreValidationContext struct {
	Score         int64
	Season        string
	UserID        uuid.UUID
	PreviousScore *leaderboardmodels.Score
	MaxScore      int64
	MinScore      int64
}

// NotificationStrategy определяет стратегию уведомлений
type NotificationStrategy interface {
	// Send отправляет уведомление
	Send(ctx context.Context, notification *Notification) error

	// Name возвращает название стратегии
	Name() string
}

// Notification уведомление
type Notification struct {
	UserID  uuid.UUID
	Type    string
	Title   string
	Message string
	Data    map[string]interface{}
}

// SortStrategy определяет стратегию сортировки
type SortStrategy interface {
	// Sort сортирует список счетов
	Sort(scores []*leaderboardmodels.Score) []*leaderboardmodels.Score

	// Name возвращает название стратегии
	Name() string
}

// FilterStrategy определяет стратегию фильтрации
type FilterStrategy interface {
	// Filter фильтрует список счетов
	Filter(scores []*leaderboardmodels.Score) []*leaderboardmodels.Score

	// Name возвращает название стратегии
	Name() string
}

// AggregationStrategy определяет стратегию агрегации данных
type AggregationStrategy interface {
	// Aggregate агрегирует счета пользователя
	Aggregate(scores []*leaderboardmodels.Score) int64

	// Name возвращает название стратегии
	Name() string
}

// RetryStrategy определяет стратегию повторных попыток
type RetryStrategy interface {
	// ShouldRetry определяет, нужно ли повторить попытку
	ShouldRetry(attempt int, err error) bool

	// NextDelay возвращает задержку перед следующей попыткой
	NextDelay(attempt int) time.Duration

	// Name возвращает название стратегии
	Name() string
}

// CompressionStrategy определяет стратегию сжатия данных
type CompressionStrategy interface {
	// Compress сжимает данные
	Compress(data []byte) ([]byte, error)

	// Decompress распаковывает данные
	Decompress(data []byte) ([]byte, error)

	// Name возвращает название стратегии
	Name() string
}

// EncryptionStrategy определяет стратегию шифрования
type EncryptionStrategy interface {
	// Encrypt шифрует данные
	Encrypt(data []byte) ([]byte, error)

	// Decrypt расшифровывает данные
	Decrypt(data []byte) ([]byte, error)

	// Name возвращает название стратегии
	Name() string
}
