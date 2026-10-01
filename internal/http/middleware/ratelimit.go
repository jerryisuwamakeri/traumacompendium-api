package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	apphttp "github.com/traumacompendium/backend/internal/http"
	"github.com/traumacompendium/backend/internal/pkg/apperror"
)

// RateLimit enforces a fixed-window request limit per client IP, backed
// by Redis so the limit holds across multiple API instances rather than
// resetting per-process. prefix namespaces the Redis keys (and therefore
// the budget) so, e.g., the auth endpoints' tighter limit doesn't share a
// bucket with general API traffic.
//
// On any Redis error the request is allowed through rather than
// rejected: an unavailable rate limiter should degrade to "no limiting",
// not take the whole API down.
func RateLimit(client *redis.Client, prefix string, limit int, window time.Duration, log zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := fmt.Sprintf("ratelimit:%s:%s", prefix, c.ClientIP())

		count, err := client.Incr(c.Request.Context(), key).Result()
		if err != nil {
			log.Warn().Err(err).Str("key", key).Msg("ratelimit: redis unavailable, allowing request")
			c.Next()
			return
		}
		if count == 1 {
			if err := client.Expire(c.Request.Context(), key, window).Err(); err != nil {
				log.Warn().Err(err).Str("key", key).Msg("ratelimit: failed to set expiry")
			}
		}

		if count > int64(limit) {
			ttl, _ := client.TTL(c.Request.Context(), key).Result()
			c.Writer.Header().Set("Retry-After", fmt.Sprintf("%.0f", ttl.Seconds()))
			apphttp.Fail(c, apperror.NewTooManyRequests("too many requests; please slow down"))
			c.Abort()
			return
		}

		c.Next()
	}
}
