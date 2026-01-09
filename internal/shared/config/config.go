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
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	RateLimit RateLimitConfig
	Supabase SupabaseConfig
	Log      LogConfig
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
	Secret       string
	ExpiryHours  int
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

func (c *Config) GetJWTExpiry() time.Duration {
	return time.Duration(c.JWT.ExpiryHours) * time.Hour
}
