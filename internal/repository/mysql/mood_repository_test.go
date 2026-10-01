package mysql

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/traumacompendium/backend/internal/domain/mood"
)

func TestMoodRepository_CreateFindDelete(t *testing.T) {
	db := openTestDB(t)
	userID := seedTestUser(t, db)
	repo := NewMoodRepository(db)
	ctx := context.Background()

	e := &mood.Entry{
		ID:        uuid.NewString(),
		UserID:    userID,
		Score:     4,
		Tags:      []string{"work stress", "poor sleep"},
		Note:      "Rough morning but recovered.",
		CreatedAt: time.Now().UTC().Truncate(time.Second),
	}

	if err := repo.Create(ctx, e); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := repo.FindByID(ctx, e.ID, userID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if got.Score != 4 {
		t.Errorf("Score = %d, want 4", got.Score)
	}
	if len(got.Tags) != 2 || got.Tags[0] != "work stress" {
		t.Errorf("Tags = %v, want [work stress poor sleep]", got.Tags)
	}

	if err := repo.SoftDelete(ctx, e.ID, userID); err != nil {
		t.Fatalf("SoftDelete() error = %v", err)
	}
	_, err = repo.FindByID(ctx, e.ID, userID)
	if !errors.Is(err, mood.ErrNotFound) {
		t.Errorf("FindByID() after delete error = %v, want %v", err, mood.ErrNotFound)
	}
}

func TestMoodRepository_FindByID_WrongUser_NotFound(t *testing.T) {
	db := openTestDB(t)
	owner := seedTestUser(t, db)
	intruder := seedTestUser(t, db)
	repo := NewMoodRepository(db)
	ctx := context.Background()

	e := &mood.Entry{ID: uuid.NewString(), UserID: owner, Score: 3, Tags: []string{}, CreatedAt: time.Now()}
	if err := repo.Create(ctx, e); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	_, err := repo.FindByID(ctx, e.ID, intruder)
	if !errors.Is(err, mood.ErrNotFound) {
		t.Errorf("FindByID() by non-owner error = %v, want %v", err, mood.ErrNotFound)
	}
}

func TestMoodRepository_List_Paginates(t *testing.T) {
	db := openTestDB(t)
	userID := seedTestUser(t, db)
	repo := NewMoodRepository(db)
	ctx := context.Background()

	base := time.Now().UTC().Add(-5 * time.Hour).Truncate(time.Second)
	for i := 0; i < 4; i++ {
		e := &mood.Entry{
			ID: uuid.NewString(), UserID: userID, Score: 3, Tags: []string{},
			CreatedAt: base.Add(time.Duration(i) * time.Hour),
		}
		if err := repo.Create(ctx, e); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	entries, total, err := repo.List(ctx, userID, mood.ListFilter{Limit: 2})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if total != 4 {
		t.Errorf("total = %d, want 4", total)
	}
	if len(entries) != 2 {
		t.Errorf("len(entries) = %d, want 2", len(entries))
	}
}

func TestMoodRepository_ListInRange(t *testing.T) {
	db := openTestDB(t)
	userID := seedTestUser(t, db)
	repo := NewMoodRepository(db)
	ctx := context.Background()

	base := time.Now().UTC().Add(-10 * time.Hour).Truncate(time.Second)
	for i := 0; i < 5; i++ {
		e := &mood.Entry{
			ID: uuid.NewString(), UserID: userID, Score: i%5 + 1, Tags: []string{"tag"},
			CreatedAt: base.Add(time.Duration(i) * time.Hour),
		}
		if err := repo.Create(ctx, e); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	from := base.Add(1 * time.Hour)
	to := base.Add(3 * time.Hour)
	entries, err := repo.ListInRange(ctx, userID, from, to)
	if err != nil {
		t.Fatalf("ListInRange() error = %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("len(entries) = %d, want 3", len(entries))
	}
	if entries[0].CreatedAt.After(entries[len(entries)-1].CreatedAt) {
		t.Error("ListInRange() did not return entries in ascending order")
	}
}
