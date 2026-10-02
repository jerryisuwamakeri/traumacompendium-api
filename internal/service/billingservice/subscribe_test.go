package billingservice

import (
	"context"
	"testing"

	"github.com/traumacompendium/backend/internal/domain/billing"
	"github.com/traumacompendium/backend/internal/pkg/apperror"
)

func TestSubscribe_Success_ReturnsCheckoutURL_AndRecordsPending(t *testing.T) {
	svc, subs, _, paystackClient, users := newTestService()
	ctx := context.Background()
	seedUser(t, users, "user-1", "payer@example.com")

	result, err := svc.Subscribe(ctx, "user-1")
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}
	if result.AuthorizationURL == "" || result.Reference == "" {
		t.Errorf("Subscribe() = %+v, want a populated authorization URL and reference", result)
	}

	if len(paystackClient.InitializeCalls) != 1 {
		t.Fatalf("len(InitializeCalls) = %d, want 1", len(paystackClient.InitializeCalls))
	}
	call := paystackClient.InitializeCalls[0]
	if call.Email != "payer@example.com" {
		t.Errorf("InitializeTransaction email = %q, want payer@example.com", call.Email)
	}
	if call.AmountKobo != PremiumPriceKobo {
		t.Errorf("InitializeTransaction amount = %d, want %d", call.AmountKobo, PremiumPriceKobo)
	}

	sub, err := subs.FindByUserID(ctx, "user-1")
	if err != nil {
		t.Fatalf("FindByUserID() error = %v", err)
	}
	if sub.Status != billing.StatusPending {
		t.Errorf("Status = %q, want pending (not yet confirmed)", sub.Status)
	}
}

func TestSubscribe_UnknownUser_ReturnsNotFound(t *testing.T) {
	svc, _, _, _, _ := newTestService()

	_, err := svc.Subscribe(context.Background(), "does-not-exist")
	appErr, ok := apperror.As(err)
	if !ok || appErr.Code != apperror.CodeNotFound {
		t.Errorf("Subscribe() for an unknown user error = %v, want not_found", err)
	}
}

func TestGetSubscription_NeverSubscribed_ReturnsNil(t *testing.T) {
	svc, _, _, _, _ := newTestService()

	sub, err := svc.GetSubscription(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("GetSubscription() error = %v", err)
	}
	if sub != nil {
		t.Errorf("GetSubscription() = %+v, want nil for a user who never subscribed", sub)
	}
}

func TestGetSubscription_AfterSubscribe_ReturnsPending(t *testing.T) {
	svc, _, _, _, users := newTestService()
	ctx := context.Background()
	seedUser(t, users, "user-1", "payer@example.com")

	if _, err := svc.Subscribe(ctx, "user-1"); err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	sub, err := svc.GetSubscription(ctx, "user-1")
	if err != nil {
		t.Fatalf("GetSubscription() error = %v", err)
	}
	if sub == nil || sub.Status != billing.StatusPending {
		t.Errorf("GetSubscription() = %+v, want a pending subscription", sub)
	}
}
