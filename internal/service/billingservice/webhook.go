package billingservice

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/traumacompendium/backend/internal/domain/billing"
	"github.com/traumacompendium/backend/internal/notification/paystack"
	"github.com/traumacompendium/backend/internal/pkg/apperror"
	"github.com/traumacompendium/backend/internal/pkg/hash"
)

type paystackWebhookPayload struct {
	Event string `json:"event"`
	Data  struct {
		Reference string `json:"reference"`
	} `json:"data"`
}

// HandleWebhook verifies, dedupes, and acts on a Paystack webhook
// delivery. Verification and dedup both happen before anything else is
// trusted:
//  1. the signature must be valid for rawBody (not trusted from the
//     body's own claims — this is the only thing that proves the
//     request actually came from Paystack);
//  2. the exact delivery (by raw-body hash) must not have been handled
//     before — Paystack retries webhooks, and a retry must not
//     re-extend a subscription a second time.
//
// Even after that, the transaction's success is re-verified directly
// against Paystack's API rather than trusted from the webhook body —
// the webhook says a payment happened, but only Paystack's own
// authoritative verify endpoint confirms it.
func (s *Service) HandleWebhook(ctx context.Context, rawBody []byte, signature string) error {
	if !paystack.VerifyWebhookSignature(s.paystackKey, rawBody, signature) {
		return apperror.NewUnauthorized("invalid webhook signature")
	}

	dedupeKey := hash.Token(string(rawBody))
	if err := s.webhookEvents.MarkProcessed(ctx, dedupeKey); err != nil {
		if errors.Is(err, billing.ErrAlreadyProcessed) {
			return nil // already handled this exact delivery
		}
		return apperror.NewInternal("could not process webhook").WithCause(err)
	}

	var payload paystackWebhookPayload
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		return apperror.NewBadRequest("could not parse webhook payload")
	}

	if payload.Event != "charge.success" {
		return nil // not an event this service acts on
	}
	if payload.Data.Reference == "" {
		return nil
	}

	return s.activateSubscription(ctx, payload.Data.Reference)
}

func (s *Service) activateSubscription(ctx context.Context, reference string) error {
	verify, err := s.paystack.VerifyTransaction(ctx, reference)
	if err != nil {
		return apperror.NewInternal("could not verify transaction").WithCause(err)
	}
	if verify.Status != paystack.TransactionSuccess {
		s.log.Warn().Str("reference", reference).Str("status", string(verify.Status)).
			Msg("billingservice: webhook claimed success but verification did not confirm it")
		return nil
	}

	sub, err := s.subscriptions.FindByReference(ctx, reference)
	if err != nil {
		if errors.Is(err, billing.ErrNotFound) {
			s.log.Warn().Str("reference", reference).Msg("billingservice: verified payment for an unrecognized reference")
			return nil
		}
		return apperror.NewInternal("could not activate subscription").WithCause(err)
	}

	now := s.now()
	sub.Plan = billing.PlanPremium
	sub.Status = billing.StatusActive
	sub.CurrentPeriodEnd = now.Add(subscriptionPeriod)
	sub.UpdatedAt = now

	if err := s.subscriptions.Upsert(ctx, sub); err != nil {
		return apperror.NewInternal("could not activate subscription").WithCause(err)
	}
	return nil
}
