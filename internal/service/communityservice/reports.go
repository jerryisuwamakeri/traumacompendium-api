package communityservice

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	"github.com/traumacompendium/backend/internal/domain/community"
	"github.com/traumacompendium/backend/internal/pkg/apperror"
	"github.com/traumacompendium/backend/internal/pkg/sanitize"
)

// FileReport is the user-facing "report this" action — a new
// Report row distinct from fileSystemReport (which handles moderation's
// own, reporter-less reports); both land in the same admin queue.
func (s *Service) FileReport(ctx context.Context, reporterUserID string, targetType community.TargetType, targetID, reason string) (*community.Report, error) {
	if targetType != community.TargetTypePost && targetType != community.TargetTypeComment {
		return nil, apperror.NewValidation("target_type must be post or comment")
	}
	reason = strings.TrimSpace(sanitize.StripHTML(reason))

	r := &community.Report{
		ID:             uuid.NewString(),
		TargetType:     targetType,
		TargetID:       targetID,
		ReporterUserID: reporterUserID,
		Reason:         reason,
		Status:         community.ReportStatusOpen,
		CreatedAt:      s.now(),
	}
	if err := s.reports.Create(ctx, r); err != nil {
		return nil, apperror.NewInternal("could not file report").WithCause(err)
	}
	return r, nil
}

type ListReportsResult struct {
	Reports []*community.Report
	Total   int
}

// ListOpenReports is the admin moderation queue.
func (s *Service) ListOpenReports(ctx context.Context, filter community.ListFilter) (ListReportsResult, error) {
	reports, total, err := s.reports.ListOpen(ctx, filter)
	if err != nil {
		return ListReportsResult{}, apperror.NewInternal("could not list reports").WithCause(err)
	}
	return ListReportsResult{Reports: reports, Total: total}, nil
}

// ResolveAction is an admin's decision on an open report.
type ResolveAction string

const (
	// ResolveActionDismiss closes the report without touching the
	// reported content — the admin judged it not to be a problem.
	ResolveActionDismiss ResolveAction = "dismiss"
	// ResolveActionRemove closes the report and hides the reported
	// content from the feed (Status -> removed).
	ResolveActionRemove ResolveAction = "remove"
)

func (s *Service) ResolveReport(ctx context.Context, adminUserID, reportID string, action ResolveAction) error {
	if action != ResolveActionDismiss && action != ResolveActionRemove {
		return apperror.NewValidation("action must be dismiss or remove")
	}

	report, err := s.reports.FindByID(ctx, reportID)
	if err != nil {
		return mapReportNotFound(err)
	}

	if action == ResolveActionRemove {
		if err := s.removeTarget(ctx, report.TargetType, report.TargetID); err != nil {
			return err
		}
	}

	if err := s.reports.Resolve(ctx, reportID, adminUserID); err != nil {
		return mapReportNotFound(err)
	}
	return nil
}

func (s *Service) removeTarget(ctx context.Context, targetType community.TargetType, targetID string) error {
	var err error
	switch targetType {
	case community.TargetTypePost:
		err = s.posts.UpdateStatus(ctx, targetID, community.StatusRemoved)
	case community.TargetTypeComment:
		err = s.comments.UpdateStatus(ctx, targetID, community.StatusRemoved)
	}
	if err != nil && !errors.Is(err, community.ErrPostNotFound) && !errors.Is(err, community.ErrCommentNotFound) {
		return apperror.NewInternal("could not remove reported content").WithCause(err)
	}
	// A target that's already gone (deleted by its author in the
	// meantime) isn't an error here — the report is still resolvable;
	// there's simply nothing left to hide.
	return nil
}

func mapReportNotFound(err error) error {
	if errors.Is(err, community.ErrReportNotFound) {
		return apperror.NewNotFound("report not found")
	}
	return apperror.NewInternal("could not process report").WithCause(err)
}
