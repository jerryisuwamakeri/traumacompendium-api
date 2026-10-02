// Package billingservice implements Premium subscription checkout and
// activation: starting a Paystack transaction, confirming it via a
// signed webhook (re-verified server-side, never trusted from the
// webhook body alone), and the single IsPremium check every
// premium-gated feature elsewhere in this codebase calls.
package billingservice

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog"

	"github.com/traumacompendium/backend/internal/domain/billing"
	"github.com/traumacompendium/backend/internal/domain/user"
	"github.com/traumacompendium/backend/internal/notification/paystack"
)

// PremiumPriceKobo is the brief's flat Premium price (₦2,500/mo), in
// kobo — Paystack amounts are in the currency's smallest unit.
const PremiumPriceKobo = 250_000

// subscriptionPeriod is how long a confirmed payment covers before the
// subscription needs renewing.
const subscriptionPeriod = 30 * 24 * time.Hour

type Service struct {
	subscriptions billing.Repository
	webhookEvents billing.WebhookEventRepository
	users         user.Repository
	paystack      paystack.Client
	paystackKey   string
	appBaseURL    string
	log           zerolog.Logger
	now           func() time.Time
}

func New(
	subscriptions billing.Repository,
	webhookEvents billing.WebhookEventRepository,
	users user.Repository,
	paystackClient paystack.Client,
	paystackSecretKey string,
	appBaseURL string,
	log zerolog.Logger,
) *Service {
	return &Service{
		subscriptions: subscriptions,
		webhookEvents: webhookEvents,
		users:         users,
		paystack:      paystackClient,
		paystackKey:   paystackSecretKey,
		appBaseURL:    appBaseURL,
		log:           log,
		now:           time.Now,
	}
}

// IsPremium is the single gate every premium-only feature elsewhere in
// this codebase (AI assistant, Levels 2-6, certificate download) calls.
// A user who has never subscribed simply isn't premium — that's not an
// error condition.
func (s *Service) IsPremium(ctx context.Context, userID string) (bool, error) {
	sub, err := s.subscriptions.FindByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, billing.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	return sub.IsActivePremium(s.now()), nil
}
