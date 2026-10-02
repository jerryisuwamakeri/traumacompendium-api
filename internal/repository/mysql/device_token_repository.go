package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/traumacompendium/backend/internal/domain/devicetoken"
)

type DeviceTokenRepository struct {
	db *sql.DB
}

func NewDeviceTokenRepository(db *sql.DB) *DeviceTokenRepository {
	return &DeviceTokenRepository{db: db}
}

// Register upserts on the unique token column: INSERT ... ON DUPLICATE
// KEY UPDATE, the same idempotent-write pattern already used for
// MarkLessonComplete/AwardBadge — re-registering a token is always
// safe, never a conflict.
func (r *DeviceTokenRepository) Register(ctx context.Context, d *devicetoken.DeviceToken) error {
	const q = `
		INSERT INTO device_tokens (id, user_id, token, platform, created_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON DUPLICATE KEY UPDATE user_id = VALUES(user_id), platform = VALUES(platform)
	`
	_, err := r.db.ExecContext(ctx, q, d.ID, d.UserID, d.Token, d.Platform)
	if err != nil {
		return fmt.Errorf("mysql: register device token: %w", err)
	}
	return nil
}

func (r *DeviceTokenRepository) ListByUser(ctx context.Context, userID string) ([]*devicetoken.DeviceToken, error) {
	const q = `SELECT id, user_id, token, platform FROM device_tokens WHERE user_id = ?`
	rows, err := r.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("mysql: list device tokens: %w", err)
	}
	defer rows.Close() //nolint:errcheck // read-only cursor, nothing actionable on close failure

	tokens := make([]*devicetoken.DeviceToken, 0)
	for rows.Next() {
		var d devicetoken.DeviceToken
		if err := rows.Scan(&d.ID, &d.UserID, &d.Token, &d.Platform); err != nil {
			return nil, fmt.Errorf("mysql: scan device token: %w", err)
		}
		tokens = append(tokens, &d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("mysql: iterate device tokens: %w", err)
	}
	return tokens, nil
}
