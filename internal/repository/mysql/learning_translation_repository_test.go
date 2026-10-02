package mysql

import (
	"context"
	"database/sql"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
)

// These tests verify the JOIN/COALESCE fallback logic itself, using
// throwaway level/lesson fixtures created (and cleaned up) per test —
// never the globally-seeded Level 1 content (seededLevel1ID /
// seededLesson1ID), since migration 0021 seeds translations for every
// supported locale against that exact content, and this suite must keep
// working after that migration lands without colliding with it.

// fixtureLevelNumber hands out a unique levels.number per throwaway
// fixture (that column is UNIQUE and TINYINT, so it must stay under
// 128), starting well outside the real 1-6 range so a fixture can never
// collide with seeded content.
var fixtureLevelNumber int64 = 100

func seedThrowawayLevel(t *testing.T, db *sql.DB) string {
	t.Helper()
	id := uuid.NewString()
	number := atomic.AddInt64(&fixtureLevelNumber, 1)
	const q = `INSERT INTO levels (id, number, name, description, created_at) VALUES (?, ?, 'Fixture Level', 'For tests only', ?)`
	if _, err := db.ExecContext(context.Background(), q, id, number, time.Now().UTC()); err != nil {
		t.Fatalf("seed throwaway level: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM levels WHERE id = ?`, id) // cascades to lessons/translations
	})
	return id
}

// orderIndex is the caller's responsibility: lessons.uq_lessons_level_order
// is UNIQUE(level_id, order_index), so two fixture lessons under the same
// throwaway level need distinct values.
func seedThrowawayLesson(t *testing.T, db *sql.DB, levelID string, orderIndex int) string {
	t.Helper()
	id := uuid.NewString()
	const q = `INSERT INTO lessons (id, level_id, title, content, order_index, created_at) VALUES (?, ?, 'Fixture Lesson', 'For tests only', ?, ?)`
	if _, err := db.ExecContext(context.Background(), q, id, levelID, orderIndex, time.Now().UTC()); err != nil {
		t.Fatalf("seed throwaway lesson: %v", err)
	}
	return id
}

func insertLevelTranslation(t *testing.T, db *sql.DB, levelID, locale, name, description string) {
	t.Helper()
	const q = `
		INSERT INTO level_translations (level_id, locale, name, description, needs_review)
		VALUES (?, ?, ?, ?, TRUE)
	`
	if _, err := db.ExecContext(context.Background(), q, levelID, locale, name, description); err != nil {
		t.Fatalf("insert level translation: %v", err)
	}
}

func insertLessonTranslation(t *testing.T, db *sql.DB, lessonID, locale, title, content string) {
	t.Helper()
	const q = `
		INSERT INTO lesson_translations (lesson_id, locale, title, content, needs_review)
		VALUES (?, ?, ?, ?, TRUE)
	`
	if _, err := db.ExecContext(context.Background(), q, lessonID, locale, title, content); err != nil {
		t.Fatalf("insert lesson translation: %v", err)
	}
}

func TestLearningContentRepository_FindLevel_TranslationPresent_ReturnsTranslated(t *testing.T) {
	db := openTestDB(t)
	levelID := seedThrowawayLevel(t, db)
	insertLevelTranslation(t, db, levelID, "yo", "Olubere", "Ipile ipalara, awon aroso, imo opolo")
	repo := NewLearningContentRepository(db)

	level, err := repo.FindLevel(context.Background(), levelID, "yo")
	if err != nil {
		t.Fatalf("FindLevel() error = %v", err)
	}
	if level.Name != "Olubere" {
		t.Errorf("Name = %q, want %q", level.Name, "Olubere")
	}
	if level.Locale != "yo" {
		t.Errorf("Locale = %q, want %q", level.Locale, "yo")
	}
	if !level.NeedsReview {
		t.Error("NeedsReview = false, want true for a translated row")
	}
}

func TestLearningContentRepository_FindLevel_TranslationMissing_FallsBackToEnglish(t *testing.T) {
	db := openTestDB(t)
	levelID := seedThrowawayLevel(t, db)
	repo := NewLearningContentRepository(db)

	// No translation row was inserted for "fr" on this fixture level.
	level, err := repo.FindLevel(context.Background(), levelID, "fr")
	if err != nil {
		t.Fatalf("FindLevel() error = %v", err)
	}
	if level.Name != "Fixture Level" {
		t.Errorf("Name = %q, want canonical English %q", level.Name, "Fixture Level")
	}
	if level.Locale != "en" {
		t.Errorf("Locale = %q, want %q (fallback)", level.Locale, "en")
	}
	if level.NeedsReview {
		t.Error("NeedsReview = true, want false for canonical English fallback")
	}
}

func TestLearningContentRepository_FindLesson_TranslationPresent_ReturnsTranslated(t *testing.T) {
	db := openTestDB(t)
	levelID := seedThrowawayLevel(t, db)
	lessonID := seedThrowawayLesson(t, db, levelID, 1)
	insertLessonTranslation(t, db, lessonID, "fr", "Qu'est-ce que le traumatisme ?", "Contenu traduit.")
	repo := NewLearningContentRepository(db)

	lesson, err := repo.FindLesson(context.Background(), lessonID, "fr")
	if err != nil {
		t.Fatalf("FindLesson() error = %v", err)
	}
	if lesson.Title != "Qu'est-ce que le traumatisme ?" {
		t.Errorf("Title = %q, want the French translation", lesson.Title)
	}
	if lesson.Locale != "fr" || !lesson.NeedsReview {
		t.Errorf("Locale/NeedsReview = %q/%v, want fr/true", lesson.Locale, lesson.NeedsReview)
	}
}

func TestLearningContentRepository_ListLessonsByLevel_MixedTranslationCoverage(t *testing.T) {
	db := openTestDB(t)
	levelID := seedThrowawayLevel(t, db)
	translatedLesson := seedThrowawayLesson(t, db, levelID, 1)
	untranslatedLesson := seedThrowawayLesson(t, db, levelID, 2)
	insertLessonTranslation(t, db, translatedLesson, "ha", "Menene Rauni?", "Abun ciki.")
	repo := NewLearningContentRepository(db)

	lessons, err := repo.ListLessonsByLevel(context.Background(), levelID, "ha")
	if err != nil {
		t.Fatalf("ListLessonsByLevel() error = %v", err)
	}
	if len(lessons) != 2 {
		t.Fatalf("len(lessons) = %d, want 2 (partial translation coverage must not drop rows)", len(lessons))
	}

	byID := make(map[string]string, len(lessons)) // lessonID -> resolved locale
	for _, l := range lessons {
		byID[l.ID] = l.Locale
	}
	if byID[translatedLesson] != "ha" {
		t.Errorf("translated lesson resolved locale = %q, want ha", byID[translatedLesson])
	}
	if byID[untranslatedLesson] != "en" {
		t.Errorf("untranslated lesson resolved locale = %q, want en (fallback)", byID[untranslatedLesson])
	}
}
