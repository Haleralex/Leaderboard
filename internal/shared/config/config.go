package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Server     ServerConfig
	Database   DatabaseConfig
	Redis      RedisConfig
	JWT        JWTConfig
	RateLimit  RateLimitConfig
	Supabase   SupabaseConfig
	Log        LogConfig
	WebSocket  WebSocketConfig
	Cache      CacheConfig
	Validation ValidationConfig
}

type ServerConfig struct {
	Port string
	Env  string
}

type DatabaseConfig struct {
	URL      string
	MaxConns int
	MinConns int
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret      string
	ExpiryHours int
}

type RateLimitConfig struct {
	Requests      int
	WindowSeconds int
}

type SupabaseConfig struct {
	URL     string
	AnonKey string
}

type LogConfig struct {
	Level string
}

type WebSocketConfig struct {
	BroadcastIntervalSeconds int
	DefaultLimit             int
	WriteWaitSeconds         int
	PongWaitSeconds          int
	PingPeriodSeconds        int
	MaxMessageSize           int64
}

type CacheConfig struct {
	LeaderboardTTLMinutes  int
	UserCacheTTLMinutes    int
	ScoreCacheTTLMinutes   int
	CleanupIntervalMinutes int
}

type ValidationConfig struct {
	MaxScore int64
	MinScore int64
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists (ignore error in production)
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Env:  getEnv("ENV", "development"),
		},
		Database: DatabaseConfig{
			URL:      getEnv("DATABASE_URL", ""),
			MaxConns: getEnvAsInt("DB_MAX_CONNS", 25),
			MinConns: getEnvAsInt("DB_MIN_CONNS", 5),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret:      getEnv("JWT_SECRET", ""),
			ExpiryHours: getEnvAsInt("JWT_EXPIRY_HOURS", 24),
		},
		RateLimit: RateLimitConfig{
			Requests:      getEnvAsInt("RATE_LIMIT_REQUESTS", 100),
			WindowSeconds: getEnvAsInt("RATE_LIMIT_WINDOW_SECONDS", 60),
		},
		Supabase: SupabaseConfig{
			URL:     getEnv("SUPABASE_URL", ""),
			AnonKey: getEnv("SUPABASE_ANON_KEY", ""),
		},
		Log: LogConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
		WebSocket: WebSocketConfig{
			BroadcastIntervalSeconds: getEnvAsInt("WS_BROADCAST_INTERVAL_SEC", 3),
			DefaultLimit:             getEnvAsInt("WS_DEFAULT_LIMIT", 50),
			WriteWaitSeconds:         getEnvAsInt("WS_WRITE_WAIT_SEC", 10),
			PongWaitSeconds:          getEnvAsInt("WS_PONG_WAIT_SEC", 60),
			PingPeriodSeconds:        getEnvAsInt("WS_PING_PERIOD_SEC", 54),
			MaxMessageSize:           getEnvAsInt64("WS_MAX_MESSAGE_SIZE", 512*1024),
		},
		Cache: CacheConfig{
			LeaderboardTTLMinutes:  getEnvAsInt("CACHE_LEADERBOARD_TTL_MIN", 5),
			UserCacheTTLMinutes:    getEnvAsInt("CACHE_USER_TTL_MIN", 5),
			ScoreCacheTTLMinutes:   getEnvAsInt("CACHE_SCORE_TTL_MIN", 2),
			CleanupIntervalMinutes: getEnvAsInt("CACHE_CLEANUP_INTERVAL_MIN", 5),
		},
		Validation: ValidationConfig{
			MaxScore: getEnvAsInt64("VALIDATION_MAX_SCORE", 10000000),
			MinScore: getEnvAsInt64("VALIDATION_MIN_SCORE", 0),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks if required configuration is present
func (c *Config) Validate() error {
	if c.Database.URL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	return nil
}

// Helper functions
func getEnv(key, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultVal
}

func getEnvAsInt64(key string, defaultVal int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intVal
		}
	}
	return defaultVal
}

func (c *Config) GetJWTExpiry() time.Duration {
	return time.Duration(c.JWT.ExpiryHours) * time.Hour
}

func (c *Config) GetWebSocketBroadcastInterval() time.Duration {
	return time.Duration(c.WebSocket.BroadcastIntervalSeconds) * time.Second
}

func (c *Config) GetWebSocketWriteWait() time.Duration {
	return time.Duration(c.WebSocket.WriteWaitSeconds) * time.Second
}

func (c *Config) GetWebSocketPongWait() time.Duration {
	return time.Duration(c.WebSocket.PongWaitSeconds) * time.Second
}

func (c *Config) GetWebSocketPingPeriod() time.Duration {
	return time.Duration(c.WebSocket.PingPeriodSeconds) * time.Second
}

func (c *Config) GetCacheLeaderboardTTL() time.Duration {
	return time.Duration(c.Cache.LeaderboardTTLMinutes) * time.Minute
}

func (c *Config) GetCacheUserTTL() time.Duration {
	return time.Duration(c.Cache.UserCacheTTLMinutes) * time.Minute
}

func (c *Config) GetCacheScoreTTL() time.Duration {
	return time.Duration(c.Cache.ScoreCacheTTLMinutes) * time.Minute
}

func (c *Config) GetCacheCleanupInterval() time.Duration {
	return time.Duration(c.Cache.CleanupIntervalMinutes) * time.Minute
}
