package moodservice

import (
	"context"
	"testing"

	"github.com/traumacompendium/backend/internal/domain/mood"
	"github.com/traumacompendium/backend/internal/pkg/apperror"
	"github.com/traumacompendium/backend/internal/testsupport/fakes"
)

func TestCreate_NormalizesTagsAndPersists(t *testing.T) {
	svc := New(fakes.NewMoodRepo())

	e, err := svc.Create(context.Background(), "user-1", CreateInput{
		Score: 4,
		Tags:  []string{"  Work Stress ", "POOR SLEEP", "", "work stress"},
		Note:  "  rough day  ",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if e.Note != "rough day" {
		t.Errorf("Note = %q, want trimmed %q", e.Note, "rough day")
	}

	want := []string{"work stress", "poor sleep", "work stress"}
	if len(e.Tags) != len(want) {
		t.Fatalf("Tags = %v, want %v", e.Tags, want)
	}
	for i, tag := range want {
		if e.Tags[i] != tag {
			t.Errorf("Tags[%d] = %q, want %q", i, e.Tags[i], tag)
		}
	}
}

func TestCreate_RejectsOutOfRangeScore(t *testing.T) {
	svc := New(fakes.NewMoodRepo())

	for _, score := range []int{0, 6, -1} {
		_, err := svc.Create(context.Background(), "user-1", CreateInput{Score: score})
		appErr, ok := apperror.As(err)
		if !ok || appErr.Code != apperror.CodeValidation {
			t.Errorf("Create(score=%d) error = %v, want validation_error", score, err)
		}
	}
}

func TestGet_WrongUser_ReturnsNotFound(t *testing.T) {
	svc := New(fakes.NewMoodRepo())
	ctx := context.Background()

	e, err := svc.Create(ctx, "owner", CreateInput{Score: 3})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	_, err = svc.Get(ctx, "intruder", e.ID)
	appErr, ok := apperror.As(err)
	if !ok || appErr.Code != apperror.CodeNotFound {
		t.Errorf("Get() by non-owner error = %v, want not_found", err)
	}
}

func TestDelete_RemovesFromSubsequentReads(t *testing.T) {
	svc := New(fakes.NewMoodRepo())
	ctx := context.Background()

	e, err := svc.Create(ctx, "user-1", CreateInput{Score: 2})
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

func TestList_ReturnsOnlyCallersEntries(t *testing.T) {
	svc := New(fakes.NewMoodRepo())
	ctx := context.Background()

	if _, err := svc.Create(ctx, "user-1", CreateInput{Score: 3}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if _, err := svc.Create(ctx, "user-2", CreateInput{Score: 5}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	result, err := svc.List(ctx, "user-1", mood.ListFilter{})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if result.Total != 1 {
		t.Errorf("Total = %d, want 1", result.Total)
	}
}
