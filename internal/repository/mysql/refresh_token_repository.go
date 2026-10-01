package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/traumacompendium/backend/internal/domain/auth"
)

type RefreshTokenRepository struct {
	db *sql.DB
}

func NewRefreshTokenRepository(db *sql.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, rt *auth.RefreshToken) error {
	const q = `
		INSERT INTO refresh_tokens (id, user_id, token_hash, user_agent, ip_address, expires_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, q, rt.ID, rt.UserID, rt.TokenHash, rt.UserAgent, rt.IPAddress, rt.ExpiresAt)
	if err != nil {
		return fmt.Errorf("mysql: create refresh token: %w", err)
	}
	return nil
}

func (r *RefreshTokenRepository) FindByID(ctx context.Context, id string) (*auth.RefreshToken, error) {
	const q = `
		SELECT id, user_id, token_hash, user_agent, ip_address, expires_at, revoked_at, created_at
		FROM refresh_tokens
		WHERE id = ?
	`
	var rt auth.RefreshToken
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&rt.ID, &rt.UserID, &rt.TokenHash, &rt.UserAgent, &rt.IPAddress, &rt.ExpiresAt, &rt.RevokedAt, &rt.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, auth.ErrRefreshTokenNotFound
		}
		return nil, fmt.Errorf("mysql: find refresh token: %w", err)
	}
	return &rt, nil
}

func (r *RefreshTokenRepository) Revoke(ctx context.Context, id string) error {
	const q = `UPDATE refresh_tokens SET revoked_at = CURRENT_TIMESTAMP WHERE id = ? AND revoked_at IS NULL`
	_, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("mysql: revoke refresh token: %w", err)
	}
	return nil
}

// RevokeAllForUser invalidates every active session for a user — used on
// password reset/change so a stolen password can't be paired with a
// still-valid refresh token from before the reset.
func (r *RefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID string) error {
	const q = `UPDATE refresh_tokens SET revoked_at = CURRENT_TIMESTAMP WHERE user_id = ? AND revoked_at IS NULL`
	_, err := r.db.ExecContext(ctx, q, userID)
	if err != nil {
		return fmt.Errorf("mysql: revoke all refresh tokens for user: %w", err)
	}
	return nil
}
