package billingservice

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/traumacompendium/backend/internal/domain/billing"
	"github.com/traumacompendium/backend/internal/notification/paystack"
	"github.com/traumacompendium/backend/internal/pkg/apperror"
)

const testPaystackSecret = "test-paystack-secret"

func sign(t *testing.T, body []byte) string {
	t.Helper()
	mac := hmac.New(sha512.New, []byte(testPaystackSecret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

func chargeSuccessPayload(reference string) []byte {
	return []byte(fmt.Sprintf(`{"event":"charge.success","data":{"reference":%q}}`, reference))
}

func TestHandleWebhook_InvalidSignature_ReturnsUnauthorized(t *testing.T) {
	svc, _, _, _, _ := newTestService()
	body := chargeSuccessPayload("ref-1")

	err := svc.HandleWebhook(context.Background(), body, "not-a-real-signature")
	appErr, ok := apperror.As(err)
	if !ok || appErr.Code != apperror.CodeUnauthorized {
		t.Errorf("HandleWebhook() with a bad signature error = %v, want unauthorized", err)
	}
}

func TestHandleWebhook_ChargeSuccess_ActivatesSubscription(t *testing.T) {
	svc, subs, _, paystackClient, users := newTestService()
	ctx := context.Background()
	seedUser(t, users, "user-1", "payer@example.com")
	paystackClient.VerifyStatus = paystack.TransactionSuccess

	result, err := svc.Subscribe(ctx, "user-1")
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	body := chargeSuccessPayload(result.Reference)
	if err := svc.HandleWebhook(ctx, body, sign(t, body)); err != nil {
		t.Fatalf("HandleWebhook() error = %v", err)
	}

	sub, err := subs.FindByUserID(ctx, "user-1")
	if err != nil {
		t.Fatalf("FindByUserID() error = %v", err)
	}
	if sub.Status != billing.StatusActive {
		t.Errorf("Status = %q, want active", sub.Status)
	}

	premium, err := svc.IsPremium(ctx, "user-1")
	if err != nil {
		t.Fatalf("IsPremium() error = %v", err)
	}
	if !premium {
		t.Error("IsPremium() = false, want true after a confirmed payment")
	}
}

func TestHandleWebhook_Replay_DoesNotReExtend(t *testing.T) {
	svc, subs, _, paystackClient, users := newTestService()
	ctx := context.Background()
	seedUser(t, users, "user-1", "payer@example.com")

	result, err := svc.Subscribe(ctx, "user-1")
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}
	body := chargeSuccessPayload(result.Reference)
	signature := sign(t, body)

	if err := svc.HandleWebhook(ctx, body, signature); err != nil {
		t.Fatalf("HandleWebhook() first delivery error = %v", err)
	}
	firstEnd, err := subs.FindByUserID(ctx, "user-1")
	if err != nil {
		t.Fatalf("FindByUserID() error = %v", err)
	}

	if err := svc.HandleWebhook(ctx, body, signature); err != nil {
		t.Fatalf("HandleWebhook() replay error = %v", err)
	}

	if len(paystackClient.VerifyCalls) != 1 {
		t.Errorf("len(VerifyCalls) = %d, want 1 (replay must not re-verify)", len(paystackClient.VerifyCalls))
	}
	secondEnd, err := subs.FindByUserID(ctx, "user-1")
	if err != nil {
		t.Fatalf("FindByUserID() error = %v", err)
	}
	if !firstEnd.CurrentPeriodEnd.Equal(secondEnd.CurrentPeriodEnd) {
		t.Errorf("CurrentPeriodEnd changed on replay: %v -> %v, want unchanged", firstEnd.CurrentPeriodEnd, secondEnd.CurrentPeriodEnd)
	}
}

func TestHandleWebhook_IgnoresNonChargeSuccessEvents(t *testing.T) {
	svc, subs, _, _, users := newTestService()
	ctx := context.Background()
	seedUser(t, users, "user-1", "payer@example.com")

	result, err := svc.Subscribe(ctx, "user-1")
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	body := []byte(fmt.Sprintf(`{"event":"subscription.create","data":{"reference":%q}}`, result.Reference))
	if err := svc.HandleWebhook(ctx, body, sign(t, body)); err != nil {
		t.Fatalf("HandleWebhook() error = %v", err)
	}

	sub, err := subs.FindByUserID(ctx, "user-1")
	if err != nil {
		t.Fatalf("FindByUserID() error = %v", err)
	}
	if sub.Status != billing.StatusPending {
		t.Errorf("Status = %q, want still pending (event type not acted on)", sub.Status)
	}
}

func TestHandleWebhook_VerificationDoesNotConfirmSuccess_DoesNotActivate(t *testing.T) {
	svc, subs, _, paystackClient, users := newTestService()
	ctx := context.Background()
	seedUser(t, users, "user-1", "payer@example.com")
	paystackClient.VerifyStatus = paystack.TransactionFailed

	result, err := svc.Subscribe(ctx, "user-1")
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	body := chargeSuccessPayload(result.Reference)
	if err := svc.HandleWebhook(ctx, body, sign(t, body)); err != nil {
		t.Fatalf("HandleWebhook() error = %v", err)
	}

	sub, err := subs.FindByUserID(ctx, "user-1")
	if err != nil {
		t.Fatalf("FindByUserID() error = %v", err)
	}
	if sub.Status != billing.StatusPending {
		t.Errorf("Status = %q, want still pending — the webhook claimed success but server-side verification did not confirm it", sub.Status)
	}
}

func TestHandleWebhook_UnrecognizedReference_DoesNotError(t *testing.T) {
	svc, _, _, _, _ := newTestService()
	ctx := context.Background()

	body := chargeSuccessPayload("never-seen-this-reference")
	if err := svc.HandleWebhook(ctx, body, sign(t, body)); err != nil {
		t.Errorf("HandleWebhook() for an unrecognized reference error = %v, want nil", err)
	}
}
