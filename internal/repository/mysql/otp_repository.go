package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/traumacompendium/backend/internal/domain/auth"
)

type OTPRepository struct {
	db *sql.DB
}

func NewOTPRepository(db *sql.DB) *OTPRepository {
	return &OTPRepository{db: db}
}

func (r *OTPRepository) Create(ctx context.Context, code *auth.OTPCode) error {
	const q = `
		INSERT INTO otp_codes (id, user_id, purpose, code_hash, expires_at)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, q, code.ID, code.UserID, code.Purpose, code.CodeHash, code.ExpiresAt)
	if err != nil {
		return fmt.Errorf("mysql: create otp code: %w", err)
	}
	return nil
}

// FindActive returns the most recently created, not-yet-consumed code for
// a user+purpose — the one a verify/resend request should act against.
func (r *OTPRepository) FindActive(ctx context.Context, userID string, purpose auth.OTPPurpose) (*auth.OTPCode, error) {
	const q = `
		SELECT id, user_id, purpose, code_hash, expires_at, consumed_at, attempt_count, created_at
		FROM otp_codes
		WHERE user_id = ? AND purpose = ? AND consumed_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`
	var c auth.OTPCode
	err := r.db.QueryRowContext(ctx, q, userID, purpose).Scan(
		&c.ID, &c.UserID, &c.Purpose, &c.CodeHash, &c.ExpiresAt, &c.ConsumedAt, &c.AttemptCount, &c.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, auth.ErrOTPNotFound
		}
		return nil, fmt.Errorf("mysql: find active otp code: %w", err)
	}
	return &c, nil
}

func (r *OTPRepository) IncrementAttempts(ctx context.Context, id string) error {
	const q = `UPDATE otp_codes SET attempt_count = attempt_count + 1 WHERE id = ?`
	_, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("mysql: increment otp attempts: %w", err)
	}
	return nil
}

func (r *OTPRepository) MarkConsumed(ctx context.Context, id string) error {
	const q = `UPDATE otp_codes SET consumed_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("mysql: mark otp consumed: %w", err)
	}
	return nil
}
