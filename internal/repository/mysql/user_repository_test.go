package mysql

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/traumacompendium/backend/internal/domain/user"
)

func newTestUser() *user.User {
	suffix := uuid.NewString()[:8]
	now := time.Now().UTC().Truncate(time.Second)
	return &user.User{
		ID:           uuid.NewString(),
		FullName:     "Jane Doe",
		Email:        "jane+" + suffix + "@example.com",
		Phone:        "+23480" + suffix,
		PasswordHash: "bcrypt-hash-placeholder",
		Role:         user.RoleUser,
		Status:       user.StatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func TestUserRepository_CreateAndFindByID(t *testing.T) {
	db := openTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	u := newTestUser()
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := repo.FindByID(ctx, u.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if got.Email != u.Email {
		t.Errorf("Email = %q, want %q", got.Email, u.Email)
	}
	if got.Role != user.RoleUser {
		t.Errorf("Role = %q, want %q", got.Role, user.RoleUser)
	}
	if got.IsFullyVerified() {
		t.Error("newly created user should not be fully verified")
	}
}

func TestUserRepository_FindByID_NotFound(t *testing.T) {
	db := openTestDB(t)
	repo := NewUserRepository(db)

	_, err := repo.FindByID(context.Background(), uuid.NewString())
	if !errors.Is(err, user.ErrNotFound) {
		t.Errorf("FindByID() error = %v, want %v", err, user.ErrNotFound)
	}
}

func TestUserRepository_Create_DuplicateEmail(t *testing.T) {
	db := openTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	u1 := newTestUser()
	if err := repo.Create(ctx, u1); err != nil {
		t.Fatalf("Create() first user error = %v", err)
	}

	u2 := newTestUser()
	u2.Email = u1.Email
	err := repo.Create(ctx, u2)
	if !errors.Is(err, user.ErrEmailTaken) {
		t.Errorf("Create() duplicate email error = %v, want %v", err, user.ErrEmailTaken)
	}
}

func TestUserRepository_Update(t *testing.T) {
	db := openTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	u := newTestUser()
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	u.FullName = "Jane Updated"
	if err := repo.Update(ctx, u); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	got, err := repo.FindByID(ctx, u.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if got.FullName != "Jane Updated" {
		t.Errorf("FullName = %q, want %q", got.FullName, "Jane Updated")
	}
}

func TestUserRepository_FindByEmail_IsCaseInsensitive(t *testing.T) {
	db := openTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	u := newTestUser() // stored lowercase, by convention
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := repo.FindByEmail(ctx, strings.ToUpper(u.Email))
	if err != nil {
		t.Fatalf("FindByEmail() with upper-cased lookup error = %v", err)
	}
	if got.ID != u.ID {
		t.Errorf("FindByEmail() returned user %q, want %q", got.ID, u.ID)
	}
}
