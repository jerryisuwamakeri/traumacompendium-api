package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func TestIdempotency_ReplaysCachedResponse_WithoutRerunningHandler(t *testing.T) {
	client := openTestRedis(t)
	key := "replay-" + time.Now().Format("150405.000000000")

	var calls int32
	r := gin.New()
	r.Use(Idempotency(client, time.Minute, zerolog.Nop()))
	r.POST("/thing", func(c *gin.Context) {
		atomic.AddInt32(&calls, 1)
		c.JSON(http.StatusCreated, gin.H{"id": "abc123"})
	})

	newReq := func() *http.Request {
		req := httptest.NewRequest(http.MethodPost, "/thing", nil)
		req.Header.Set(idempotencyKeyHeader, key)
		return req
	}

	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, newReq())
	if w1.Code != http.StatusCreated {
		t.Fatalf("first request: status = %d, want %d", w1.Code, http.StatusCreated)
	}

	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, newReq())
	if w2.Code != http.StatusCreated {
		t.Errorf("replayed request: status = %d, want %d", w2.Code, http.StatusCreated)
	}
	if w2.Body.String() != w1.Body.String() {
		t.Errorf("replayed body = %q, want %q", w2.Body.String(), w1.Body.String())
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("handler called %d times, want 1 (the replay must not re-run it)", got)
	}
}

func TestIdempotency_ConcurrentDuplicate_Returns409(t *testing.T) {
	client := openTestRedis(t)
	key := "concurrent-" + time.Now().Format("150405.000000000")

	started := make(chan struct{})
	proceed := make(chan struct{})

	r := gin.New()
	r.Use(Idempotency(client, time.Minute, zerolog.Nop()))
	r.POST("/thing", func(c *gin.Context) {
		close(started)
		<-proceed
		c.JSON(http.StatusCreated, gin.H{"ok": true})
	})

	done := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/thing", nil)
		req.Header.Set(idempotencyKeyHeader, key)
		r.ServeHTTP(w, req)
		done <- w
	}()

	<-started // the first request has claimed the lock and is now blocked inside the handler

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/thing", nil)
	req2.Header.Set(idempotencyKeyHeader, key)
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusConflict {
		t.Errorf("concurrent duplicate: status = %d, want %d", w2.Code, http.StatusConflict)
	}

	close(proceed)
	w1 := <-done
	if w1.Code != http.StatusCreated {
		t.Errorf("original in-flight request: status = %d, want %d", w1.Code, http.StatusCreated)
	}
}

func TestIdempotency_FailedResponse_IsNotCached_RetryRerunsHandler(t *testing.T) {
	client := openTestRedis(t)
	key := "fail-" + time.Now().Format("150405.000000000")

	var calls int32
	r := gin.New()
	r.Use(Idempotency(client, time.Minute, zerolog.Nop()))
	r.POST("/thing", func(c *gin.Context) {
		if atomic.AddInt32(&calls, 1) == 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad input"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"ok": true})
	})

	newReq := func() *http.Request {
		req := httptest.NewRequest(http.MethodPost, "/thing", nil)
		req.Header.Set(idempotencyKeyHeader, key)
		return req
	}

	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, newReq())
	if w1.Code != http.StatusBadRequest {
		t.Fatalf("first request: status = %d, want %d", w1.Code, http.StatusBadRequest)
	}

	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, newReq())
	if w2.Code != http.StatusCreated {
		t.Errorf("retry with same key after a failure: status = %d, want %d (handler should run again)", w2.Code, http.StatusCreated)
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Errorf("handler called %d times, want 2", got)
	}
}

func TestIdempotency_NoHeader_PassesThroughEveryTime(t *testing.T) {
	client := openTestRedis(t)

	var calls int32
	r := gin.New()
	r.Use(Idempotency(client, time.Minute, zerolog.Nop()))
	r.POST("/thing", func(c *gin.Context) {
		atomic.AddInt32(&calls, 1)
		c.JSON(http.StatusCreated, gin.H{"ok": true})
	})

	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/thing", nil))
		if w.Code != http.StatusCreated {
			t.Fatalf("request %d: status = %d, want %d", i, w.Code, http.StatusCreated)
		}
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Errorf("handler called %d times, want 2 (no key means no replay protection)", got)
	}
}

func TestIdempotency_DifferentUsers_SameKey_DontCollide(t *testing.T) {
	client := openTestRedis(t)
	key := "scope-" + time.Now().Format("150405.000000000")

	var calls int32
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(contextKeyUserID, c.GetHeader("X-Test-User"))
		c.Next()
	})
	r.Use(Idempotency(client, time.Minute, zerolog.Nop()))
	r.POST("/thing", func(c *gin.Context) {
		atomic.AddInt32(&calls, 1)
		c.JSON(http.StatusCreated, gin.H{"ok": true})
	})

	for _, user := range []string{"user-a", "user-b"} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/thing", nil)
		req.Header.Set(idempotencyKeyHeader, key)
		req.Header.Set("X-Test-User", user)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("user %q: status = %d, want %d", user, w.Code, http.StatusCreated)
		}
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Errorf("handler called %d times, want 2 (different users must not share an idempotency key)", got)
	}
}
