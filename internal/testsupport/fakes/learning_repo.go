package fakes

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/traumacompendium/backend/internal/domain/learning"
)

// LearningContentRepo is an in-memory learning.ContentRepository, built
// from whatever levels/lessons/quizzes/badges the test seeds into it via
// the exported fields — there is no production write path for content,
// so the fake exposes direct field access instead of Create methods
// mirroring a nonexistent interface.
//
// It ignores the locale argument and always returns the seeded content
// as-is: the real locale-fallback behavior (COALESCE to English when no
// translation row exists) is a SQL concern tested directly against MySQL
// in internal/repository/mysql's learning_repository_test.go, not
// something worth re-simulating here. Service-level tests only need to
// confirm the requested locale is threaded through, which they do by
// asserting on what this fake received.
type LearningContentRepo struct {
	mu        sync.Mutex
	Levels    map[string]learning.Level
	Lessons   map[string]learning.Lesson
	Quizzes   map[string]learning.Quiz       // keyed by lessonID
	Questions map[string][]learning.Question // keyed by quizID
	Badges    map[string]learning.Badge      // keyed by levelID
}

func NewLearningContentRepo() *LearningContentRepo {
	return &LearningContentRepo{
		Levels:    make(map[string]learning.Level),
		Lessons:   make(map[string]learning.Lesson),
		Quizzes:   make(map[string]learning.Quiz),
		Questions: make(map[string][]learning.Question),
		Badges:    make(map[string]learning.Badge),
	}
}

func (f *LearningContentRepo) ListLevels(_ context.Context, _ learning.Locale) ([]learning.Level, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]learning.Level, 0, len(f.Levels))
	for _, l := range f.Levels {
		out = append(out, l)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Number < out[j].Number })
	return out, nil
}

func (f *LearningContentRepo) FindLevel(_ context.Context, id string, _ learning.Locale) (learning.Level, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	l, ok := f.Levels[id]
	if !ok {
		return learning.Level{}, learning.ErrLevelNotFound
	}
	return l, nil
}

func (f *LearningContentRepo) ListLessonsByLevel(_ context.Context, levelID string, _ learning.Locale) ([]learning.Lesson, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []learning.Lesson
	for _, l := range f.Lessons {
		if l.LevelID == levelID {
			out = append(out, l)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].OrderIndex < out[j].OrderIndex })
	return out, nil
}

func (f *LearningContentRepo) FindLesson(_ context.Context, id string, _ learning.Locale) (learning.Lesson, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	l, ok := f.Lessons[id]
	if !ok {
		return learning.Lesson{}, learning.ErrLessonNotFound
	}
	return l, nil
}

func (f *LearningContentRepo) FindQuizByLesson(_ context.Context, lessonID string, _ learning.Locale) (learning.Quiz, []learning.Question, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	quiz, ok := f.Quizzes[lessonID]
	if !ok {
		return learning.Quiz{}, nil, learning.ErrQuizNotFound
	}
	return quiz, f.Questions[quiz.ID], nil
}

func (f *LearningContentRepo) FindBadgeByLevel(_ context.Context, levelID string, _ learning.Locale) (learning.Badge, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	b, ok := f.Badges[levelID]
	if !ok {
		return learning.Badge{}, learning.ErrBadgeNotFound
	}
	return b, nil
}

func (f *LearningContentRepo) ListBadgesByIDs(_ context.Context, ids []string, _ learning.Locale) ([]learning.Badge, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	wanted := make(map[string]bool, len(ids))
	for _, id := range ids {
		wanted[id] = true
	}
	out := make([]learning.Badge, 0, len(ids))
	for _, b := range f.Badges {
		if wanted[b.ID] {
			out = append(out, b)
		}
	}
	return out, nil
}

// LearningProgressRepo is an in-memory learning.ProgressRepository.
type LearningProgressRepo struct {
	mu        sync.Mutex
	completed map[string]map[string]struct{} // userID -> lessonID set
	badges    map[string]map[string]struct{} // userID -> badgeID set
}

func NewLearningProgressRepo() *LearningProgressRepo {
	return &LearningProgressRepo{
		completed: make(map[string]map[string]struct{}),
		badges:    make(map[string]map[string]struct{}),
	}
}

func (f *LearningProgressRepo) MarkLessonComplete(_ context.Context, userID, lessonID string, _ time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.completed[userID] == nil {
		f.completed[userID] = make(map[string]struct{})
	}
	f.completed[userID][lessonID] = struct{}{}
	return nil
}

func (f *LearningProgressRepo) ListCompletedLessonIDs(_ context.Context, userID string, lessonIDs []string) ([]string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []string
	for _, id := range lessonIDs {
		if _, ok := f.completed[userID][id]; ok {
			out = append(out, id)
		}
	}
	if out == nil {
		out = []string{}
	}
	return out, nil
}

func (f *LearningProgressRepo) AwardBadge(_ context.Context, userID, badgeID string, _ time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.badges[userID] == nil {
		f.badges[userID] = make(map[string]struct{})
	}
	f.badges[userID][badgeID] = struct{}{}
	return nil
}

func (f *LearningProgressRepo) ListUserBadgeIDs(_ context.Context, userID string) ([]string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]string, 0, len(f.badges[userID]))
	for id := range f.badges[userID] {
		out = append(out, id)
	}
	return out, nil
}
