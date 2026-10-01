package fakes

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/traumacompendium/backend/internal/domain/mood"
)

// MoodRepo is an in-memory mood.Repository.
type MoodRepo struct {
	mu   sync.Mutex
	byID map[string]*mood.Entry
}

func NewMoodRepo() *MoodRepo {
	return &MoodRepo{byID: make(map[string]*mood.Entry)}
}

func (f *MoodRepo) Create(_ context.Context, e *mood.Entry) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	cp := *e
	cp.Tags = append([]string(nil), e.Tags...)
	f.byID[e.ID] = &cp
	return nil
}

func (f *MoodRepo) SoftDelete(_ context.Context, id, userID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	e, ok := f.byID[id]
	if !ok || e.UserID != userID || e.DeletedAt != nil {
		return mood.ErrNotFound
	}
	now := time.Now()
	e.DeletedAt = &now
	return nil
}

func (f *MoodRepo) FindByID(_ context.Context, id, userID string) (*mood.Entry, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	e, ok := f.byID[id]
	if !ok || e.UserID != userID || e.DeletedAt != nil {
		return nil, mood.ErrNotFound
	}
	cp := *e
	return &cp, nil
}

func (f *MoodRepo) List(_ context.Context, userID string, filter mood.ListFilter) ([]*mood.Entry, int, error) {
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

	page := make([]*mood.Entry, len(matched[start:end]))
	for i, e := range matched[start:end] {
		cp := *e
		page[i] = &cp
	}

	return page, total, nil
}

func (f *MoodRepo) ListInRange(_ context.Context, userID string, from, to time.Time) ([]*mood.Entry, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	matched := f.matchForUser(userID, &from, &to)
	sort.Slice(matched, func(i, j int) bool { return matched[i].CreatedAt.Before(matched[j].CreatedAt) })

	out := make([]*mood.Entry, len(matched))
	for i, e := range matched {
		cp := *e
		out[i] = &cp
	}
	return out, nil
}

// matchForUser must be called with f.mu held.
func (f *MoodRepo) matchForUser(userID string, from, to *time.Time) []*mood.Entry {
	var matched []*mood.Entry
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
