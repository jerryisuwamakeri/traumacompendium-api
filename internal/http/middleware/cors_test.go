package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newCORSTestRouter(allowed []string) *gin.Engine {
	r := gin.New()
	r.Use(CORS(allowed))
	r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })
	return r
}

func TestCORS_AllowedOrigin_ReflectsIt(t *testing.T) {
	r := newCORSTestRouter([]string{"https://app.traumacompendium.com"})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Origin", "https://app.traumacompendium.com")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://app.traumacompendium.com" {
		t.Errorf("Access-Control-Allow-Origin = %q, want the allowed origin", got)
	}
}

func TestCORS_DisallowedOrigin_NotReflected(t *testing.T) {
	r := newCORSTestRouter([]string{"https://app.traumacompendium.com"})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Access-Control-Allow-Origin = %q, want empty for a disallowed origin", got)
	}
}

func TestCORS_PreflightRequest_RespondsWithoutInvokingHandler(t *testing.T) {
	called := false
	r := gin.New()
	r.Use(CORS([]string{"https://app.traumacompendium.com"}))
	r.GET("/ping", func(c *gin.Context) {
		called = true
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodOptions, "/ping", nil)
	req.Header.Set("Origin", "https://app.traumacompendium.com")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNoContent)
	}
	if called {
		t.Error("preflight request should not reach the route handler")
	}
}
