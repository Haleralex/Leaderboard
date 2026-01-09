package database

import (
	"context"
	"fmt"
	"time"

	"leaderboard-service/internal/shared/config"

	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// PostgresDB wraps the GORM database connection
type PostgresDB struct {
	DB *gorm.DB
}

// NewPostgresDB creates a new PostgreSQL connection using GORM
func NewPostgresDB(cfg *config.Config) (*PostgresDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Configure GORM logger
	gormLogger := logger.Default.LogMode(logger.Info)
	if cfg.Server.Env == "production" {
		gormLogger = logger.Default.LogMode(logger.Error)
	}

	// Open GORM connection
	db, err := gorm.Open(postgres.Open(cfg.Database.URL), &gorm.Config{
		Logger:                 gormLogger,
		SkipDefaultTransaction: true, // Better performance
		PrepareStmt:            true, // Prepare statements for reuse
	})
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	// Get underlying SQL DB for connection pool settings
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("unable to get underlying sql.DB: %w", err)
	}

	// Set connection pool limits
	sqlDB.SetMaxOpenConns(cfg.Database.MaxConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MinConns)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(30 * time.Minute)

	// Test connection
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	// NOTE: Skip auto-migration - using existing schema.sql
	// Schema must be applied manually via psql < sql/schema.sql

	log.Info().Msg("PostgreSQL connection established (GORM)")

	return &PostgresDB{DB: db}, nil
}

// Close closes the database connection
func (db *PostgresDB) Close() error {
	if db.DB != nil {
		sqlDB, err := db.DB.DB()
		if err != nil {
			return err
		}
		if err := sqlDB.Close(); err != nil {
			return err
		}
		log.Info().Msg("PostgreSQL connection closed")
	}
	return nil
}

// Health checks the database connection
func (db *PostgresDB) Health(ctx context.Context) error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return sqlDB.PingContext(ctx)
}
