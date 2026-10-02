package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	apphttp "github.com/traumacompendium/backend/internal/http"
	"github.com/traumacompendium/backend/internal/pkg/apperror"
)

const (
	idempotencyKeyHeader        = "Idempotency-Key"
	idempotencyProcessingMarker = "processing"
	idempotencyLockTTL          = 30 * time.Second
)

// cachedIdempotentResponse is what gets stored in Redis once a request
// finishes: enough to replay the exact response a retry would have
// produced, without re-running the handler.
type cachedIdempotentResponse struct {
	Status int             `json:"status"`
	Body   json.RawMessage `json:"body"`
}

// Idempotency makes a POST handler safe to retry: a client that resends
// the same request (same Idempotency-Key header, same route, same
// caller) after a timeout gets back the original response instead of
// creating a duplicate resource. It only activates when the header is
// present, so it is purely opt-in and fully backward compatible with
// clients that don't send one.
//
// Every response on this API is a single JSON envelope written via
// internal/http's OK/Created/Fail helpers (see response.go), so caching
// status+body is enough to byte-for-byte replay a prior response —
// there are no other per-handler headers to capture.
//
// On any Redis error the request is allowed through unprotected rather
// than rejected, the same failure posture as RateLimit: an unavailable
// idempotency store should degrade to "no replay protection", not take
// the API down.
func Idempotency(client *redis.Client, ttl time.Duration, log zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader(idempotencyKeyHeader)
		if key == "" {
			c.Next()
			return
		}

		ctx := c.Request.Context()
		redisKey := fmt.Sprintf("idempotency:%s:%s:%s:%s", idempotencyScope(c), c.FullPath(), c.Request.Method, key)

		claimed, err := client.SetNX(ctx, redisKey, idempotencyProcessingMarker, idempotencyLockTTL).Result()
		if err != nil {
			log.Warn().Err(err).Str("key", redisKey).Msg("idempotency: redis unavailable, allowing request unprotected")
			c.Next()
			return
		}

		if !claimed {
			replayCachedOrReject(c, client, log, redisKey)
			return
		}

		rec := &responseRecorder{ResponseWriter: c.Writer, body: &bytes.Buffer{}}
		c.Writer = rec
		c.Next()

		finalizeIdempotencyRecord(c, client, log, redisKey, rec, ttl)
	}
}

// idempotencyScope identifies the caller the key belongs to, so two
// different users (or two anonymous clients, for an endpoint like
// register that runs before RequireAuth) can't collide on the same key.
func idempotencyScope(c *gin.Context) string {
	if userID := UserID(c); userID != "" {
		return "user:" + userID
	}
	return "ip:" + c.ClientIP()
}

// replayCachedOrReject handles a key that's already claimed: a finished
// request replays its cached response; a request still in flight gets a
// 409 rather than running the handler a second time concurrently.
func replayCachedOrReject(c *gin.Context, client *redis.Client, log zerolog.Logger, redisKey string) {
	ctx := c.Request.Context()

	existing, err := client.Get(ctx, redisKey).Result()
	if err != nil {
		log.Warn().Err(err).Str("key", redisKey).Msg("idempotency: redis unavailable reading existing entry, allowing request unprotected")
		c.Next()
		return
	}

	if existing == idempotencyProcessingMarker {
		apphttp.Fail(c, apperror.NewConflict("a request with this idempotency key is still being processed"))
		c.Abort()
		return
	}

	var cached cachedIdempotentResponse
	if err := json.Unmarshal([]byte(existing), &cached); err != nil {
		log.Warn().Err(err).Str("key", redisKey).Msg("idempotency: failed to decode cached response, allowing request unprotected")
		c.Next()
		return
	}

	c.Data(cached.Status, "application/json; charset=utf-8", cached.Body)
	c.Abort()
}

// finalizeIdempotencyRecord runs after the handler: a successful (2xx)
// response is cached so a retry can replay it; anything else releases
// the lock instead, since failures aren't remembered — the client can
// fix its input and retry with the same key.
func finalizeIdempotencyRecord(c *gin.Context, client *redis.Client, log zerolog.Logger, redisKey string, rec *responseRecorder, ttl time.Duration) {
	ctx := c.Request.Context()

	if rec.status < http.StatusOK || rec.status >= http.StatusMultipleChoices {
		if err := client.Del(ctx, redisKey).Err(); err != nil {
			log.Warn().Err(err).Str("key", redisKey).Msg("idempotency: failed to release lock after non-2xx response")
		}
		return
	}

	payload, err := json.Marshal(cachedIdempotentResponse{Status: rec.status, Body: rec.body.Bytes()})
	if err != nil {
		log.Warn().Err(err).Str("key", redisKey).Msg("idempotency: failed to encode response for caching")
		_ = client.Del(ctx, redisKey).Err()
		return
	}

	if err := client.Set(ctx, redisKey, payload, ttl).Err(); err != nil {
		log.Warn().Err(err).Str("key", redisKey).Msg("idempotency: failed to cache response")
	}
}

// responseRecorder captures the status code and body gin writes so
// Idempotency can cache the exact response bytes for a later replay,
// while still streaming the original response to the client unchanged.
type responseRecorder struct {
	gin.ResponseWriter
	body   *bytes.Buffer
	status int
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}
