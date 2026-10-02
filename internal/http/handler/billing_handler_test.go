package handler

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/traumacompendium/backend/internal/config"
	"github.com/traumacompendium/backend/internal/domain/user"
	"github.com/traumacompendium/backend/internal/http/middleware"
	"github.com/traumacompendium/backend/internal/pkg/jwtutil"
	"github.com/traumacompendium/backend/internal/service/billingservice"
	"github.com/traumacompendium/backend/internal/testsupport/fakes"
)

const testBillingWebhookSecret = "test-paystack-secret"

func newTestBillingRouter() (*gin.Engine, *jwtutil.Manager, *fakes.SubscriptionRepo, *fakes.PaystackClient) {
	jwt := jwtutil.NewManager(config.JWTConfig{
		AccessSecret:  "test-access-secret",
		RefreshSecret: "test-refresh-secret",
		AccessTTL:     15 * time.Minute,
		RefreshTTL:    30 * 24 * time.Hour,
		Issuer:        "traumacompendium-test",
	})
	subs := fakes.NewSubscriptionRepo()
	events := fakes.NewWebhookEventRepo()
	paystackClient := fakes.NewPaystackClient()
	users := fakes.NewUserRepo()
	_ = users.Create(context.Background(), &user.User{
		ID: "user-1", FullName: "Test User", Email: "payer@example.com", Phone: "+2348000000000",
		Role: user.RoleUser, Status: user.StatusActive,
	})
	svc := billingservice.New(subs, events, users, paystackClient, testBillingWebhookSecret, "https://app.test", zerolog.Nop())
	h := NewBillingHandler(svc)

	r := gin.New()
	billing := r.Group("/billing")
	billing.POST("/webhook", h.Webhook)
	billing.Use(middleware.RequireAuth(jwt))
	billing.POST("/subscribe", h.Subscribe)
	billing.GET("/subscription", h.GetSubscription)

	return r, jwt, subs, paystackClient
}

func signBillingBody(body []byte) string {
	mac := hmac.New(sha512.New, []byte(testBillingWebhookSecret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

func TestBillingSubscribe_ReturnsCheckoutURL(t *testing.T) {
	r, jwt, _, _ := newTestBillingRouter()

	req := authedRequest(t, jwt, "user-1", http.MethodPost, "/billing/subscribe", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Subscribe status = %d, want %d, body=%s", w.Code, http.StatusOK, w.Body.String())
	}

	resp := decodeData[struct {
		AuthorizationURL string `json:"authorization_url"`
		Reference        string `json:"reference"`
	}](t, w.Body.Bytes())
	if resp.AuthorizationURL == "" || resp.Reference == "" {
		t.Errorf("Subscribe() response = %+v, want populated fields", resp)
	}
}

func TestBillingGetSubscription_NeverSubscribed_ReturnsFree(t *testing.T) {
	r, jwt, _, _ := newTestBillingRouter()

	req := authedRequest(t, jwt, "user-1", http.MethodGet, "/billing/subscription", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GetSubscription status = %d, want %d, body=%s", w.Code, http.StatusOK, w.Body.String())
	}

	resp := decodeData[struct {
		Plan string `json:"plan"`
	}](t, w.Body.Bytes())
	if resp.Plan != "free" {
		t.Errorf("Plan = %q, want free", resp.Plan)
	}
}

func TestBillingSubscribe_RequiresAuth(t *testing.T) {
	r, _, _, _ := newTestBillingRouter()

	req := httptest.NewRequest(http.MethodPost, "/billing/subscribe", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestBillingWebhook_NoAuthRequired_ValidSignature_Activates(t *testing.T) {
	r, jwt, subs, _ := newTestBillingRouter()

	subscribeReq := authedRequest(t, jwt, "user-1", http.MethodPost, "/billing/subscribe", nil)
	subscribeW := httptest.NewRecorder()
	r.ServeHTTP(subscribeW, subscribeReq)
	subscribed := decodeData[struct {
		Reference string `json:"reference"`
	}](t, subscribeW.Body.Bytes())

	body := []byte(fmt.Sprintf(`{"event":"charge.success","data":{"reference":%q}}`, subscribed.Reference))
	webhookReq := httptest.NewRequest(http.MethodPost, "/billing/webhook", bytes.NewReader(body))
	webhookReq.Header.Set("X-Paystack-Signature", signBillingBody(body))
	webhookW := httptest.NewRecorder()
	r.ServeHTTP(webhookW, webhookReq)
	if webhookW.Code != http.StatusOK {
		t.Fatalf("Webhook status = %d, want %d, body=%s", webhookW.Code, http.StatusOK, webhookW.Body.String())
	}

	sub, err := subs.FindByUserID(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("FindByUserID() error = %v", err)
	}
	if sub.Status != "active" {
		t.Errorf("Status = %q, want active", sub.Status)
	}
}

func TestBillingWebhook_InvalidSignature_401(t *testing.T) {
	r, _, _, _ := newTestBillingRouter()

	body := []byte(`{"event":"charge.success","data":{"reference":"whatever"}}`)
	req := httptest.NewRequest(http.MethodPost, "/billing/webhook", bytes.NewReader(body))
	req.Header.Set("X-Paystack-Signature", "not-a-real-signature")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d, body=%s", w.Code, http.StatusUnauthorized, w.Body.String())
	}
}

func TestBillingSubscription_AfterActivation_ShowsPremium(t *testing.T) {
	r, jwt, _, _ := newTestBillingRouter()

	subscribeReq := authedRequest(t, jwt, "user-1", http.MethodPost, "/billing/subscribe", nil)
	subscribeW := httptest.NewRecorder()
	r.ServeHTTP(subscribeW, subscribeReq)
	subscribed := decodeData[struct {
		Reference string `json:"reference"`
	}](t, subscribeW.Body.Bytes())

	body := []byte(fmt.Sprintf(`{"event":"charge.success","data":{"reference":%q}}`, subscribed.Reference))
	webhookReq := httptest.NewRequest(http.MethodPost, "/billing/webhook", bytes.NewReader(body))
	webhookReq.Header.Set("X-Paystack-Signature", signBillingBody(body))
	r.ServeHTTP(httptest.NewRecorder(), webhookReq)

	subReq := authedRequest(t, jwt, "user-1", http.MethodGet, "/billing/subscription", nil)
	subW := httptest.NewRecorder()
	r.ServeHTTP(subW, subReq)
	var resp struct {
		Data struct {
			Plan   string `json:"plan"`
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(subW.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Data.Plan != "premium" || resp.Data.Status != "active" {
		t.Errorf("subscription = %+v, want Plan=premium Status=active", resp.Data)
	}
}
