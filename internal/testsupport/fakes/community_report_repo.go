package fakes

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/traumacompendium/backend/internal/domain/community"
)

// CommunityReportRepo is an in-memory community.ReportRepository.
type CommunityReportRepo struct {
	mu   sync.Mutex
	byID map[string]*community.Report
}

func NewCommunityReportRepo() *CommunityReportRepo {
	return &CommunityReportRepo{byID: make(map[string]*community.Report)}
}

func (f *CommunityReportRepo) Create(_ context.Context, r *community.Report) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	cp := *r
	f.byID[r.ID] = &cp
	return nil
}

func (f *CommunityReportRepo) FindByID(_ context.Context, id string) (*community.Report, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.byID[id]
	if !ok {
		return nil, community.ErrReportNotFound
	}
	cp := *r
	return &cp, nil
}

func (f *CommunityReportRepo) ListOpen(_ context.Context, filter community.ListFilter) ([]*community.Report, int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	var matched []*community.Report
	for _, r := range f.byID {
		if r.Status == community.ReportStatusOpen {
			cp := *r
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

func (f *CommunityReportRepo) Resolve(_ context.Context, id, resolvedByUserID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.byID[id]
	if !ok || r.Status != community.ReportStatusOpen {
		return community.ErrReportNotFound
	}
	now := time.Now()
	r.Status = community.ReportStatusResolved
	r.ResolvedByUserID = resolvedByUserID
	r.ResolvedAt = &now
	return nil
}
