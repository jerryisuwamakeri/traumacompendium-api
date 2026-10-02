package fakes

import (
	"context"
	"sort"
	"sync"

	"github.com/traumacompendium/backend/internal/domain/assistant"
)

// AssistantRepo is an in-memory assistant.Repository.
type AssistantRepo struct {
	mu            sync.Mutex
	conversations map[string]*assistant.Conversation
	messages      map[string][]*assistant.Message // conversationID -> messages, oldest-first
}

func NewAssistantRepo() *AssistantRepo {
	return &AssistantRepo{
		conversations: make(map[string]*assistant.Conversation),
		messages:      make(map[string][]*assistant.Message),
	}
}

func (f *AssistantRepo) CreateConversation(_ context.Context, c *assistant.Conversation) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	cp := *c
	f.conversations[c.ID] = &cp
	return nil
}

func (f *AssistantRepo) FindConversation(_ context.Context, id, userID string) (*assistant.Conversation, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	c, ok := f.conversations[id]
	if !ok || c.UserID != userID {
		return nil, assistant.ErrConversationNotFound
	}
	cp := *c
	return &cp, nil
}

func (f *AssistantRepo) ListConversations(_ context.Context, userID string, filter assistant.ListFilter) ([]*assistant.Conversation, int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	var matched []*assistant.Conversation
	for _, c := range f.conversations {
		if c.UserID == userID {
			cp := *c
			matched = append(matched, &cp)
		}
	}
	sort.Slice(matched, func(i, j int) bool { return matched[i].CreatedAt.After(matched[j].CreatedAt) })
	total := len(matched)

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	start := filter.Offset
	if start > len(matched) {
		start = len(matched)
	}
	end := start + limit
	if end > len(matched) {
		end = len(matched)
	}
	return matched[start:end], total, nil
}

func (f *AssistantRepo) AddMessage(_ context.Context, m *assistant.Message) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	cp := *m
	f.messages[m.ConversationID] = append(f.messages[m.ConversationID], &cp)
	return nil
}

func (f *AssistantRepo) ListMessages(_ context.Context, conversationID string, filter assistant.ListFilter) ([]*assistant.Message, int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	all := f.messages[conversationID]
	total := len(all)

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	start := filter.Offset
	if start > len(all) {
		start = len(all)
	}
	end := start + limit
	if end > len(all) {
		end = len(all)
	}

	out := make([]*assistant.Message, end-start)
	for i, m := range all[start:end] {
		cp := *m
		out[i] = &cp
	}
	return out, total, nil
}

func (f *AssistantRepo) RecentMessages(_ context.Context, conversationID string, limit int) ([]*assistant.Message, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	all := f.messages[conversationID]
	start := len(all) - limit
	if start < 0 {
		start = 0
	}

	out := make([]*assistant.Message, len(all)-start)
	for i, m := range all[start:] {
		cp := *m
		out[i] = &cp
	}
	return out, nil
}
