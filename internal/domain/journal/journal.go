// Package journal defines the journaling recovery tool: free-form,
// private entries owned by exactly one user.
package journal

import (
	"context"
	"time"
)

type Entry struct {
	ID        string
	UserID    string
	Title     string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// ListFilter scopes and paginates a journal listing. Zero-value Limit
// means "use the repository's default page size" (see mysql package).
type ListFilter struct {
	From   *time.Time
	To     *time.Time
	Limit  int
	Offset int
}

// Repository persists and retrieves journal entries. Every method is
// scoped by UserID so one user's entries are never reachable through
// another user's requests — ownership is enforced at the query level, not
// just checked after the fact in the service layer.
type Repository interface {
	Create(ctx context.Context, e *Entry) error
	Update(ctx context.Context, e *Entry) error
	SoftDelete(ctx context.Context, id, userID string) error
	FindByID(ctx context.Context, id, userID string) (*Entry, error)
	List(ctx context.Context, userID string, filter ListFilter) ([]*Entry, int, error)
	// ListAll returns every entry for userID, unpaginated, ordered oldest
	// first — used only by the export flow.
	ListAll(ctx context.Context, userID string) ([]*Entry, error)
}

var ErrNotFound = errNotFound{}

type errNotFound struct{}

func (errNotFound) Error() string { return "journal: entry not found" }
