// Package communityservice implements the peer-support feed: posting,
// commenting, and the moderation pipeline that sits in front of both.
//
// Every post/comment body is sanitized (internal/pkg/sanitize) and then
// run through ai.Client.Moderate before it's stored. Flagged content is
// still created — just with Status = under_review instead of visible —
// and a system-filed Report is created alongside it, landing in the
// same admin queue a user-filed report would. If the moderation call
// itself fails (provider down, timeout), content publishes normally:
// this fails open rather than blocking the whole feature on one
// external call, the same posture RateLimit and Idempotency already
// take on a Redis error elsewhere in this codebase.
package communityservice

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/traumacompendium/backend/internal/domain/community"
	"github.com/traumacompendium/backend/internal/notification/ai"
)

type Service struct {
	posts    community.PostRepository
	comments community.CommentRepository
	reports  community.ReportRepository
	ai       ai.Client
	log      zerolog.Logger
	now      func() time.Time
}

func New(
	posts community.PostRepository,
	comments community.CommentRepository,
	reports community.ReportRepository,
	aiClient ai.Client,
	log zerolog.Logger,
) *Service {
	return &Service{posts: posts, comments: comments, reports: reports, ai: aiClient, log: log, now: time.Now}
}

// classify runs body through moderation and returns the status the new
// content should be created with, plus a reason string (empty unless
// flagged) for the accompanying system report.
func (s *Service) classify(ctx context.Context, body string) (community.Status, string) {
	result, err := s.ai.Moderate(ctx, body)
	if err != nil {
		s.log.Warn().Err(err).Msg("communityservice: moderation check failed, publishing without review")
		return community.StatusVisible, ""
	}
	if !result.Flagged {
		return community.StatusVisible, ""
	}

	reason := "AI moderation flagged this content"
	if len(result.Categories) > 0 {
		reason += ": " + strings.Join(result.Categories, ", ")
	}
	return community.StatusUnderReview, reason
}

// fileSystemReport records a system-filed report (empty ReporterUserID)
// for content that moderation flagged at creation time. Failures here
// are logged, not returned — the content itself was already created
// successfully, and losing the review-queue entry shouldn't roll that
// back or fail the request the user is waiting on.
func (s *Service) fileSystemReport(ctx context.Context, targetType community.TargetType, targetID, reason string) {
	r := &community.Report{
		ID:         uuid.NewString(),
		TargetType: targetType,
		TargetID:   targetID,
		Reason:     reason,
		Status:     community.ReportStatusOpen,
		CreatedAt:  s.now(),
	}
	if err := s.reports.Create(ctx, r); err != nil {
		s.log.Error().Err(err).Str("target_id", targetID).Msg("communityservice: failed to file system moderation report")
	}
}
