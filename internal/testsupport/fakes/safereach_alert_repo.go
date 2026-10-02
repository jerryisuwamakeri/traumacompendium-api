package fakes

import (
	"context"
	"sort"
	"sync"

	"github.com/traumacompendium/backend/internal/domain/safereach"
)

// SafeReachAlertRepo is an in-memory safereach.AlertRepository.
type SafeReachAlertRepo struct {
	mu        sync.Mutex
	byID      map[string]*safereach.Alert
	locations map[string][]*safereach.LocationUpdate // alertID -> updates, oldest first
}

func NewSafeReachAlertRepo() *SafeReachAlertRepo {
	return &SafeReachAlertRepo{
		byID:      make(map[string]*safereach.Alert),
		locations: make(map[string][]*safereach.LocationUpdate),
	}
}

func (f *SafeReachAlertRepo) Create(_ context.Context, a *safereach.Alert) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	cp := *a
	f.byID[a.ID] = &cp
	return nil
}

func (f *SafeReachAlertRepo) Update(_ context.Context, a *safereach.Alert) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	existing, ok := f.byID[a.ID]
	if !ok || existing.UserID != a.UserID {
		return safereach.ErrAlertNotFound
	}
	cp := *a
	f.byID[a.ID] = &cp
	return nil
}

func (f *SafeReachAlertRepo) FindByID(_ context.Context, id string) (*safereach.Alert, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	a, ok := f.byID[id]
	if !ok {
		return nil, safereach.ErrAlertNotFound
	}
	cp := *a
	return &cp, nil
}

func (f *SafeReachAlertRepo) FindOwnedByID(_ context.Context, id, userID string) (*safereach.Alert, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	a, ok := f.byID[id]
	if !ok || a.UserID != userID {
		return nil, safereach.ErrAlertNotFound
	}
	cp := *a
	return &cp, nil
}

func (f *SafeReachAlertRepo) List(_ context.Context, userID string, filter safereach.ListFilter) ([]*safereach.Alert, int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	var matched []*safereach.Alert
	for _, a := range f.byID {
		if a.UserID == userID {
			cp := *a
			matched = append(matched, &cp)
		}
	}
	sort.Slice(matched, func(i, j int) bool { return matched[i].TriggeredAt.After(matched[j].TriggeredAt) })
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

	return matched[start:end], total, nil
}

func (f *SafeReachAlertRepo) AddLocationUpdate(_ context.Context, u *safereach.LocationUpdate) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	cp := *u
	f.locations[u.AlertID] = append(f.locations[u.AlertID], &cp)
	return nil
}

func (f *SafeReachAlertRepo) LatestLocation(_ context.Context, alertID string) (*safereach.LocationUpdate, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	updates := f.locations[alertID]
	if len(updates) == 0 {
		return nil, nil
	}
	cp := *updates[len(updates)-1]
	return &cp, nil
}
