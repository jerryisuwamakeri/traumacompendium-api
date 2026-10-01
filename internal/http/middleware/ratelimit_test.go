package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

// openTestRedis connects to a real Redis instance for rate-limit tests. It
// is skipped if no Redis is reachable at REDIS_ADDR (default
// 127.0.0.1:6379) so `go test ./...` still passes in an environment with
// no Redis running.
func openTestRedis(t *testing.T) *redis.Client {
	t.Helper()

	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "127.0.0.1:6379"
	}

	client := redis.NewClient(&redis.Options{Addr: addr})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("no redis reachable at %s; skipping rate-limit integration test: %v", addr, err)
	}

	t.Cleanup(func() { _ = client.Close() })
	return client
}

func newRateLimitTestRouter(client *redis.Client, prefix string, limit int, window time.Duration) *gin.Engine {
	r := gin.New()
	r.Use(RateLimit(client, prefix, limit, window, zerolog.Nop()))
	r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })
	return r
}

func TestRateLimit_AllowsUpToLimit_ThenRejects(t *testing.T) {
	client := openTestRedis(t)
	// Unique prefix per test run so repeated test executions don't share
	// a counter left over (with a TTL) from a previous run.
	prefix := "test-" + time.Now().Format("150405.000000000")

	r := newRateLimitTestRouter(client, prefix, 3, time.Minute)

	for i := 1; i <= 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("request %d: status = %d, want %d", i, w.Code, http.StatusOK)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("4th request: status = %d, want %d", w.Code, http.StatusTooManyRequests)
	}
	if w.Header().Get("Retry-After") == "" {
		t.Error("expected a Retry-After header on a rate-limited response")
	}
}

func TestRateLimit_DifferentPrefixes_HaveIndependentBudgets(t *testing.T) {
	client := openTestRedis(t)
	suffix := time.Now().Format("150405.000000000")

	general := newRateLimitTestRouter(client, "general-"+suffix, 1, time.Minute)
	auth := newRateLimitTestRouter(client, "auth-"+suffix, 1, time.Minute)

	// Exhaust the "general" budget for this IP.
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	general.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("general request 1: status = %d, want %d", w.Code, http.StatusOK)
	}
	w = httptest.NewRecorder()
	general.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/ping", nil))
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("general request 2: status = %d, want %d", w.Code, http.StatusTooManyRequests)
	}

	// The "auth" limiter for the same client must be unaffected.
	w = httptest.NewRecorder()
	auth.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/ping", nil))
	if w.Code != http.StatusOK {
		t.Errorf("auth request 1 (independent budget): status = %d, want %d", w.Code, http.StatusOK)
	}
}
