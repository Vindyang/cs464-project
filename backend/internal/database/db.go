package database

import (
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Config holds database configuration
type Config struct {
	DatabaseURL string
}

// Connect establishes a connection to the PostgreSQL database
func Connect(cfg Config) (*sqlx.DB, error) {
	// Connect to database
	db, err := sqlx.Connect("pgx", cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// ConnectFromEnv connects using DATABASE_URL environment variable
func ConnectFromEnv() (*sqlx.DB, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is not set")
	}

	return Connect(Config{DatabaseURL: databaseURL})
}
