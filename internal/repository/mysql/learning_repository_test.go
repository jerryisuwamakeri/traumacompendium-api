package mysql

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/traumacompendium/backend/internal/domain/learning"
)

// These tests read against the catalog seeded by migration 0014 (6
// levels; Level 1 has 3 lessons, each with a quiz) rather than inserting
// their own fixtures — see testdb_test.go's truncateAll doc comment for
// why that data is treated as stable reference content across the test
// run.
const (
	seededLevel1ID  = "00000000-0000-0000-0000-000000000001"
	seededLesson1ID = "00000000-0000-0000-0000-000000000101"
)

func TestLearningContentRepository_ListLevels(t *testing.T) {
	db := openTestDB(t)
	repo := NewLearningContentRepository(db)

	levels, err := repo.ListLevels(context.Background())
	if err != nil {
		t.Fatalf("ListLevels() error = %v", err)
	}
	if len(levels) != 6 {
		t.Fatalf("len(levels) = %d, want 6", len(levels))
	}
	if levels[0].Number != 1 || levels[5].Number != 6 {
		t.Errorf("levels not ordered by number: first=%d last=%d", levels[0].Number, levels[5].Number)
	}
}

func TestLearningContentRepository_FindLevel_NotFound(t *testing.T) {
	db := openTestDB(t)
	repo := NewLearningContentRepository(db)

	_, err := repo.FindLevel(context.Background(), "does-not-exist")
	if !errors.Is(err, learning.ErrLevelNotFound) {
		t.Errorf("FindLevel() error = %v, want %v", err, learning.ErrLevelNotFound)
	}
}

func TestLearningContentRepository_ListLessonsByLevel(t *testing.T) {
	db := openTestDB(t)
	repo := NewLearningContentRepository(db)

	lessons, err := repo.ListLessonsByLevel(context.Background(), seededLevel1ID)
	if err != nil {
		t.Fatalf("ListLessonsByLevel() error = %v", err)
	}
	if len(lessons) != 3 {
		t.Fatalf("len(lessons) = %d, want 3", len(lessons))
	}
	if lessons[0].Title != "What Is Trauma?" {
		t.Errorf("lessons[0].Title = %q, want %q", lessons[0].Title, "What Is Trauma?")
	}
	if lessons[0].OrderIndex >= lessons[1].OrderIndex {
		t.Error("lessons not ordered by order_index")
	}
}

func TestLearningContentRepository_FindQuizByLesson_IncludesOptionsWithCorrectness(t *testing.T) {
	db := openTestDB(t)
	repo := NewLearningContentRepository(db)

	quiz, questions, err := repo.FindQuizByLesson(context.Background(), seededLesson1ID)
	if err != nil {
		t.Fatalf("FindQuizByLesson() error = %v", err)
	}
	if quiz.LessonID != seededLesson1ID {
		t.Errorf("quiz.LessonID = %q, want %q", quiz.LessonID, seededLesson1ID)
	}
	if len(questions) != 1 {
		t.Fatalf("len(questions) = %d, want 1", len(questions))
	}
	if len(questions[0].Options) != 4 {
		t.Fatalf("len(options) = %d, want 4", len(questions[0].Options))
	}

	correctCount := 0
	for _, o := range questions[0].Options {
		if o.IsCorrect {
			correctCount++
		}
	}
	if correctCount != 1 {
		t.Errorf("correctCount = %d, want exactly 1 correct option", correctCount)
	}
}

func TestLearningContentRepository_FindBadgeByLevel(t *testing.T) {
	db := openTestDB(t)
	repo := NewLearningContentRepository(db)

	badge, err := repo.FindBadgeByLevel(context.Background(), seededLevel1ID)
	if err != nil {
		t.Fatalf("FindBadgeByLevel() error = %v", err)
	}
	if badge.Name != "Beginner" {
		t.Errorf("badge.Name = %q, want %q", badge.Name, "Beginner")
	}
}

func TestLearningProgressRepository_MarkCompleteIsIdempotent(t *testing.T) {
	db := openTestDB(t)
	userID := seedTestUser(t, db)
	repo := NewLearningProgressRepository(db)
	ctx := context.Background()

	first := time.Now().UTC().Truncate(time.Second)
	if err := repo.MarkLessonComplete(ctx, userID, seededLesson1ID, first); err != nil {
		t.Fatalf("MarkLessonComplete() error = %v", err)
	}

	second := first.Add(1 * time.Hour)
	if err := repo.MarkLessonComplete(ctx, userID, seededLesson1ID, second); err != nil {
		t.Fatalf("MarkLessonComplete() (again) error = %v", err)
	}

	completed, err := repo.ListCompletedLessonIDs(ctx, userID, []string{seededLesson1ID})
	if err != nil {
		t.Fatalf("ListCompletedLessonIDs() error = %v", err)
	}
	if len(completed) != 1 {
		t.Fatalf("len(completed) = %d, want 1 (marking complete twice should not duplicate)", len(completed))
	}
}

func TestLearningProgressRepository_ListCompletedLessonIDs_OnlyForCaller(t *testing.T) {
	db := openTestDB(t)
	userA := seedTestUser(t, db)
	userB := seedTestUser(t, db)
	repo := NewLearningProgressRepository(db)
	ctx := context.Background()

	if err := repo.MarkLessonComplete(ctx, userA, seededLesson1ID, time.Now()); err != nil {
		t.Fatalf("MarkLessonComplete() error = %v", err)
	}

	completed, err := repo.ListCompletedLessonIDs(ctx, userB, []string{seededLesson1ID})
	if err != nil {
		t.Fatalf("ListCompletedLessonIDs() error = %v", err)
	}
	if len(completed) != 0 {
		t.Errorf("len(completed) for userB = %d, want 0 (userA's completion must not leak)", len(completed))
	}
}

func TestLearningProgressRepository_AwardBadgeIsIdempotent(t *testing.T) {
	db := openTestDB(t)
	userID := seedTestUser(t, db)
	content := NewLearningContentRepository(db)
	progress := NewLearningProgressRepository(db)
	ctx := context.Background()

	badge, err := content.FindBadgeByLevel(ctx, seededLevel1ID)
	if err != nil {
		t.Fatalf("FindBadgeByLevel() error = %v", err)
	}

	if err := progress.AwardBadge(ctx, userID, badge.ID, time.Now()); err != nil {
		t.Fatalf("AwardBadge() error = %v", err)
	}
	if err := progress.AwardBadge(ctx, userID, badge.ID, time.Now()); err != nil {
		t.Fatalf("AwardBadge() (again) error = %v", err)
	}

	ids, err := progress.ListUserBadgeIDs(ctx, userID)
	if err != nil {
		t.Fatalf("ListUserBadgeIDs() error = %v", err)
	}
	if len(ids) != 1 {
		t.Fatalf("len(ids) = %d, want 1 (awarding twice should not duplicate)", len(ids))
	}
}
