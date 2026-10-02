package fakes

import (
	"context"
	"sync"

	"github.com/traumacompendium/backend/internal/domain/safereach"
)

// SafeReachContactRepo is an in-memory safereach.TrustedContactRepository.
type SafeReachContactRepo struct {
	mu   sync.Mutex
	byID map[string]*safereach.TrustedContact
}

func NewSafeReachContactRepo() *SafeReachContactRepo {
	return &SafeReachContactRepo{byID: make(map[string]*safereach.TrustedContact)}
}

func (f *SafeReachContactRepo) Create(_ context.Context, c *safereach.TrustedContact) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	cp := *c
	f.byID[c.ID] = &cp
	return nil
}

func (f *SafeReachContactRepo) Delete(_ context.Context, id, userID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	existing, ok := f.byID[id]
	if !ok || existing.UserID != userID {
		return safereach.ErrContactNotFound
	}
	delete(f.byID, id)
	return nil
}

func (f *SafeReachContactRepo) FindByID(_ context.Context, id, userID string) (*safereach.TrustedContact, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	c, ok := f.byID[id]
	if !ok || c.UserID != userID {
		return nil, safereach.ErrContactNotFound
	}
	cp := *c
	return &cp, nil
}

func (f *SafeReachContactRepo) ListByUser(_ context.Context, userID string) ([]*safereach.TrustedContact, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []*safereach.TrustedContact
	for _, c := range f.byID {
		if c.UserID == userID {
			cp := *c
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (f *SafeReachContactRepo) CountByUser(_ context.Context, userID string) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	count := 0
	for _, c := range f.byID {
		if c.UserID == userID {
			count++
		}
	}
	return count, nil
}
