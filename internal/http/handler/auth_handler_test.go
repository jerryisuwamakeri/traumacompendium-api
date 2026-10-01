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
	"github.com/traumacompendium/backend/internal/pkg/jwtutil"
	"github.com/traumacompendium/backend/internal/service/authservice"
	"github.com/traumacompendium/backend/internal/testsupport/fakes"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newTestAuthHandler() *AuthHandler {
	jwtManager := jwtutil.NewManager(config.JWTConfig{
		AccessSecret:  "test-access-secret",
		RefreshSecret: "test-refresh-secret",
		AccessTTL:     15 * time.Minute,
		RefreshTTL:    30 * 24 * time.Hour,
		Issuer:        "traumacompendium-test",
	})
	svc := authservice.New(
		fakes.NewUserRepo(), fakes.NewOTPRepo(), fakes.NewRefreshTokenRepo(),
		jwtManager, &fakes.EmailSender{}, &fakes.SMSSender{},
		config.OTPConfig{Length: 6, TTL: 10 * time.Minute, MaxAttempts: 5},
		zerolog.Nop(),
	)
	return NewAuthHandler(svc)
}

func TestRegister_EndToEnd_ReturnsCreatedUser(t *testing.T) {
	h := newTestAuthHandler()
	r := gin.New()
	r.POST("/register", h.Register)

	body := `{"full_name":"Jane Doe","email":"jane@example.com","phone":"+2348012345678","password":"correct horse battery staple"}`
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusCreated, w.Body.String())
	}

	var resp struct {
		Data struct {
			Email string `json:"email"`
			ID    string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Data.Email != "jane@example.com" {
		t.Errorf("Email = %q, want %q", resp.Data.Email, "jane@example.com")
	}
	if resp.Data.ID == "" {
		t.Error("expected a non-empty user ID in the response")
	}
}

func TestRegister_InvalidPayload_Returns400WithFieldErrors(t *testing.T) {
	h := newTestAuthHandler()
	r := gin.New()
	r.POST("/register", h.Register)

	body := `{"full_name":"J","email":"not-an-email","phone":"not-a-phone","password":"short"}`
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusBadRequest, w.Body.String())
	}
}

func TestRegister_DuplicateEmail_Returns409(t *testing.T) {
	h := newTestAuthHandler()
	r := gin.New()
	r.POST("/register", h.Register)

	body := `{"full_name":"Jane Doe","email":"jane@example.com","phone":"+2348012345678","password":"correct horse battery staple"}`
	req1 := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(body))
	req1.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(httptest.NewRecorder(), req1)

	body2 := `{"full_name":"Jane Two","email":"jane@example.com","phone":"+2348000000000","password":"correct horse battery staple"}`
	req2 := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(body2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d, body=%s", w2.Code, http.StatusConflict, w2.Body.String())
	}
}
