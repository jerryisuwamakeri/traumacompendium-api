// Package safereachservice implements the SafeReach emergency-alert
// system's business logic: registering trusted contacts, triggering an
// alert and dispatching it to those contacts (SMS + email — see the
// package doc below for why push goes to the owner's own devices
// instead), recording live location during an open alert, and serving
// the token-gated public view trusted contacts use to check status.
//
// Channel design: the brief names SMS/push/email as the dispatch
// channels. Trusted contacts here are plain contact info (see
// safereach.TrustedContact), not necessarily platform accounts with
// their own device tokens — so SMS and email are what actually reaches
// them, each carrying a share link to the public view endpoint. Push
// goes to the triggering user's own registered devices instead, as a
// "your alert was sent" / live-status confirmation channel, most useful
// alongside the authenticated WebSocket stream (also owner-only). This
// is a deliberate reinterpretation of "push to contacts" into something
// this backend can actually deliver without requiring every trusted
// contact to also be a registered user.
package safereachservice

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/rs/zerolog"

	"github.com/traumacompendium/backend/internal/domain/devicetoken"
	"github.com/traumacompendium/backend/internal/domain/safereach"
	"github.com/traumacompendium/backend/internal/notification/email"
	"github.com/traumacompendium/backend/internal/notification/push"
	"github.com/traumacompendium/backend/internal/notification/sms"
)

// Publisher is the one Redis operation this service needs: broadcasting
// a live location update to anyone subscribed to an alert's channel.
// Kept as a narrow, go-redis-free interface so tests can fake it without
// a real Redis; internal/platform/cache.Publisher adapts a real
// *redis.Client to this shape for production wiring.
type Publisher interface {
	Publish(ctx context.Context, channel string, message any) error
}

type Service struct {
	contacts     safereach.TrustedContactRepository
	alerts       safereach.AlertRepository
	deviceTokens devicetoken.Repository
	smsSender    sms.Sender
	emailSender  email.Sender
	pushSender   push.Sender
	publisher    Publisher
	appBaseURL   string
	log          zerolog.Logger
	now          func() time.Time
}

func New(
	contacts safereach.TrustedContactRepository,
	alerts safereach.AlertRepository,
	deviceTokens devicetoken.Repository,
	smsSender sms.Sender,
	emailSender email.Sender,
	pushSender push.Sender,
	publisher Publisher,
	appBaseURL string,
	log zerolog.Logger,
) *Service {
	return &Service{
		contacts:     contacts,
		alerts:       alerts,
		deviceTokens: deviceTokens,
		smsSender:    smsSender,
		emailSender:  emailSender,
		pushSender:   pushSender,
		publisher:    publisher,
		appBaseURL:   appBaseURL,
		log:          log,
		now:          time.Now,
	}
}

// generateShareToken returns a fresh, cryptographically random raw
// token — 32 bytes, base64url-encoded (43 characters). Only its hash is
// ever persisted; see safereach.Alert's doc comment.
func generateShareToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("safereachservice: generate share token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (s *Service) shareURL(alertID, rawToken string) string {
	return fmt.Sprintf("%s/api/v1/safereach/public/alerts/%s?token=%s", s.appBaseURL, alertID, rawToken)
}

func locationChannel(alertID string) string {
	return "safereach:alert:" + alertID + ":location"
}

// LocationChannel returns the Redis pub/sub channel name for alertID's
// live location updates — exported so the HTTP layer's WebSocket handler
// can subscribe to exactly what RecordLocation publishes to.
func LocationChannel(alertID string) string { return locationChannel(alertID) }

// LocationEvent is the JSON payload published to an alert's location
// channel (see RecordLocation) and consumed by the WebSocket stream
// handler.
type LocationEvent struct {
	Lat        float64   `json:"lat"`
	Lng        float64   `json:"lng"`
	RecordedAt time.Time `json:"recorded_at"`
}
