package mysql

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/traumacompendium/backend/internal/domain/community"
)

func newTestPost(userID string, status community.Status) *community.Post {
	return &community.Post{
		ID: uuid.NewString(), UserID: userID, Body: "Hang in there, you're not alone.",
		Status: status, CreatedAt: time.Now().UTC().Truncate(time.Second),
	}
}

func TestCommunityPostRepository_CreateFindDelete(t *testing.T) {
	db := openTestDB(t)
	userID := seedTestUser(t, db)
	repo := NewCommunityPostRepository(db)
	ctx := context.Background()

	p := newTestPost(userID, community.StatusVisible)
	if err := repo.Create(ctx, p); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := repo.FindByID(ctx, p.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if got.Body != p.Body || got.Status != community.StatusVisible {
		t.Errorf("FindByID() = %+v, want Body=%q Status=visible", got, p.Body)
	}

	if err := repo.SoftDelete(ctx, p.ID, userID); err != nil {
		t.Fatalf("SoftDelete() error = %v", err)
	}
	_, err = repo.FindByID(ctx, p.ID)
	if !errors.Is(err, community.ErrPostNotFound) {
		t.Errorf("FindByID() after delete error = %v, want %v", err, community.ErrPostNotFound)
	}
}

func TestCommunityPostRepository_SoftDelete_WrongUser_NotFound(t *testing.T) {
	db := openTestDB(t)
	owner := seedTestUser(t, db)
	intruder := seedTestUser(t, db)
	repo := NewCommunityPostRepository(db)
	ctx := context.Background()

	p := newTestPost(owner, community.StatusVisible)
	if err := repo.Create(ctx, p); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := repo.SoftDelete(ctx, p.ID, intruder); !errors.Is(err, community.ErrPostNotFound) {
		t.Errorf("SoftDelete() by non-owner error = %v, want %v", err, community.ErrPostNotFound)
	}
}

func TestCommunityPostRepository_List_OnlyVisible_NewestFirst(t *testing.T) {
	db := openTestDB(t)
	userID := seedTestUser(t, db)
	repo := NewCommunityPostRepository(db)
	ctx := context.Background()

	base := time.Now().UTC().Add(-3 * time.Hour).Truncate(time.Second)

	visible1 := newTestPost(userID, community.StatusVisible)
	visible1.CreatedAt = base
	visible2 := newTestPost(userID, community.StatusVisible)
	visible2.CreatedAt = base.Add(time.Hour)
	underReview := newTestPost(userID, community.StatusUnderReview)
	underReview.CreatedAt = base.Add(2 * time.Hour)

	for _, p := range []*community.Post{visible1, visible2, underReview} {
		if err := repo.Create(ctx, p); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	posts, total, err := repo.List(ctx, community.ListFilter{Limit: 10})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2 (under_review excluded)", total)
	}
	if len(posts) != 2 {
		t.Fatalf("len(posts) = %d, want 2", len(posts))
	}
	if posts[0].ID != visible2.ID {
		t.Errorf("posts[0].ID = %q, want the newest visible post %q", posts[0].ID, visible2.ID)
	}
}

func TestCommunityPostRepository_UpdateStatus(t *testing.T) {
	db := openTestDB(t)
	userID := seedTestUser(t, db)
	repo := NewCommunityPostRepository(db)
	ctx := context.Background()

	p := newTestPost(userID, community.StatusVisible)
	if err := repo.Create(ctx, p); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := repo.UpdateStatus(ctx, p.ID, community.StatusRemoved); err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}

	got, err := repo.FindByID(ctx, p.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if got.Status != community.StatusRemoved {
		t.Errorf("Status = %q, want removed", got.Status)
	}
}
