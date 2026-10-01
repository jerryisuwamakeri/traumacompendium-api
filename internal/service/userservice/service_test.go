package userservice

import (
	"context"
	"testing"

	"github.com/traumacompendium/backend/internal/domain/user"
	"github.com/traumacompendium/backend/internal/pkg/apperror"
)

// fakeUserRepo is a tiny stand-in for user.Repository — intentionally
// re-declared here (rather than shared with authservice's fake) since
// each package's test double should stay free to diverge as each
// service's needs diverge.
type fakeUserRepo struct {
	byID map[string]*user.User
}

func newFakeUserRepo(users ...*user.User) *fakeUserRepo {
	r := &fakeUserRepo{byID: make(map[string]*user.User)}
	for _, u := range users {
		cp := *u
		r.byID[u.ID] = &cp
	}
	return r
}

func (f *fakeUserRepo) Create(_ context.Context, u *user.User) error {
	f.byID[u.ID] = u
	return nil
}

func (f *fakeUserRepo) Update(_ context.Context, u *user.User) error {
	if _, ok := f.byID[u.ID]; !ok {
		return user.ErrNotFound
	}
	cp := *u
	f.byID[u.ID] = &cp
	return nil
}

func (f *fakeUserRepo) FindByID(_ context.Context, id string) (*user.User, error) {
	u, ok := f.byID[id]
	if !ok {
		return nil, user.ErrNotFound
	}
	cp := *u
	return &cp, nil
}

func (f *fakeUserRepo) FindByEmail(context.Context, string) (*user.User, error) {
	return nil, user.ErrNotFound
}

func (f *fakeUserRepo) FindByPhone(context.Context, string) (*user.User, error) {
	return nil, user.ErrNotFound
}

func TestGetProfile_Found(t *testing.T) {
	repo := newFakeUserRepo(&user.User{ID: "u1", FullName: "Jane Doe"})
	svc := New(repo)

	got, err := svc.GetProfile(context.Background(), "u1")
	if err != nil {
		t.Fatalf("GetProfile() error = %v", err)
	}
	if got.FullName != "Jane Doe" {
		t.Errorf("FullName = %q, want %q", got.FullName, "Jane Doe")
	}
}

func TestGetProfile_NotFound(t *testing.T) {
	svc := New(newFakeUserRepo())

	_, err := svc.GetProfile(context.Background(), "missing")
	appErr, ok := apperror.As(err)
	if !ok || appErr.Code != apperror.CodeNotFound {
		t.Errorf("GetProfile() error = %v, want not_found", err)
	}
}

func TestUpdateProfile_ChangesFullNameAndTrims(t *testing.T) {
	repo := newFakeUserRepo(&user.User{ID: "u1", FullName: "Old Name"})
	svc := New(repo)

	got, err := svc.UpdateProfile(context.Background(), "u1", UpdateProfileInput{FullName: "  New Name  "})
	if err != nil {
		t.Fatalf("UpdateProfile() error = %v", err)
	}
	if got.FullName != "New Name" {
		t.Errorf("FullName = %q, want %q", got.FullName, "New Name")
	}

	persisted, _ := repo.FindByID(context.Background(), "u1")
	if persisted.FullName != "New Name" {
		t.Errorf("persisted FullName = %q, want %q", persisted.FullName, "New Name")
	}
}

func TestUpdateProfile_NotFound(t *testing.T) {
	svc := New(newFakeUserRepo())

	_, err := svc.UpdateProfile(context.Background(), "missing", UpdateProfileInput{FullName: "X"})
	appErr, ok := apperror.As(err)
	if !ok || appErr.Code != apperror.CodeNotFound {
		t.Errorf("UpdateProfile() error = %v, want not_found", err)
	}
}
