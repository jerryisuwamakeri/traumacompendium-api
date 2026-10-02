package fakes

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/traumacompendium/backend/internal/domain/community"
)

// CommunityCommentRepo is an in-memory community.CommentRepository.
type CommunityCommentRepo struct {
	mu   sync.Mutex
	byID map[string]*community.Comment
}

func NewCommunityCommentRepo() *CommunityCommentRepo {
	return &CommunityCommentRepo{byID: make(map[string]*community.Comment)}
}

func (f *CommunityCommentRepo) Create(_ context.Context, c *community.Comment) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	cp := *c
	f.byID[c.ID] = &cp
	return nil
}

func (f *CommunityCommentRepo) FindByID(_ context.Context, id string) (*community.Comment, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	c, ok := f.byID[id]
	if !ok || c.DeletedAt != nil {
		return nil, community.ErrCommentNotFound
	}
	cp := *c
	return &cp, nil
}

func (f *CommunityCommentRepo) SoftDelete(_ context.Context, id, userID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	c, ok := f.byID[id]
	if !ok || c.UserID != userID || c.DeletedAt != nil {
		return community.ErrCommentNotFound
	}
	now := time.Now()
	c.DeletedAt = &now
	return nil
}

func (f *CommunityCommentRepo) ListByPost(_ context.Context, postID string, filter community.ListFilter) ([]*community.Comment, int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	var matched []*community.Comment
	for _, c := range f.byID {
		if c.PostID == postID && c.DeletedAt == nil && c.Status == community.StatusVisible {
			cp := *c
			matched = append(matched, &cp)
		}
	}
	sort.Slice(matched, func(i, j int) bool { return matched[i].CreatedAt.Before(matched[j].CreatedAt) })
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

func (f *CommunityCommentRepo) UpdateStatus(_ context.Context, id string, status community.Status) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	c, ok := f.byID[id]
	if !ok || c.DeletedAt != nil {
		return community.ErrCommentNotFound
	}
	c.Status = status
	return nil
}
