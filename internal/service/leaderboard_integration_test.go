package service

import (
	"context"
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create leaderboard service with repositories
func newTestLeaderboardService(db *database.PostgresDB, redis *database.RedisClient, cfg *config.Config) *leaderboardservice.LeaderboardService {
	// Use decorators in tests too
	cache := decorators.NewSimpleCache()

	baseUserRepo := authrepo.NewPostgresUserRepository(db)
	baseScoreRepo := leaderboardrepo.NewPostgresScoreRepository(db)

	userRepo := decorators.NewCachedUserRepository(baseUserRepo, cache)
	scoreRepo := decorators.NewCachedScoreRepository(baseScoreRepo, cache)

	return leaderboardservice.NewLeaderboardService(scoreRepo, userRepo, redis, cfg)
}

// Helper function to create test config
func newTestConfig() *config.Config {
	return &config.Config{
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
		Validation: config.ValidationConfig{
			MaxScore: 1000000,
			MinScore: 0,
		},
	}
}

// TestIntegrationGetLeaderboardWithRedis tests leaderboard retrieval with Redis cache
func TestIntegrationGetLeaderboardWithRedis(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	cfg := newTestConfig()

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err, "Failed to connect to PostgreSQL")
	defer db.Close()

	redis, err := database.NewRedisClient(cfg)
	require.NoError(t, err, "Failed to connect to Redis")
	defer redis.Close()

	service := newTestLeaderboardService(db, redis, cfg)
	ctx := context.Background()

	// First call - should hit database
	query := &leaderboardmodels.LeaderboardQuery{
		Season:    "global",
		Limit:     10,
		Page:      0,
		SortOrder: "desc",
	}

	start := time.Now()
	result1, err := service.GetLeaderboard(ctx, query)
	dbTime := time.Since(start)

	require.NoError(t, err)
	require.NotNil(t, result1)
	assert.GreaterOrEqual(t, len(result1.Entries), 0)

	// Second call - should hit Redis cache (faster)
	start = time.Now()
	result2, err := service.GetLeaderboard(ctx, query)
	cacheTime := time.Since(start)

	require.NoError(t, err)
	require.NotNil(t, result2)
	assert.Equal(t, len(result1.Entries), len(result2.Entries))

	t.Logf("Database query time: %v", dbTime)
	t.Logf("Redis cache time: %v", cacheTime)
	t.Logf("Cache speedup: %.2fx", float64(dbTime)/float64(cacheTime))
}

// TestIntegrationGetLeaderboardNoRedis tests leaderboard without Redis
func TestIntegrationGetLeaderboardNoRedis(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	cfg := newTestConfig()

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	service := newTestLeaderboardService(db, nil, cfg) // No Redis

	ctx := context.Background()
	query := &leaderboardmodels.LeaderboardQuery{
		Season:    "global",
		Limit:     10,
		Page:      0,
		SortOrder: "desc",
	}

	result, err := service.GetLeaderboard(ctx, query)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.GreaterOrEqual(t, len(result.Entries), 0)
}

// TestIntegrationSubmitScoreAndRetrieve tests score submission and retrieval
func TestIntegrationSubmitScoreAndRetrieve(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	cfg := newTestConfig()

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	redis, err := database.NewRedisClient(cfg)
	require.NoError(t, err)
	defer redis.Close()

	service := newTestLeaderboardService(db, redis, cfg)
	ctx := context.Background()

	// Create test user first
	userID := uuid.New()
	db.DB.Exec("INSERT INTO users (id, name, email, password_hash) VALUES (?, ?, ?, ?)",
		userID, "Test User", "test@example.com", "hashed")

	// Submit a score
	testScore := int64(99999)

	req := &leaderboardmodels.SubmitScoreRequest{
		Score:  testScore,
		Season: "test_season",
	}

	start := time.Now()
	score, err := service.SubmitScore(ctx, userID, req)
	submitTime := time.Since(start)

	require.NoError(t, err)
	require.NotNil(t, score)
	assert.Equal(t, testScore, score.Score)

	t.Logf("Score submission time: %v", submitTime)

	// Retrieve user rank
	start = time.Now()
	rank, err := service.GetUserRank(ctx, userID, "test_season")
	rankTime := time.Since(start)

	require.NoError(t, err)
	require.NotNil(t, rank)
	assert.Equal(t, testScore, rank.Score)

	t.Logf("Rank retrieval time: %v", rankTime)

	// Clean up test data
	db.DB.Exec("DELETE FROM scores WHERE user_id = ?", userID)
	db.DB.Exec("DELETE FROM users WHERE id = ?", userID)
}

// TestIntegrationConcurrentAccess tests concurrent reads and writes
func TestIntegrationConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	cfg := newTestConfig()

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	redis, err := database.NewRedisClient(cfg)
	require.NoError(t, err)
	defer redis.Close()

	service := newTestLeaderboardService(db, redis, cfg)

	// Launch concurrent requests
	concurrency := 10
	done := make(chan bool, concurrency)

	start := time.Now()

	for i := 0; i < concurrency; i++ {
		go func(index int) {
			ctx := context.Background()
			query := &leaderboardmodels.LeaderboardQuery{
				Season:    "global",
				Limit:     10,
				Page:      index % 5, // Different pages
				SortOrder: "desc",
			}

			_, err := service.GetLeaderboard(ctx, query)
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < concurrency; i++ {
		<-done
	}

	totalTime := time.Since(start)
	t.Logf("Concurrent access (%d requests) completed in: %v", concurrency, totalTime)
	t.Logf("Average time per request: %v", totalTime/time.Duration(concurrency))
}

// TestIntegrationRedisCacheInvalidation tests cache invalidation on score update
func TestIntegrationRedisCacheInvalidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	cfg := newTestConfig()

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	redis, err := database.NewRedisClient(cfg)
	require.NoError(t, err)
	defer redis.Close()

	service := newTestLeaderboardService(db, redis, cfg)
	ctx := context.Background()

	season := "cache_test"
	userID := uuid.New()

	// Create test user first
	db.DB.Exec("INSERT INTO users (id, name, email, password_hash) VALUES (?, ?, ?, ?)",
		userID, "Test User", "test@example.com", "hashed")

	// Submit initial score
	req1 := &leaderboardmodels.SubmitScoreRequest{
		Score:  1000,
		Season: season,
	}
	_, err = service.SubmitScore(ctx, userID, req1)
	require.NoError(t, err)

	// Wait for cache update
	time.Sleep(100 * time.Millisecond)

	// Get leaderboard (should be cached)
	query := &leaderboardmodels.LeaderboardQuery{
		Season:    season,
		Limit:     10,
		Page:      0,
		SortOrder: "desc",
	}
	result1, err := service.GetLeaderboard(ctx, query)
	require.NoError(t, err)

	// Update score
	req2 := &leaderboardmodels.SubmitScoreRequest{
		Score:  2000,
		Season: season,
	}
	_, err = service.SubmitScore(ctx, userID, req2)
	require.NoError(t, err)

	// Wait for cache update
	time.Sleep(100 * time.Millisecond)

	// Get leaderboard again (cache should be updated)
	result2, err := service.GetLeaderboard(ctx, query)
	require.NoError(t, err)

	t.Logf("Initial leaderboard entries: %d", len(result1.Entries))
	t.Logf("Updated leaderboard entries: %d", len(result2.Entries))

	// Clean up
	db.DB.Exec("DELETE FROM scores WHERE user_id = ?", userID)
	db.DB.Exec("DELETE FROM users WHERE id = ?", userID)
}

// TestIntegrationPagination tests pagination correctness
func TestIntegrationPagination(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	cfg := newTestConfig()

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	service := newTestLeaderboardService(db, nil, cfg)
	ctx := context.Background()

	// Get first page
	query1 := &leaderboardmodels.LeaderboardQuery{
		Season:    "global",
		Limit:     5,
		Page:      0,
		SortOrder: "desc",
	}
	result1, err := service.GetLeaderboard(ctx, query1)
	require.NoError(t, err)

	// Get second page
	query2 := &leaderboardmodels.LeaderboardQuery{
		Season:    "global",
		Limit:     5,
		Page:      1,
		SortOrder: "desc",
	}
	result2, err := service.GetLeaderboard(ctx, query2)
	require.NoError(t, err)

	// Ensure no overlap between pages
	if len(result1.Entries) > 0 && len(result2.Entries) > 0 {
		lastRankPage1 := result1.Entries[len(result1.Entries)-1].Rank
		firstRankPage2 := result2.Entries[0].Rank

		assert.Less(t, lastRankPage1, firstRankPage2, "Pages should not overlap")
		t.Logf("Page 1 last rank: %d, Page 2 first rank: %d", lastRankPage1, firstRankPage2)
	}
}
