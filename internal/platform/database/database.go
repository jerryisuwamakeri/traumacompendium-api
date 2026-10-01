// Package database owns the MySQL connection pool and the lightweight
// schema migrator used at startup.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql" // registers the "mysql" driver
	"github.com/rs/zerolog"

	"github.com/traumacompendium/backend/internal/config"
)

// Open establishes the MySQL connection pool and blocks (with backoff)
// until the database is reachable or the context is done. Waiting here
// instead of failing immediately matters in containerized environments
// (docker-compose, Kubernetes) where the API container can start before
// the database is ready to accept connections.
func Open(ctx context.Context, cfg config.DatabaseConfig, log zerolog.Logger) (*sql.DB, error) {
	db, err := sql.Open("mysql", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("database: open: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	const (
		maxAttempts = 10
		baseDelay   = 500 * time.Millisecond
	)

	var pingErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		pingErr = db.PingContext(ctx)
		if pingErr == nil {
			return db, nil
		}

		log.Warn().
			Err(pingErr).
			Int("attempt", attempt).
			Int("max_attempts", maxAttempts).
			Msg("database: ping failed, retrying")

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("database: ping: %w", ctx.Err())
		case <-time.After(baseDelay * time.Duration(attempt)):
		}
	}

	return nil, fmt.Errorf("database: unreachable after %d attempts: %w", maxAttempts, pingErr)
}
