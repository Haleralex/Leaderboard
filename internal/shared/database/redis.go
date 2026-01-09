package database

import (
	"context"
	"fmt"
	"time"

	"leaderboard-service/internal/shared/config"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// RedisClient wraps the Redis client
type RedisClient struct {
	Client *redis.Client
}

// NewRedisClient creates a new Redis client
func NewRedisClient(cfg *config.Config) (*RedisClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.Addr,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
	})

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("unable to connect to Redis: %w", err)
	}

	log.Info().Msg("Redis connection established")

	return &RedisClient{Client: client}, nil
}

// Close closes the Redis connection
func (r *RedisClient) Close() error {
	if r.Client != nil {
		if err := r.Client.Close(); err != nil {
			return err
		}
		log.Info().Msg("Redis connection closed")
	}
	return nil
}

// Health checks the Redis connection
func (r *RedisClient) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return r.Client.Ping(ctx).Err()
}
