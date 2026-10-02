package mysql

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/traumacompendium/backend/internal/domain/community"
)

func TestCommunityCommentRepository_CreateFindDelete(t *testing.T) {
	db := openTestDB(t)
	userID := seedTestUser(t, db)
	postRepo := NewCommunityPostRepository(db)
	repo := NewCommunityCommentRepository(db)
	ctx := context.Background()

	p := newTestPost(userID, community.StatusVisible)
	if err := postRepo.Create(ctx, p); err != nil {
		t.Fatalf("create post error = %v", err)
	}

	c := &community.Comment{
		ID: uuid.NewString(), PostID: p.ID, UserID: userID, Body: "Sending strength.",
		Status: community.StatusVisible, CreatedAt: time.Now().UTC().Truncate(time.Second),
	}
	if err := repo.Create(ctx, c); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := repo.FindByID(ctx, c.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if got.Body != c.Body {
		t.Errorf("Body = %q, want %q", got.Body, c.Body)
	}

	if err := repo.SoftDelete(ctx, c.ID, userID); err != nil {
		t.Fatalf("SoftDelete() error = %v", err)
	}
	_, err = repo.FindByID(ctx, c.ID)
	if !errors.Is(err, community.ErrCommentNotFound) {
		t.Errorf("FindByID() after delete error = %v, want %v", err, community.ErrCommentNotFound)
	}
}

func TestCommunityCommentRepository_ListByPost_OnlyVisible_OldestFirst(t *testing.T) {
	db := openTestDB(t)
	userID := seedTestUser(t, db)
	postRepo := NewCommunityPostRepository(db)
	repo := NewCommunityCommentRepository(db)
	ctx := context.Background()

	p := newTestPost(userID, community.StatusVisible)
	if err := postRepo.Create(ctx, p); err != nil {
		t.Fatalf("create post error = %v", err)
	}

	base := time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Second)
	first := &community.Comment{ID: uuid.NewString(), PostID: p.ID, UserID: userID, Body: "First", Status: community.StatusVisible, CreatedAt: base}
	second := &community.Comment{ID: uuid.NewString(), PostID: p.ID, UserID: userID, Body: "Second", Status: community.StatusVisible, CreatedAt: base.Add(time.Hour)}
	flagged := &community.Comment{ID: uuid.NewString(), PostID: p.ID, UserID: userID, Body: "Flagged", Status: community.StatusUnderReview, CreatedAt: base.Add(90 * time.Minute)}

	for _, c := range []*community.Comment{first, second, flagged} {
		if err := repo.Create(ctx, c); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	comments, total, err := repo.ListByPost(ctx, p.ID, community.ListFilter{Limit: 10})
	if err != nil {
		t.Fatalf("ListByPost() error = %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2 (under_review excluded)", total)
	}
	if len(comments) != 2 || comments[0].ID != first.ID || comments[1].ID != second.ID {
		t.Errorf("ListByPost() = %+v, want [First, Second] oldest-first", comments)
	}
}

func TestCommunityCommentRepository_UpdateStatus(t *testing.T) {
	db := openTestDB(t)
	userID := seedTestUser(t, db)
	postRepo := NewCommunityPostRepository(db)
	repo := NewCommunityCommentRepository(db)
	ctx := context.Background()

	p := newTestPost(userID, community.StatusVisible)
	if err := postRepo.Create(ctx, p); err != nil {
		t.Fatalf("create post error = %v", err)
	}
	c := &community.Comment{ID: uuid.NewString(), PostID: p.ID, UserID: userID, Body: "x", Status: community.StatusVisible, CreatedAt: time.Now().UTC()}
	if err := repo.Create(ctx, c); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := repo.UpdateStatus(ctx, c.ID, community.StatusRemoved); err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}
	got, err := repo.FindByID(ctx, c.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if got.Status != community.StatusRemoved {
		t.Errorf("Status = %q, want removed", got.Status)
	}
}
