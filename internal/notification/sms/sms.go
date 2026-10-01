// Package sms sends SMS (phone OTP codes, in Phase 1) via Termii, the SMS
// provider the user specified for this project.
//
// Like the email package, Sender is an interface with a sandbox and a
// production implementation selected by config.TermiiConfig.Mode, so
// nothing ever reaches Termii's live API outside of a deliberately
// configured production environment.
//
// Design note: Termii also offers a stateful "Token" API (send-token /
// verify-token, with Termii holding the pin server-side). This package
// deliberately uses Termii's plain /api/sms/send endpoint instead and lets
// our own otp package + database own the OTP lifecycle — so email and
// phone verification share one code path (generate, hash, store, expire,
// verify) instead of two different state machines.
package sms

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog"

	"github.com/traumacompendium/backend/internal/config"
)

// Message is a single outbound SMS.
type Message struct {
	To   string // E.164 phone number, e.g. +2348012345678
	Body string
}

// Sender dispatches a single SMS.
type Sender interface {
	Send(ctx context.Context, msg Message) error
}

// New returns the Sender appropriate for cfg.Mode.
func New(cfg config.TermiiConfig, log zerolog.Logger) Sender {
	if cfg.Mode == config.ModeProduction {
		return &termiiSender{
			cfg:        cfg,
			httpClient: &http.Client{Timeout: 10 * time.Second},
		}
	}
	return &sandboxSender{cfg: cfg, log: log}
}

// sandboxSender logs the message instead of calling Termii — used in
// local development, tests, and CI.
type sandboxSender struct {
	cfg config.TermiiConfig
	log zerolog.Logger
}

func (s *sandboxSender) Send(_ context.Context, msg Message) error {
	s.log.Info().
		Str("channel", "sms").
		Str("provider", "termii").
		Str("mode", "sandbox").
		Str("to", msg.To).
		Str("body", msg.Body).
		Msg("notification: sms suppressed in sandbox mode (would have been sent)")
	return nil
}

// termiiSender calls Termii's generic SMS send endpoint.
// https://developers.termii.com — POST /api/sms/send
type termiiSender struct {
	cfg        config.TermiiConfig
	httpClient *http.Client
}

type termiiSendRequest struct {
	APIKey  string `json:"api_key"`
	To      string `json:"to"`
	From    string `json:"from"`
	SMS     string `json:"sms"`
	Type    string `json:"type"`
	Channel string `json:"channel"`
}

type termiiSendResponse struct {
	MessageID string `json:"message_id"`
	Message   string `json:"message"`
	Balance   float64 `json:"balance"`
	User      string  `json:"user"`
}

func buildTermiiRequest(cfg config.TermiiConfig, msg Message) termiiSendRequest {
	return termiiSendRequest{
		APIKey:  cfg.APIKey,
		To:      msg.To,
		From:    cfg.SenderID,
		SMS:     msg.Body,
		Type:    "plain",
		Channel: "generic",
	}
}

func (s *termiiSender) Send(ctx context.Context, msg Message) error {
	payload, err := json.Marshal(buildTermiiRequest(s.cfg, msg))
	if err != nil {
		return fmt.Errorf("sms: encode request: %w", err)
	}

	endpoint := s.cfg.BaseURL + "/api/sms/send"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("sms: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("sms: request termii: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("sms: read termii response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("sms: termii returned %d: %s", resp.StatusCode, string(body))
	}

	var parsed termiiSendResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return fmt.Errorf("sms: decode termii response: %w", err)
	}
	if parsed.MessageID == "" {
		return fmt.Errorf("sms: termii did not return a message_id: %s", string(body))
	}

	return nil
}
