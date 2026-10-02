package mysql

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/traumacompendium/backend/internal/domain/assistant"
)

func TestAssistantRepository_CreateFindConversation(t *testing.T) {
	db := openTestDB(t)
	userID := seedTestUser(t, db)
	repo := NewAssistantRepository(db)
	ctx := context.Background()

	c := &assistant.Conversation{ID: uuid.NewString(), UserID: userID, CreatedAt: time.Now().UTC().Truncate(time.Second)}
	if err := repo.CreateConversation(ctx, c); err != nil {
		t.Fatalf("CreateConversation() error = %v", err)
	}

	got, err := repo.FindConversation(ctx, c.ID, userID)
	if err != nil {
		t.Fatalf("FindConversation() error = %v", err)
	}
	if got.ID != c.ID {
		t.Errorf("FindConversation() ID = %q, want %q", got.ID, c.ID)
	}
}

func TestAssistantRepository_FindConversation_WrongUser_NotFound(t *testing.T) {
	db := openTestDB(t)
	owner := seedTestUser(t, db)
	intruder := seedTestUser(t, db)
	repo := NewAssistantRepository(db)
	ctx := context.Background()

	c := &assistant.Conversation{ID: uuid.NewString(), UserID: owner, CreatedAt: time.Now().UTC().Truncate(time.Second)}
	if err := repo.CreateConversation(ctx, c); err != nil {
		t.Fatalf("CreateConversation() error = %v", err)
	}

	_, err := repo.FindConversation(ctx, c.ID, intruder)
	if !errors.Is(err, assistant.ErrConversationNotFound) {
		t.Errorf("FindConversation() by non-owner error = %v, want %v", err, assistant.ErrConversationNotFound)
	}
}

func TestAssistantRepository_ListConversations_NewestFirst(t *testing.T) {
	db := openTestDB(t)
	userID := seedTestUser(t, db)
	repo := NewAssistantRepository(db)
	ctx := context.Background()

	base := time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Second)
	for i := 0; i < 3; i++ {
		c := &assistant.Conversation{ID: uuid.NewString(), UserID: userID, CreatedAt: base.Add(time.Duration(i) * time.Hour)}
		if err := repo.CreateConversation(ctx, c); err != nil {
			t.Fatalf("CreateConversation() error = %v", err)
		}
	}

	conversations, total, err := repo.ListConversations(ctx, userID, assistant.ListFilter{Limit: 2})
	if err != nil {
		t.Fatalf("ListConversations() error = %v", err)
	}
	if total != 3 {
		t.Errorf("total = %d, want 3", total)
	}
	if len(conversations) != 2 {
		t.Fatalf("len(conversations) = %d, want 2", len(conversations))
	}
	if !conversations[0].CreatedAt.After(conversations[1].CreatedAt) {
		t.Error("ListConversations() did not return newest-first")
	}
}

func TestAssistantRepository_AddAndListMessages_OldestFirst(t *testing.T) {
	db := openTestDB(t)
	userID := seedTestUser(t, db)
	repo := NewAssistantRepository(db)
	ctx := context.Background()

	c := &assistant.Conversation{ID: uuid.NewString(), UserID: userID, CreatedAt: time.Now().UTC()}
	if err := repo.CreateConversation(ctx, c); err != nil {
		t.Fatalf("CreateConversation() error = %v", err)
	}

	base := time.Now().UTC().Add(-1 * time.Hour).Truncate(time.Second)
	userMsg := &assistant.Message{ID: uuid.NewString(), ConversationID: c.ID, Role: assistant.RoleUser, Content: "What is grounding?", CreatedAt: base}
	assistantMsg := &assistant.Message{ID: uuid.NewString(), ConversationID: c.ID, Role: assistant.RoleAssistant, Content: "Grounding is...", CreatedAt: base.Add(time.Minute)}
	for _, m := range []*assistant.Message{userMsg, assistantMsg} {
		if err := repo.AddMessage(ctx, m); err != nil {
			t.Fatalf("AddMessage() error = %v", err)
		}
	}

	messages, total, err := repo.ListMessages(ctx, c.ID, assistant.ListFilter{Limit: 10})
	if err != nil {
		t.Fatalf("ListMessages() error = %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	if len(messages) != 2 || messages[0].ID != userMsg.ID || messages[1].ID != assistantMsg.ID {
		t.Errorf("ListMessages() = %+v, want [userMsg, assistantMsg] oldest-first", messages)
	}
}

func TestAssistantRepository_RecentMessages_BoundedAndOldestFirst(t *testing.T) {
	db := openTestDB(t)
	userID := seedTestUser(t, db)
	repo := NewAssistantRepository(db)
	ctx := context.Background()

	c := &assistant.Conversation{ID: uuid.NewString(), UserID: userID, CreatedAt: time.Now().UTC()}
	if err := repo.CreateConversation(ctx, c); err != nil {
		t.Fatalf("CreateConversation() error = %v", err)
	}

	base := time.Now().UTC().Add(-5 * time.Minute).Truncate(time.Second)
	var want []string
	for i := 0; i < 5; i++ {
		m := &assistant.Message{
			ID: uuid.NewString(), ConversationID: c.ID, Role: assistant.RoleUser,
			Content: "msg", CreatedAt: base.Add(time.Duration(i) * time.Minute),
		}
		if err := repo.AddMessage(ctx, m); err != nil {
			t.Fatalf("AddMessage() error = %v", err)
		}
		want = append(want, m.ID)
	}
	// Only the 3 most recent, still oldest-first among themselves.
	want = want[2:]

	recent, err := repo.RecentMessages(ctx, c.ID, 3)
	if err != nil {
		t.Fatalf("RecentMessages() error = %v", err)
	}
	if len(recent) != 3 {
		t.Fatalf("len(recent) = %d, want 3", len(recent))
	}
	for i, m := range recent {
		if m.ID != want[i] {
			t.Errorf("recent[%d].ID = %q, want %q", i, m.ID, want[i])
		}
	}
}
