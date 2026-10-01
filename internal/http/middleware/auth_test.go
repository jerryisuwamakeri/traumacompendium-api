package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/traumacompendium/backend/internal/config"
	"github.com/traumacompendium/backend/internal/pkg/jwtutil"
)

func testJWTManager() *jwtutil.Manager {
	return jwtutil.NewManager(config.JWTConfig{
		AccessSecret:  "test-access-secret",
		RefreshSecret: "test-refresh-secret",
		AccessTTL:     15 * time.Minute,
		RefreshTTL:    30 * 24 * time.Hour,
		Issuer:        "traumacompendium-test",
	})
}

func newAuthTestRouter(jwt *jwtutil.Manager, roles ...string) *gin.Engine {
	r := gin.New()
	group := r.Group("/")
	group.Use(RequireAuth(jwt))
	if len(roles) > 0 {
		group.Use(RequireRole(roles...))
	}
	group.GET("/me", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"user_id": UserID(c), "role": Role(c)})
	})
	return r
}

func TestRequireAuth_MissingHeader_401(t *testing.T) {
	r := newAuthTestRouter(testJWTManager())

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestRequireAuth_MalformedHeader_401(t *testing.T) {
	r := newAuthTestRouter(testJWTManager())

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("Authorization", "not-a-bearer-token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestRequireAuth_ValidToken_SetsUserIDAndRole(t *testing.T) {
	jwt := testJWTManager()
	r := newAuthTestRouter(jwt)

	issued, err := jwt.GenerateAccessToken("user-123", "user")
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("Authorization", "Bearer "+issued.Token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusOK, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"user_id":"user-123"`) {
		t.Errorf("body = %s, want it to contain the authenticated user_id", w.Body.String())
	}
}

func TestRequireAuth_RefreshTokenRejected(t *testing.T) {
	jwt := testJWTManager()
	r := newAuthTestRouter(jwt)

	// A refresh token must never authorize API access, even though it's a
	// validly-signed JWT from the same issuer.
	issued, err := jwt.GenerateRefreshToken("user-123")
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("Authorization", "Bearer "+issued.Token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestRequireRole_WrongRole_403(t *testing.T) {
	jwt := testJWTManager()
	r := newAuthTestRouter(jwt, "admin")

	issued, err := jwt.GenerateAccessToken("user-123", "user")
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("Authorization", "Bearer "+issued.Token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestRequireRole_CorrectRole_Allowed(t *testing.T) {
	jwt := testJWTManager()
	r := newAuthTestRouter(jwt, "admin")

	issued, err := jwt.GenerateAccessToken("admin-1", "admin")
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("Authorization", "Bearer "+issued.Token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}
