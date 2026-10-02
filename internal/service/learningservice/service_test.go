package learningservice

import (
	"context"
	"testing"

	"github.com/traumacompendium/backend/internal/pkg/apperror"
)

func TestListLevels_ReturnsSeededLevels(t *testing.T) {
	svc, content, _ := newTestService()
	seedTwoLessonLevel(content)

	levels, err := svc.ListLevels(context.Background())
	if err != nil {
		t.Fatalf("ListLevels() error = %v", err)
	}
	if len(levels) != 1 {
		t.Fatalf("len(levels) = %d, want 1", len(levels))
	}
}

func TestListLessons_UnknownLevel_ReturnsNotFound(t *testing.T) {
	svc, _, _ := newTestService()

	_, err := svc.ListLessons(context.Background(), "nope")
	appErr, ok := apperror.As(err)
	if !ok || appErr.Code != apperror.CodeNotFound {
		t.Errorf("ListLessons() error = %v, want not_found", err)
	}
}

func TestListLessons_ReturnsLessonsForLevel(t *testing.T) {
	svc, content, _ := newTestService()
	levelID, _, _, _, _, _, _ := seedTwoLessonLevel(content)

	lessons, err := svc.ListLessons(context.Background(), levelID)
	if err != nil {
		t.Fatalf("ListLessons() error = %v", err)
	}
	if len(lessons) != 2 {
		t.Fatalf("len(lessons) = %d, want 2", len(lessons))
	}
}

func TestGetLesson_NotFound(t *testing.T) {
	svc, _, _ := newTestService()

	_, err := svc.GetLesson(context.Background(), "nope")
	appErr, ok := apperror.As(err)
	if !ok || appErr.Code != apperror.CodeNotFound {
		t.Errorf("GetLesson() error = %v, want not_found", err)
	}
}
