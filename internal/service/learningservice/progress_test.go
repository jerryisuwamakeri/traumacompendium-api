package learningservice

import (
	"context"
	"testing"
)

func TestCompleteLesson_FirstOfTwo_LevelNotYetComplete_NoBadge(t *testing.T) {
	svc, content, _ := newTestService()
	_, lesson1ID, _, _, _, _, _ := seedTwoLessonLevel(content)

	result, err := svc.CompleteLesson(context.Background(), "user-1", lesson1ID)
	if err != nil {
		t.Fatalf("CompleteLesson() error = %v", err)
	}
	if result.LevelCompleted {
		t.Error("LevelCompleted = true, want false (only 1 of 2 lessons done)")
	}
	if result.BadgeAwarded != nil {
		t.Errorf("BadgeAwarded = %+v, want nil", result.BadgeAwarded)
	}
}

func TestCompleteLesson_LastOfLevel_AwardsBadge(t *testing.T) {
	svc, content, _ := newTestService()
	_, lesson1ID, lesson2ID, _, _, _, _ := seedTwoLessonLevel(content)
	ctx := context.Background()

	if _, err := svc.CompleteLesson(ctx, "user-1", lesson1ID); err != nil {
		t.Fatalf("CompleteLesson(lesson1) error = %v", err)
	}

	result, err := svc.CompleteLesson(ctx, "user-1", lesson2ID)
	if err != nil {
		t.Fatalf("CompleteLesson(lesson2) error = %v", err)
	}
	if !result.LevelCompleted {
		t.Error("LevelCompleted = false, want true (both lessons done)")
	}
	if result.BadgeAwarded == nil {
		t.Fatal("BadgeAwarded = nil, want the level's badge")
	}
	if result.BadgeAwarded.Name != "Beginner" {
		t.Errorf("BadgeAwarded.Name = %q, want %q", result.BadgeAwarded.Name, "Beginner")
	}
}

func TestCompleteLesson_ReCompletingAfterBadgeEarned_DoesNotReportBadgeAgain(t *testing.T) {
	// A user re-visiting an already-completed lesson (or re-submitting
	// completion) must not see the "you earned a badge!" celebration a
	// second time.
	svc, content, _ := newTestService()
	_, lesson1ID, lesson2ID, _, _, _, _ := seedTwoLessonLevel(content)
	ctx := context.Background()

	if _, err := svc.CompleteLesson(ctx, "user-1", lesson1ID); err != nil {
		t.Fatalf("CompleteLesson(lesson1) error = %v", err)
	}
	if _, err := svc.CompleteLesson(ctx, "user-1", lesson2ID); err != nil {
		t.Fatalf("CompleteLesson(lesson2) error = %v", err)
	}

	// Re-complete lesson2 (level already fully done and badge already
	// awarded).
	result, err := svc.CompleteLesson(ctx, "user-1", lesson2ID)
	if err != nil {
		t.Fatalf("CompleteLesson(lesson2, again) error = %v", err)
	}
	if !result.LevelCompleted {
		t.Error("LevelCompleted = false, want true")
	}
	if result.BadgeAwarded != nil {
		t.Errorf("BadgeAwarded = %+v, want nil on a re-completion (badge already owned)", result.BadgeAwarded)
	}
}

func TestCompleteLesson_DoesNotAwardBadgeToOtherUsers(t *testing.T) {
	svc, content, _ := newTestService()
	_, lesson1ID, lesson2ID, _, _, _, _ := seedTwoLessonLevel(content)
	ctx := context.Background()

	if _, err := svc.CompleteLesson(ctx, "user-1", lesson1ID); err != nil {
		t.Fatalf("CompleteLesson() error = %v", err)
	}
	if _, err := svc.CompleteLesson(ctx, "user-1", lesson2ID); err != nil {
		t.Fatalf("CompleteLesson() error = %v", err)
	}

	badges, err := svc.ListUserBadges(ctx, "user-2")
	if err != nil {
		t.Fatalf("ListUserBadges() error = %v", err)
	}
	if len(badges) != 0 {
		t.Errorf("user-2's badges = %+v, want none (user-1 earned the badge, not user-2)", badges)
	}
}

func TestGetProgress_ReflectsPartialAndFullCompletion(t *testing.T) {
	svc, content, _ := newTestService()
	_, lesson1ID, _, _, _, _, _ := seedTwoLessonLevel(content)
	ctx := context.Background()

	if _, err := svc.CompleteLesson(ctx, "user-1", lesson1ID); err != nil {
		t.Fatalf("CompleteLesson() error = %v", err)
	}

	progress, err := svc.GetProgress(ctx, "user-1")
	if err != nil {
		t.Fatalf("GetProgress() error = %v", err)
	}
	if len(progress) != 1 {
		t.Fatalf("len(progress) = %d, want 1", len(progress))
	}
	if progress[0].TotalLessons != 2 {
		t.Errorf("TotalLessons = %d, want 2", progress[0].TotalLessons)
	}
	if progress[0].CompletedLessons != 1 {
		t.Errorf("CompletedLessons = %d, want 1", progress[0].CompletedLessons)
	}
	if progress[0].BadgeEarned {
		t.Error("BadgeEarned = true, want false (level not yet fully completed)")
	}
}

func TestListUserBadges_ReturnsEarnedBadges(t *testing.T) {
	svc, content, _ := newTestService()
	_, lesson1ID, lesson2ID, _, _, _, _ := seedTwoLessonLevel(content)
	ctx := context.Background()

	if _, err := svc.CompleteLesson(ctx, "user-1", lesson1ID); err != nil {
		t.Fatalf("CompleteLesson() error = %v", err)
	}
	if _, err := svc.CompleteLesson(ctx, "user-1", lesson2ID); err != nil {
		t.Fatalf("CompleteLesson() error = %v", err)
	}

	badges, err := svc.ListUserBadges(ctx, "user-1")
	if err != nil {
		t.Fatalf("ListUserBadges() error = %v", err)
	}
	if len(badges) != 1 || badges[0].Name != "Beginner" {
		t.Errorf("badges = %+v, want [Beginner]", badges)
	}
}
