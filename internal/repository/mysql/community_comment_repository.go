package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/traumacompendium/backend/internal/domain/community"
)

type CommunityCommentRepository struct {
	db *sql.DB
}

func NewCommunityCommentRepository(db *sql.DB) *CommunityCommentRepository {
	return &CommunityCommentRepository{db: db}
}

func (r *CommunityCommentRepository) Create(ctx context.Context, c *community.Comment) error {
	const q = `
		INSERT INTO community_comments (id, post_id, user_id, body, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, q, c.ID, c.PostID, c.UserID, c.Body, c.Status, c.CreatedAt)
	if err != nil {
		return fmt.Errorf("mysql: create community comment: %w", err)
	}
	return nil
}

func (r *CommunityCommentRepository) FindByID(ctx context.Context, id string) (*community.Comment, error) {
	const q = `
		SELECT id, post_id, user_id, body, status, created_at, deleted_at
		FROM community_comments
		WHERE id = ? AND deleted_at IS NULL
	`
	var c community.Comment
	err := r.db.QueryRowContext(ctx, q, id).Scan(&c.ID, &c.PostID, &c.UserID, &c.Body, &c.Status, &c.CreatedAt, &c.DeletedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, community.ErrCommentNotFound
		}
		return nil, fmt.Errorf("mysql: find community comment: %w", err)
	}
	return &c, nil
}

func (r *CommunityCommentRepository) SoftDelete(ctx context.Context, id, userID string) error {
	const q = `
		UPDATE community_comments SET deleted_at = CURRENT_TIMESTAMP
		WHERE id = ? AND user_id = ? AND deleted_at IS NULL
	`
	res, err := r.db.ExecContext(ctx, q, id, userID)
	if err != nil {
		return fmt.Errorf("mysql: soft delete community comment: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("mysql: soft delete community comment rows affected: %w", err)
	}
	if rows == 0 {
		return community.ErrCommentNotFound
	}
	return nil
}

func (r *CommunityCommentRepository) ListByPost(ctx context.Context, postID string, filter community.ListFilter) ([]*community.Comment, int, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = defaultCommunityPageSize
	}

	const countQ = `SELECT COUNT(*) FROM community_comments WHERE post_id = ? AND status = ? AND deleted_at IS NULL`
	var total int
	if err := r.db.QueryRowContext(ctx, countQ, postID, community.StatusVisible).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("mysql: count community comments: %w", err)
	}

	const listQ = `
		SELECT id, post_id, user_id, body, status, created_at, deleted_at
		FROM community_comments
		WHERE post_id = ? AND status = ? AND deleted_at IS NULL
		ORDER BY created_at ASC
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.QueryContext(ctx, listQ, postID, community.StatusVisible, limit, filter.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("mysql: list community comments: %w", err)
	}
	defer rows.Close() //nolint:errcheck // read-only cursor, nothing actionable on close failure

	comments := make([]*community.Comment, 0)
	for rows.Next() {
		var c community.Comment
		if err := rows.Scan(&c.ID, &c.PostID, &c.UserID, &c.Body, &c.Status, &c.CreatedAt, &c.DeletedAt); err != nil {
			return nil, 0, fmt.Errorf("mysql: scan community comment: %w", err)
		}
		comments = append(comments, &c)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("mysql: iterate community comments: %w", err)
	}

	return comments, total, nil
}

func (r *CommunityCommentRepository) UpdateStatus(ctx context.Context, id string, status community.Status) error {
	const q = `UPDATE community_comments SET status = ? WHERE id = ? AND deleted_at IS NULL`
	res, err := r.db.ExecContext(ctx, q, status, id)
	if err != nil {
		return fmt.Errorf("mysql: update community comment status: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("mysql: update community comment status rows affected: %w", err)
	}
	if rows == 0 {
		return community.ErrCommentNotFound
	}
	return nil
}
