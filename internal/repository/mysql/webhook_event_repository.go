package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"

	"github.com/traumacompendium/backend/internal/domain/billing"
)

type WebhookEventRepository struct {
	db *sql.DB
}

func NewWebhookEventRepository(db *sql.DB) *WebhookEventRepository {
	return &WebhookEventRepository{db: db}
}

// MarkProcessed is an insert-first dedupe check: the unique primary key
// on dedupe_key makes a concurrent or retried delivery for the same key
// fail with a duplicate-entry error, which is translated to
// billing.ErrAlreadyProcessed — race-free, unlike a separate
// check-then-insert.
func (r *WebhookEventRepository) MarkProcessed(ctx context.Context, dedupeKey string) error {
	const q = `INSERT INTO webhook_events (dedupe_key, provider, processed_at) VALUES (?, 'paystack', CURRENT_TIMESTAMP)`
	_, err := r.db.ExecContext(ctx, q, dedupeKey)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 { // ER_DUP_ENTRY
			return billing.ErrAlreadyProcessed
		}
		return fmt.Errorf("mysql: mark webhook event processed: %w", err)
	}
	return nil
}
