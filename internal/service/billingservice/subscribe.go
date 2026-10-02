package billingservice

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/traumacompendium/backend/internal/domain/billing"
	"github.com/traumacompendium/backend/internal/domain/user"
	"github.com/traumacompendium/backend/internal/notification/paystack"
	"github.com/traumacompendium/backend/internal/pkg/apperror"
)

// SubscribeResult is what the client redirects the payer to.
type SubscribeResult struct {
	AuthorizationURL string
	Reference        string
}

// Subscribe starts a Premium checkout: a fresh Paystack reference,
// recorded as a Pending subscription so the webhook can later find
// which user a payment confirmation belongs to. It does NOT grant
// premium access — that only happens once HandleWebhook confirms a
// successful payment.
func (s *Service) Subscribe(ctx context.Context, userID string) (SubscribeResult, error) {
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return SubscribeResult{}, apperror.NewNotFound("account not found")
		}
		return SubscribeResult{}, apperror.NewInternal("could not start checkout").WithCause(err)
	}

	reference := uuid.NewString()
	now := s.now()

	pending := &billing.Subscription{
		ID:                uuid.NewString(),
		UserID:            userID,
		Plan:              billing.PlanPremium,
		Status:            billing.StatusPending,
		PaystackReference: reference,
		CurrentPeriodEnd:  now,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if err := s.subscriptions.Upsert(ctx, pending); err != nil {
		return SubscribeResult{}, apperror.NewInternal("could not start checkout").WithCause(err)
	}

	result, err := s.paystack.InitializeTransaction(ctx, paystack.InitializeRequest{
		Email:       u.Email,
		AmountKobo:  PremiumPriceKobo,
		Reference:   reference,
		CallbackURL: s.appBaseURL + "/billing/callback",
	})
	if err != nil {
		return SubscribeResult{}, apperror.NewInternal("could not start checkout").WithCause(err)
	}

	return SubscribeResult{AuthorizationURL: result.AuthorizationURL, Reference: result.Reference}, nil
}

// GetSubscription returns the caller's subscription, or nil if they've
// never subscribed — callers treat nil the same as an explicit Free
// plan, not an error.
func (s *Service) GetSubscription(ctx context.Context, userID string) (*billing.Subscription, error) {
	sub, err := s.subscriptions.FindByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, billing.ErrNotFound) {
			return nil, nil
		}
		return nil, apperror.NewInternal("could not load subscription").WithCause(err)
	}
	return sub, nil
}
