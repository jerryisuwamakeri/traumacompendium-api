package safereachservice

import (
	"context"
	"net/url"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/traumacompendium/backend/internal/domain/devicetoken"
	"github.com/traumacompendium/backend/internal/domain/safereach"
	"github.com/traumacompendium/backend/internal/pkg/apperror"
	"github.com/traumacompendium/backend/internal/testsupport/fakes"
)

func TestTriggerAlert_InvalidLevel_ReturnsValidationError(t *testing.T) {
	svc := newTestService()

	_, err := svc.TriggerAlert(context.Background(), "user-1", TriggerAlertInput{Level: "not-a-real-level"})
	appErr, ok := apperror.As(err)
	if !ok || appErr.Code != apperror.CodeValidation {
		t.Errorf("TriggerAlert() with invalid level error = %v, want validation_error", err)
	}
}

func TestTriggerAlert_DispatchesToTrustedContactsAndOwnDevices(t *testing.T) {
	contacts := fakes.NewSafeReachContactRepo()
	alerts := fakes.NewSafeReachAlertRepo()
	devices := fakes.NewDeviceTokenRepo()
	smsSender := &fakes.SMSSender{}
	emailSender := &fakes.EmailSender{}
	pushSender := &fakes.PushSender{}
	svc := New(contacts, alerts, devices, smsSender, emailSender, pushSender, &fakes.Publisher{}, "https://app.test", zerolog.Nop())
	ctx := context.Background()

	if _, err := svc.RegisterContact(ctx, "user-1", RegisterContactInput{Name: "Ada", Phone: "+2348012345678", Email: "ada@example.com"}); err != nil {
		t.Fatalf("RegisterContact() error = %v", err)
	}
	if err := devices.Register(ctx, &devicetoken.DeviceToken{ID: uuid.NewString(), UserID: "user-1", Token: "device-1"}); err != nil {
		t.Fatalf("Register() device token error = %v", err)
	}

	a, err := svc.TriggerAlert(ctx, "user-1", TriggerAlertInput{Level: safereach.LevelUrgent, Message: "Need help"})
	if err != nil {
		t.Fatalf("TriggerAlert() error = %v", err)
	}
	if a.Status != safereach.StatusActive {
		t.Errorf("Status = %q, want active", a.Status)
	}

	if len(smsSender.Sent) != 1 || smsSender.Sent[0].To != "+2348012345678" {
		t.Errorf("sms sent = %+v, want one message to +2348012345678", smsSender.Sent)
	}
	if !strings.Contains(smsSender.Sent[0].Body, "Need help") {
		t.Errorf("sms body = %q, want it to include the alert message", smsSender.Sent[0].Body)
	}
	if len(emailSender.Sent) != 1 || emailSender.Sent[0].To != "ada@example.com" {
		t.Errorf("email sent = %+v, want one message to ada@example.com", emailSender.Sent)
	}
	if len(pushSender.Sent) != 1 || pushSender.Sent[0].DeviceToken != "device-1" {
		t.Errorf("push sent = %+v, want one message to device-1", pushSender.Sent)
	}
}

func TestTriggerAlert_WithInitialLocation_RecordsAndPublishes(t *testing.T) {
	publisher := &fakes.Publisher{}
	svc := New(
		fakes.NewSafeReachContactRepo(), fakes.NewSafeReachAlertRepo(), fakes.NewDeviceTokenRepo(),
		&fakes.SMSSender{}, &fakes.EmailSender{}, &fakes.PushSender{}, publisher,
		"https://app.test", zerolog.Nop(),
	)
	ctx := context.Background()
	lat, lng := 6.5244, 3.3792

	a, err := svc.TriggerAlert(ctx, "user-1", TriggerAlertInput{Level: safereach.LevelCritical, Lat: &lat, Lng: &lng})
	if err != nil {
		t.Fatalf("TriggerAlert() error = %v", err)
	}

	if len(publisher.Published) != 1 {
		t.Fatalf("len(publisher.Published) = %d, want 1", len(publisher.Published))
	}
	if publisher.Published[0].Channel != LocationChannel(a.ID) {
		t.Errorf("published channel = %q, want %q", publisher.Published[0].Channel, LocationChannel(a.ID))
	}
}

func TestCancelAlert_ClosesAlert_AndIsIdempotent(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	a, err := svc.TriggerAlert(ctx, "user-1", TriggerAlertInput{Level: safereach.LevelCheckIn})
	if err != nil {
		t.Fatalf("TriggerAlert() error = %v", err)
	}

	cancelled, err := svc.CancelAlert(ctx, "user-1", a.ID)
	if err != nil {
		t.Fatalf("CancelAlert() error = %v", err)
	}
	if cancelled.Status != safereach.StatusCancelled {
		t.Errorf("Status = %q, want cancelled", cancelled.Status)
	}

	// Cancelling again must be a no-op, not an error.
	again, err := svc.CancelAlert(ctx, "user-1", a.ID)
	if err != nil {
		t.Fatalf("CancelAlert() second call error = %v", err)
	}
	if again.Status != safereach.StatusCancelled {
		t.Errorf("Status after re-cancel = %q, want still cancelled", again.Status)
	}
}

func TestResolveAlert_WrongUser_ReturnsNotFound(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	a, err := svc.TriggerAlert(ctx, "owner", TriggerAlertInput{Level: safereach.LevelCheckIn})
	if err != nil {
		t.Fatalf("TriggerAlert() error = %v", err)
	}

	_, err = svc.ResolveAlert(ctx, "intruder", a.ID)
	appErr, ok := apperror.As(err)
	if !ok || appErr.Code != apperror.CodeNotFound {
		t.Errorf("ResolveAlert() by non-owner error = %v, want not_found", err)
	}
}

func TestRecordLocation_OnClosedAlert_ReturnsUnprocessable(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	a, err := svc.TriggerAlert(ctx, "user-1", TriggerAlertInput{Level: safereach.LevelCheckIn})
	if err != nil {
		t.Fatalf("TriggerAlert() error = %v", err)
	}
	if _, err := svc.ResolveAlert(ctx, "user-1", a.ID); err != nil {
		t.Fatalf("ResolveAlert() error = %v", err)
	}

	err = svc.RecordLocation(ctx, "user-1", a.ID, 6.5, 3.4)
	appErr, ok := apperror.As(err)
	if !ok || appErr.Code != apperror.CodeUnprocessable {
		t.Errorf("RecordLocation() on a closed alert error = %v, want unprocessable_entity", err)
	}
}

// rawTokenFromLastSMS recovers the raw share token from the share link
// embedded in the most recently dispatched SMS — TriggerAlert never
// returns the raw token directly (only its hash is persisted), so a
// test that needs it reads it the same way a real trusted contact
// would: from the message they received.
func rawTokenFromLastSMS(t *testing.T, sender *fakes.SMSSender) string {
	t.Helper()
	body := sender.Last().Body
	idx := strings.Index(body, "token=")
	if idx == -1 {
		t.Fatalf("sms body %q does not contain a share token", body)
	}
	rest := body[idx+len("token="):]
	end := strings.IndexByte(rest, ' ')
	if end == -1 {
		end = len(rest)
	}
	token, err := url.QueryUnescape(rest[:end])
	if err != nil {
		t.Fatalf("decode share token: %v", err)
	}
	return token
}

func TestGetPublicView_CorrectToken_ReturnsStatusAndLocation(t *testing.T) {
	smsSender := &fakes.SMSSender{}
	svc := New(
		fakes.NewSafeReachContactRepo(), fakes.NewSafeReachAlertRepo(), fakes.NewDeviceTokenRepo(),
		smsSender, &fakes.EmailSender{}, &fakes.PushSender{}, &fakes.Publisher{},
		"https://app.test", zerolog.Nop(),
	)
	ctx := context.Background()
	lat, lng := 6.5244, 3.3792

	if _, err := svc.RegisterContact(ctx, "user-1", RegisterContactInput{Name: "Ada", Phone: "+2348012345678"}); err != nil {
		t.Fatalf("RegisterContact() error = %v", err)
	}

	a, err := svc.TriggerAlert(ctx, "user-1", TriggerAlertInput{Level: safereach.LevelUrgent, Lat: &lat, Lng: &lng})
	if err != nil {
		t.Fatalf("TriggerAlert() error = %v", err)
	}
	rawToken := rawTokenFromLastSMS(t, smsSender)

	view, err := svc.GetPublicView(ctx, a.ID, rawToken)
	if err != nil {
		t.Fatalf("GetPublicView() error = %v", err)
	}
	if view.Status != safereach.StatusActive {
		t.Errorf("Status = %q, want active", view.Status)
	}
	if view.Lat == nil || *view.Lat != lat {
		t.Errorf("Lat = %v, want %v", view.Lat, lat)
	}
}

func TestGetPublicView_WrongToken_ReturnsNotFound(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	a, err := svc.TriggerAlert(ctx, "user-1", TriggerAlertInput{Level: safereach.LevelUrgent})
	if err != nil {
		t.Fatalf("TriggerAlert() error = %v", err)
	}

	_, err = svc.GetPublicView(ctx, a.ID, "definitely-the-wrong-token")
	appErr, ok := apperror.As(err)
	if !ok || appErr.Code != apperror.CodeNotFound {
		t.Errorf("GetPublicView() with wrong token error = %v, want not_found", err)
	}
}

func TestGetPublicView_ClosedAlert_OmitsLocation(t *testing.T) {
	smsSender := &fakes.SMSSender{}
	svc := New(
		fakes.NewSafeReachContactRepo(), fakes.NewSafeReachAlertRepo(), fakes.NewDeviceTokenRepo(),
		smsSender, &fakes.EmailSender{}, &fakes.PushSender{}, &fakes.Publisher{},
		"https://app.test", zerolog.Nop(),
	)
	ctx := context.Background()
	lat, lng := 6.5244, 3.3792

	if _, err := svc.RegisterContact(ctx, "user-1", RegisterContactInput{Name: "Ada", Phone: "+2348012345678"}); err != nil {
		t.Fatalf("RegisterContact() error = %v", err)
	}

	a, err := svc.TriggerAlert(ctx, "user-1", TriggerAlertInput{Level: safereach.LevelUrgent, Lat: &lat, Lng: &lng})
	if err != nil {
		t.Fatalf("TriggerAlert() error = %v", err)
	}
	rawToken := rawTokenFromLastSMS(t, smsSender)

	if _, err := svc.ResolveAlert(ctx, "user-1", a.ID); err != nil {
		t.Fatalf("ResolveAlert() error = %v", err)
	}

	view, err := svc.GetPublicView(ctx, a.ID, rawToken)
	if err != nil {
		t.Fatalf("GetPublicView() error = %v", err)
	}
	if view.Status != safereach.StatusResolved {
		t.Errorf("Status = %q, want resolved", view.Status)
	}
	if view.Lat != nil {
		t.Errorf("Lat = %v, want nil once the alert is closed", view.Lat)
	}
}
