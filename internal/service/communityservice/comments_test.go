package communityservice

import (
	"context"
	"testing"

	"github.com/traumacompendium/backend/internal/domain/community"
	"github.com/traumacompendium/backend/internal/pkg/apperror"
)

func TestCreateComment_Success(t *testing.T) {
	svc, _, _, _, _ := newTestService()
	ctx := context.Background()

	p, err := svc.CreatePost(ctx, "author", "original post")
	if err != nil {
		t.Fatalf("CreatePost() error = %v", err)
	}

	c, err := svc.CreateComment(ctx, "commenter", p.ID, `<b>Sending</b> strength`)
	if err != nil {
		t.Fatalf("CreateComment() error = %v", err)
	}
	if c.Body != "Sending strength" {
		t.Errorf("Body = %q, want HTML stripped", c.Body)
	}
	if c.Status != community.StatusVisible {
		t.Errorf("Status = %q, want visible", c.Status)
	}
}

func TestCreateComment_OnNonexistentPost_ReturnsNotFound(t *testing.T) {
	svc, _, _, _, _ := newTestService()

	_, err := svc.CreateComment(context.Background(), "user-1", "does-not-exist", "hello")
	appErr, ok := apperror.As(err)
	if !ok || appErr.Code != apperror.CodeNotFound {
		t.Errorf("CreateComment() on a nonexistent post error = %v, want not_found", err)
	}
}

func TestCreateComment_OnRemovedPost_ReturnsNotFound(t *testing.T) {
	svc, posts, _, _, _ := newTestService()
	ctx := context.Background()

	p, err := svc.CreatePost(ctx, "author", "original post")
	if err != nil {
		t.Fatalf("CreatePost() error = %v", err)
	}
	if err := posts.UpdateStatus(ctx, p.ID, community.StatusRemoved); err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}

	_, err = svc.CreateComment(ctx, "user-1", p.ID, "hello")
	appErr, ok := apperror.As(err)
	if !ok || appErr.Code != apperror.CodeNotFound {
		t.Errorf("CreateComment() on a removed post error = %v, want not_found", err)
	}
}

func TestCreateComment_Flagged_UnderReview_AndFilesSystemReport(t *testing.T) {
	svc, _, _, reports, aiClient := newTestService()
	ctx := context.Background()
	aiClient.FlaggedBodies["risky comment"] = true

	p, err := svc.CreatePost(ctx, "author", "original post")
	if err != nil {
		t.Fatalf("CreatePost() error = %v", err)
	}
	c, err := svc.CreateComment(ctx, "commenter", p.ID, "risky comment")
	if err != nil {
		t.Fatalf("CreateComment() error = %v", err)
	}
	if c.Status != community.StatusUnderReview {
		t.Errorf("Status = %q, want under_review", c.Status)
	}

	open, _, err := reports.ListOpen(ctx, community.ListFilter{Limit: 10})
	if err != nil {
		t.Fatalf("ListOpen() error = %v", err)
	}
	if len(open) != 1 || open[0].TargetType != community.TargetTypeComment || open[0].TargetID != c.ID {
		t.Errorf("open reports = %+v, want one targeting the flagged comment", open)
	}
}

func TestListComments_OnlyVisible(t *testing.T) {
	svc, _, _, _, aiClient := newTestService()
	ctx := context.Background()
	aiClient.FlaggedBodies["hidden comment"] = true

	p, err := svc.CreatePost(ctx, "author", "original post")
	if err != nil {
		t.Fatalf("CreatePost() error = %v", err)
	}
	if _, err := svc.CreateComment(ctx, "user-1", p.ID, "visible comment"); err != nil {
		t.Fatalf("CreateComment() error = %v", err)
	}
	if _, err := svc.CreateComment(ctx, "user-1", p.ID, "hidden comment"); err != nil {
		t.Fatalf("CreateComment() error = %v", err)
	}

	result, err := svc.ListComments(ctx, p.ID, community.ListFilter{Limit: 10})
	if err != nil {
		t.Fatalf("ListComments() error = %v", err)
	}
	if result.Total != 1 {
		t.Errorf("Total = %d, want 1 (flagged comment excluded)", result.Total)
	}
}

func TestDeleteComment_WrongUser_ReturnsNotFound(t *testing.T) {
	svc, _, _, _, _ := newTestService()
	ctx := context.Background()

	p, err := svc.CreatePost(ctx, "author", "original post")
	if err != nil {
		t.Fatalf("CreatePost() error = %v", err)
	}
	c, err := svc.CreateComment(ctx, "owner", p.ID, "my comment")
	if err != nil {
		t.Fatalf("CreateComment() error = %v", err)
	}

	err = svc.DeleteComment(ctx, "intruder", c.ID)
	appErr, ok := apperror.As(err)
	if !ok || appErr.Code != apperror.CodeNotFound {
		t.Errorf("DeleteComment() by non-owner error = %v, want not_found", err)
	}
}
