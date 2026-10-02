package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/traumacompendium/backend/internal/domain/journal"
)

const defaultJournalPageSize = 20

type JournalRepository struct {
	db *sql.DB
}

func NewJournalRepository(db *sql.DB) *JournalRepository {
	return &JournalRepository{db: db}
}

func (r *JournalRepository) Create(ctx context.Context, e *journal.Entry) error {
	const q = `
		INSERT INTO journal_entries (id, user_id, title, content, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, q, e.ID, e.UserID, e.Title, e.Content, e.CreatedAt, e.UpdatedAt)
	if err != nil {
		return fmt.Errorf("mysql: create journal entry: %w", err)
	}
	return nil
}

func (r *JournalRepository) Update(ctx context.Context, e *journal.Entry) error {
	const q = `
		UPDATE journal_entries SET title = ?, content = ?, updated_at = ?
		WHERE id = ? AND user_id = ? AND deleted_at IS NULL
	`
	res, err := r.db.ExecContext(ctx, q, e.Title, e.Content, e.UpdatedAt, e.ID, e.UserID)
	if err != nil {
		return fmt.Errorf("mysql: update journal entry: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("mysql: update journal entry rows affected: %w", err)
	}
	if rows == 0 {
		return journal.ErrNotFound
	}
	return nil
}

func (r *JournalRepository) SoftDelete(ctx context.Context, id, userID string) error {
	const q = `
		UPDATE journal_entries SET deleted_at = CURRENT_TIMESTAMP
		WHERE id = ? AND user_id = ? AND deleted_at IS NULL
	`
	res, err := r.db.ExecContext(ctx, q, id, userID)
	if err != nil {
		return fmt.Errorf("mysql: soft delete journal entry: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("mysql: soft delete journal entry rows affected: %w", err)
	}
	if rows == 0 {
		return journal.ErrNotFound
	}
	return nil
}

func (r *JournalRepository) FindByID(ctx context.Context, id, userID string) (*journal.Entry, error) {
	const q = `
		SELECT id, user_id, title, content, created_at, updated_at, deleted_at
		FROM journal_entries
		WHERE id = ? AND user_id = ? AND deleted_at IS NULL
	`
	var e journal.Entry
	err := r.db.QueryRowContext(ctx, q, id, userID).Scan(
		&e.ID, &e.UserID, &e.Title, &e.Content, &e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, journal.ErrNotFound
		}
		return nil, fmt.Errorf("mysql: find journal entry: %w", err)
	}
	return &e, nil
}

func (r *JournalRepository) List(ctx context.Context, userID string, filter journal.ListFilter) ([]*journal.Entry, int, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = defaultJournalPageSize
	}

	where := "WHERE user_id = ? AND deleted_at IS NULL"
	args := []any{userID}

	if filter.From != nil {
		where += " AND created_at >= ?"
		args = append(args, *filter.From)
	}
	if filter.To != nil {
		where += " AND created_at <= ?"
		args = append(args, *filter.To)
	}

	var total int
	countQ := "SELECT COUNT(*) FROM journal_entries " + where
	if err := r.db.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("mysql: count journal entries: %w", err)
	}

	listQ := `
		SELECT id, user_id, title, content, created_at, updated_at, deleted_at
		FROM journal_entries ` + where + `
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	args = append(args, limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, listQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("mysql: list journal entries: %w", err)
	}
	defer rows.Close() //nolint:errcheck // read-only cursor, nothing actionable on close failure

	entries, err := scanJournalEntries(rows)
	if err != nil {
		return nil, 0, err
	}

	return entries, total, nil
}

func (r *JournalRepository) ListAll(ctx context.Context, userID string) ([]*journal.Entry, error) {
	const q = `
		SELECT id, user_id, title, content, created_at, updated_at, deleted_at
		FROM journal_entries
		WHERE user_id = ? AND deleted_at IS NULL
		ORDER BY created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("mysql: list all journal entries: %w", err)
	}
	defer rows.Close() //nolint:errcheck // read-only cursor, nothing actionable on close failure

	return scanJournalEntries(rows)
}

func scanJournalEntries(rows *sql.Rows) ([]*journal.Entry, error) {
	entries := make([]*journal.Entry, 0)
	for rows.Next() {
		var e journal.Entry
		if err := rows.Scan(&e.ID, &e.UserID, &e.Title, &e.Content, &e.CreatedAt, &e.UpdatedAt, &e.DeletedAt); err != nil {
			return nil, fmt.Errorf("mysql: scan journal entry: %w", err)
		}
		entries = append(entries, &e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("mysql: iterate journal entries: %w", err)
	}
	return entries, nil
}
