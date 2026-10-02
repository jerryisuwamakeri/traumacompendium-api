// Package moodservice implements the mood & trigger tracking recovery
// tool: logging a quick daily entry and summarizing a user's recent
// trend.
package moodservice

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/traumacompendium/backend/internal/domain/mood"
	"github.com/traumacompendium/backend/internal/pkg/apperror"
	"github.com/traumacompendium/backend/internal/pkg/sanitize"
)

type Service struct {
	entries mood.Repository
	now     func() time.Time
}

func New(entries mood.Repository) *Service {
	return &Service{entries: entries, now: time.Now}
}

type CreateInput struct {
	Score int
	Tags  []string
	Note  string
}

func (s *Service) Create(ctx context.Context, userID string, in CreateInput) (*mood.Entry, error) {
	if in.Score < mood.MinScore || in.Score > mood.MaxScore {
		return nil, apperror.NewValidation("score must be between 1 and 5")
	}

	e := &mood.Entry{
		ID:        uuid.NewString(),
		UserID:    userID,
		Score:     in.Score,
		Tags:      normalizeTags(in.Tags),
		Note:      strings.TrimSpace(sanitize.StripHTML(in.Note)),
		CreatedAt: s.now(),
	}

	if err := s.entries.Create(ctx, e); err != nil {
		return nil, apperror.NewInternal("could not log mood entry").WithCause(err)
	}

	return e, nil
}

func (s *Service) Delete(ctx context.Context, userID, entryID string) error {
	if err := s.entries.SoftDelete(ctx, entryID, userID); err != nil {
		return mapNotFound(err)
	}
	return nil
}

func (s *Service) Get(ctx context.Context, userID, entryID string) (*mood.Entry, error) {
	e, err := s.entries.FindByID(ctx, entryID, userID)
	if err != nil {
		return nil, mapNotFound(err)
	}
	return e, nil
}

type ListResult struct {
	Entries []*mood.Entry
	Total   int
}

func (s *Service) List(ctx context.Context, userID string, filter mood.ListFilter) (ListResult, error) {
	entries, total, err := s.entries.List(ctx, userID, filter)
	if err != nil {
		return ListResult{}, apperror.NewInternal("could not list mood entries").WithCause(err)
	}
	return ListResult{Entries: entries, Total: total}, nil
}

// normalizeTags trims whitespace, drops empties, and lower-cases each tag
// so "Work Stress" and "work stress" aggregate together in the summary
// instead of being counted as two different triggers.
func normalizeTags(tags []string) []string {
	out := make([]string, 0, len(tags))
	for _, t := range tags {
		t = strings.ToLower(strings.TrimSpace(sanitize.StripHTML(t)))
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}

func mapNotFound(err error) error {
	if errors.Is(err, mood.ErrNotFound) {
		return apperror.NewNotFound("mood entry not found")
	}
	return apperror.NewInternal("could not process mood entry").WithCause(err)
}
