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

type Level struct {
	ID          string
	Number      int
	Name        string
	Description string
	CreatedAt   time.Time
}

type Lesson struct {
	ID         string
	LevelID    string
	Title      string
	Content    string
	OrderIndex int
	CreatedAt  time.Time
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
}

type Badge struct {
	ID          string
	LevelID     string
	Name        string
	Description string
	IconKey     string
}

// ContentRepository reads the learning catalog. There is deliberately no
// write method here in Phase 1 — content is authored via migration, not
// an in-app CMS; an admin content API is a later-phase concern.
type ContentRepository interface {
	ListLevels(ctx context.Context) ([]Level, error)
	FindLevel(ctx context.Context, id string) (Level, error)
	ListLessonsByLevel(ctx context.Context, levelID string) ([]Lesson, error)
	FindLesson(ctx context.Context, id string) (Lesson, error)
	FindQuizByLesson(ctx context.Context, lessonID string) (Quiz, []Question, error)
	FindBadgeByLevel(ctx context.Context, levelID string) (Badge, error)
	ListBadgesByIDs(ctx context.Context, ids []string) ([]Badge, error)
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
