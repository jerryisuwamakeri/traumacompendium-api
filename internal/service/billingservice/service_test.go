package billingservice

import (
	"context"
	"testing"

	"github.com/rs/zerolog"

	"github.com/traumacompendium/backend/internal/domain/user"
	"github.com/traumacompendium/backend/internal/testsupport/fakes"
)

func newTestService() (*Service, *fakes.SubscriptionRepo, *fakes.WebhookEventRepo, *fakes.PaystackClient, *fakes.UserRepo) {
	subs := fakes.NewSubscriptionRepo()
	events := fakes.NewWebhookEventRepo()
	paystackClient := fakes.NewPaystackClient()
	users := fakes.NewUserRepo()
	svc := New(subs, events, users, paystackClient, "test-paystack-secret", "https://app.test", zerolog.Nop())
	return svc, subs, events, paystackClient, users
}

func seedUser(t *testing.T, users *fakes.UserRepo, id, email string) {
	t.Helper()
	u := &user.User{ID: id, FullName: "Test User", Email: email, Phone: "+2348000000000", Role: user.RoleUser, Status: user.StatusActive}
	if err := users.Create(context.Background(), u); err != nil {
		t.Fatalf("seed user: %v", err)
	}
}

func TestIsPremium_NeverSubscribed_False(t *testing.T) {
	svc, _, _, _, _ := newTestService()

	premium, err := svc.IsPremium(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("IsPremium() error = %v", err)
	}
	if premium {
		t.Error("IsPremium() = true, want false for a user who never subscribed")
	}
}
