// Package paystack integrates with Paystack for Premium subscription
// checkout, mirroring internal/notification/tts exactly: a Client
// interface, a sandboxClient that never calls out (used in dev/CI/tests
// — fabricates a successful checkout/verification without reaching
// Paystack), and a production implementation for config.PaystackConfig.Mode
// == production, calling Paystack's REST API over plain net/http.
//
// Card data never touches this backend: InitializeTransaction returns
// Paystack's own hosted-checkout URL, and the client is redirected there
// directly — the standard, correct shape for a Paystack integration.
package paystack

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/rs/zerolog"

	"github.com/traumacompendium/backend/internal/config"
)

// InitializeRequest starts a new checkout.
type InitializeRequest struct {
	Email       string
	AmountKobo  int64 // Paystack amounts are in the currency's smallest unit (kobo for NGN)
	Reference   string
	CallbackURL string
}

// InitializeResult is what the client redirects the payer to.
type InitializeResult struct {
	AuthorizationURL string
	AccessCode       string
	Reference        string
}

// TransactionStatus mirrors Paystack's transaction status values.
type TransactionStatus string

const (
	TransactionSuccess   TransactionStatus = "success"
	TransactionFailed    TransactionStatus = "failed"
	TransactionAbandoned TransactionStatus = "abandoned"
)

// VerifyResult is the server-side-confirmed outcome of a transaction —
// always re-checked against Paystack directly (never trusted from a
// webhook body alone) before a subscription is activated.
type VerifyResult struct {
	Reference  string
	Status     TransactionStatus
	AmountKobo int64
}

// Client is what this codebase needs from Paystack.
type Client interface {
	InitializeTransaction(ctx context.Context, req InitializeRequest) (InitializeResult, error)
	VerifyTransaction(ctx context.Context, reference string) (VerifyResult, error)
}

// New returns the Client appropriate for cfg.Mode.
func New(cfg config.PaystackConfig, log zerolog.Logger) Client {
	if cfg.Mode == config.ModeProduction {
		return &paystackClient{
			cfg:        cfg,
			httpClient: &http.Client{Timeout: 15 * time.Second},
		}
	}
	return &sandboxClient{log: log}
}

// sandboxClient never calls Paystack. InitializeTransaction fabricates a
// checkout URL carrying the reference; VerifyTransaction always reports
// success for any reference, so the full "subscribe -> webhook ->
// activate" pipeline is exercisable end to end in dev/CI without a real
// Paystack account.
type sandboxClient struct {
	log zerolog.Logger
}

func (s *sandboxClient) InitializeTransaction(_ context.Context, req InitializeRequest) (InitializeResult, error) {
	s.log.Info().
		Str("channel", "payment").
		Str("provider", "paystack").
		Str("mode", "sandbox").
		Str("reference", req.Reference).
		Int64("amount_kobo", req.AmountKobo).
		Msg("notification: paystack checkout suppressed in sandbox mode (would have initialized a real transaction)")

	return InitializeResult{
		AuthorizationURL: "https://checkout.paystack.test/" + req.Reference,
		AccessCode:       "sandbox_" + req.Reference,
		Reference:        req.Reference,
	}, nil
}

func (s *sandboxClient) VerifyTransaction(_ context.Context, reference string) (VerifyResult, error) {
	s.log.Info().
		Str("channel", "payment").
		Str("provider", "paystack").
		Str("mode", "sandbox").
		Str("reference", reference).
		Msg("notification: paystack verification suppressed in sandbox mode (reporting success)")

	return VerifyResult{Reference: reference, Status: TransactionSuccess}, nil
}

// paystackClient calls Paystack's REST API directly.
// https://paystack.com/docs/api/transaction/
type paystackClient struct {
	cfg        config.PaystackConfig
	httpClient *http.Client
}

type initializeRequestBody struct {
	Email       string `json:"email"`
	Amount      int64  `json:"amount"`
	Reference   string `json:"reference"`
	CallbackURL string `json:"callback_url,omitempty"`
}

type initializeResponseBody struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		AuthorizationURL string `json:"authorization_url"`
		AccessCode       string `json:"access_code"`
		Reference        string `json:"reference"`
	} `json:"data"`
}

func (c *paystackClient) InitializeTransaction(ctx context.Context, req InitializeRequest) (InitializeResult, error) {
	payload, err := json.Marshal(initializeRequestBody{
		Email: req.Email, Amount: req.AmountKobo, Reference: req.Reference, CallbackURL: req.CallbackURL,
	})
	if err != nil {
		return InitializeResult{}, fmt.Errorf("paystack: encode request: %w", err)
	}

	body, err := c.do(ctx, http.MethodPost, "/transaction/initialize", bytes.NewReader(payload))
	if err != nil {
		return InitializeResult{}, err
	}

	var parsed initializeResponseBody
	if err := json.Unmarshal(body, &parsed); err != nil {
		return InitializeResult{}, fmt.Errorf("paystack: decode initialize response: %w", err)
	}
	if !parsed.Status {
		return InitializeResult{}, fmt.Errorf("paystack: initialize failed: %s", parsed.Message)
	}

	return InitializeResult{
		AuthorizationURL: parsed.Data.AuthorizationURL,
		AccessCode:       parsed.Data.AccessCode,
		Reference:        parsed.Data.Reference,
	}, nil
}

type verifyResponseBody struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Reference string `json:"reference"`
		Status    string `json:"status"`
		Amount    int64  `json:"amount"`
	} `json:"data"`
}

func (c *paystackClient) VerifyTransaction(ctx context.Context, reference string) (VerifyResult, error) {
	path := "/transaction/verify/" + url.PathEscape(reference)
	body, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return VerifyResult{}, err
	}

	var parsed verifyResponseBody
	if err := json.Unmarshal(body, &parsed); err != nil {
		return VerifyResult{}, fmt.Errorf("paystack: decode verify response: %w", err)
	}
	if !parsed.Status {
		return VerifyResult{}, fmt.Errorf("paystack: verify failed: %s", parsed.Message)
	}

	return VerifyResult{
		Reference:  parsed.Data.Reference,
		Status:     TransactionStatus(parsed.Data.Status),
		AmountKobo: parsed.Data.Amount,
	}, nil
}

func (c *paystackClient) do(ctx context.Context, method, path string, body io.Reader) ([]byte, error) {
	httpReq, err := http.NewRequestWithContext(ctx, method, c.cfg.BaseURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("paystack: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.cfg.SecretKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("paystack: request paystack: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck // response already fully read below

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("paystack: read paystack response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("paystack: returned %d for %s: %s", resp.StatusCode, path, string(respBody))
	}

	return respBody, nil
}

// VerifyWebhookSignature reports whether signature (the X-Paystack-Signature
// header) is a valid HMAC-SHA512 of body using secretKey — Paystack's
// documented webhook authenticity check. Call this before trusting
// anything in a webhook payload.
func VerifyWebhookSignature(secretKey string, body []byte, signature string) bool {
	mac := hmac.New(sha512.New, []byte(secretKey))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}
