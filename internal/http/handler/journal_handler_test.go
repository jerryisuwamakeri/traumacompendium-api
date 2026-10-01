package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/traumacompendium/backend/internal/config"
	"github.com/traumacompendium/backend/internal/http/middleware"
	"github.com/traumacompendium/backend/internal/pkg/jwtutil"
	"github.com/traumacompendium/backend/internal/service/journalservice"
	"github.com/traumacompendium/backend/internal/testsupport/fakes"
)

// newTestJournalRouter wires the real RequireAuth middleware (not a
// stand-in), so these tests exercise exactly the same auth path
// production traffic goes through — the handler reads the user ID via
// middleware.UserID, populated from a genuine, verified access token.
func newTestJournalRouter() (router *gin.Engine, jwt *jwtutil.Manager) {
	jwt = jwtutil.NewManager(config.JWTConfig{
		AccessSecret:  "test-access-secret",
		RefreshSecret: "test-refresh-secret",
		AccessTTL:     15 * time.Minute,
		RefreshTTL:    30 * 24 * time.Hour,
		Issuer:        "traumacompendium-test",
	})
	h := NewJournalHandler(journalservice.New(fakes.NewJournalRepo()))

	r := gin.New()
	group := r.Group("/journal")
	group.Use(middleware.RequireAuth(jwt))
	group.POST("/entries", h.Create)
	group.GET("/entries", h.List)
	group.GET("/entries/:id", h.Get)
	group.PATCH("/entries/:id", h.Update)
	group.DELETE("/entries/:id", h.Delete)
	group.GET("/export", h.Export)

	return r, jwt
}

func authedRequest(t *testing.T, jwt *jwtutil.Manager, userID, method, target string, body []byte) *http.Request {
	t.Helper()
	issued, err := jwt.GenerateAccessToken(userID, "user")
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, target, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, target, nil)
	}
	req.Header.Set("Authorization", "Bearer "+issued.Token)
	return req
}

func TestJournalCreate_And_Get(t *testing.T) {
	r, jwt := newTestJournalRouter()

	req := authedRequest(t, jwt, "user-1", http.MethodPost, "/journal/entries",
		[]byte(`{"title":"A hard day","content":"Today was difficult."}`))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Create status = %d, want %d, body=%s", w.Code, http.StatusCreated, w.Body.String())
	}

	var created struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}

	getReq := authedRequest(t, jwt, "user-1", http.MethodGet, "/journal/entries/"+created.Data.ID, nil)
	getW := httptest.NewRecorder()
	r.ServeHTTP(getW, getReq)

	if getW.Code != http.StatusOK {
		t.Fatalf("Get status = %d, want %d, body=%s", getW.Code, http.StatusOK, getW.Body.String())
	}
}

func TestJournalGet_AnotherUsersEntry_404(t *testing.T) {
	r, jwt := newTestJournalRouter()

	createReq := authedRequest(t, jwt, "owner", http.MethodPost, "/journal/entries",
		[]byte(`{"title":"Private","content":"..."}`))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, createReq)

	var created struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &created)

	getReq := authedRequest(t, jwt, "intruder", http.MethodGet, "/journal/entries/"+created.Data.ID, nil)
	getW := httptest.NewRecorder()
	r.ServeHTTP(getW, getReq)

	if getW.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", getW.Code, http.StatusNotFound)
	}
}

func TestJournalList_Pagination(t *testing.T) {
	r, jwt := newTestJournalRouter()

	for i := 0; i < 3; i++ {
		req := authedRequest(t, jwt, "user-1", http.MethodPost, "/journal/entries",
			[]byte(`{"title":"Entry","content":"..."}`))
		r.ServeHTTP(httptest.NewRecorder(), req)
	}

	req := authedRequest(t, jwt, "user-1", http.MethodGet, "/journal/entries?limit=2", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp struct {
		Data []json.RawMessage `json:"data"`
		Meta struct {
			Total int `json:"total"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Errorf("len(Data) = %d, want 2", len(resp.Data))
	}
	if resp.Meta.Total != 3 {
		t.Errorf("Meta.Total = %d, want 3", resp.Meta.Total)
	}
}

func TestJournalExport_CSV(t *testing.T) {
	r, jwt := newTestJournalRouter()

	createReq := authedRequest(t, jwt, "user-1", http.MethodPost, "/journal/entries",
		[]byte(`{"title":"Entry","content":"Some content"}`))
	r.ServeHTTP(httptest.NewRecorder(), createReq)

	exportReq := authedRequest(t, jwt, "user-1", http.MethodGet, "/journal/export?format=csv", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, exportReq)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if ct := w.Header().Get("Content-Type"); ct != "text/csv" {
		t.Errorf("Content-Type = %q, want text/csv", ct)
	}
	if w.Header().Get("Content-Disposition") == "" {
		t.Error("expected a Content-Disposition header on the export response")
	}
}

func TestJournalRoutes_RejectUnauthenticated(t *testing.T) {
	r, _ := newTestJournalRouter()

	req := httptest.NewRequest(http.MethodGet, "/journal/entries", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}
