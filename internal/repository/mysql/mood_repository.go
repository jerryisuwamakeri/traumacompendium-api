package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/traumacompendium/backend/internal/domain/mood"
)

const defaultMoodPageSize = 20

type MoodRepository struct {
	db *sql.DB
}

func NewMoodRepository(db *sql.DB) *MoodRepository {
	return &MoodRepository{db: db}
}

func (r *MoodRepository) Create(ctx context.Context, e *mood.Entry) error {
	tagsJSON, err := json.Marshal(e.Tags)
	if err != nil {
		return fmt.Errorf("mysql: encode mood tags: %w", err)
	}

	const q = `
		INSERT INTO mood_entries (id, user_id, score, tags, note, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err = r.db.ExecContext(ctx, q, e.ID, e.UserID, e.Score, tagsJSON, e.Note, e.CreatedAt)
	if err != nil {
		return fmt.Errorf("mysql: create mood entry: %w", err)
	}
	return nil
}

func (r *MoodRepository) SoftDelete(ctx context.Context, id, userID string) error {
	const q = `
		UPDATE mood_entries SET deleted_at = CURRENT_TIMESTAMP
		WHERE id = ? AND user_id = ? AND deleted_at IS NULL
	`
	res, err := r.db.ExecContext(ctx, q, id, userID)
	if err != nil {
		return fmt.Errorf("mysql: soft delete mood entry: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("mysql: soft delete mood entry rows affected: %w", err)
	}
	if rows == 0 {
		return mood.ErrNotFound
	}
	return nil
}

func (r *MoodRepository) FindByID(ctx context.Context, id, userID string) (*mood.Entry, error) {
	const q = `
		SELECT id, user_id, score, tags, note, created_at, deleted_at
		FROM mood_entries
		WHERE id = ? AND user_id = ? AND deleted_at IS NULL
	`
	return scanMoodEntry(r.db.QueryRowContext(ctx, q, id, userID))
}

func (r *MoodRepository) List(ctx context.Context, userID string, filter mood.ListFilter) ([]*mood.Entry, int, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = defaultMoodPageSize
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
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM mood_entries "+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("mysql: count mood entries: %w", err)
	}

	listQ := `
		SELECT id, user_id, score, tags, note, created_at, deleted_at
		FROM mood_entries ` + where + `
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	args = append(args, limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, listQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("mysql: list mood entries: %w", err)
	}
	defer rows.Close()

	entries, err := scanMoodEntries(rows)
	if err != nil {
		return nil, 0, err
	}

	return entries, total, nil
}

func (r *MoodRepository) ListInRange(ctx context.Context, userID string, from, to time.Time) ([]*mood.Entry, error) {
	const q = `
		SELECT id, user_id, score, tags, note, created_at, deleted_at
		FROM mood_entries
		WHERE user_id = ? AND deleted_at IS NULL AND created_at >= ? AND created_at <= ?
		ORDER BY created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, q, userID, from, to)
	if err != nil {
		return nil, fmt.Errorf("mysql: list mood entries in range: %w", err)
	}
	defer rows.Close()

	return scanMoodEntries(rows)
}

func scanMoodEntry(row interface{ Scan(dest ...any) error }) (*mood.Entry, error) {
	var e mood.Entry
	var tagsJSON []byte

	err := row.Scan(&e.ID, &e.UserID, &e.Score, &tagsJSON, &e.Note, &e.CreatedAt, &e.DeletedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, mood.ErrNotFound
		}
		return nil, fmt.Errorf("mysql: scan mood entry: %w", err)
	}

	if err := json.Unmarshal(tagsJSON, &e.Tags); err != nil {
		return nil, fmt.Errorf("mysql: decode mood tags: %w", err)
	}

	return &e, nil
}

func scanMoodEntries(rows *sql.Rows) ([]*mood.Entry, error) {
	entries := make([]*mood.Entry, 0)
	for rows.Next() {
		e, err := scanMoodEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("mysql: iterate mood entries: %w", err)
	}
	return entries, nil
}
