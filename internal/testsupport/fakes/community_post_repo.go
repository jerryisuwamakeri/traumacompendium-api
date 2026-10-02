package fakes

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/traumacompendium/backend/internal/domain/community"
)

// CommunityPostRepo is an in-memory community.PostRepository.
type CommunityPostRepo struct {
	mu   sync.Mutex
	byID map[string]*community.Post
}

func NewCommunityPostRepo() *CommunityPostRepo {
	return &CommunityPostRepo{byID: make(map[string]*community.Post)}
}

func (f *CommunityPostRepo) Create(_ context.Context, p *community.Post) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	cp := *p
	f.byID[p.ID] = &cp
	return nil
}

func (f *CommunityPostRepo) FindByID(_ context.Context, id string) (*community.Post, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	p, ok := f.byID[id]
	if !ok || p.DeletedAt != nil {
		return nil, community.ErrPostNotFound
	}
	cp := *p
	return &cp, nil
}

func (f *CommunityPostRepo) SoftDelete(_ context.Context, id, userID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	p, ok := f.byID[id]
	if !ok || p.UserID != userID || p.DeletedAt != nil {
		return community.ErrPostNotFound
	}
	now := time.Now()
	p.DeletedAt = &now
	return nil
}

func (f *CommunityPostRepo) List(_ context.Context, filter community.ListFilter) ([]*community.Post, int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	var matched []*community.Post
	for _, p := range f.byID {
		if p.DeletedAt == nil && p.Status == community.StatusVisible {
			cp := *p
			matched = append(matched, &cp)
		}
	}
	sort.Slice(matched, func(i, j int) bool { return matched[i].CreatedAt.After(matched[j].CreatedAt) })
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

func (f *CommunityPostRepo) UpdateStatus(_ context.Context, id string, status community.Status) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	p, ok := f.byID[id]
	if !ok || p.DeletedAt != nil {
		return community.ErrPostNotFound
	}
	p.Status = status
	return nil
}
