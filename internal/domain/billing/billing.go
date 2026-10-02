// Package billing implements Premium subscription state. Checkout and
// payment verification live in internal/notification/paystack and
// internal/service/billingservice — this package is just the persisted
// subscription record and the one question everything else asks of it:
// is this user currently premium.
package billing

import (
	"context"
	"time"
)

// Plan is a fixed enum — one paid tier, per the brief's MVP pricing (a
// flat ₦2,500/mo Premium). Enterprise is a manual/sales process, not a
// self-serve plan (see the Phase 2/3 plan's explicit scope note).
type Plan string

const (
	PlanFree    Plan = "free"
	PlanPremium Plan = "premium"
)

type Status string

const (
	// StatusPending is set the moment checkout is initiated (see
	// billingservice.Subscribe) — a reference exists at Paystack, but
	// no payment has been confirmed yet. Never counts as premium.
	StatusPending Status = "pending"
	StatusActive  Status = "active"
	StatusExpired Status = "expired"
)

// Subscription is one user's billing state — at most one row per user.
type Subscription struct {
	ID                string
	UserID            string
	Plan              Plan
	Status            Status
	PaystackReference string
	CurrentPeriodEnd  time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// IsActivePremium is the single check every premium-gated feature in
// this codebase uses (see billingservice.IsPremium).
func (s Subscription) IsActivePremium(now time.Time) bool {
	return s.Plan == PlanPremium && s.Status == StatusActive && now.Before(s.CurrentPeriodEnd)
}

// Repository persists subscription state. Upsert is insert-or-update on
// the unique user_id column — a user has at most one subscription row,
// repeatedly replaced as they subscribe/renew/expire.
type Repository interface {
	Upsert(ctx context.Context, s *Subscription) error
	FindByUserID(ctx context.Context, userID string) (*Subscription, error)
	// FindByReference looks up the subscription checkout started with
	// the given Paystack reference — used by the webhook handler to
	// find which user a payment confirmation belongs to.
	FindByReference(ctx context.Context, reference string) (*Subscription, error)
}

// ErrNotFound is returned by FindByUserID when the user has never
// subscribed — callers treat that the same as an explicit Free plan,
// not an error condition.
var ErrNotFound = errNotFound{}

type errNotFound struct{}

func (errNotFound) Error() string { return "billing: subscription not found" }

// WebhookEventRepository dedupes Paystack webhook deliveries — a
// provider is free to retry a webhook, and this must not double-activate
// or double-extend a subscription on a replay.
type WebhookEventRepository interface {
	// MarkProcessed atomically records dedupeKey as handled. It returns
	// ErrAlreadyProcessed if that key was already recorded, so the
	// caller can tell "first delivery, go ahead" from "replay, skip"
	// without a separate check-then-insert race.
	MarkProcessed(ctx context.Context, dedupeKey string) error
}

var ErrAlreadyProcessed = errAlreadyProcessed{}

type errAlreadyProcessed struct{}

func (errAlreadyProcessed) Error() string { return "billing: webhook event already processed" }
