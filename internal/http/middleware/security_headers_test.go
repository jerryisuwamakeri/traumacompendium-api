package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSecurityHeaders_SetsExpectedHeaders(t *testing.T) {
	r := gin.New()
	r.Use(SecurityHeaders())
	r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	want := map[string]string{
		"X-Content-Type-Options":  "nosniff",
		"X-Frame-Options":         "DENY",
		"Referrer-Policy":         "strict-origin-when-cross-origin",
		"Content-Security-Policy": "default-src 'none'",
		"Permissions-Policy":      "geolocation=(), camera=(), microphone=()",
	}
	for header, wantValue := range want {
		if got := w.Header().Get(header); got != wantValue {
			t.Errorf("%s = %q, want %q", header, got, wantValue)
		}
	}
}

func TestSecurityHeaders_DoesNotSetHSTS(t *testing.T) {
	// Strict-Transport-Security is deliberately left to the reverse
	// proxy — asserting its absence here guards against someone adding
	// it back without reading why it isn't here.
	r := gin.New()
	r.Use(SecurityHeaders())
	r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if got := w.Header().Get("Strict-Transport-Security"); got != "" {
		t.Errorf("Strict-Transport-Security = %q, want unset (belongs on the reverse proxy)", got)
	}
}
