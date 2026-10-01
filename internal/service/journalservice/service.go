// Package journalservice implements the journaling recovery tool:
// create/read/update/soft-delete entries and export a user's full
// history. Every operation is scoped to the acting user — an entry ID
// that belongs to someone else behaves exactly like a nonexistent one.
package journalservice

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/traumacompendium/backend/internal/domain/journal"
	"github.com/traumacompendium/backend/internal/pkg/apperror"
)

type Service struct {
	entries journal.Repository
	now     func() time.Time
}

func New(entries journal.Repository) *Service {
	return &Service{entries: entries, now: time.Now}
}

type CreateInput struct {
	Title   string
	Content string
}

func (s *Service) Create(ctx context.Context, userID string, in CreateInput) (*journal.Entry, error) {
	now := s.now()
	e := &journal.Entry{
		ID:        uuid.NewString(),
		UserID:    userID,
		Title:     strings.TrimSpace(in.Title),
		Content:   in.Content,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.entries.Create(ctx, e); err != nil {
		return nil, apperror.NewInternal("could not create journal entry").WithCause(err)
	}

	return e, nil
}

type UpdateInput struct {
	Title   string
	Content string
}

func (s *Service) Update(ctx context.Context, userID, entryID string, in UpdateInput) (*journal.Entry, error) {
	e, err := s.entries.FindByID(ctx, entryID, userID)
	if err != nil {
		return nil, mapNotFound(err)
	}

	e.Title = strings.TrimSpace(in.Title)
	e.Content = in.Content
	e.UpdatedAt = s.now()

	if err := s.entries.Update(ctx, e); err != nil {
		return nil, mapNotFound(err)
	}

	return e, nil
}

func (s *Service) Delete(ctx context.Context, userID, entryID string) error {
	if err := s.entries.SoftDelete(ctx, entryID, userID); err != nil {
		return mapNotFound(err)
	}
	return nil
}

func (s *Service) Get(ctx context.Context, userID, entryID string) (*journal.Entry, error) {
	e, err := s.entries.FindByID(ctx, entryID, userID)
	if err != nil {
		return nil, mapNotFound(err)
	}
	return e, nil
}

type ListResult struct {
	Entries []*journal.Entry
	Total   int
}

func (s *Service) List(ctx context.Context, userID string, filter journal.ListFilter) (ListResult, error) {
	entries, total, err := s.entries.List(ctx, userID, filter)
	if err != nil {
		return ListResult{}, apperror.NewInternal("could not list journal entries").WithCause(err)
	}
	return ListResult{Entries: entries, Total: total}, nil
}

func mapNotFound(err error) error {
	if errors.Is(err, journal.ErrNotFound) {
		return apperror.NewNotFound("journal entry not found")
	}
	return apperror.NewInternal("could not process journal entry").WithCause(err)
}
