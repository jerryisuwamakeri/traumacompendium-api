package mysql

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/traumacompendium/backend/internal/domain/journal"
)

func TestJournalRepository_CreateFindUpdateDelete(t *testing.T) {
	db := openTestDB(t)
	userID := seedTestUser(t, db)
	repo := NewJournalRepository(db)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Second)
	e := &journal.Entry{
		ID:        uuid.NewString(),
		UserID:    userID,
		Title:     "A hard day",
		Content:   "Today was difficult, but I used a grounding technique.",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := repo.Create(ctx, e); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := repo.FindByID(ctx, e.ID, userID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if got.Title != e.Title {
		t.Errorf("Title = %q, want %q", got.Title, e.Title)
	}

	e.Title = "A hard day (updated)"
	e.UpdatedAt = time.Now().UTC().Truncate(time.Second)
	if err := repo.Update(ctx, e); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	got, _ = repo.FindByID(ctx, e.ID, userID)
	if got.Title != "A hard day (updated)" {
		t.Errorf("Title after update = %q, want %q", got.Title, "A hard day (updated)")
	}

	if err := repo.SoftDelete(ctx, e.ID, userID); err != nil {
		t.Fatalf("SoftDelete() error = %v", err)
	}
	_, err = repo.FindByID(ctx, e.ID, userID)
	if !errors.Is(err, journal.ErrNotFound) {
		t.Errorf("FindByID() after delete error = %v, want %v", err, journal.ErrNotFound)
	}
}

func TestJournalRepository_FindByID_WrongUser_NotFound(t *testing.T) {
	db := openTestDB(t)
	owner := seedTestUser(t, db)
	intruder := seedTestUser(t, db)
	repo := NewJournalRepository(db)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Second)
	e := &journal.Entry{ID: uuid.NewString(), UserID: owner, Title: "Private", Content: "...", CreatedAt: now, UpdatedAt: now}
	if err := repo.Create(ctx, e); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	_, err := repo.FindByID(ctx, e.ID, intruder)
	if !errors.Is(err, journal.ErrNotFound) {
		t.Errorf("FindByID() by non-owner error = %v, want %v", err, journal.ErrNotFound)
	}
}

func TestJournalRepository_List_PaginatesAndFilters(t *testing.T) {
	db := openTestDB(t)
	userID := seedTestUser(t, db)
	repo := NewJournalRepository(db)
	ctx := context.Background()

	base := time.Now().UTC().Add(-24 * time.Hour).Truncate(time.Second)
	for i := 0; i < 5; i++ {
		ts := base.Add(time.Duration(i) * time.Hour)
		e := &journal.Entry{
			ID: uuid.NewString(), UserID: userID,
			Title: "Entry", Content: "...",
			CreatedAt: ts, UpdatedAt: ts,
		}
		if err := repo.Create(ctx, e); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	entries, total, err := repo.List(ctx, userID, journal.ListFilter{Limit: 2, Offset: 0})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if total != 5 {
		t.Errorf("total = %d, want 5", total)
	}
	if len(entries) != 2 {
		t.Errorf("len(entries) = %d, want 2", len(entries))
	}

	from := base.Add(2 * time.Hour)
	entries, total, err = repo.List(ctx, userID, journal.ListFilter{From: &from, Limit: 10})
	if err != nil {
		t.Fatalf("List() with From filter error = %v", err)
	}
	if total != 3 {
		t.Errorf("total with From filter = %d, want 3", total)
	}
	if len(entries) != 3 {
		t.Errorf("len(entries) with From filter = %d, want 3", len(entries))
	}
}

func TestJournalRepository_ListAll_ReturnsEverythingOldestFirst(t *testing.T) {
	db := openTestDB(t)
	userID := seedTestUser(t, db)
	repo := NewJournalRepository(db)
	ctx := context.Background()

	base := time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Second)
	for i := 0; i < 3; i++ {
		ts := base.Add(time.Duration(i) * time.Hour)
		e := &journal.Entry{ID: uuid.NewString(), UserID: userID, Title: "E", Content: "C", CreatedAt: ts, UpdatedAt: ts}
		if err := repo.Create(ctx, e); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	all, err := repo.ListAll(ctx, userID)
	if err != nil {
		t.Fatalf("ListAll() error = %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("len(all) = %d, want 3", len(all))
	}
	if !all[0].CreatedAt.Before(all[1].CreatedAt) || !all[1].CreatedAt.Before(all[2].CreatedAt) {
		t.Error("ListAll() did not return entries oldest-first")
	}
}
