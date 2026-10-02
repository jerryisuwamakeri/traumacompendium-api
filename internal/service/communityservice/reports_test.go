package communityservice

import (
	"context"
	"testing"

	"github.com/traumacompendium/backend/internal/domain/community"
	"github.com/traumacompendium/backend/internal/pkg/apperror"
)

func TestFileReport_Success(t *testing.T) {
	svc, _, _, _, _ := newTestService()
	ctx := context.Background()

	p, err := svc.CreatePost(ctx, "author", "a post")
	if err != nil {
		t.Fatalf("CreatePost() error = %v", err)
	}

	r, err := svc.FileReport(ctx, "reporter", community.TargetTypePost, p.ID, `<b>This</b> seems harmful`)
	if err != nil {
		t.Fatalf("FileReport() error = %v", err)
	}
	if r.ReporterUserID != "reporter" {
		t.Errorf("ReporterUserID = %q, want reporter", r.ReporterUserID)
	}
	if r.Reason != "This seems harmful" {
		t.Errorf("Reason = %q, want HTML stripped", r.Reason)
	}
	if r.Status != community.ReportStatusOpen {
		t.Errorf("Status = %q, want open", r.Status)
	}
}

func TestFileReport_InvalidTargetType_ReturnsValidationError(t *testing.T) {
	svc, _, _, _, _ := newTestService()

	_, err := svc.FileReport(context.Background(), "reporter", "not-a-type", "some-id", "reason")
	appErr, ok := apperror.As(err)
	if !ok || appErr.Code != apperror.CodeValidation {
		t.Errorf("FileReport() with bad target type error = %v, want validation_error", err)
	}
}

func TestResolveReport_Dismiss_LeavesContentVisible(t *testing.T) {
	svc, posts, _, _, _ := newTestService()
	ctx := context.Background()

	p, err := svc.CreatePost(ctx, "author", "a post")
	if err != nil {
		t.Fatalf("CreatePost() error = %v", err)
	}
	r, err := svc.FileReport(ctx, "reporter", community.TargetTypePost, p.ID, "reason")
	if err != nil {
		t.Fatalf("FileReport() error = %v", err)
	}

	if err := svc.ResolveReport(ctx, "admin", r.ID, ResolveActionDismiss); err != nil {
		t.Fatalf("ResolveReport() error = %v", err)
	}

	got, err := posts.FindByID(ctx, p.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if got.Status != community.StatusVisible {
		t.Errorf("post Status after dismiss = %q, want still visible", got.Status)
	}
}

func TestResolveReport_Remove_HidesContent(t *testing.T) {
	svc, posts, _, _, _ := newTestService()
	ctx := context.Background()

	p, err := svc.CreatePost(ctx, "author", "a post")
	if err != nil {
		t.Fatalf("CreatePost() error = %v", err)
	}
	r, err := svc.FileReport(ctx, "reporter", community.TargetTypePost, p.ID, "reason")
	if err != nil {
		t.Fatalf("FileReport() error = %v", err)
	}

	if err := svc.ResolveReport(ctx, "admin", r.ID, ResolveActionRemove); err != nil {
		t.Fatalf("ResolveReport() error = %v", err)
	}

	got, err := posts.FindByID(ctx, p.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if got.Status != community.StatusRemoved {
		t.Errorf("post Status after remove = %q, want removed", got.Status)
	}

	// The feed must no longer surface it.
	feed, err := svc.ListFeed(ctx, community.ListFilter{Limit: 10})
	if err != nil {
		t.Fatalf("ListFeed() error = %v", err)
	}
	if feed.Total != 0 {
		t.Errorf("feed Total = %d, want 0 after removal", feed.Total)
	}
}

func TestResolveReport_InvalidAction_ReturnsValidationError(t *testing.T) {
	svc, _, _, _, _ := newTestService()
	ctx := context.Background()

	p, err := svc.CreatePost(ctx, "author", "a post")
	if err != nil {
		t.Fatalf("CreatePost() error = %v", err)
	}
	r, err := svc.FileReport(ctx, "reporter", community.TargetTypePost, p.ID, "reason")
	if err != nil {
		t.Fatalf("FileReport() error = %v", err)
	}

	err = svc.ResolveReport(ctx, "admin", r.ID, "not-a-real-action")
	appErr, ok := apperror.As(err)
	if !ok || appErr.Code != apperror.CodeValidation {
		t.Errorf("ResolveReport() with a bad action error = %v, want validation_error", err)
	}
}

func TestResolveReport_AlreadyResolved_ReturnsNotFound(t *testing.T) {
	svc, _, _, _, _ := newTestService()
	ctx := context.Background()

	p, err := svc.CreatePost(ctx, "author", "a post")
	if err != nil {
		t.Fatalf("CreatePost() error = %v", err)
	}
	r, err := svc.FileReport(ctx, "reporter", community.TargetTypePost, p.ID, "reason")
	if err != nil {
		t.Fatalf("FileReport() error = %v", err)
	}
	if err := svc.ResolveReport(ctx, "admin", r.ID, ResolveActionDismiss); err != nil {
		t.Fatalf("ResolveReport() error = %v", err)
	}

	err = svc.ResolveReport(ctx, "admin", r.ID, ResolveActionDismiss)
	appErr, ok := apperror.As(err)
	if !ok || appErr.Code != apperror.CodeNotFound {
		t.Errorf("ResolveReport() a second time error = %v, want not_found", err)
	}
}

func TestListOpenReports_ReturnsFiledReports(t *testing.T) {
	svc, _, _, _, _ := newTestService()
	ctx := context.Background()

	p, err := svc.CreatePost(ctx, "author", "a post")
	if err != nil {
		t.Fatalf("CreatePost() error = %v", err)
	}
	if _, err := svc.FileReport(ctx, "reporter", community.TargetTypePost, p.ID, "reason"); err != nil {
		t.Fatalf("FileReport() error = %v", err)
	}

	result, err := svc.ListOpenReports(ctx, community.ListFilter{Limit: 10})
	if err != nil {
		t.Fatalf("ListOpenReports() error = %v", err)
	}
	if result.Total != 1 {
		t.Errorf("Total = %d, want 1", result.Total)
	}
}
