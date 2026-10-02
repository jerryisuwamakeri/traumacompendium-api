package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/traumacompendium/backend/internal/domain/billing"
)

type SubscriptionRepository struct {
	db *sql.DB
}

func NewSubscriptionRepository(db *sql.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

// Upsert inserts a user's first subscription row or replaces their
// existing one — the same idempotent-write pattern already used
// elsewhere in this codebase (MarkLessonComplete, device token
// registration), appropriate here since a user has at most one
// subscription record that subscribe/renew/expire all just overwrite.
func (r *SubscriptionRepository) Upsert(ctx context.Context, s *billing.Subscription) error {
	const q = `
		INSERT INTO subscriptions (id, user_id, plan, status, paystack_reference, current_period_end, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			plan = VALUES(plan),
			status = VALUES(status),
			paystack_reference = VALUES(paystack_reference),
			current_period_end = VALUES(current_period_end),
			updated_at = VALUES(updated_at)
	`
	_, err := r.db.ExecContext(ctx, q,
		s.ID, s.UserID, s.Plan, s.Status, s.PaystackReference, s.CurrentPeriodEnd, s.CreatedAt, s.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("mysql: upsert subscription: %w", err)
	}
	return nil
}

func (r *SubscriptionRepository) FindByUserID(ctx context.Context, userID string) (*billing.Subscription, error) {
	const q = `
		SELECT id, user_id, plan, status, paystack_reference, current_period_end, created_at, updated_at
		FROM subscriptions
		WHERE user_id = ?
	`
	var s billing.Subscription
	err := r.db.QueryRowContext(ctx, q, userID).Scan(
		&s.ID, &s.UserID, &s.Plan, &s.Status, &s.PaystackReference, &s.CurrentPeriodEnd, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, billing.ErrNotFound
		}
		return nil, fmt.Errorf("mysql: find subscription: %w", err)
	}
	return &s, nil
}
