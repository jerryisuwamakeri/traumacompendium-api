package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type LearningProgressRepository struct {
	db *sql.DB
}

func NewLearningProgressRepository(db *sql.DB) *LearningProgressRepository {
	return &LearningProgressRepository{db: db}
}

// MarkLessonComplete is idempotent: completing an already-completed
// lesson again leaves the original completed_at untouched (the ON
// DUPLICATE KEY clause is a harmless no-op self-assignment, not an
// overwrite) and returns no error.
func (r *LearningProgressRepository) MarkLessonComplete(ctx context.Context, userID, lessonID string, completedAt time.Time) error {
	const q = `
		INSERT INTO user_lesson_progress (user_id, lesson_id, completed_at)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE user_id = user_id
	`
	_, err := r.db.ExecContext(ctx, q, userID, lessonID, completedAt)
	if err != nil {
		return fmt.Errorf("mysql: mark lesson complete: %w", err)
	}
	return nil
}

func (r *LearningProgressRepository) ListCompletedLessonIDs(ctx context.Context, userID string, lessonIDs []string) ([]string, error) {
	if len(lessonIDs) == 0 {
		return []string{}, nil
	}

	placeholders := make([]string, len(lessonIDs))
	args := make([]any, 0, len(lessonIDs)+1)
	args = append(args, userID)
	for i, id := range lessonIDs {
		placeholders[i] = "?"
		args = append(args, id)
	}

	q := `
		SELECT lesson_id FROM user_lesson_progress
		WHERE user_id = ? AND lesson_id IN (` + strings.Join(placeholders, ",") + `)
	`
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("mysql: list completed lesson ids: %w", err)
	}
	defer rows.Close()

	completed := make([]string, 0, len(lessonIDs))
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("mysql: scan completed lesson id: %w", err)
		}
		completed = append(completed, id)
	}
	return completed, rows.Err()
}

// AwardBadge is idempotent for the same reason MarkLessonComplete is.
func (r *LearningProgressRepository) AwardBadge(ctx context.Context, userID, badgeID string, awardedAt time.Time) error {
	const q = `
		INSERT INTO user_badges (user_id, badge_id, awarded_at)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE user_id = user_id
	`
	_, err := r.db.ExecContext(ctx, q, userID, badgeID, awardedAt)
	if err != nil {
		return fmt.Errorf("mysql: award badge: %w", err)
	}
	return nil
}

func (r *LearningProgressRepository) ListUserBadgeIDs(ctx context.Context, userID string) ([]string, error) {
	const q = `SELECT badge_id FROM user_badges WHERE user_id = ? ORDER BY awarded_at ASC`
	rows, err := r.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("mysql: list user badge ids: %w", err)
	}
	defer rows.Close()

	ids := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("mysql: scan user badge id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
