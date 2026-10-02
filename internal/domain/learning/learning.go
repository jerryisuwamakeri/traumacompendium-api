// Package learning defines the structured-learning domain: a fixed
// catalog of levels/lessons/quizzes (authored via migration, not an
// in-app CMS, in this phase) and the per-user progress and badges
// layered on top of it.
//
// The catalog and progress concerns are split into two repository
// interfaces (ContentRepository, ProgressRepository) because they are
// genuinely different responsibilities with different write patterns:
// content is effectively read-only reference data in Phase 1, while
// progress is per-user state written constantly as users move through
// lessons.
package learning

import (
	"context"
	"time"
)

// Locale is a BCP-47-ish language tag. DefaultLocale ("en") is the
// canonical language every entity's own columns hold directly; every
// other supported locale lives in that entity's *_translations table and
// is resolved with a fallback to English when a translation doesn't
// exist yet (see mysql.LearningContentRepository).
type Locale = string

const DefaultLocale Locale = "en"

// SupportedLocales mirrors the brief's localization roadmap: MVP-launch
// languages (Pidgin, Yoruba, Hausa, Igbo) plus the later-wave languages
// (French, Swahili, Arabic) requested to be built out now too.
var SupportedLocales = []Locale{DefaultLocale, "pcm", "yo", "ha", "ig", "fr", "sw", "ar"}

func IsSupportedLocale(l Locale) bool {
	for _, s := range SupportedLocales {
		if s == l {
			return true
		}
	}
	return false
}

// Translation is embedded in every content type to report which locale
// the returned text is actually in (after fallback) and whether it's an
// unreviewed draft — see the package doc comment on NeedsReview's
// purpose. Canonical English content is always Locale: "en",
// NeedsReview: false.
type Translation struct {
	Locale      Locale
	NeedsReview bool
}

type Level struct {
	ID          string
	Number      int
	Name        string
	Description string
	CreatedAt   time.Time
	Translation
}

type Lesson struct {
	ID         string
	LevelID    string
	Title      string
	Content    string
	OrderIndex int
	CreatedAt  time.Time
	Translation
}

type Quiz struct {
	ID       string
	LessonID string
}

type Question struct {
	ID         string
	QuizID     string
	Text       string
	OrderIndex int
	Options    []Option
	Translation
}

type Option struct {
	ID         string
	QuestionID string
	Text       string
	// IsCorrect must never be serialized into a "take this quiz" HTTP
	// response — it is only read internally by the service layer to
	// score a submission. See http/dto.NewQuizResponse, which
	// deliberately omits it.
	IsCorrect  bool
	OrderIndex int
	Translation
}

type Badge struct {
	ID          string
	LevelID     string
	Name        string
	Description string
	IconKey     string
	Translation
}

// ContentRepository reads the learning catalog. There is deliberately no
// write method here in Phase 1 — content is authored via migration, not
// an in-app CMS; an admin content API is a later-phase concern.
//
// Every read takes a locale and resolves it with fallback to
// DefaultLocale when no translation row exists — callers never need to
// branch on "did the translation exist," they just always get usable
// content, annotated via Translation for whether it was actually in the
// requested language.
type ContentRepository interface {
	ListLevels(ctx context.Context, locale Locale) ([]Level, error)
	FindLevel(ctx context.Context, id string, locale Locale) (Level, error)
	ListLessonsByLevel(ctx context.Context, levelID string, locale Locale) ([]Lesson, error)
	FindLesson(ctx context.Context, id string, locale Locale) (Lesson, error)
	FindQuizByLesson(ctx context.Context, lessonID string, locale Locale) (Quiz, []Question, error)
	FindBadgeByLevel(ctx context.Context, levelID string, locale Locale) (Badge, error)
	ListBadgesByIDs(ctx context.Context, ids []string, locale Locale) ([]Badge, error)
}

// ProgressRepository persists per-user progress through the catalog.
type ProgressRepository interface {
	MarkLessonComplete(ctx context.Context, userID, lessonID string, completedAt time.Time) error
	// ListCompletedLessonIDs returns the subset of lessonIDs the user has
	// completed — used to compute per-level progress without the
	// progress store needing to know about levels at all.
	ListCompletedLessonIDs(ctx context.Context, userID string, lessonIDs []string) ([]string, error)
	// AwardBadge is idempotent: awarding the same badge to the same user
	// twice is a no-op, not an error.
	AwardBadge(ctx context.Context, userID, badgeID string, awardedAt time.Time) error
	ListUserBadgeIDs(ctx context.Context, userID string) ([]string, error)
}

var (
	ErrLevelNotFound  = errLevelNotFound{}
	ErrLessonNotFound = errLessonNotFound{}
	ErrQuizNotFound   = errQuizNotFound{}
	ErrBadgeNotFound  = errBadgeNotFound{}
)

type errLevelNotFound struct{}

func (errLevelNotFound) Error() string { return "learning: level not found" }

type errLessonNotFound struct{}

func (errLessonNotFound) Error() string { return "learning: lesson not found" }

type errQuizNotFound struct{}

func (errQuizNotFound) Error() string { return "learning: quiz not found" }

type errBadgeNotFound struct{}

func (errBadgeNotFound) Error() string { return "learning: badge not found" }
