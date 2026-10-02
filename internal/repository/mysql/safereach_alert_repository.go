package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/traumacompendium/backend/internal/domain/safereach"
)

const defaultSafeReachAlertPageSize = 20

type SafeReachAlertRepository struct {
	db *sql.DB
}

func NewSafeReachAlertRepository(db *sql.DB) *SafeReachAlertRepository {
	return &SafeReachAlertRepository{db: db}
}

func (r *SafeReachAlertRepository) Create(ctx context.Context, a *safereach.Alert) error {
	const q = `
		INSERT INTO safereach_alerts (id, user_id, level, status, message, share_token_hash, triggered_at, resolved_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, q, a.ID, a.UserID, a.Level, a.Status, a.Message, a.ShareTokenHash, a.TriggeredAt, a.ResolvedAt)
	if err != nil {
		return fmt.Errorf("mysql: create safereach alert: %w", err)
	}
	return nil
}

// Update persists Status and ResolvedAt — the only fields that change
// after an alert is created (cancel/resolve). Scoped to the owning user,
// matching how every caller reaches this (via FindOwnedByID first).
func (r *SafeReachAlertRepository) Update(ctx context.Context, a *safereach.Alert) error {
	const q = `
		UPDATE safereach_alerts SET status = ?, resolved_at = ?
		WHERE id = ? AND user_id = ?
	`
	res, err := r.db.ExecContext(ctx, q, a.Status, a.ResolvedAt, a.ID, a.UserID)
	if err != nil {
		return fmt.Errorf("mysql: update safereach alert: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("mysql: update safereach alert rows affected: %w", err)
	}
	if rows == 0 {
		return safereach.ErrAlertNotFound
	}
	return nil
}

func (r *SafeReachAlertRepository) FindByID(ctx context.Context, id string) (*safereach.Alert, error) {
	const q = `
		SELECT id, user_id, level, status, message, share_token_hash, triggered_at, resolved_at
		FROM safereach_alerts
		WHERE id = ?
	`
	return scanSafeReachAlert(r.db.QueryRowContext(ctx, q, id))
}

func (r *SafeReachAlertRepository) FindOwnedByID(ctx context.Context, id, userID string) (*safereach.Alert, error) {
	const q = `
		SELECT id, user_id, level, status, message, share_token_hash, triggered_at, resolved_at
		FROM safereach_alerts
		WHERE id = ? AND user_id = ?
	`
	return scanSafeReachAlert(r.db.QueryRowContext(ctx, q, id, userID))
}

func scanSafeReachAlert(row *sql.Row) (*safereach.Alert, error) {
	var a safereach.Alert
	err := row.Scan(&a.ID, &a.UserID, &a.Level, &a.Status, &a.Message, &a.ShareTokenHash, &a.TriggeredAt, &a.ResolvedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, safereach.ErrAlertNotFound
		}
		return nil, fmt.Errorf("mysql: find safereach alert: %w", err)
	}
	return &a, nil
}

func (r *SafeReachAlertRepository) List(ctx context.Context, userID string, filter safereach.ListFilter) ([]*safereach.Alert, int, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = defaultSafeReachAlertPageSize
	}

	var total int
	const countQ = `SELECT COUNT(*) FROM safereach_alerts WHERE user_id = ?`
	if err := r.db.QueryRowContext(ctx, countQ, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("mysql: count safereach alerts: %w", err)
	}

	const listQ = `
		SELECT id, user_id, level, status, message, share_token_hash, triggered_at, resolved_at
		FROM safereach_alerts
		WHERE user_id = ?
		ORDER BY triggered_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.QueryContext(ctx, listQ, userID, limit, filter.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("mysql: list safereach alerts: %w", err)
	}
	defer rows.Close() //nolint:errcheck // read-only cursor, nothing actionable on close failure

	alerts := make([]*safereach.Alert, 0)
	for rows.Next() {
		var a safereach.Alert
		if err := rows.Scan(&a.ID, &a.UserID, &a.Level, &a.Status, &a.Message, &a.ShareTokenHash, &a.TriggeredAt, &a.ResolvedAt); err != nil {
			return nil, 0, fmt.Errorf("mysql: scan safereach alert: %w", err)
		}
		alerts = append(alerts, &a)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("mysql: iterate safereach alerts: %w", err)
	}

	return alerts, total, nil
}

func (r *SafeReachAlertRepository) AddLocationUpdate(ctx context.Context, u *safereach.LocationUpdate) error {
	const q = `
		INSERT INTO safereach_location_updates (id, alert_id, lat, lng, recorded_at)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, q, u.ID, u.AlertID, u.Lat, u.Lng, u.RecordedAt)
	if err != nil {
		return fmt.Errorf("mysql: add safereach location update: %w", err)
	}
	return nil
}

// LatestLocation returns the most recent fix for alertID, or nil if none
// has been recorded yet — "no location yet" is a normal state for a
// fresh alert, not a not-found error.
func (r *SafeReachAlertRepository) LatestLocation(ctx context.Context, alertID string) (*safereach.LocationUpdate, error) {
	const q = `
		SELECT id, alert_id, lat, lng, recorded_at
		FROM safereach_location_updates
		WHERE alert_id = ?
		ORDER BY recorded_at DESC
		LIMIT 1
	`
	var u safereach.LocationUpdate
	err := r.db.QueryRowContext(ctx, q, alertID).Scan(&u.ID, &u.AlertID, &u.Lat, &u.Lng, &u.RecordedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("mysql: find latest safereach location: %w", err)
	}
	return &u, nil
}
