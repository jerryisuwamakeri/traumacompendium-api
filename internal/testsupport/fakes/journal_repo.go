package fakes

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/traumacompendium/backend/internal/domain/journal"
)

// JournalRepo is an in-memory journal.Repository.
type JournalRepo struct {
	mu   sync.Mutex
	byID map[string]*journal.Entry
}

func NewJournalRepo() *JournalRepo {
	return &JournalRepo{byID: make(map[string]*journal.Entry)}
}

func (f *JournalRepo) Create(_ context.Context, e *journal.Entry) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	cp := *e
	f.byID[e.ID] = &cp
	return nil
}

func (f *JournalRepo) Update(_ context.Context, e *journal.Entry) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	existing, ok := f.byID[e.ID]
	if !ok || existing.UserID != e.UserID || existing.DeletedAt != nil {
		return journal.ErrNotFound
	}
	cp := *e
	f.byID[e.ID] = &cp
	return nil
}

func (f *JournalRepo) SoftDelete(_ context.Context, id, userID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	existing, ok := f.byID[id]
	if !ok || existing.UserID != userID || existing.DeletedAt != nil {
		return journal.ErrNotFound
	}
	now := time.Now()
	existing.DeletedAt = &now
	return nil
}

func (f *JournalRepo) FindByID(_ context.Context, id, userID string) (*journal.Entry, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	e, ok := f.byID[id]
	if !ok || e.UserID != userID || e.DeletedAt != nil {
		return nil, journal.ErrNotFound
	}
	cp := *e
	return &cp, nil
}

func (f *JournalRepo) List(_ context.Context, userID string, filter journal.ListFilter) ([]*journal.Entry, int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	matched := f.matchForUser(userID, filter.From, filter.To)
	total := len(matched)

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	start := filter.Offset
	if start > len(matched) {
		start = len(matched)
	}
	end := start + limit
	if end > len(matched) {
		end = len(matched)
	}

	page := make([]*journal.Entry, len(matched[start:end]))
	for i, e := range matched[start:end] {
		cp := *e
		page[i] = &cp
	}

	return page, total, nil
}

func (f *JournalRepo) ListAll(_ context.Context, userID string) ([]*journal.Entry, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	matched := f.matchForUser(userID, nil, nil)
	sort.Slice(matched, func(i, j int) bool { return matched[i].CreatedAt.Before(matched[j].CreatedAt) })

	out := make([]*journal.Entry, len(matched))
	for i, e := range matched {
		cp := *e
		out[i] = &cp
	}
	return out, nil
}

// matchForUser must be called with f.mu held.
func (f *JournalRepo) matchForUser(userID string, from, to *time.Time) []*journal.Entry {
	var matched []*journal.Entry
	for _, e := range f.byID {
		if e.UserID != userID || e.DeletedAt != nil {
			continue
		}
		if from != nil && e.CreatedAt.Before(*from) {
			continue
		}
		if to != nil && e.CreatedAt.After(*to) {
			continue
		}
		matched = append(matched, e)
	}
	sort.Slice(matched, func(i, j int) bool { return matched[i].CreatedAt.After(matched[j].CreatedAt) })
	return matched
}
