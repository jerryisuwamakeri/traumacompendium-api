package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/traumacompendium/backend/internal/domain/assistant"
)

const defaultAssistantPageSize = 20

type AssistantRepository struct {
	db *sql.DB
}

func NewAssistantRepository(db *sql.DB) *AssistantRepository {
	return &AssistantRepository{db: db}
}

func (r *AssistantRepository) CreateConversation(ctx context.Context, c *assistant.Conversation) error {
	const q = `INSERT INTO assistant_conversations (id, user_id, created_at) VALUES (?, ?, ?)`
	_, err := r.db.ExecContext(ctx, q, c.ID, c.UserID, c.CreatedAt)
	if err != nil {
		return fmt.Errorf("mysql: create assistant conversation: %w", err)
	}
	return nil
}

func (r *AssistantRepository) FindConversation(ctx context.Context, id, userID string) (*assistant.Conversation, error) {
	const q = `SELECT id, user_id, created_at FROM assistant_conversations WHERE id = ? AND user_id = ?`
	var c assistant.Conversation
	err := r.db.QueryRowContext(ctx, q, id, userID).Scan(&c.ID, &c.UserID, &c.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, assistant.ErrConversationNotFound
		}
		return nil, fmt.Errorf("mysql: find assistant conversation: %w", err)
	}
	return &c, nil
}

func (r *AssistantRepository) ListConversations(ctx context.Context, userID string, filter assistant.ListFilter) ([]*assistant.Conversation, int, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = defaultAssistantPageSize
	}

	const countQ = `SELECT COUNT(*) FROM assistant_conversations WHERE user_id = ?`
	var total int
	if err := r.db.QueryRowContext(ctx, countQ, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("mysql: count assistant conversations: %w", err)
	}

	const listQ = `
		SELECT id, user_id, created_at
		FROM assistant_conversations
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.QueryContext(ctx, listQ, userID, limit, filter.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("mysql: list assistant conversations: %w", err)
	}
	defer rows.Close() //nolint:errcheck // read-only cursor, nothing actionable on close failure

	conversations := make([]*assistant.Conversation, 0)
	for rows.Next() {
		var c assistant.Conversation
		if err := rows.Scan(&c.ID, &c.UserID, &c.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("mysql: scan assistant conversation: %w", err)
		}
		conversations = append(conversations, &c)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("mysql: iterate assistant conversations: %w", err)
	}

	return conversations, total, nil
}

func (r *AssistantRepository) AddMessage(ctx context.Context, m *assistant.Message) error {
	const q = `
		INSERT INTO assistant_messages (id, conversation_id, role, content, created_at)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, q, m.ID, m.ConversationID, m.Role, m.Content, m.CreatedAt)
	if err != nil {
		return fmt.Errorf("mysql: add assistant message: %w", err)
	}
	return nil
}

func (r *AssistantRepository) ListMessages(ctx context.Context, conversationID string, filter assistant.ListFilter) ([]*assistant.Message, int, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = defaultAssistantPageSize
	}

	const countQ = `SELECT COUNT(*) FROM assistant_messages WHERE conversation_id = ?`
	var total int
	if err := r.db.QueryRowContext(ctx, countQ, conversationID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("mysql: count assistant messages: %w", err)
	}

	const listQ = `
		SELECT id, conversation_id, role, content, created_at
		FROM assistant_messages
		WHERE conversation_id = ?
		ORDER BY created_at ASC
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.QueryContext(ctx, listQ, conversationID, limit, filter.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("mysql: list assistant messages: %w", err)
	}
	defer rows.Close() //nolint:errcheck // read-only cursor, nothing actionable on close failure

	messages, err := scanAssistantMessages(rows)
	if err != nil {
		return nil, 0, err
	}
	return messages, total, nil
}

// RecentMessages returns the most recent limit messages, oldest-first —
// queried newest-first (so LIMIT bounds the right end of a long
// conversation) and reversed in Go, rather than reversed in SQL.
func (r *AssistantRepository) RecentMessages(ctx context.Context, conversationID string, limit int) ([]*assistant.Message, error) {
	const q = `
		SELECT id, conversation_id, role, content, created_at
		FROM assistant_messages
		WHERE conversation_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`
	rows, err := r.db.QueryContext(ctx, q, conversationID, limit)
	if err != nil {
		return nil, fmt.Errorf("mysql: list recent assistant messages: %w", err)
	}
	defer rows.Close() //nolint:errcheck // read-only cursor, nothing actionable on close failure

	messages, err := scanAssistantMessages(rows)
	if err != nil {
		return nil, err
	}

	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	return messages, nil
}

func scanAssistantMessages(rows *sql.Rows) ([]*assistant.Message, error) {
	messages := make([]*assistant.Message, 0)
	for rows.Next() {
		var m assistant.Message
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.Role, &m.Content, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("mysql: scan assistant message: %w", err)
		}
		messages = append(messages, &m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("mysql: iterate assistant messages: %w", err)
	}
	return messages, nil
}
