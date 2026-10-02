package journalservice

import (
	"context"
	"testing"

	"github.com/traumacompendium/backend/internal/domain/journal"
	"github.com/traumacompendium/backend/internal/pkg/apperror"
	"github.com/traumacompendium/backend/internal/testsupport/fakes"
)

func TestCreate_TrimsTitleAndPersists(t *testing.T) {
	svc := New(fakes.NewJournalRepo())

	e, err := svc.Create(context.Background(), "user-1", CreateInput{Title: "  A hard day  ", Content: "..."})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if e.Title != "A hard day" {
		t.Errorf("Title = %q, want %q", e.Title, "A hard day")
	}
	if e.UserID != "user-1" {
		t.Errorf("UserID = %q, want %q", e.UserID, "user-1")
	}
}

func TestCreate_StripsHTMLFromTitleAndContent(t *testing.T) {
	svc := New(fakes.NewJournalRepo())

	e, err := svc.Create(context.Background(), "user-1", CreateInput{
		Title:   `<script>alert(1)</script>My day`,
		Content: `Today was <b>hard</b> but okay.`,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if e.Title != "My day" {
		t.Errorf("Title = %q, want HTML stripped", e.Title)
	}
	if e.Content != "Today was hard but okay." {
		t.Errorf("Content = %q, want HTML stripped", e.Content)
	}
}

func TestUpdate_StripsHTMLFromTitleAndContent(t *testing.T) {
	svc := New(fakes.NewJournalRepo())
	ctx := context.Background()

	e, err := svc.Create(ctx, "user-1", CreateInput{Title: "Old", Content: "old"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	updated, err := svc.Update(ctx, "user-1", e.ID, UpdateInput{
		Title:   `<img src=x onerror="alert(1)">New title`,
		Content: "New content",
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Title != "New title" {
		t.Errorf("Title = %q, want HTML stripped", updated.Title)
	}
}

func TestGet_WrongUser_ReturnsNotFound(t *testing.T) {
	svc := New(fakes.NewJournalRepo())
	ctx := context.Background()

	e, err := svc.Create(ctx, "owner", CreateInput{Title: "Private", Content: "..."})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	_, err = svc.Get(ctx, "intruder", e.ID)
	appErr, ok := apperror.As(err)
	if !ok || appErr.Code != apperror.CodeNotFound {
		t.Errorf("Get() by non-owner error = %v, want not_found", err)
	}
}

func TestUpdate_ChangesTitleAndContent(t *testing.T) {
	svc := New(fakes.NewJournalRepo())
	ctx := context.Background()

	e, err := svc.Create(ctx, "user-1", CreateInput{Title: "Old", Content: "old content"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	updated, err := svc.Update(ctx, "user-1", e.ID, UpdateInput{Title: "New", Content: "new content"})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Title != "New" || updated.Content != "new content" {
		t.Errorf("Update() = %+v, want Title=New Content=new content", updated)
	}
}

func TestUpdate_WrongUser_ReturnsNotFound(t *testing.T) {
	svc := New(fakes.NewJournalRepo())
	ctx := context.Background()

	e, err := svc.Create(ctx, "owner", CreateInput{Title: "Private", Content: "..."})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	_, err = svc.Update(ctx, "intruder", e.ID, UpdateInput{Title: "Hacked", Content: "..."})
	appErr, ok := apperror.As(err)
	if !ok || appErr.Code != apperror.CodeNotFound {
		t.Errorf("Update() by non-owner error = %v, want not_found", err)
	}
}

func TestDelete_RemovesFromSubsequentReads(t *testing.T) {
	svc := New(fakes.NewJournalRepo())
	ctx := context.Background()

	e, err := svc.Create(ctx, "user-1", CreateInput{Title: "Temp", Content: "..."})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := svc.Delete(ctx, "user-1", e.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err = svc.Get(ctx, "user-1", e.ID)
	appErr, ok := apperror.As(err)
	if !ok || appErr.Code != apperror.CodeNotFound {
		t.Errorf("Get() after delete error = %v, want not_found", err)
	}
}

func TestDelete_WrongUser_ReturnsNotFound_EntryStillExists(t *testing.T) {
	svc := New(fakes.NewJournalRepo())
	ctx := context.Background()

	e, err := svc.Create(ctx, "owner", CreateInput{Title: "Private", Content: "..."})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	err = svc.Delete(ctx, "intruder", e.ID)
	appErr, ok := apperror.As(err)
	if !ok || appErr.Code != apperror.CodeNotFound {
		t.Errorf("Delete() by non-owner error = %v, want not_found", err)
	}

	// The owner must still be able to read it — the intruder's attempt
	// must not have deleted it.
	if _, err := svc.Get(ctx, "owner", e.ID); err != nil {
		t.Errorf("Get() by owner after a failed intruder delete error = %v, want nil", err)
	}
}

func TestList_ReturnsOnlyCallersEntries(t *testing.T) {
	svc := New(fakes.NewJournalRepo())
	ctx := context.Background()

	if _, err := svc.Create(ctx, "user-1", CreateInput{Title: "A", Content: "."}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if _, err := svc.Create(ctx, "user-2", CreateInput{Title: "B", Content: "."}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	result, err := svc.List(ctx, "user-1", journal.ListFilter{})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if result.Total != 1 {
		t.Errorf("Total = %d, want 1", result.Total)
	}
	if len(result.Entries) != 1 || result.Entries[0].Title != "A" {
		t.Errorf("Entries = %+v, want just user-1's entry", result.Entries)
	}
}
