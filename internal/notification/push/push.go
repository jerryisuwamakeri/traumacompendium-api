// Package push sends mobile push notifications via Firebase Cloud
// Messaging, mirroring internal/notification/sms exactly: a Sender
// interface, a sandbox implementation (used in dev/CI/tests — no real
// provider call) and a production implementation, selected by
// config.PushConfig.Mode.
//
// Design note: this targets FCM's legacy server-key HTTP endpoint
// (https://fcm.googleapis.com/fcm/send, "Authorization: key=<server
// key>") rather than the newer OAuth2-based HTTP v1 API. The v1 API
// needs a signed service-account JWT exchanged for an access token —
// real functionality this codebase would otherwise have to hand-roll
// from scratch for one integration. The legacy endpoint trades that
// away for a static server key, the same shape as every other
// integration here (Termii, SMTP, TTS, Translate, OpenAI) — a
// deliberate simplicity choice, flagged here rather than silently made.
package push

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

// Message is a single push notification targeted at one device token.
type Message struct {
	DeviceToken string
	Title       string
	Body        string
	// Data carries small out-of-band fields a client might act on
	// (e.g. {"alert_id": "...", "type": "safereach_alert"}) without
	// needing to parse them out of the display text.
	Data map[string]string
}

// Sender dispatches a single push notification.
type Sender interface {
	Send(ctx context.Context, msg Message) error
}

// New returns the Sender appropriate for cfg.Mode.
func New(cfg config.PushConfig, log zerolog.Logger) Sender {
	if cfg.Mode == config.ModeProduction {
		return &fcmSender{
			cfg:        cfg,
			httpClient: &http.Client{Timeout: 10 * time.Second},
		}
	}
	return &sandboxSender{log: log}
}

// sandboxSender logs the notification instead of calling FCM — used in
// local development, tests, and CI.
type sandboxSender struct {
	log zerolog.Logger
}

func (s *sandboxSender) Send(_ context.Context, msg Message) error {
	s.log.Info().
		Str("channel", "push").
		Str("provider", "fcm").
		Str("mode", "sandbox").
		Str("device_token", msg.DeviceToken).
		Str("title", msg.Title).
		Msg("notification: push suppressed in sandbox mode (would have been sent)")
	return nil
}

// fcmSender calls FCM's legacy HTTP send endpoint.
// https://firebase.google.com/docs/cloud-messaging/http-server-ref
type fcmSender struct {
	cfg        config.PushConfig
	httpClient *http.Client
}

type fcmSendRequest struct {
	To           string            `json:"to"`
	Notification fcmNotification   `json:"notification"`
	Data         map[string]string `json:"data,omitempty"`
}

type fcmNotification struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

type fcmSendResponse struct {
	Success int `json:"success"`
	Failure int `json:"failure"`
}

func (s *fcmSender) Send(ctx context.Context, msg Message) error {
	payload, err := json.Marshal(fcmSendRequest{
		To:           msg.DeviceToken,
		Notification: fcmNotification{Title: msg.Title, Body: msg.Body},
		Data:         msg.Data,
	})
	if err != nil {
		return fmt.Errorf("push: encode request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.cfg.BaseURL+"/fcm/send", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("push: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "key="+s.cfg.ServerKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("push: request fcm: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck // response already fully read below

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("push: read fcm response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("push: fcm returned %d: %s", resp.StatusCode, string(body))
	}

	var parsed fcmSendResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return fmt.Errorf("push: decode fcm response: %w", err)
	}
	if parsed.Success != 1 {
		return fmt.Errorf("push: fcm did not report success: %s", string(body))
	}

	return nil
}
