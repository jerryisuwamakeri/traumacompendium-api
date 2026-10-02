package mysql

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/traumacompendium/backend/internal/domain/safereach"
)

func TestSafeReachContactRepository_CreateFindListDelete(t *testing.T) {
	db := openTestDB(t)
	userID := seedTestUser(t, db)
	repo := NewSafeReachContactRepository(db)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Second)
	c := &safereach.TrustedContact{
		ID:           uuid.NewString(),
		UserID:       userID,
		Name:         "Ada Okafor",
		Relationship: "Sister",
		Phone:        "+2348012345678",
		Email:        "ada@example.com",
		CreatedAt:    now,
	}

	if err := repo.Create(ctx, c); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := repo.FindByID(ctx, c.ID, userID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if got.Name != c.Name || got.Phone != c.Phone {
		t.Errorf("FindByID() = %+v, want Name=%q Phone=%q", got, c.Name, c.Phone)
	}

	list, err := repo.ListByUser(ctx, userID)
	if err != nil {
		t.Fatalf("ListByUser() error = %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("len(ListByUser()) = %d, want 1", len(list))
	}

	count, err := repo.CountByUser(ctx, userID)
	if err != nil {
		t.Fatalf("CountByUser() error = %v", err)
	}
	if count != 1 {
		t.Errorf("CountByUser() = %d, want 1", count)
	}

	if err := repo.Delete(ctx, c.ID, userID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	_, err = repo.FindByID(ctx, c.ID, userID)
	if !errors.Is(err, safereach.ErrContactNotFound) {
		t.Errorf("FindByID() after delete error = %v, want %v", err, safereach.ErrContactNotFound)
	}
}

func TestSafeReachContactRepository_FindByID_WrongUser_NotFound(t *testing.T) {
	db := openTestDB(t)
	owner := seedTestUser(t, db)
	intruder := seedTestUser(t, db)
	repo := NewSafeReachContactRepository(db)
	ctx := context.Background()

	c := &safereach.TrustedContact{
		ID: uuid.NewString(), UserID: owner, Name: "Owner's Contact", Relationship: "Friend",
		Phone: "+2348011111111", CreatedAt: time.Now().UTC().Truncate(time.Second),
	}
	if err := repo.Create(ctx, c); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	_, err := repo.FindByID(ctx, c.ID, intruder)
	if !errors.Is(err, safereach.ErrContactNotFound) {
		t.Errorf("FindByID() by non-owner error = %v, want %v", err, safereach.ErrContactNotFound)
	}

	if err := repo.Delete(ctx, c.ID, intruder); !errors.Is(err, safereach.ErrContactNotFound) {
		t.Errorf("Delete() by non-owner error = %v, want %v", err, safereach.ErrContactNotFound)
	}
}

func TestSafeReachContactRepository_CountByUser_OnlyCountsOwnContacts(t *testing.T) {
	db := openTestDB(t)
	userA := seedTestUser(t, db)
	userB := seedTestUser(t, db)
	repo := NewSafeReachContactRepository(db)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	for i := 0; i < 3; i++ {
		c := &safereach.TrustedContact{ID: uuid.NewString(), UserID: userA, Name: "A", Relationship: "Friend", Phone: "+234800000000" + string(rune('0'+i)), CreatedAt: now}
		if err := repo.Create(ctx, c); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}
	c := &safereach.TrustedContact{ID: uuid.NewString(), UserID: userB, Name: "B", Relationship: "Friend", Phone: "+2348099999999", CreatedAt: now}
	if err := repo.Create(ctx, c); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	count, err := repo.CountByUser(ctx, userA)
	if err != nil {
		t.Fatalf("CountByUser() error = %v", err)
	}
	if count != 3 {
		t.Errorf("CountByUser(userA) = %d, want 3", count)
	}
}
