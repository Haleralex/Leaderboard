package main

import (
	"context"
	"math/rand"

	"leaderboard-service/internal/shared/config"
	"leaderboard-service/internal/shared/database"
	"leaderboard-service/internal/shared/middleware"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

const (
	maxUsers = 100
	minScore = 500
	maxScore = 50000
)

func main() {
	log.Info().Msg("🧹 Starting database cleanup and score redistribution")

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

	ctx := context.Background()

	// Count current users
	var currentCount int64
	err = db.DB.WithContext(ctx).Model(&struct {
		ID uuid.UUID `gorm:"column:id"`
	}{}).Table("users").Count(&currentCount).Error
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to count users")
	}

	log.Info().Int64("current", currentCount).Int("max", maxUsers).Msg("Current user count")

	if currentCount <= int64(maxUsers) {
		log.Info().Msg("✅ User count is within limit, no deletion needed")
	} else {
		// Delete excess users
		usersToDelete := currentCount - int64(maxUsers)
		log.Info().Int64("to_delete", usersToDelete).Msg("Deleting excess users...")

		// Keep the most recently created users, delete the oldest
		deleteQuery := `
			DELETE FROM users 
			WHERE id IN (
				SELECT id FROM users 
				ORDER BY created_at ASC 
				LIMIT ?
			)
		`

		result := db.DB.WithContext(ctx).Exec(deleteQuery, usersToDelete)
		if result.Error != nil {
			log.Fatal().Err(result.Error).Msg("Failed to delete users")
		}

		rowsDeleted := result.RowsAffected
		log.Info().Int64("deleted", rowsDeleted).Msg("✅ Users deleted")
	}

	// Get remaining users
	var remainingCount int64
	err = db.DB.WithContext(ctx).Model(&struct {
		ID uuid.UUID `gorm:"column:id"`
	}{}).Table("users").Count(&remainingCount).Error
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to count remaining users")
	}

	log.Info().Int64("remaining", remainingCount).Msg("Remaining users")

	// Redistribute scores
	log.Info().Msg("🎲 Redistributing scores for better variety...")

	// Get all user IDs
	var userIDs []uuid.UUID
	err = db.DB.WithContext(ctx).Table("users").Select("id").Pluck("id", &userIDs).Error
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to query users")
	}

	log.Info().Int("count", len(userIDs)).Msg("Loaded user IDs")

	// Update scores with diverse values
	updatedCount := 0
	for _, userID := range userIDs {
		// Generate diverse score using power distribution
		// This creates more variety with some high scores and many lower scores
		randomValue := rand.Float64()

		var newScore int64
		if randomValue < 0.01 {
			// 1% - very high scores (top players)
			newScore = int64(40000 + rand.Intn(10000))
		} else if randomValue < 0.05 {
			// 4% - high scores
			newScore = int64(30000 + rand.Intn(10000))
		} else if randomValue < 0.15 {
			// 10% - above average
			newScore = int64(20000 + rand.Intn(10000))
		} else if randomValue < 0.40 {
			// 25% - average
			newScore = int64(10000 + rand.Intn(10000))
		} else if randomValue < 0.70 {
			// 30% - below average
			newScore = int64(5000 + rand.Intn(5000))
		} else {
			// 30% - low scores
			newScore = int64(500 + rand.Intn(4500))
		}

		// Update or insert score for global season
		upsertQuery := `
			INSERT INTO scores (user_id, score, season, metadata, timestamp)
			VALUES (?, ?, 'global', '{"redistributed": "true"}', NOW())
			ON CONFLICT (user_id, season)
			DO UPDATE SET score = ?, timestamp = NOW()
		`

		err := db.DB.WithContext(ctx).Exec(upsertQuery, userID, newScore, newScore).Error
		if err != nil {
			log.Warn().Err(err).Str("user_id", userID.String()).Msg("Failed to update score")
			continue
		}

		updatedCount++

		if updatedCount%100 == 0 {
			log.Info().Int("updated", updatedCount).Msg("Progress...")
		}
	}

	log.Info().Int("updated", updatedCount).Msg("✅ Scores redistributed")

	// Show statistics
	log.Info().Msg("📊 Final statistics:")

	var stats struct {
		TotalUsers  int64
		TotalScores int64
		AvgScore    float64
		MinScore    int64
		MaxScore    int64
	}

	db.DB.WithContext(ctx).Table("users").Count(&stats.TotalUsers)
	db.DB.WithContext(ctx).Table("scores").Where("season = 'global'").Count(&stats.TotalScores)
	db.DB.WithContext(ctx).Table("scores").Where("season = 'global'").Select("AVG(score)").Scan(&stats.AvgScore)
	db.DB.WithContext(ctx).Table("scores").Where("season = 'global'").Select("MIN(score)").Scan(&stats.MinScore)
	db.DB.WithContext(ctx).Table("scores").Where("season = 'global'").Select("MAX(score)").Scan(&stats.MaxScore)

	log.Info().
		Int64("total_users", stats.TotalUsers).
		Int64("total_scores", stats.TotalScores).
		Float64("avg_score", stats.AvgScore).
		Int64("min_score", stats.MinScore).
		Int64("max_score", stats.MaxScore).
		Msg("Database statistics")

	// Show top 10
	log.Info().Msg("🏆 Top 10 players:")

	var topPlayers []struct {
		Name  string `gorm:"column:name"`
		Score int64  `gorm:"column:score"`
	}

	err = db.DB.WithContext(ctx).Raw(`
		SELECT u.name, s.score
		FROM scores s
		JOIN users u ON s.user_id = u.id
		WHERE s.season = 'global'
		ORDER BY s.score DESC
		LIMIT 10
	`).Scan(&topPlayers).Error
	if err != nil {
		log.Error().Err(err).Msg("Failed to query top players")
	} else {
		for i, player := range topPlayers {
			rank := i + 1
			log.Info().Int("rank", rank).Str("name", player.Name).Int64("score", player.Score).Msgf("   #%d: %s - %d points", rank, player.Name, player.Score)
		}
	}

	log.Info().Msg("✅ Cleanup and redistribution complete!")
}
