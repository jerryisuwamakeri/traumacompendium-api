package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/traumacompendium/backend/internal/config"
	"github.com/traumacompendium/backend/internal/http/middleware"
	"github.com/traumacompendium/backend/internal/pkg/jwtutil"
	"github.com/traumacompendium/backend/internal/service/moodservice"
	"github.com/traumacompendium/backend/internal/testsupport/fakes"
)

func newTestMoodRouter() (*gin.Engine, *jwtutil.Manager) {
	jwt := jwtutil.NewManager(config.JWTConfig{
		AccessSecret:  "test-access-secret",
		RefreshSecret: "test-refresh-secret",
		AccessTTL:     15 * time.Minute,
		RefreshTTL:    30 * 24 * time.Hour,
		Issuer:        "traumacompendium-test",
	})
	h := NewMoodHandler(moodservice.New(fakes.NewMoodRepo()))

	r := gin.New()
	group := r.Group("/mood")
	group.Use(middleware.RequireAuth(jwt))
	group.POST("/entries", h.Create)
	group.GET("/entries", h.List)
	group.GET("/entries/:id", h.Get)
	group.DELETE("/entries/:id", h.Delete)
	group.GET("/summary", h.Summary)

	return r, jwt
}

func TestMoodCreate_And_Get(t *testing.T) {
	r, jwt := newTestMoodRouter()

	req := authedRequest(t, jwt, "user-1", http.MethodPost, "/mood/entries",
		[]byte(`{"score":4,"tags":["work stress"],"note":"rough day"}`))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Create status = %d, want %d, body=%s", w.Code, http.StatusCreated, w.Body.String())
	}

	var created struct {
		Data struct {
			ID    string `json:"id"`
			Score int    `json:"score"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if created.Data.Score != 4 {
		t.Errorf("Score = %d, want 4", created.Data.Score)
	}

	getReq := authedRequest(t, jwt, "user-1", http.MethodGet, "/mood/entries/"+created.Data.ID, nil)
	getW := httptest.NewRecorder()
	r.ServeHTTP(getW, getReq)
	if getW.Code != http.StatusOK {
		t.Fatalf("Get status = %d, want %d", getW.Code, http.StatusOK)
	}
}

func TestMoodCreate_InvalidScore_400(t *testing.T) {
	r, jwt := newTestMoodRouter()

	req := authedRequest(t, jwt, "user-1", http.MethodPost, "/mood/entries", []byte(`{"score":9}`))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d, body=%s", w.Code, http.StatusBadRequest, w.Body.String())
	}
}

func TestMoodGet_AnotherUsersEntry_404(t *testing.T) {
	r, jwt := newTestMoodRouter()

	createReq := authedRequest(t, jwt, "owner", http.MethodPost, "/mood/entries", []byte(`{"score":3}`))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, createReq)

	var created struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &created)

	getReq := authedRequest(t, jwt, "intruder", http.MethodGet, "/mood/entries/"+created.Data.ID, nil)
	getW := httptest.NewRecorder()
	r.ServeHTTP(getW, getReq)

	if getW.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", getW.Code, http.StatusNotFound)
	}
}

func TestMoodSummary_ReturnsAggregatedData(t *testing.T) {
	r, jwt := newTestMoodRouter()

	for _, score := range []int{2, 4} {
		body := []byte(`{"score":` + strconv.Itoa(score) + `,"tags":["work stress"]}`)
		req := authedRequest(t, jwt, "user-1", http.MethodPost, "/mood/entries", body)
		r.ServeHTTP(httptest.NewRecorder(), req)
	}

	req := authedRequest(t, jwt, "user-1", http.MethodGet, "/mood/summary", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp struct {
		Data struct {
			EntryCount   int     `json:"entry_count"`
			AverageScore float64 `json:"average_score"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Data.EntryCount != 2 {
		t.Errorf("EntryCount = %d, want 2", resp.Data.EntryCount)
	}
	if resp.Data.AverageScore != 3 {
		t.Errorf("AverageScore = %v, want 3", resp.Data.AverageScore)
	}
}

func TestMoodRoutes_RejectUnauthenticated(t *testing.T) {
	r, _ := newTestMoodRouter()

	req := httptest.NewRequest(http.MethodGet, "/mood/entries", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}
