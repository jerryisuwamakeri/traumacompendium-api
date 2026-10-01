// Package cache owns the Redis client used for rate limiting, OTP attempt
// counters, and refresh-token revocation (denylist).
package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"github.com/traumacompendium/backend/internal/config"
)

// Open connects to Redis and blocks (with backoff) until it responds to
// PING or the context is done — mirrors database.Open's startup-ordering
// tolerance for containerized environments.
func Open(ctx context.Context, cfg config.RedisConfig, log zerolog.Logger) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	const (
		maxAttempts = 10
		baseDelay   = 500 * time.Millisecond
	)

	var pingErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		pingErr = client.Ping(ctx).Err()
		if pingErr == nil {
			return client, nil
		}

		log.Warn().
			Err(pingErr).
			Int("attempt", attempt).
			Int("max_attempts", maxAttempts).
			Msg("cache: redis ping failed, retrying")

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("cache: ping: %w", ctx.Err())
		case <-time.After(baseDelay * time.Duration(attempt)):
		}
	}

	return nil, fmt.Errorf("cache: redis unreachable after %d attempts: %w", maxAttempts, pingErr)
}
