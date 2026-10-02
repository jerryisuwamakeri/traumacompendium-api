package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/traumacompendium/backend/internal/domain/safereach"
)

type SafeReachContactRepository struct {
	db *sql.DB
}

func NewSafeReachContactRepository(db *sql.DB) *SafeReachContactRepository {
	return &SafeReachContactRepository{db: db}
}

func (r *SafeReachContactRepository) Create(ctx context.Context, c *safereach.TrustedContact) error {
	const q = `
		INSERT INTO trusted_contacts (id, user_id, name, relationship, phone, email, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, q, c.ID, c.UserID, c.Name, c.Relationship, c.Phone, c.Email, c.CreatedAt)
	if err != nil {
		return fmt.Errorf("mysql: create trusted contact: %w", err)
	}
	return nil
}

func (r *SafeReachContactRepository) Delete(ctx context.Context, id, userID string) error {
	const q = `DELETE FROM trusted_contacts WHERE id = ? AND user_id = ?`
	res, err := r.db.ExecContext(ctx, q, id, userID)
	if err != nil {
		return fmt.Errorf("mysql: delete trusted contact: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("mysql: delete trusted contact rows affected: %w", err)
	}
	if rows == 0 {
		return safereach.ErrContactNotFound
	}
	return nil
}

func (r *SafeReachContactRepository) FindByID(ctx context.Context, id, userID string) (*safereach.TrustedContact, error) {
	const q = `
		SELECT id, user_id, name, relationship, phone, email, created_at
		FROM trusted_contacts
		WHERE id = ? AND user_id = ?
	`
	var c safereach.TrustedContact
	err := r.db.QueryRowContext(ctx, q, id, userID).Scan(
		&c.ID, &c.UserID, &c.Name, &c.Relationship, &c.Phone, &c.Email, &c.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, safereach.ErrContactNotFound
		}
		return nil, fmt.Errorf("mysql: find trusted contact: %w", err)
	}
	return &c, nil
}

func (r *SafeReachContactRepository) ListByUser(ctx context.Context, userID string) ([]*safereach.TrustedContact, error) {
	const q = `
		SELECT id, user_id, name, relationship, phone, email, created_at
		FROM trusted_contacts
		WHERE user_id = ?
		ORDER BY created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("mysql: list trusted contacts: %w", err)
	}
	defer rows.Close() //nolint:errcheck // read-only cursor, nothing actionable on close failure

	contacts := make([]*safereach.TrustedContact, 0)
	for rows.Next() {
		var c safereach.TrustedContact
		if err := rows.Scan(&c.ID, &c.UserID, &c.Name, &c.Relationship, &c.Phone, &c.Email, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("mysql: scan trusted contact: %w", err)
		}
		contacts = append(contacts, &c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("mysql: iterate trusted contacts: %w", err)
	}
	return contacts, nil
}

func (r *SafeReachContactRepository) CountByUser(ctx context.Context, userID string) (int, error) {
	const q = `SELECT COUNT(*) FROM trusted_contacts WHERE user_id = ?`
	var count int
	if err := r.db.QueryRowContext(ctx, q, userID).Scan(&count); err != nil {
		return 0, fmt.Errorf("mysql: count trusted contacts: %w", err)
	}
	return count, nil
}
