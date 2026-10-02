package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"github.com/traumacompendium/backend/internal/config"
	"github.com/traumacompendium/backend/internal/http/middleware"
	"github.com/traumacompendium/backend/internal/pkg/jwtutil"
	"github.com/traumacompendium/backend/internal/platform/cache"
	"github.com/traumacompendium/backend/internal/service/safereachservice"
	"github.com/traumacompendium/backend/internal/testsupport/fakes"
)

// openTestRedis connects to a real local Redis instance, skipping the
// test if none is reachable — same pattern used throughout this
// codebase's Redis-touching tests (see internal/http/middleware).
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
		t.Skipf("no redis reachable at %s; skipping safereach stream test: %v", addr, err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func newTestSafeReachRouter(redisClient *redis.Client) (*gin.Engine, *jwtutil.Manager) {
	jwt := jwtutil.NewManager(config.JWTConfig{
		AccessSecret:  "test-access-secret",
		RefreshSecret: "test-refresh-secret",
		AccessTTL:     15 * time.Minute,
		RefreshTTL:    30 * 24 * time.Hour,
		Issuer:        "traumacompendium-test",
	})

	// The fake in-memory Publisher is fine for REST-only tests, but the
	// WebSocket stream tests subscribe on the real redisClient — the
	// service's publish side has to go through that same real Redis, or
	// a published location update never reaches the stream.
	var publisher safereachservice.Publisher = &fakes.Publisher{}
	if redisClient != nil {
		publisher = cache.Publisher{Client: redisClient}
	}

	svc := safereachservice.New(
		fakes.NewSafeReachContactRepo(), fakes.NewSafeReachAlertRepo(), fakes.NewDeviceTokenRepo(),
		&fakes.SMSSender{}, &fakes.EmailSender{}, &fakes.PushSender{}, publisher,
		"https://app.test", zerolog.Nop(),
	)
	h := NewSafeReachHandler(svc, redisClient, []string{"*"}, zerolog.Nop())

	r := gin.New()
	group := r.Group("/safereach")
	group.GET("/public/alerts/:id", h.PublicView)
	group.Use(middleware.RequireAuth(jwt))
	group.POST("/contacts", h.RegisterContact)
	group.GET("/contacts", h.ListContacts)
	group.DELETE("/contacts/:id", h.DeleteContact)
	group.POST("/alerts", h.TriggerAlert)
	group.GET("/alerts", h.ListAlerts)
	group.GET("/alerts/:id", h.GetAlert)
	group.POST("/alerts/:id/cancel", h.CancelAlert)
	group.POST("/alerts/:id/resolve", h.ResolveAlert)
	group.POST("/alerts/:id/location", h.RecordLocation)
	group.GET("/alerts/:id/stream", h.Stream)
	group.POST("/device-tokens", h.RegisterDeviceToken)

	return r, jwt
}

type dataEnvelope[T any] struct {
	Data T `json:"data"`
}

// accessToken issues a raw access token string for userID — used by the
// WebSocket tests, which need the bare token for a Dial() header rather
// than a full *http.Request the way authedRequest provides.
func accessToken(t *testing.T, jwt *jwtutil.Manager, userID string) string {
	t.Helper()
	issued, err := jwt.GenerateAccessToken(userID, "user")
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}
	return issued.Token
}

func decodeData[T any](t *testing.T, body []byte) T {
	t.Helper()
	var env dataEnvelope[T]
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("decode response: %v, body=%s", err, body)
	}
	return env.Data
}

func TestSafeReachContacts_RegisterListDelete(t *testing.T) {
	r, jwt := newTestSafeReachRouter(nil)

	createReq := authedRequest(t, jwt, "user-1", http.MethodPost, "/safereach/contacts",
		[]byte(`{"name":"Ada Okafor","relationship":"Sister","phone":"+2348012345678"}`))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, createReq)
	if w.Code != http.StatusCreated {
		t.Fatalf("RegisterContact status = %d, want %d, body=%s", w.Code, http.StatusCreated, w.Body.String())
	}
	created := decodeData[struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}](t, w.Body.Bytes())
	if created.Name != "Ada Okafor" {
		t.Errorf("Name = %q, want Ada Okafor", created.Name)
	}

	listReq := authedRequest(t, jwt, "user-1", http.MethodGet, "/safereach/contacts", nil)
	listW := httptest.NewRecorder()
	r.ServeHTTP(listW, listReq)
	if listW.Code != http.StatusOK {
		t.Fatalf("ListContacts status = %d, want %d", listW.Code, http.StatusOK)
	}

	delReq := authedRequest(t, jwt, "user-1", http.MethodDelete, "/safereach/contacts/"+created.ID, nil)
	delW := httptest.NewRecorder()
	r.ServeHTTP(delW, delReq)
	if delW.Code != http.StatusNoContent {
		t.Fatalf("DeleteContact status = %d, want %d", delW.Code, http.StatusNoContent)
	}
}

func TestSafeReachContacts_MissingNameAndPhone_400(t *testing.T) {
	r, jwt := newTestSafeReachRouter(nil)

	req := authedRequest(t, jwt, "user-1", http.MethodPost, "/safereach/contacts", []byte(`{}`))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d, body=%s", w.Code, http.StatusBadRequest, w.Body.String())
	}
}

func TestSafeReachAlerts_TriggerGetCancel(t *testing.T) {
	r, jwt := newTestSafeReachRouter(nil)

	triggerReq := authedRequest(t, jwt, "user-1", http.MethodPost, "/safereach/alerts",
		[]byte(`{"level":"urgent","message":"Need help"}`))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, triggerReq)
	if w.Code != http.StatusCreated {
		t.Fatalf("TriggerAlert status = %d, want %d, body=%s", w.Code, http.StatusCreated, w.Body.String())
	}
	created := decodeData[struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}](t, w.Body.Bytes())
	if created.Status != "active" {
		t.Errorf("Status = %q, want active", created.Status)
	}

	getReq := authedRequest(t, jwt, "user-1", http.MethodGet, "/safereach/alerts/"+created.ID, nil)
	getW := httptest.NewRecorder()
	r.ServeHTTP(getW, getReq)
	if getW.Code != http.StatusOK {
		t.Fatalf("GetAlert status = %d, want %d", getW.Code, http.StatusOK)
	}

	cancelReq := authedRequest(t, jwt, "user-1", http.MethodPost, "/safereach/alerts/"+created.ID+"/cancel", nil)
	cancelW := httptest.NewRecorder()
	r.ServeHTTP(cancelW, cancelReq)
	if cancelW.Code != http.StatusOK {
		t.Fatalf("CancelAlert status = %d, want %d, body=%s", cancelW.Code, http.StatusOK, cancelW.Body.String())
	}
	cancelled := decodeData[struct {
		Status string `json:"status"`
	}](t, cancelW.Body.Bytes())
	if cancelled.Status != "cancelled" {
		t.Errorf("Status after cancel = %q, want cancelled", cancelled.Status)
	}
}

func TestSafeReachAlerts_TriggerInvalidLevel_400(t *testing.T) {
	r, jwt := newTestSafeReachRouter(nil)

	req := authedRequest(t, jwt, "user-1", http.MethodPost, "/safereach/alerts", []byte(`{"level":"not-a-level"}`))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d, body=%s", w.Code, http.StatusBadRequest, w.Body.String())
	}
}

func TestSafeReachAlerts_GetAnotherUsersAlert_404(t *testing.T) {
	r, jwt := newTestSafeReachRouter(nil)

	triggerReq := authedRequest(t, jwt, "owner", http.MethodPost, "/safereach/alerts", []byte(`{"level":"check_in"}`))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, triggerReq)
	created := decodeData[struct {
		ID string `json:"id"`
	}](t, w.Body.Bytes())

	getReq := authedRequest(t, jwt, "intruder", http.MethodGet, "/safereach/alerts/"+created.ID, nil)
	getW := httptest.NewRecorder()
	r.ServeHTTP(getW, getReq)
	if getW.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", getW.Code, http.StatusNotFound)
	}
}

func TestSafeReachAlerts_RecordLocationOnClosedAlert_422(t *testing.T) {
	r, jwt := newTestSafeReachRouter(nil)

	triggerReq := authedRequest(t, jwt, "user-1", http.MethodPost, "/safereach/alerts", []byte(`{"level":"check_in"}`))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, triggerReq)
	created := decodeData[struct {
		ID string `json:"id"`
	}](t, w.Body.Bytes())

	resolveReq := authedRequest(t, jwt, "user-1", http.MethodPost, "/safereach/alerts/"+created.ID+"/resolve", nil)
	r.ServeHTTP(httptest.NewRecorder(), resolveReq)

	locReq := authedRequest(t, jwt, "user-1", http.MethodPost, "/safereach/alerts/"+created.ID+"/location", []byte(`{"lat":6.5,"lng":3.4}`))
	locW := httptest.NewRecorder()
	r.ServeHTTP(locW, locReq)
	if locW.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want %d, body=%s", locW.Code, http.StatusUnprocessableEntity, locW.Body.String())
	}
}

func TestSafeReachPublicView_WrongToken_404(t *testing.T) {
	r, jwt := newTestSafeReachRouter(nil)

	triggerReq := authedRequest(t, jwt, "user-1", http.MethodPost, "/safereach/alerts", []byte(`{"level":"urgent"}`))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, triggerReq)
	created := decodeData[struct {
		ID string `json:"id"`
	}](t, w.Body.Bytes())

	viewReq := httptest.NewRequest(http.MethodGet, "/safereach/public/alerts/"+created.ID+"?token=wrong", nil)
	viewW := httptest.NewRecorder()
	r.ServeHTTP(viewW, viewReq)
	if viewW.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", viewW.Code, http.StatusNotFound)
	}
}

func TestSafeReachPublicView_NoAuthRequired(t *testing.T) {
	r, _ := newTestSafeReachRouter(nil)

	// No Authorization header at all — the public view route must not
	// be behind RequireAuth. A missing/wrong token still 404s, but that
	// must come from GetPublicView's own check, not from the auth
	// middleware rejecting the request with 401 first.
	req := httptest.NewRequest(http.MethodGet, "/safereach/public/alerts/does-not-exist?token=x", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusUnauthorized {
		t.Error("public view route rejected an unauthenticated request with 401 — it must not require auth")
	}
}

func TestSafeReachRoutes_RejectUnauthenticated(t *testing.T) {
	r, _ := newTestSafeReachRouter(nil)

	req := httptest.NewRequest(http.MethodGet, "/safereach/alerts", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestSafeReachDeviceTokens_Register(t *testing.T) {
	r, jwt := newTestSafeReachRouter(nil)

	req := authedRequest(t, jwt, "user-1", http.MethodPost, "/safereach/device-tokens", []byte(`{"token":"fcm-abc","platform":"android"}`))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d, body=%s", w.Code, http.StatusNoContent, w.Body.String())
	}
}

func TestSafeReachStream_WrongOwner_RejectedBeforeUpgrade(t *testing.T) {
	redisClient := openTestRedis(t)
	r, jwt := newTestSafeReachRouter(redisClient)
	server := httptest.NewServer(r)
	defer server.Close()

	triggerReq := authedRequest(t, jwt, "owner", http.MethodPost, "/safereach/alerts", []byte(`{"level":"critical"}`))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, triggerReq)
	created := decodeData[struct {
		ID string `json:"id"`
	}](t, w.Body.Bytes())

	token := accessToken(t, jwt, "intruder")
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/safereach/alerts/" + created.ID + "/stream"
	headers := http.Header{"Authorization": {"Bearer " + token}}

	_, resp, err := websocket.DefaultDialer.Dial(wsURL, headers)
	if resp != nil {
		defer resp.Body.Close() //nolint:errcheck // test cleanup
	}
	if err == nil {
		t.Fatal("Dial() by a non-owner succeeded, want it rejected before upgrade")
	}
	if resp == nil || resp.StatusCode != http.StatusNotFound {
		status := -1
		if resp != nil {
			status = resp.StatusCode
		}
		t.Errorf("response status = %d, want %d", status, http.StatusNotFound)
	}
}

func TestSafeReachStream_OwnerReceivesPublishedLocation(t *testing.T) {
	redisClient := openTestRedis(t)
	r, jwt := newTestSafeReachRouter(redisClient)
	server := httptest.NewServer(r)
	defer server.Close()

	triggerReq := authedRequest(t, jwt, "owner", http.MethodPost, "/safereach/alerts", []byte(`{"level":"critical"}`))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, triggerReq)
	created := decodeData[struct {
		ID string `json:"id"`
	}](t, w.Body.Bytes())

	token := accessToken(t, jwt, "owner")
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/safereach/alerts/" + created.ID + "/stream"
	headers := http.Header{"Authorization": {"Bearer " + token}}

	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, headers)
	if resp != nil {
		defer resp.Body.Close() //nolint:errcheck // test cleanup
	}
	if err != nil {
		status := -1
		if resp != nil {
			status = resp.StatusCode
		}
		t.Fatalf("Dial() error = %v (status %d)", err, status)
	}
	defer conn.Close() //nolint:errcheck // test cleanup

	locReq := authedRequest(t, jwt, "owner", http.MethodPost, "/safereach/alerts/"+created.ID+"/location", []byte(`{"lat":6.5244,"lng":3.3792}`))
	locW := httptest.NewRecorder()
	r.ServeHTTP(locW, locReq)
	if locW.Code != http.StatusNoContent {
		t.Fatalf("RecordLocation status = %d, want %d, body=%s", locW.Code, http.StatusNoContent, locW.Body.String())
	}

	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, message, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage() error = %v", err)
	}

	var event struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	}
	if err := json.Unmarshal(message, &event); err != nil {
		t.Fatalf("decode stream message: %v, message=%s", err, message)
	}
	if event.Lat != 6.5244 || event.Lng != 3.3792 {
		t.Errorf("event = %+v, want lat=6.5244 lng=3.3792", event)
	}
}
