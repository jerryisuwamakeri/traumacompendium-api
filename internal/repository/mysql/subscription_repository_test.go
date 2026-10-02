package mysql

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/traumacompendium/backend/internal/domain/billing"
)

func TestSubscriptionRepository_FindByUserID_NotFound(t *testing.T) {
	db := openTestDB(t)
	userID := seedTestUser(t, db)
	repo := NewSubscriptionRepository(db)

	_, err := repo.FindByUserID(context.Background(), userID)
	if !errors.Is(err, billing.ErrNotFound) {
		t.Errorf("FindByUserID() for a user who never subscribed error = %v, want %v", err, billing.ErrNotFound)
	}
}

func TestSubscriptionRepository_Upsert_CreateThenRenew(t *testing.T) {
	db := openTestDB(t)
	userID := seedTestUser(t, db)
	repo := NewSubscriptionRepository(db)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Second)
	s := &billing.Subscription{
		ID: uuid.NewString(), UserID: userID, Plan: billing.PlanPremium, Status: billing.StatusActive,
		PaystackReference: "ref-1", CurrentPeriodEnd: now.Add(30 * 24 * time.Hour),
		CreatedAt: now, UpdatedAt: now,
	}
	if err := repo.Upsert(ctx, s); err != nil {
		t.Fatalf("Upsert() create error = %v", err)
	}

	got, err := repo.FindByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("FindByUserID() error = %v", err)
	}
	if got.PaystackReference != "ref-1" || got.Status != billing.StatusActive {
		t.Errorf("FindByUserID() = %+v, want PaystackReference=ref-1 Status=active", got)
	}

	// Renewal: upsert again for the same user with a new reference and
	// a later period end — must replace, not create a second row (the
	// unique key on user_id enforces that; this proves the replace
	// actually took effect).
	renewed := now.Add(60 * 24 * time.Hour)
	s.PaystackReference = "ref-2"
	s.CurrentPeriodEnd = renewed
	s.UpdatedAt = now.Add(time.Hour)
	if err := repo.Upsert(ctx, s); err != nil {
		t.Fatalf("Upsert() renew error = %v", err)
	}

	got, err = repo.FindByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("FindByUserID() after renew error = %v", err)
	}
	if got.PaystackReference != "ref-2" {
		t.Errorf("PaystackReference after renew = %q, want ref-2", got.PaystackReference)
	}
	if !got.CurrentPeriodEnd.Equal(renewed) {
		t.Errorf("CurrentPeriodEnd after renew = %v, want %v", got.CurrentPeriodEnd, renewed)
	}
}
