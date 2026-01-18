package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load .env
	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	// Disable SSL for local development
	if !strings.Contains(dbURL, "sslmode") {
		if strings.Contains(dbURL, "?") {
			dbURL += "&sslmode=disable"
		} else {
			dbURL += "?sslmode=disable"
		}
	}

	numUsers := 1000000
	if len(os.Args) > 1 {
		if n, err := strconv.Atoi(os.Args[1]); err == nil {
			numUsers = n
		}
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Failed to close database: %v", err)
		}
	}()

	if err := db.Ping(); err != nil {
		log.Fatal("Failed to connect:", err)
	}

	fmt.Printf("Starting seed with %d users...\n", numUsers)

	// Seed users and scores
	if err := seedData(db, numUsers); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Seeding complete!")
}

func seedData(db *sql.DB, numUsers int) error {
	ctx := context.Background()
	batchSize := 10000
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	names := []string{
		"Alex", "Bailey", "Casey", "Dakota", "Evan", "Finley", "Graham", "Harper",
		"Iris", "Jordan", "Kai", "Logan", "Morgan", "Noah", "Owen", "Parker",
		"Quinn", "Riley", "Sam", "Taylor", "Uma", "Vale", "Waite", "Xavier",
		"Yara", "Zane", "Ace", "Blake", "Cody", "Drew", "Echo", "Fox",
	}

	surnames := []string{
		"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis",
		"Rodriguez", "Martinez", "Hernandez", "Lopez", "Gonzalez", "Wilson", "Anderson",
		"Thomas", "Taylor", "Moore", "Jackson", "Martin", "Lee", "Perez", "Thompson",
	}

	domains := []string{
		"gmail.com", "yahoo.com", "hotmail.com", "outlook.com", "proton.me", "test.local",
	}

	seasons := []string{"global", "season_1", "season_2", "season_3"}
	passwordHash := "$2a$10$dummy_hash_for_load_test"

	for batch := 0; batch*batchSize < numUsers; batch++ {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}

		batchStart := batch * batchSize
		batchEnd := min(batchStart+batchSize, numUsers)

		for i := batchStart; i < batchEnd; i++ {
			// Generate realistic user data
			name := names[rng.Intn(len(names))] + " " + surnames[rng.Intn(len(surnames))]
			email := fmt.Sprintf("player_%d@%s", i, domains[rng.Intn(len(domains))])

			// Insert user
			var userID string
			err := tx.QueryRowContext(ctx,
				`INSERT INTO users (name, email, password_hash) 
				 VALUES ($1, $2, $3) 
				 RETURNING id`,
				name, email, passwordHash,
			).Scan(&userID)
			if err != nil {
				_ = tx.Rollback()
				// Skip duplicate email
				if err.Error() == "pq: duplicate key value violates unique constraint \"users_email_key\"" {
					continue
				}
				return err
			}

			// Insert scores for multiple seasons (realistic distribution)
			for _, season := range seasons {
				// Use normal-like distribution for scores (most players mid-range, few at extremes)
				score := gaussianScore(rng, 0, 500000, 1000000)
				if score < 0 {
					score = 0
				}

				_, err := tx.ExecContext(ctx,
					`INSERT INTO scores (user_id, score, season, timestamp)
					 VALUES ($1, $2, $3, NOW())`,
					userID, score, season,
				)
				if err != nil {
					// Ignore unique constraint violations
					if err.Error() != "pq: duplicate key value violates unique constraint \"unique_user_season\"" {
						_ = tx.Rollback()
						return err
					}
				}
			}
		}

		if err := tx.Commit(); err != nil {
			return err
		}

		progress := min(batchEnd, numUsers)
		fmt.Printf("Progress: %d/%d users inserted (%.1f%%)\n", progress, numUsers, float64(progress)*100/float64(numUsers))
	}

	return nil
}

// gaussianScore generates score with normal distribution
// min: minimum score, mean: center of distribution, max: maximum possible
func gaussianScore(rng *rand.Rand, min, mean, max float64) int64 {
	// Normal distribution with std dev = range/4
	stdDev := (max - min) / 4
	value := rng.NormFloat64()*stdDev + mean

	if value < min {
		value = min
	}
	if value > max {
		value = max
	}

	return int64(math.Round(value))
}
