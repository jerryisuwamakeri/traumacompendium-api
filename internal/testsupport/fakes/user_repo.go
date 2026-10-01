// Package fakes provides in-memory implementations of this project's
// repository and notification-sender interfaces, shared across
// service-level and handler-level tests so each new module doesn't
// reinvent the same test doubles. It is imported only from _test.go
// files — cmd/api and every other production code path never reference
// it, so none of this ships in the compiled server binary's reachable
// graph from main.
package fakes

import (
	"context"
	"sync"

	"github.com/traumacompendium/backend/internal/domain/user"
)

// UserRepo is an in-memory user.Repository.
type UserRepo struct {
	mu   sync.Mutex
	byID map[string]*user.User
}

func NewUserRepo() *UserRepo {
	return &UserRepo{byID: make(map[string]*user.User)}
}

func (f *UserRepo) Create(_ context.Context, u *user.User) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, existing := range f.byID {
		if existing.Email == u.Email {
			return user.ErrEmailTaken
		}
		if existing.Phone == u.Phone {
			return user.ErrPhoneTaken
		}
	}
	cp := *u
	f.byID[u.ID] = &cp
	return nil
}

func (f *UserRepo) Update(_ context.Context, u *user.User) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.byID[u.ID]; !ok {
		return user.ErrNotFound
	}
	cp := *u
	f.byID[u.ID] = &cp
	return nil
}

func (f *UserRepo) FindByID(_ context.Context, id string) (*user.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.byID[id]
	if !ok {
		return nil, user.ErrNotFound
	}
	cp := *u
	return &cp, nil
}

func (f *UserRepo) FindByEmail(_ context.Context, email string) (*user.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, u := range f.byID {
		if u.Email == email {
			cp := *u
			return &cp, nil
		}
	}
	return nil, user.ErrNotFound
}

func (f *UserRepo) FindByPhone(_ context.Context, phone string) (*user.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, u := range f.byID {
		if u.Phone == phone {
			cp := *u
			return &cp, nil
		}
	}
	return nil, user.ErrNotFound
}

// Seed inserts users directly, bypassing uniqueness checks — useful for
// handler tests that need a pre-existing account without going through
// the full Register flow.
func (f *UserRepo) Seed(users ...*user.User) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, u := range users {
		cp := *u
		f.byID[u.ID] = &cp
	}
}
