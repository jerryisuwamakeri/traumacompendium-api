// Package mood defines the mood & trigger tracking recovery tool: a
// quick daily log of how a user is feeling (1-5 score), what triggered
// it (free-form tags), and an optional note.
package mood

import (
	"context"
	"time"
)

const (
	MinScore = 1
	MaxScore = 5
)

type Entry struct {
	ID        string
	UserID    string
	Score     int
	Tags      []string
	Note      string
	CreatedAt time.Time
	DeletedAt *time.Time
}

// ListFilter scopes and paginates a mood listing, mirroring
// journal.ListFilter.
type ListFilter struct {
	From   *time.Time
	To     *time.Time
	Limit  int
	Offset int
}

// Repository persists and retrieves mood entries, scoped by UserID at
// the query level — see journal.Repository's doc comment for why that
// matters.
//
// There is deliberately no Update: a mood entry is a timestamped record
// of how someone felt in the moment, not a document to revise later — if
// a user logs the wrong thing, the product action is "delete and log
// again," not "edit history."
type Repository interface {
	Create(ctx context.Context, e *Entry) error
	SoftDelete(ctx context.Context, id, userID string) error
	FindByID(ctx context.Context, id, userID string) (*Entry, error)
	List(ctx context.Context, userID string, filter ListFilter) ([]*Entry, int, error)
	// ListInRange returns every (non-deleted) entry for userID between
	// from and to, unpaginated — used to compute Summary.
	ListInRange(ctx context.Context, userID string, from, to time.Time) ([]*Entry, error)
}

var ErrNotFound = errNotFound{}

type errNotFound struct{}

func (errNotFound) Error() string { return "mood: entry not found" }
