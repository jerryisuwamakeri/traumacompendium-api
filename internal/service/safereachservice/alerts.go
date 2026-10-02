package safereachservice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/traumacompendium/backend/internal/domain/devicetoken"
	"github.com/traumacompendium/backend/internal/domain/safereach"
	"github.com/traumacompendium/backend/internal/notification/email"
	"github.com/traumacompendium/backend/internal/notification/push"
	"github.com/traumacompendium/backend/internal/notification/sms"
	"github.com/traumacompendium/backend/internal/pkg/apperror"
	"github.com/traumacompendium/backend/internal/pkg/hash"
	"github.com/traumacompendium/backend/internal/pkg/sanitize"
)

type TriggerAlertInput struct {
	Level   safereach.Level
	Message string
	// Lat/Lng are optional — a client may not have a fix yet at the
	// instant the alert fires; RecordLocation can add updates afterward.
	Lat *float64
	Lng *float64
}

// TriggerAlert creates the alert and immediately dispatches it to every
// trusted contact (SMS + email) and to the owner's own registered
// devices (push) — see the package doc for why push targets the owner,
// not the contacts.
func (s *Service) TriggerAlert(ctx context.Context, userID string, in TriggerAlertInput) (*safereach.Alert, error) {
	if !safereach.IsValidLevel(in.Level) {
		return nil, apperror.NewValidation("level must be one of check_in, urgent, critical")
	}

	rawToken, err := generateShareToken()
	if err != nil {
		return nil, apperror.NewInternal("could not create alert").WithCause(err)
	}

	a := &safereach.Alert{
		ID:             uuid.NewString(),
		UserID:         userID,
		Level:          in.Level,
		Status:         safereach.StatusActive,
		Message:        strings.TrimSpace(sanitize.StripHTML(in.Message)),
		ShareTokenHash: hash.Token(rawToken),
		TriggeredAt:    s.now(),
	}
	if err := s.alerts.Create(ctx, a); err != nil {
		return nil, apperror.NewInternal("could not create alert").WithCause(err)
	}

	if in.Lat != nil && in.Lng != nil {
		if err := s.recordLocationUnchecked(ctx, a.ID, *in.Lat, *in.Lng); err != nil {
			s.log.Error().Err(err).Str("alert_id", a.ID).Msg("safereachservice: failed to record initial location")
		}
	}

	s.dispatchAlertNotifications(ctx, userID, a, rawToken)

	return a, nil
}

func (s *Service) dispatchAlertNotifications(ctx context.Context, userID string, a *safereach.Alert, rawToken string) {
	body := alertMessageBody(a, s.shareURL(a.ID, rawToken))

	contacts, err := s.contacts.ListByUser(ctx, userID)
	if err != nil {
		s.log.Error().Err(err).Str("alert_id", a.ID).Msg("safereachservice: failed to load trusted contacts for dispatch")
		contacts = nil
	}
	for _, c := range contacts {
		if c.Phone != "" {
			if err := s.smsSender.Send(ctx, sms.Message{To: c.Phone, Body: body}); err != nil {
				s.log.Error().Err(err).Str("alert_id", a.ID).Str("contact_id", c.ID).Msg("safereachservice: failed to send sms alert")
			}
		}
		if c.Email != "" {
			if err := s.emailSender.Send(ctx, email.Message{To: c.Email, Subject: "TraumaCompendium SafeReach alert", Body: body}); err != nil {
				s.log.Error().Err(err).Str("alert_id", a.ID).Str("contact_id", c.ID).Msg("safereachservice: failed to send email alert")
			}
		}
	}

	devices, err := s.deviceTokens.ListByUser(ctx, userID)
	if err != nil {
		s.log.Error().Err(err).Str("alert_id", a.ID).Msg("safereachservice: failed to load device tokens for dispatch")
		devices = nil
	}
	for _, d := range devices {
		if err := s.pushSender.Send(ctx, push.Message{
			DeviceToken: d.Token,
			Title:       "SafeReach alert sent",
			Body:        body,
			Data:        map[string]string{"alert_id": a.ID, "type": "safereach_alert"},
		}); err != nil {
			s.log.Error().Err(err).Str("alert_id", a.ID).Str("device_id", d.ID).Msg("safereachservice: failed to send push confirmation")
		}
	}
}

func alertMessageBody(a *safereach.Alert, shareURL string) string {
	levelText := map[safereach.Level]string{
		safereach.LevelCheckIn:  "wants to check in",
		safereach.LevelUrgent:   "needs urgent support",
		safereach.LevelCritical: "has a critical emergency",
	}[a.Level]

	msg := fmt.Sprintf("TraumaCompendium SafeReach: a contact %s.", levelText)
	if a.Message != "" {
		msg += " Message: " + a.Message
	}
	msg += " View live status: " + shareURL
	msg += " This is not emergency dispatch — if this is life-threatening, contact local emergency services immediately."
	return msg
}

// CancelAlert and ResolveAlert both close an open alert and immediately
// cut off its share link; closing an already-closed alert is a no-op,
// not an error, so a client retry or a double-tap never fails loudly.
func (s *Service) CancelAlert(ctx context.Context, userID, alertID string) (*safereach.Alert, error) {
	return s.closeAlert(ctx, userID, alertID, safereach.StatusCancelled)
}

func (s *Service) ResolveAlert(ctx context.Context, userID, alertID string) (*safereach.Alert, error) {
	return s.closeAlert(ctx, userID, alertID, safereach.StatusResolved)
}

func (s *Service) closeAlert(ctx context.Context, userID, alertID string, status safereach.Status) (*safereach.Alert, error) {
	a, err := s.alerts.FindOwnedByID(ctx, alertID, userID)
	if err != nil {
		return nil, mapAlertNotFound(err)
	}
	if !a.IsOpen() {
		return a, nil
	}

	now := s.now()
	a.Status = status
	a.ResolvedAt = &now
	if err := s.alerts.Update(ctx, a); err != nil {
		return nil, apperror.NewInternal("could not update alert").WithCause(err)
	}
	return a, nil
}

// GetAlert returns a single alert the caller owns — used both for a
// plain "get alert detail" read and to check ownership before upgrading
// the WebSocket live-location stream (an upgraded connection can't
// return a normal JSON error response, so that check has to happen
// first).
func (s *Service) GetAlert(ctx context.Context, userID, alertID string) (*safereach.Alert, error) {
	a, err := s.alerts.FindOwnedByID(ctx, alertID, userID)
	if err != nil {
		return nil, mapAlertNotFound(err)
	}
	return a, nil
}

type ListAlertsResult struct {
	Alerts []*safereach.Alert
	Total  int
}

// ListAlerts is the event log: every alert the user has triggered,
// newest first.
func (s *Service) ListAlerts(ctx context.Context, userID string, filter safereach.ListFilter) (ListAlertsResult, error) {
	alerts, total, err := s.alerts.List(ctx, userID, filter)
	if err != nil {
		return ListAlertsResult{}, apperror.NewInternal("could not list alerts").WithCause(err)
	}
	return ListAlertsResult{Alerts: alerts, Total: total}, nil
}

// RecordLocation appends a GPS fix to an open alert the caller owns, and
// publishes it on the alert's Redis pub/sub channel for the WebSocket
// stream to forward live.
func (s *Service) RecordLocation(ctx context.Context, userID, alertID string, lat, lng float64) error {
	a, err := s.alerts.FindOwnedByID(ctx, alertID, userID)
	if err != nil {
		return mapAlertNotFound(err)
	}
	if !a.IsOpen() {
		return apperror.NewUnprocessable("this alert is no longer active")
	}
	return s.recordLocationUnchecked(ctx, alertID, lat, lng)
}

func (s *Service) recordLocationUnchecked(ctx context.Context, alertID string, lat, lng float64) error {
	u := &safereach.LocationUpdate{
		ID: uuid.NewString(), AlertID: alertID,
		Lat: lat, Lng: lng, RecordedAt: s.now(),
	}
	if err := s.alerts.AddLocationUpdate(ctx, u); err != nil {
		return apperror.NewInternal("could not record location").WithCause(err)
	}

	payload, err := json.Marshal(LocationEvent{Lat: lat, Lng: lng, RecordedAt: u.RecordedAt})
	if err != nil {
		s.log.Warn().Err(err).Str("alert_id", alertID).Msg("safereachservice: failed to encode location update for publish")
		return nil
	}
	if err := s.publisher.Publish(ctx, locationChannel(alertID), payload); err != nil {
		// Publishing is best-effort: the fix is already durably stored
		// above, so a Redis hiccup only costs a live-stream update, not
		// the data itself.
		s.log.Warn().Err(err).Str("alert_id", alertID).Msg("safereachservice: failed to publish location update")
	}
	return nil
}

// PublicAlertView is what the unauthenticated, token-gated share link
// exposes — deliberately minimal: enough for a trusted contact to see
// what's going on, nothing more.
type PublicAlertView struct {
	Level     safereach.Level
	Status    safereach.Status
	Message   string
	Lat       *float64
	Lng       *float64
	UpdatedAt *time.Time
}

// GetPublicView verifies rawToken against the alert's stored hash before
// returning anything — a wrong or missing token is reported the same as
// "not found" so the endpoint doesn't leak whether a given alert ID
// exists to someone guessing IDs without the token.
func (s *Service) GetPublicView(ctx context.Context, alertID, rawToken string) (PublicAlertView, error) {
	a, err := s.alerts.FindByID(ctx, alertID)
	if err != nil {
		return PublicAlertView{}, mapAlertNotFound(err)
	}
	if rawToken == "" || !hash.TokenMatches(a.ShareTokenHash, rawToken) {
		return PublicAlertView{}, apperror.NewNotFound("alert not found")
	}

	view := PublicAlertView{Level: a.Level, Status: a.Status, Message: a.Message}
	if !a.IsOpen() {
		return view, nil
	}

	loc, err := s.alerts.LatestLocation(ctx, alertID)
	if err != nil {
		return PublicAlertView{}, apperror.NewInternal("could not load alert status").WithCause(err)
	}
	if loc != nil {
		view.Lat, view.Lng = &loc.Lat, &loc.Lng
		view.UpdatedAt = &loc.RecordedAt
	}
	return view, nil
}

// RegisterDeviceToken records a device token for push delivery of the
// caller's own SafeReach alert confirmations/live updates.
func (s *Service) RegisterDeviceToken(ctx context.Context, userID, token, platform string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return apperror.NewValidation("device token is required")
	}

	d := &devicetoken.DeviceToken{ID: uuid.NewString(), UserID: userID, Token: token, Platform: platform}
	if err := s.deviceTokens.Register(ctx, d); err != nil {
		return apperror.NewInternal("could not register device token").WithCause(err)
	}
	return nil
}

func mapAlertNotFound(err error) error {
	if errors.Is(err, safereach.ErrAlertNotFound) {
		return apperror.NewNotFound("alert not found")
	}
	return apperror.NewInternal("could not process alert").WithCause(err)
}
