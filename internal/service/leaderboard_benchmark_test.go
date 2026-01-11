package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	authrepo "leaderboard-service/internal/auth/repository"
	leaderboardmodels "leaderboard-service/internal/leaderboard/models"
	leaderboardrepo "leaderboard-service/internal/leaderboard/repository"
	leaderboardservice "leaderboard-service/internal/leaderboard/service"
	"leaderboard-service/internal/shared/config"
	"leaderboard-service/internal/shared/database"
	"leaderboard-service/internal/shared/repository/decorators"

	"github.com/google/uuid"
)

// Helper function to create leaderboard service with repositories for benchmarks
func newBenchLeaderboardService(db *database.PostgresDB, redis *database.RedisClient, cfg *config.Config) *leaderboardservice.LeaderboardService {
	// Use decorators in benchmarks
	cache := decorators.NewSimpleCache()

	baseUserRepo := authrepo.NewPostgresUserRepository(db)
	baseScoreRepo := leaderboardrepo.NewPostgresScoreRepository(db)

	userRepo := decorators.NewCachedUserRepository(baseUserRepo, cache)
	scoreRepo := decorators.NewCachedScoreRepository(baseScoreRepo, cache)

	return leaderboardservice.NewLeaderboardService(scoreRepo, userRepo, redis, cfg)
}

// BenchmarkGetLeaderboardWithRedis measures performance with Redis cache
func BenchmarkGetLeaderboardWithRedis(b *testing.B) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			URL:      "postgres://postgres:postgres@localhost:5432/leaderboard?sslmode=disable",
			MaxConns: 25,
			MinConns: 5,
		},
		Redis: config.RedisConfig{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
	}

	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		b.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer db.Close()

	redis, err := database.NewRedisClient(cfg)
	if err != nil {
		b.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redis.Close()

	service := newBenchLeaderboardService(db, redis, cfg)

	// Prepare test data
	ctx := context.Background()
	query := &leaderboardmodels.LeaderboardQuery{
		Season:    "global",
		Limit:     10,
		Page:      0,
		SortOrder: "desc",
	}

	// Reset timer after setup
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := service.GetLeaderboard(ctx, query)
		if err != nil {
			b.Errorf("GetLeaderboard failed: %v", err)
		}
	}
}

// BenchmarkGetLeaderboardNoRedis measures performance without Redis cache
func BenchmarkGetLeaderboardNoRedis(b *testing.B) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			URL:      "postgres://postgres:postgres@localhost:5432/leaderboard?sslmode=disable",
			MaxConns: 25,
			MinConns: 5,
		},
	}

	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		b.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer db.Close()

	service := newBenchLeaderboardService(db, nil, cfg) // No Redis

	ctx := context.Background()
	query := &leaderboardmodels.LeaderboardQuery{
		Season:    "global",
		Limit:     10,
		Page:      0,
		SortOrder: "desc",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := service.GetLeaderboard(ctx, query)
		if err != nil {
			b.Errorf("GetLeaderboard failed: %v", err)
		}
	}
}

// BenchmarkSubmitScoreWithRedis measures score submission with Redis
func BenchmarkSubmitScoreWithRedis(b *testing.B) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			URL:      "postgres://postgres:postgres@localhost:5432/leaderboard?sslmode=disable",
			MaxConns: 25,
			MinConns: 5,
		},
		Redis: config.RedisConfig{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
	}

	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		b.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer db.Close()

	redis, err := database.NewRedisClient(cfg)
	if err != nil {
		b.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redis.Close()

	service := newBenchLeaderboardService(db, redis, cfg)
	ctx := context.Background()

	// Create test user
	userID := uuid.New()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := &leaderboardmodels.SubmitScoreRequest{
			Score:  int64(1000 + i),
			Season: "global",
		}
		_, err := service.SubmitScore(ctx, userID, req)
		if err != nil {
			b.Errorf("SubmitScore failed: %v", err)
		}
	}
}

// BenchmarkSubmitScoreNoRedis measures score submission without Redis
func BenchmarkSubmitScoreNoRedis(b *testing.B) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			URL:      "postgres://postgres:postgres@localhost:5432/leaderboard?sslmode=disable",
			MaxConns: 25,
			MinConns: 5,
		},
	}

	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		b.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer db.Close()

	service := newBenchLeaderboardService(db, nil, cfg) // No Redis

	ctx := context.Background()
	userID := uuid.New()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := &leaderboardmodels.SubmitScoreRequest{
			Score:  int64(1000 + i),
			Season: "global",
		}
		_, err := service.SubmitScore(ctx, userID, req)
		if err != nil {
			b.Errorf("SubmitScore failed: %v", err)
		}
	}
}

// BenchmarkGetUserRank measures rank lookup performance
func BenchmarkGetUserRank(b *testing.B) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			URL:      "postgres://postgres:postgres@localhost:5432/leaderboard?sslmode=disable",
			MaxConns: 25,
			MinConns: 5,
		},
	}

	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		b.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer db.Close()

	service := newBenchLeaderboardService(db, nil, cfg)
	ctx := context.Background()

	// Get a valid user ID from database
	var userID uuid.UUID
	err = db.DB.Raw("SELECT user_id FROM scores LIMIT 1").Scan(&userID).Error
	if err != nil {
		b.Skip("No test data available")
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := service.GetUserRank(ctx, userID, "global")
		if err != nil {
			b.Errorf("GetUserRank failed: %v", err)
		}
	}
}

// BenchmarkParallelGetLeaderboard measures concurrent access performance
func BenchmarkParallelGetLeaderboard(b *testing.B) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			URL:      "postgres://postgres:postgres@localhost:5432/leaderboard?sslmode=disable",
			MaxConns: 25,
			MinConns: 5,
		},
		Redis: config.RedisConfig{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
	}

	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		b.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer db.Close()

	redis, err := database.NewRedisClient(cfg)
	if err != nil {
		b.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redis.Close()

	service := newBenchLeaderboardService(db, redis, cfg)

	query := &leaderboardmodels.LeaderboardQuery{
		Season:    "global",
		Limit:     10,
		Page:      0,
		SortOrder: "desc",
	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			_, err := service.GetLeaderboard(ctx, query)
			if err != nil {
				b.Errorf("GetLeaderboard failed: %v", err)
			}
		}
	})
}

// BenchmarkLeaderboardPagination measures pagination performance
func BenchmarkLeaderboardPagination(b *testing.B) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			URL:      "postgres://postgres:postgres@localhost:5432/leaderboard?sslmode=disable",
			MaxConns: 25,
			MinConns: 5,
		},
	}

	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		b.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer db.Close()

	service := newBenchLeaderboardService(db, nil, cfg)
	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		page := i % 100 // Test first 100 pages
		query := &leaderboardmodels.LeaderboardQuery{
			Season:    "global",
			Limit:     10,
			Page:      page,
			SortOrder: "desc",
		}
		_, err := service.GetLeaderboard(ctx, query)
		if err != nil {
			b.Errorf("GetLeaderboard page %d failed: %v", page, err)
		}
	}
}

// BenchmarkRedisCacheHit measures Redis cache hit performance
func BenchmarkRedisCacheHit(b *testing.B) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			URL:      "postgres://postgres:postgres@localhost:5432/leaderboard?sslmode=disable",
			MaxConns: 25,
			MinConns: 5,
		},
		Redis: config.RedisConfig{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
	}

	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		b.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer db.Close()

	redis, err := database.NewRedisClient(cfg)
	if err != nil {
		b.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redis.Close()

	service := newBenchLeaderboardService(db, redis, cfg)
	ctx := context.Background()

	// Warm up cache
	query := &leaderboardmodels.LeaderboardQuery{
		Season:    "global",
		Limit:     10,
		Page:      0,
		SortOrder: "desc",
	}
	_, err = service.GetLeaderboard(ctx, query)
	if err != nil {
		b.Fatalf("Failed to warm up cache: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := service.GetLeaderboard(ctx, query)
		if err != nil {
			b.Errorf("GetLeaderboard failed: %v", err)
		}
	}
}

// BenchmarkDifferentSeasons measures performance across different seasons
func BenchmarkDifferentSeasons(b *testing.B) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			URL:      "postgres://postgres:postgres@localhost:5432/leaderboard?sslmode=disable",
			MaxConns: 25,
			MinConns: 5,
		},
		Redis: config.RedisConfig{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
	}

	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		b.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer db.Close()

	redis, err := database.NewRedisClient(cfg)
	if err != nil {
		b.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redis.Close()

	service := newBenchLeaderboardService(db, redis, cfg)
	ctx := context.Background()

	seasons := []string{"global", "2024", "2025", "january", "february"}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		season := seasons[i%len(seasons)]
		query := &leaderboardmodels.LeaderboardQuery{
			Season:    season,
			Limit:     10,
			Page:      0,
			SortOrder: "desc",
		}
		_, err := service.GetLeaderboard(ctx, query)
		if err != nil {
			b.Errorf("GetLeaderboard for season %s failed: %v", season, err)
		}
	}
}

// Helper function to print benchmark comparison
func printBenchmarkResults(b *testing.B) {
	fmt.Printf("\n=== Benchmark Results ===\n")
	fmt.Printf("Operations: %d\n", b.N)
	fmt.Printf("Time per op: %v\n", b.Elapsed()/time.Duration(b.N))
	fmt.Printf("Total time: %v\n", b.Elapsed())
}
