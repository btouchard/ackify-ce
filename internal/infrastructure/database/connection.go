package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// Config holds database configuration
type Config struct {
	DSN string
}

// InitDB initializes the database connection and runs migrations
func InitDB(ctx context.Context, config Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", config.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection with timeout
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Run migrations
	if err := runMigrations(ctx, db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

// runMigrations creates the necessary tables
func runMigrations(ctx context.Context, db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS signatures (
			id BIGSERIAL PRIMARY KEY,
			doc_id TEXT NOT NULL,
			user_sub TEXT NOT NULL,
			user_email TEXT NOT NULL,
			user_name TEXT,
			signed_at TIMESTAMPTZ NOT NULL,
			payload_hash TEXT NOT NULL,
			signature TEXT NOT NULL,
			nonce TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			referer TEXT,
			UNIQUE (doc_id, user_sub)
		);
		
		-- Migration: Add prev_hash column if it doesn't exist
		ALTER TABLE signatures ADD COLUMN IF NOT EXISTS prev_hash TEXT;
		
		CREATE INDEX IF NOT EXISTS idx_signatures_user ON signatures(user_sub);
		
		CREATE OR REPLACE FUNCTION prevent_created_at_update()
		RETURNS TRIGGER AS $$
		BEGIN
			IF OLD.created_at IS DISTINCT FROM NEW.created_at THEN
				RAISE EXCEPTION 'Cannot modify created_at timestamp';
			END IF;
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;
		
		DROP TRIGGER IF EXISTS trigger_prevent_created_at_update ON signatures;
		CREATE TRIGGER trigger_prevent_created_at_update
			BEFORE UPDATE ON signatures
			FOR EACH ROW
			EXECUTE FUNCTION prevent_created_at_update();
	`

	_, err := db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to execute migrations: %w", err)
	}

	return nil
}
