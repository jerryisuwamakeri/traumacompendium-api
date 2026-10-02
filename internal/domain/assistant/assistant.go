// Package assistant implements the AI assistant's conversation history:
// a user's conversations and the messages within them. The actual
// reply generation (the constrained system prompt, the crisis-language
// backstop, the call to ai.Client) lives in internal/service/assistantservice
// — this package is just persistence.
package assistant

import (
	"context"
	"time"
)

// Role mirrors who sent a message.
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

type Conversation struct {
	ID        string
	UserID    string
	CreatedAt time.Time
}

type Message struct {
	ID             string
	ConversationID string
	Role           Role
	Content        string
	CreatedAt      time.Time
}

// ListFilter paginates a conversation or message listing.
type ListFilter struct {
	Limit  int
	Offset int
}

// Repository persists conversations and their messages.
type Repository interface {
	CreateConversation(ctx context.Context, c *Conversation) error
	// FindConversation is owner-scoped — a user may only read their own
	// conversations.
	FindConversation(ctx context.Context, id, userID string) (*Conversation, error)
	ListConversations(ctx context.Context, userID string, filter ListFilter) ([]*Conversation, int, error)

	AddMessage(ctx context.Context, m *Message) error
	ListMessages(ctx context.Context, conversationID string, filter ListFilter) ([]*Message, int, error)
	// RecentMessages returns up to limit of a conversation's most recent
	// messages, oldest-first — the bounded context window
	// assistantservice sends the model, so a long-running conversation
	// doesn't grow its prompt without bound.
	RecentMessages(ctx context.Context, conversationID string, limit int) ([]*Message, error)
}

var ErrConversationNotFound = errConversationNotFound{}

type errConversationNotFound struct{}

func (errConversationNotFound) Error() string { return "assistant: conversation not found" }
