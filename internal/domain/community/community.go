// Package community implements the peer-support feed: a flat, shared
// feed where users post and comment, backed by AI content moderation
// (flagging risky content into a review queue, see communityservice)
// plus user reporting and admin resolution — the brief's named
// mitigation for "harmful community content."
package community

import (
	"context"
	"time"
)

// Status is shared by Post and Comment: content starts Visible, or
// StatusUnderReview if AI moderation flagged it at creation time, and
// can be moved to StatusRemoved by an admin resolving a report. Only
// Visible content appears in the public feed.
type Status string

const (
	StatusVisible     Status = "visible"
	StatusUnderReview Status = "under_review"
	StatusRemoved     Status = "removed"
)

type Post struct {
	ID        string
	UserID    string
	Body      string
	Status    Status
	CreatedAt time.Time
	DeletedAt *time.Time
}

type Comment struct {
	ID        string
	PostID    string
	UserID    string
	Body      string
	Status    Status
	CreatedAt time.Time
	DeletedAt *time.Time
}

// TargetType is what a Report points at.
type TargetType string

const (
	TargetTypePost    TargetType = "post"
	TargetTypeComment TargetType = "comment"
)

// ReportStatus is a report's resolution state.
type ReportStatus string

const (
	ReportStatusOpen     ReportStatus = "open"
	ReportStatusResolved ReportStatus = "resolved"
)

// Report is a single flag against a post or comment — either
// user-filed (ReporterUserID set, Reason is their free-text explanation)
// or system-filed by AI moderation at creation time (ReporterUserID
// empty, Reason explains which categories were flagged).
type Report struct {
	ID               string
	TargetType       TargetType
	TargetID         string
	ReporterUserID   string // empty for a system-filed (AI moderation) report
	Reason           string
	Status           ReportStatus
	ResolvedByUserID string
	ResolvedAt       *time.Time
	CreatedAt        time.Time
}

// ListFilter paginates a feed/comment/report listing.
type ListFilter struct {
	Limit  int
	Offset int
}

// PostRepository persists posts.
type PostRepository interface {
	Create(ctx context.Context, p *Post) error
	FindByID(ctx context.Context, id string) (*Post, error)
	// SoftDelete is owner-scoped — a user may only delete their own post.
	SoftDelete(ctx context.Context, id, userID string) error
	// List returns only StatusVisible posts, newest first — the public
	// feed. Used for both the feed listing and (filtered to one author
	// elsewhere) nowhere else in Phase 1 scope; kept simple on purpose.
	List(ctx context.Context, filter ListFilter) ([]*Post, int, error)
	UpdateStatus(ctx context.Context, id string, status Status) error
}

// CommentRepository persists comments on a post.
type CommentRepository interface {
	Create(ctx context.Context, c *Comment) error
	FindByID(ctx context.Context, id string) (*Comment, error)
	SoftDelete(ctx context.Context, id, userID string) error
	ListByPost(ctx context.Context, postID string, filter ListFilter) ([]*Comment, int, error)
	UpdateStatus(ctx context.Context, id string, status Status) error
}

// ReportRepository persists moderation reports.
type ReportRepository interface {
	Create(ctx context.Context, r *Report) error
	FindByID(ctx context.Context, id string) (*Report, error)
	// ListOpen is the admin moderation queue.
	ListOpen(ctx context.Context, filter ListFilter) ([]*Report, int, error)
	Resolve(ctx context.Context, id, resolvedByUserID string) error
}

var (
	ErrPostNotFound    = errPostNotFound{}
	ErrCommentNotFound = errCommentNotFound{}
	ErrReportNotFound  = errReportNotFound{}
)

type errPostNotFound struct{}

func (errPostNotFound) Error() string { return "community: post not found" }

type errCommentNotFound struct{}

func (errCommentNotFound) Error() string { return "community: comment not found" }

type errReportNotFound struct{}

func (errReportNotFound) Error() string { return "community: report not found" }
