package communityservice

import (
	"context"
	"errors"
	"testing"

	"github.com/rs/zerolog"

	"github.com/traumacompendium/backend/internal/domain/community"
	"github.com/traumacompendium/backend/internal/pkg/apperror"
	"github.com/traumacompendium/backend/internal/testsupport/fakes"
)

var errSimulated = errors.New("communityservice: simulated failure")

func newTestService() (*Service, *fakes.CommunityPostRepo, *fakes.CommunityCommentRepo, *fakes.CommunityReportRepo, *fakes.AIClient) {
	posts := fakes.NewCommunityPostRepo()
	comments := fakes.NewCommunityCommentRepo()
	reports := fakes.NewCommunityReportRepo()
	aiClient := fakes.NewAIClient()
	svc := New(posts, comments, reports, aiClient, zerolog.Nop())
	return svc, posts, comments, reports, aiClient
}

func TestCreatePost_CleanContent_Visible(t *testing.T) {
	svc, _, _, reports, _ := newTestService()

	p, err := svc.CreatePost(context.Background(), "user-1", "Hang in there, you are not alone.")
	if err != nil {
		t.Fatalf("CreatePost() error = %v", err)
	}
	if p.Status != community.StatusVisible {
		t.Errorf("Status = %q, want visible", p.Status)
	}

	result, _, err := reports.ListOpen(context.Background(), community.ListFilter{})
	if err != nil {
		t.Fatalf("ListOpen() error = %v", err)
	}
	if len(result) != 0 {
		t.Errorf("open reports = %+v, want none for clean content", result)
	}
}

func TestCreatePost_StripsHTML(t *testing.T) {
	svc, _, _, _, _ := newTestService()

	p, err := svc.CreatePost(context.Background(), "user-1", `<script>alert(1)</script>Stay strong`)
	if err != nil {
		t.Fatalf("CreatePost() error = %v", err)
	}
	if p.Body != "Stay strong" {
		t.Errorf("Body = %q, want HTML stripped", p.Body)
	}
}

func TestCreatePost_EmptyBody_ReturnsValidationError(t *testing.T) {
	svc, _, _, _, _ := newTestService()

	_, err := svc.CreatePost(context.Background(), "user-1", "   ")
	appErr, ok := apperror.As(err)
	if !ok || appErr.Code != apperror.CodeValidation {
		t.Errorf("CreatePost() with empty body error = %v, want validation_error", err)
	}
}

func TestCreatePost_FlaggedContent_UnderReview_AndFilesSystemReport(t *testing.T) {
	svc, _, _, reports, aiClient := newTestService()
	aiClient.FlaggedBodies["risky content here"] = true
	aiClient.Categories = []string{"self-harm"}
	ctx := context.Background()

	p, err := svc.CreatePost(ctx, "user-1", "risky content here")
	if err != nil {
		t.Fatalf("CreatePost() error = %v", err)
	}
	if p.Status != community.StatusUnderReview {
		t.Errorf("Status = %q, want under_review", p.Status)
	}

	reportsResult, _, err := reports.ListOpen(ctx, community.ListFilter{Limit: 10})
	if err != nil {
		t.Fatalf("ListOpen() error = %v", err)
	}
	if len(reportsResult) != 1 {
		t.Fatalf("len(open reports) = %d, want 1", len(reportsResult))
	}
	if reportsResult[0].TargetID != p.ID || reportsResult[0].TargetType != community.TargetTypePost {
		t.Errorf("report = %+v, want it to target the flagged post", reportsResult[0])
	}
	if reportsResult[0].ReporterUserID != "" {
		t.Errorf("ReporterUserID = %q, want empty for a system-filed report", reportsResult[0].ReporterUserID)
	}
}

func TestCreatePost_ModerationErrors_FailsOpen(t *testing.T) {
	svc, _, _, _, aiClient := newTestService()
	aiClient.ModerateErr = errSimulated

	p, err := svc.CreatePost(context.Background(), "user-1", "some content")
	if err != nil {
		t.Fatalf("CreatePost() error = %v, want it to succeed despite the moderation failure", err)
	}
	if p.Status != community.StatusVisible {
		t.Errorf("Status = %q, want visible (fail open on moderation error)", p.Status)
	}
}

func TestListFeed_OnlyVisiblePosts(t *testing.T) {
	svc, _, _, _, aiClient := newTestService()
	ctx := context.Background()
	aiClient.FlaggedBodies["hidden"] = true

	if _, err := svc.CreatePost(ctx, "user-1", "visible post"); err != nil {
		t.Fatalf("CreatePost() error = %v", err)
	}
	if _, err := svc.CreatePost(ctx, "user-1", "hidden"); err != nil {
		t.Fatalf("CreatePost() error = %v", err)
	}

	result, err := svc.ListFeed(ctx, community.ListFilter{Limit: 10})
	if err != nil {
		t.Fatalf("ListFeed() error = %v", err)
	}
	if result.Total != 1 {
		t.Errorf("Total = %d, want 1 (flagged post excluded)", result.Total)
	}
}

func TestDeletePost_WrongUser_ReturnsNotFound(t *testing.T) {
	svc, _, _, _, _ := newTestService()
	ctx := context.Background()

	p, err := svc.CreatePost(ctx, "owner", "my post")
	if err != nil {
		t.Fatalf("CreatePost() error = %v", err)
	}

	err = svc.DeletePost(ctx, "intruder", p.ID)
	appErr, ok := apperror.As(err)
	if !ok || appErr.Code != apperror.CodeNotFound {
		t.Errorf("DeletePost() by non-owner error = %v, want not_found", err)
	}
}
