package mysql

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/traumacompendium/backend/internal/domain/community"
)

func TestCommunityReportRepository_CreateFindResolve_UserFiled(t *testing.T) {
	db := openTestDB(t)
	userID := seedTestUser(t, db)
	postRepo := NewCommunityPostRepository(db)
	reporter := seedTestUser(t, db)
	admin := seedTestUser(t, db)
	repo := NewCommunityReportRepository(db)
	ctx := context.Background()

	p := newTestPost(userID, community.StatusVisible)
	if err := postRepo.Create(ctx, p); err != nil {
		t.Fatalf("create post error = %v", err)
	}

	r := &community.Report{
		ID: uuid.NewString(), TargetType: community.TargetTypePost, TargetID: p.ID,
		ReporterUserID: reporter, Reason: "This seems harmful", Status: community.ReportStatusOpen,
		CreatedAt: time.Now().UTC().Truncate(time.Second),
	}
	if err := repo.Create(ctx, r); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := repo.FindByID(ctx, r.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if got.ReporterUserID != reporter {
		t.Errorf("ReporterUserID = %q, want %q", got.ReporterUserID, reporter)
	}
	if got.Status != community.ReportStatusOpen {
		t.Errorf("Status = %q, want open", got.Status)
	}

	if err := repo.Resolve(ctx, r.ID, admin); err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	resolved, err := repo.FindByID(ctx, r.ID)
	if err != nil {
		t.Fatalf("FindByID() after resolve error = %v", err)
	}
	if resolved.Status != community.ReportStatusResolved {
		t.Errorf("Status after resolve = %q, want resolved", resolved.Status)
	}
	if resolved.ResolvedByUserID != admin {
		t.Errorf("ResolvedByUserID = %q, want %q", resolved.ResolvedByUserID, admin)
	}
	if resolved.ResolvedAt == nil {
		t.Error("ResolvedAt = nil, want set")
	}
}

func TestCommunityReportRepository_Create_SystemFiled_NoReporter(t *testing.T) {
	db := openTestDB(t)
	userID := seedTestUser(t, db)
	postRepo := NewCommunityPostRepository(db)
	repo := NewCommunityReportRepository(db)
	ctx := context.Background()

	p := newTestPost(userID, community.StatusUnderReview)
	if err := postRepo.Create(ctx, p); err != nil {
		t.Fatalf("create post error = %v", err)
	}

	r := &community.Report{
		ID: uuid.NewString(), TargetType: community.TargetTypePost, TargetID: p.ID,
		ReporterUserID: "", Reason: "AI moderation flagged: self-harm", Status: community.ReportStatusOpen,
		CreatedAt: time.Now().UTC().Truncate(time.Second),
	}
	if err := repo.Create(ctx, r); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := repo.FindByID(ctx, r.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if got.ReporterUserID != "" {
		t.Errorf("ReporterUserID = %q, want empty for a system-filed report", got.ReporterUserID)
	}
}

func TestCommunityReportRepository_ListOpen_ExcludesResolved(t *testing.T) {
	db := openTestDB(t)
	userID := seedTestUser(t, db)
	postRepo := NewCommunityPostRepository(db)
	admin := seedTestUser(t, db)
	repo := NewCommunityReportRepository(db)
	ctx := context.Background()

	p := newTestPost(userID, community.StatusVisible)
	if err := postRepo.Create(ctx, p); err != nil {
		t.Fatalf("create post error = %v", err)
	}

	open := &community.Report{ID: uuid.NewString(), TargetType: community.TargetTypePost, TargetID: p.ID, Reason: "open one", Status: community.ReportStatusOpen, CreatedAt: time.Now().UTC()}
	toResolve := &community.Report{ID: uuid.NewString(), TargetType: community.TargetTypePost, TargetID: p.ID, Reason: "will be resolved", Status: community.ReportStatusOpen, CreatedAt: time.Now().UTC()}
	for _, r := range []*community.Report{open, toResolve} {
		if err := repo.Create(ctx, r); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}
	if err := repo.Resolve(ctx, toResolve.ID, admin); err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	reports, total, err := repo.ListOpen(ctx, community.ListFilter{Limit: 10})
	if err != nil {
		t.Fatalf("ListOpen() error = %v", err)
	}
	if total != 1 {
		t.Errorf("total = %d, want 1", total)
	}
	if len(reports) != 1 || reports[0].ID != open.ID {
		t.Errorf("ListOpen() = %+v, want just the still-open report", reports)
	}
}

func TestCommunityReportRepository_Resolve_AlreadyResolved_NotFound(t *testing.T) {
	db := openTestDB(t)
	userID := seedTestUser(t, db)
	postRepo := NewCommunityPostRepository(db)
	admin := seedTestUser(t, db)
	repo := NewCommunityReportRepository(db)
	ctx := context.Background()

	p := newTestPost(userID, community.StatusVisible)
	if err := postRepo.Create(ctx, p); err != nil {
		t.Fatalf("create post error = %v", err)
	}
	r := &community.Report{ID: uuid.NewString(), TargetType: community.TargetTypePost, TargetID: p.ID, Reason: "x", Status: community.ReportStatusOpen, CreatedAt: time.Now().UTC()}
	if err := repo.Create(ctx, r); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := repo.Resolve(ctx, r.ID, admin); err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if err := repo.Resolve(ctx, r.ID, admin); !errors.Is(err, community.ErrReportNotFound) {
		t.Errorf("Resolve() a second time error = %v, want %v", err, community.ErrReportNotFound)
	}
}
