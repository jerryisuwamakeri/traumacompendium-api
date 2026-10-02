package paystack

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"

	"github.com/traumacompendium/backend/internal/config"
)

// hmacSHA512Hex independently computes what VerifyWebhookSignature
// checks against, so the test proves the algorithm rather than the
// function merely agreeing with itself.
func hmacSHA512Hex(t *testing.T, secret string, body []byte) string {
	t.Helper()
	mac := hmac.New(sha512.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

func TestNew_SelectsImplementationByMode(t *testing.T) {
	sandbox := New(config.PaystackConfig{Mode: config.ModeSandbox}, zerolog.Nop())
	if _, ok := sandbox.(*sandboxClient); !ok {
		t.Errorf("New(sandbox mode) = %T, want *sandboxClient", sandbox)
	}

	production := New(config.PaystackConfig{Mode: config.ModeProduction}, zerolog.Nop())
	if _, ok := production.(*paystackClient); !ok {
		t.Errorf("New(production mode) = %T, want *paystackClient", production)
	}
}

func TestSandboxClient_InitializeAndVerify_WithoutCallingOut(t *testing.T) {
	c := New(config.PaystackConfig{Mode: config.ModeSandbox}, zerolog.Nop())
	ctx := context.Background()

	init, err := c.InitializeTransaction(ctx, InitializeRequest{Email: "a@example.com", AmountKobo: 250000, Reference: "ref-123"})
	if err != nil {
		t.Fatalf("InitializeTransaction() error = %v", err)
	}
	if init.Reference != "ref-123" || init.AuthorizationURL == "" {
		t.Errorf("InitializeTransaction() = %+v, want a populated authorization URL for ref-123", init)
	}

	verify, err := c.VerifyTransaction(ctx, "ref-123")
	if err != nil {
		t.Fatalf("VerifyTransaction() error = %v", err)
	}
	if verify.Status != TransactionSuccess {
		t.Errorf("Status = %q, want success", verify.Status)
	}
}

func TestPaystackClient_InitializeTransaction_Success(t *testing.T) {
	var received initializeRequestBody

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/transaction/initialize" {
			t.Errorf("path = %s, want /transaction/initialize", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-secret-key" {
			t.Errorf("Authorization = %q, want Bearer test-secret-key", r.Header.Get("Authorization"))
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(initializeResponseBody{
			Status: true, Message: "Authorization URL created",
			Data: struct {
				AuthorizationURL string `json:"authorization_url"`
				AccessCode       string `json:"access_code"`
				Reference        string `json:"reference"`
			}{AuthorizationURL: "https://checkout.paystack.com/abc", AccessCode: "access-abc", Reference: "ref-123"},
		})
	}))
	defer server.Close()

	c := &paystackClient{cfg: config.PaystackConfig{BaseURL: server.URL, SecretKey: "test-secret-key"}, httpClient: server.Client()}

	result, err := c.InitializeTransaction(context.Background(), InitializeRequest{Email: "a@example.com", AmountKobo: 250000, Reference: "ref-123"})
	if err != nil {
		t.Fatalf("InitializeTransaction() error = %v", err)
	}
	if result.AuthorizationURL != "https://checkout.paystack.com/abc" {
		t.Errorf("AuthorizationURL = %q, want https://checkout.paystack.com/abc", result.AuthorizationURL)
	}
	if received.Amount != 250000 || received.Email != "a@example.com" {
		t.Errorf("request = %+v, want Amount=250000 Email=a@example.com", received)
	}
}

func TestPaystackClient_InitializeTransaction_StatusFalse_ReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(initializeResponseBody{Status: false, Message: "Invalid amount"})
	}))
	defer server.Close()

	c := &paystackClient{cfg: config.PaystackConfig{BaseURL: server.URL, SecretKey: "test-secret-key"}, httpClient: server.Client()}

	_, err := c.InitializeTransaction(context.Background(), InitializeRequest{Email: "a@example.com", AmountKobo: -1, Reference: "ref-123"})
	if err == nil {
		t.Fatal("InitializeTransaction() error = nil, want error when paystack reports status:false")
	}
}

func TestPaystackClient_VerifyTransaction_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/transaction/verify/ref-123" {
			t.Errorf("path = %s, want /transaction/verify/ref-123", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(verifyResponseBody{
			Status: true,
			Data: struct {
				Reference string `json:"reference"`
				Status    string `json:"status"`
				Amount    int64  `json:"amount"`
			}{Reference: "ref-123", Status: "success", Amount: 250000},
		})
	}))
	defer server.Close()

	c := &paystackClient{cfg: config.PaystackConfig{BaseURL: server.URL, SecretKey: "test-secret-key"}, httpClient: server.Client()}

	result, err := c.VerifyTransaction(context.Background(), "ref-123")
	if err != nil {
		t.Fatalf("VerifyTransaction() error = %v", err)
	}
	if result.Status != TransactionSuccess || result.AmountKobo != 250000 {
		t.Errorf("result = %+v, want Status=success AmountKobo=250000", result)
	}
}

func TestPaystackClient_NonOKStatus_ReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"status":false,"message":"Invalid key"}`))
	}))
	defer server.Close()

	c := &paystackClient{cfg: config.PaystackConfig{BaseURL: server.URL, SecretKey: "bad-key"}, httpClient: server.Client()}

	_, err := c.VerifyTransaction(context.Background(), "ref-123")
	if err == nil {
		t.Fatal("VerifyTransaction() error = nil, want error for non-2xx response")
	}
}

func TestVerifyWebhookSignature(t *testing.T) {
	secret := "whsec_test"
	body := []byte(`{"event":"charge.success","data":{"reference":"ref-123"}}`)

	validSignature := hmacSHA512Hex(t, secret, body)
	if !VerifyWebhookSignature(secret, body, validSignature) {
		t.Error("VerifyWebhookSignature() = false for a correctly computed signature, want true")
	}

	if VerifyWebhookSignature(secret, body, "not-a-real-signature") {
		t.Error("VerifyWebhookSignature() = true for a bogus signature, want false")
	}

	tampered := append([]byte{}, body...)
	tampered[0] = '!'
	if VerifyWebhookSignature(secret, tampered, validSignature) {
		t.Error("VerifyWebhookSignature() = true after the body was tampered with, want false")
	}
}
