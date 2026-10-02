package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/traumacompendium/backend/internal/domain/community"
)

const defaultCommunityPageSize = 20

type CommunityPostRepository struct {
	db *sql.DB
}

func NewCommunityPostRepository(db *sql.DB) *CommunityPostRepository {
	return &CommunityPostRepository{db: db}
}

func (r *CommunityPostRepository) Create(ctx context.Context, p *community.Post) error {
	const q = `
		INSERT INTO community_posts (id, user_id, body, status, created_at)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, q, p.ID, p.UserID, p.Body, p.Status, p.CreatedAt)
	if err != nil {
		return fmt.Errorf("mysql: create community post: %w", err)
	}
	return nil
}

func (r *CommunityPostRepository) FindByID(ctx context.Context, id string) (*community.Post, error) {
	const q = `
		SELECT id, user_id, body, status, created_at, deleted_at
		FROM community_posts
		WHERE id = ? AND deleted_at IS NULL
	`
	var p community.Post
	err := r.db.QueryRowContext(ctx, q, id).Scan(&p.ID, &p.UserID, &p.Body, &p.Status, &p.CreatedAt, &p.DeletedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, community.ErrPostNotFound
		}
		return nil, fmt.Errorf("mysql: find community post: %w", err)
	}
	return &p, nil
}

func (r *CommunityPostRepository) SoftDelete(ctx context.Context, id, userID string) error {
	const q = `
		UPDATE community_posts SET deleted_at = CURRENT_TIMESTAMP
		WHERE id = ? AND user_id = ? AND deleted_at IS NULL
	`
	res, err := r.db.ExecContext(ctx, q, id, userID)
	if err != nil {
		return fmt.Errorf("mysql: soft delete community post: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("mysql: soft delete community post rows affected: %w", err)
	}
	if rows == 0 {
		return community.ErrPostNotFound
	}
	return nil
}

func (r *CommunityPostRepository) List(ctx context.Context, filter community.ListFilter) ([]*community.Post, int, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = defaultCommunityPageSize
	}

	const countQ = `SELECT COUNT(*) FROM community_posts WHERE status = ? AND deleted_at IS NULL`
	var total int
	if err := r.db.QueryRowContext(ctx, countQ, community.StatusVisible).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("mysql: count community posts: %w", err)
	}

	const listQ = `
		SELECT id, user_id, body, status, created_at, deleted_at
		FROM community_posts
		WHERE status = ? AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.QueryContext(ctx, listQ, community.StatusVisible, limit, filter.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("mysql: list community posts: %w", err)
	}
	defer rows.Close() //nolint:errcheck // read-only cursor, nothing actionable on close failure

	posts := make([]*community.Post, 0)
	for rows.Next() {
		var p community.Post
		if err := rows.Scan(&p.ID, &p.UserID, &p.Body, &p.Status, &p.CreatedAt, &p.DeletedAt); err != nil {
			return nil, 0, fmt.Errorf("mysql: scan community post: %w", err)
		}
		posts = append(posts, &p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("mysql: iterate community posts: %w", err)
	}

	return posts, total, nil
}

func (r *CommunityPostRepository) UpdateStatus(ctx context.Context, id string, status community.Status) error {
	const q = `UPDATE community_posts SET status = ? WHERE id = ? AND deleted_at IS NULL`
	res, err := r.db.ExecContext(ctx, q, status, id)
	if err != nil {
		return fmt.Errorf("mysql: update community post status: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("mysql: update community post status rows affected: %w", err)
	}
	if rows == 0 {
		return community.ErrPostNotFound
	}
	return nil
}
