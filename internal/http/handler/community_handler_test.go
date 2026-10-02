package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/traumacompendium/backend/internal/config"
	"github.com/traumacompendium/backend/internal/http/middleware"
	"github.com/traumacompendium/backend/internal/pkg/jwtutil"
	"github.com/traumacompendium/backend/internal/service/communityservice"
	"github.com/traumacompendium/backend/internal/testsupport/fakes"
)

func newTestCommunityRouter(aiClient *fakes.AIClient) (*gin.Engine, *jwtutil.Manager) {
	jwt := jwtutil.NewManager(config.JWTConfig{
		AccessSecret:  "test-access-secret",
		RefreshSecret: "test-refresh-secret",
		AccessTTL:     15 * time.Minute,
		RefreshTTL:    30 * 24 * time.Hour,
		Issuer:        "traumacompendium-test",
	})
	svc := communityservice.New(
		fakes.NewCommunityPostRepo(), fakes.NewCommunityCommentRepo(), fakes.NewCommunityReportRepo(),
		aiClient, zerolog.Nop(),
	)
	h := NewCommunityHandler(svc)

	r := gin.New()
	community := r.Group("/community")
	community.Use(middleware.RequireAuth(jwt))
	community.POST("/posts", h.CreatePost)
	community.GET("/posts", h.ListFeed)
	community.DELETE("/posts/:id", h.DeletePost)
	community.POST("/posts/:id/comments", h.CreateComment)
	community.GET("/posts/:id/comments", h.ListComments)
	community.DELETE("/comments/:id", h.DeleteComment)
	community.POST("/reports", h.FileReport)

	admin := r.Group("/admin/community")
	admin.Use(middleware.RequireAuth(jwt), middleware.RequireRole("admin"))
	admin.GET("/reports", h.ListOpenReports)
	admin.POST("/reports/:id/resolve", h.ResolveReport)

	return r, jwt
}

func adminRequest(t *testing.T, jwt *jwtutil.Manager, userID, method, target string, body []byte) *http.Request {
	t.Helper()
	issued, err := jwt.GenerateAccessToken(userID, "admin")
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

func TestCommunityPosts_CreateListDelete(t *testing.T) {
	r, jwt := newTestCommunityRouter(fakes.NewAIClient())

	createReq := authedRequest(t, jwt, "user-1", http.MethodPost, "/community/posts", []byte(`{"body":"Hang in there."}`))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, createReq)
	if w.Code != http.StatusCreated {
		t.Fatalf("CreatePost status = %d, want %d, body=%s", w.Code, http.StatusCreated, w.Body.String())
	}
	created := decodeData[struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}](t, w.Body.Bytes())
	if created.Status != "visible" {
		t.Errorf("Status = %q, want visible", created.Status)
	}

	listReq := authedRequest(t, jwt, "user-2", http.MethodGet, "/community/posts", nil)
	listW := httptest.NewRecorder()
	r.ServeHTTP(listW, listReq)
	if listW.Code != http.StatusOK {
		t.Fatalf("ListFeed status = %d, want %d", listW.Code, http.StatusOK)
	}

	delReq := authedRequest(t, jwt, "user-1", http.MethodDelete, "/community/posts/"+created.ID, nil)
	delW := httptest.NewRecorder()
	r.ServeHTTP(delW, delReq)
	if delW.Code != http.StatusNoContent {
		t.Fatalf("DeletePost status = %d, want %d", delW.Code, http.StatusNoContent)
	}
}

func TestCommunityPosts_EmptyBody_400(t *testing.T) {
	r, jwt := newTestCommunityRouter(fakes.NewAIClient())

	req := authedRequest(t, jwt, "user-1", http.MethodPost, "/community/posts", []byte(`{"body":""}`))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d, body=%s", w.Code, http.StatusBadRequest, w.Body.String())
	}
}

func TestCommunityPosts_DeleteAnotherUsersPost_404(t *testing.T) {
	r, jwt := newTestCommunityRouter(fakes.NewAIClient())

	createReq := authedRequest(t, jwt, "owner", http.MethodPost, "/community/posts", []byte(`{"body":"my post"}`))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, createReq)
	created := decodeData[struct {
		ID string `json:"id"`
	}](t, w.Body.Bytes())

	delReq := authedRequest(t, jwt, "intruder", http.MethodDelete, "/community/posts/"+created.ID, nil)
	delW := httptest.NewRecorder()
	r.ServeHTTP(delW, delReq)
	if delW.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", delW.Code, http.StatusNotFound)
	}
}

func TestCommunityComments_CreateAndList(t *testing.T) {
	r, jwt := newTestCommunityRouter(fakes.NewAIClient())

	postReq := authedRequest(t, jwt, "author", http.MethodPost, "/community/posts", []byte(`{"body":"original post"}`))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, postReq)
	post := decodeData[struct {
		ID string `json:"id"`
	}](t, w.Body.Bytes())

	commentReq := authedRequest(t, jwt, "commenter", http.MethodPost, "/community/posts/"+post.ID+"/comments", []byte(`{"body":"Sending strength"}`))
	commentW := httptest.NewRecorder()
	r.ServeHTTP(commentW, commentReq)
	if commentW.Code != http.StatusCreated {
		t.Fatalf("CreateComment status = %d, want %d, body=%s", commentW.Code, http.StatusCreated, commentW.Body.String())
	}

	listReq := authedRequest(t, jwt, "author", http.MethodGet, "/community/posts/"+post.ID+"/comments", nil)
	listW := httptest.NewRecorder()
	r.ServeHTTP(listW, listReq)
	if listW.Code != http.StatusOK {
		t.Fatalf("ListComments status = %d, want %d", listW.Code, http.StatusOK)
	}
}

func TestCommunityReports_FileAndAdminResolve(t *testing.T) {
	r, jwt := newTestCommunityRouter(fakes.NewAIClient())

	postReq := authedRequest(t, jwt, "author", http.MethodPost, "/community/posts", []byte(`{"body":"original post"}`))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, postReq)
	post := decodeData[struct {
		ID string `json:"id"`
	}](t, w.Body.Bytes())

	reportReq := authedRequest(t, jwt, "reporter", http.MethodPost, "/community/reports",
		[]byte(`{"target_type":"post","target_id":"`+post.ID+`","reason":"concerning"}`))
	reportW := httptest.NewRecorder()
	r.ServeHTTP(reportW, reportReq)
	if reportW.Code != http.StatusCreated {
		t.Fatalf("FileReport status = %d, want %d, body=%s", reportW.Code, http.StatusCreated, reportW.Body.String())
	}
	report := decodeData[struct {
		ID string `json:"id"`
	}](t, reportW.Body.Bytes())

	// A plain user must not reach the admin queue.
	forbiddenReq := authedRequest(t, jwt, "reporter", http.MethodGet, "/admin/community/reports", nil)
	forbiddenW := httptest.NewRecorder()
	r.ServeHTTP(forbiddenW, forbiddenReq)
	if forbiddenW.Code != http.StatusForbidden {
		t.Errorf("non-admin ListOpenReports status = %d, want %d", forbiddenW.Code, http.StatusForbidden)
	}

	listReq := adminRequest(t, jwt, "admin-1", http.MethodGet, "/admin/community/reports", nil)
	listW := httptest.NewRecorder()
	r.ServeHTTP(listW, listReq)
	if listW.Code != http.StatusOK {
		t.Fatalf("admin ListOpenReports status = %d, want %d, body=%s", listW.Code, http.StatusOK, listW.Body.String())
	}

	resolveReq := adminRequest(t, jwt, "admin-1", http.MethodPost, "/admin/community/reports/"+report.ID+"/resolve", []byte(`{"action":"remove"}`))
	resolveW := httptest.NewRecorder()
	r.ServeHTTP(resolveW, resolveReq)
	if resolveW.Code != http.StatusNoContent {
		t.Fatalf("ResolveReport status = %d, want %d, body=%s", resolveW.Code, http.StatusNoContent, resolveW.Body.String())
	}

	// The removed post must no longer be in the feed.
	feedReq := authedRequest(t, jwt, "author", http.MethodGet, "/community/posts", nil)
	feedW := httptest.NewRecorder()
	r.ServeHTTP(feedW, feedReq)
	var feed struct {
		Meta struct {
			Total int `json:"total"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(feedW.Body.Bytes(), &feed); err != nil {
		t.Fatalf("decode feed: %v", err)
	}
	if feed.Meta.Total != 0 {
		t.Errorf("feed total = %d, want 0 after removal", feed.Meta.Total)
	}
}

func TestCommunityPosts_FlaggedByModeration_HiddenFromFeed(t *testing.T) {
	aiClient := fakes.NewAIClient()
	aiClient.FlaggedBodies["risky content"] = true
	r, jwt := newTestCommunityRouter(aiClient)

	createReq := authedRequest(t, jwt, "user-1", http.MethodPost, "/community/posts", []byte(`{"body":"risky content"}`))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, createReq)
	created := decodeData[struct {
		Status string `json:"status"`
	}](t, w.Body.Bytes())
	if created.Status != "under_review" {
		t.Errorf("Status = %q, want under_review", created.Status)
	}

	feedReq := authedRequest(t, jwt, "user-2", http.MethodGet, "/community/posts", nil)
	feedW := httptest.NewRecorder()
	r.ServeHTTP(feedW, feedReq)
	var feed struct {
		Meta struct {
			Total int `json:"total"`
		} `json:"meta"`
	}
	_ = json.Unmarshal(feedW.Body.Bytes(), &feed)
	if feed.Meta.Total != 0 {
		t.Errorf("feed total = %d, want 0 (flagged post excluded)", feed.Meta.Total)
	}
}

func TestCommunityRoutes_RejectUnauthenticated(t *testing.T) {
	r, _ := newTestCommunityRouter(fakes.NewAIClient())

	req := httptest.NewRequest(http.MethodGet, "/community/posts", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}
