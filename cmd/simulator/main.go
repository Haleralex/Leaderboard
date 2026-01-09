package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	authrepo "leaderboard-service/internal/auth/repository"
	leaderboardmodels "leaderboard-service/internal/leaderboard/models"
	leaderboardrepo "leaderboard-service/internal/leaderboard/repository"
	leaderboardservice "leaderboard-service/internal/leaderboard/service"
	"leaderboard-service/internal/shared/config"
	"leaderboard-service/internal/shared/database"
	"leaderboard-service/internal/shared/middleware"
	"leaderboard-service/internal/shared/repository/decorators"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

const (
	updateInterval   = 5 * time.Second
	minScore         = 100
	maxScore         = 10000
	scoreIncrement   = 50
	minUsersRequired = 2
)

type existingUser struct {
	ID   uuid.UUID
	Name string
}

func main() {
	log.Info().Msg("🎮 Starting Leaderboard Score Simulator")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Setup logger
	middleware.SetupLogger(cfg.Log.Level)

	// Initialize database
	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to PostgreSQL")
	}
	defer db.Close()

	// Initialize Redis (optional)
	redis, err := database.NewRedisClient(cfg)
	if err != nil {
		log.Warn().Err(err).Msg("Redis not available, running without cache")
		redis = nil
	} else {
		defer redis.Close()
	}

	// Create cache for decorators
	cache := decorators.NewSimpleCache()

	// Initialize repositories with decorators
	baseUserRepo := authrepo.NewPostgresUserRepository(db)
	baseScoreRepo := leaderboardrepo.NewPostgresScoreRepository(db)

	userRepo := decorators.NewCachedUserRepository(baseUserRepo, cache)
	scoreRepo := decorators.NewCachedScoreRepository(baseScoreRepo, cache)

	// Initialize services
	leaderboardService := leaderboardservice.NewLeaderboardService(scoreRepo, userRepo, redis)

	// Load existing users from database
	users := loadExistingUsers(db)
	if len(users) < minUsersRequired {
		log.Fatal().Int("found", len(users)).Int("required", minUsersRequired).Msg("Not enough users in database. Please register users first!")
	}

	log.Info().Int("count", len(users)).Msg("✅ Loaded existing users from database")

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start simulation ticker
	ticker := time.NewTicker(updateInterval)
	defer ticker.Stop()

	log.Info().Dur("interval", updateInterval).Msg("🚀 Starting score simulation loop")

	// Initial scores
	simulateScoreUpdates(leaderboardService, users)

	for {
		select {
		case <-ticker.C:
			simulateScoreUpdates(leaderboardService, users)
		case <-quit:
			log.Info().Msg("🛑 Shutting down simulator...")
			return
		}
	}
}

// loadExistingUsers loads all existing users from the database
func loadExistingUsers(db *database.PostgresDB) []existingUser {
	users := make([]existingUser, 0)
	ctx := context.Background()

	// Use GORM to query users
	var dbUsers []struct {
		ID   uuid.UUID `gorm:"column:id"`
		Name string    `gorm:"column:name"`
	}

	err := db.DB.WithContext(ctx).Table("users").Select("id, name").Order("created_at DESC").Find(&dbUsers).Error
	if err != nil {
		log.Error().Err(err).Msg("Failed to query users from database")
		return users
	}

	for _, u := range dbUsers {
		users = append(users, existingUser{
			ID:   u.ID,
			Name: u.Name,
		})
		log.Debug().Str("name", u.Name).Str("id", u.ID.String()).Msg("Loaded user")
	}

	return users
}

// simulateScoreUpdates randomly updates scores for existing users
func simulateScoreUpdates(leaderboardService *leaderboardservice.LeaderboardService, users []existingUser) {
	ctx := context.Background()

	// Pick 30-50% of users to update (minimum 2, maximum all)
	numUpdates := len(users) * (30 + rand.Intn(21)) / 100
	if numUpdates < 2 {
		numUpdates = 2
	}
	if numUpdates > len(users) {
		numUpdates = len(users)
	}

	// Shuffle and pick
	shuffled := make([]existingUser, len(users))
	copy(shuffled, users)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	log.Info().Int("users_to_update", numUpdates).Msg("📊 Simulating score updates")

	for i := 0; i < numUpdates; i++ {
		user := shuffled[i]

		// Generate random score change
		var newScore int64
		if rand.Float32() < 0.7 {
			// 70% chance: increment score
			increment := int64(rand.Intn(5)+1) * scoreIncrement
			newScore = increment
		} else {
			// 30% chance: set random score
			newScore = int64(rand.Intn(maxScore-minScore) + minScore)
		}

		// Choose season: 60% global, 40% random season1-5
		season := "global"
		if rand.Float32() < 0.4 {
			// Pick random season from season1-5
			seasonNum := rand.Intn(5) + 1
			season = fmt.Sprintf("season%d", seasonNum)
		}

		// Submit score
		submitReq := &leaderboardmodels.SubmitScoreRequest{
			Score:  newScore,
			Season: season,
			Metadata: map[string]interface{}{
				"simulated":  true,
				"timestamp":  time.Now().Unix(),
				"session_id": uuid.New().String()[:8],
			},
		}

		_, err := leaderboardService.SubmitScore(ctx, user.ID, submitReq)
		if err != nil {
			log.Error().Err(err).Str("user", user.Name).Msg("❌ Failed to submit score")
			continue
		}

		// Get user's rank
		entry, err := leaderboardService.GetUserRank(ctx, user.ID, season)
		if err != nil {
			log.Warn().Err(err).Str("user", user.Name).Msg("Could not get rank")
		} else {
			log.Info().
				Str("user", user.Name).
				Int64("score", newScore).
				Int("rank", entry.Rank).
				Str("season", season).
				Msg("🎯 Score updated")
		}

		// Small delay between updates
		time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)
	}

	// Show current top 5
	showTopPlayers(leaderboardService, "global")
}

// showTopPlayers displays the current top 5 leaderboard
func showTopPlayers(leaderboardService *leaderboardservice.LeaderboardService, season string) {
	ctx := context.Background()

	query := &leaderboardmodels.LeaderboardQuery{
		Season:    season,
		Limit:     5,
		Page:      0,
		SortOrder: "desc",
	}

	resp, err := leaderboardService.GetLeaderboard(ctx, query)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get leaderboard")
		return
	}

	if len(resp.Entries) == 0 {
		log.Info().Msg("📋 Leaderboard is empty")
		return
	}

	log.Info().Str("season", season).Msg("🏆 Current Top 5:")
	for _, entry := range resp.Entries {
		log.Info().
			Int("rank", entry.Rank).
			Str("player", entry.UserName).
			Int64("score", entry.Score).
			Msgf("   #%d: %s - %d points", entry.Rank, entry.UserName, entry.Score)
	}
}
