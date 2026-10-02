// Package learningservice implements the structured-learning pillar:
// browsing levels/lessons, taking a lesson's quiz, marking a lesson
// complete, and the progress/badges that fall out of completion.
package learningservice

import (
	"context"
	"errors"
	"time"

	"github.com/traumacompendium/backend/internal/domain/learning"
	"github.com/traumacompendium/backend/internal/notification/tts"
	"github.com/traumacompendium/backend/internal/pkg/apperror"
)

type Service struct {
	content  learning.ContentRepository
	progress learning.ProgressRepository
	audio    learning.AudioRepository
	synth    tts.Synthesizer
	now      func() time.Time
}

func New(
	content learning.ContentRepository,
	progress learning.ProgressRepository,
	audio learning.AudioRepository,
	synth tts.Synthesizer,
) *Service {
	return &Service{content: content, progress: progress, audio: audio, synth: synth, now: time.Now}
}

func (s *Service) ListLevels(ctx context.Context, locale learning.Locale) ([]learning.Level, error) {
	levels, err := s.content.ListLevels(ctx, resolveLocale(locale))
	if err != nil {
		return nil, apperror.NewInternal("could not list levels").WithCause(err)
	}
	return levels, nil
}

// GetLevel is used by the HTTP layer to check a level's Number before
// deciding whether its content requires Premium (see
// handler.LearningHandler's premium-gating helpers) — level/badge
// metadata itself stays free to browse; only lesson content is gated.
func (s *Service) GetLevel(ctx context.Context, id string, locale learning.Locale) (learning.Level, error) {
	level, err := s.content.FindLevel(ctx, id, resolveLocale(locale))
	if err != nil {
		return learning.Level{}, mapContentNotFound(err)
	}
	return level, nil
}

func (s *Service) ListLessons(ctx context.Context, levelID string, locale learning.Locale) ([]learning.Lesson, error) {
	locale = resolveLocale(locale)

	if _, err := s.content.FindLevel(ctx, levelID, locale); err != nil {
		return nil, mapContentNotFound(err)
	}

	lessons, err := s.content.ListLessonsByLevel(ctx, levelID, locale)
	if err != nil {
		return nil, apperror.NewInternal("could not list lessons").WithCause(err)
	}
	return lessons, nil
}

func (s *Service) GetLesson(ctx context.Context, lessonID string, locale learning.Locale) (learning.Lesson, error) {
	lesson, err := s.content.FindLesson(ctx, lessonID, resolveLocale(locale))
	if err != nil {
		return learning.Lesson{}, mapContentNotFound(err)
	}
	return lesson, nil
}

// resolveLocale is the service layer's own safety net: handlers are
// expected to have already validated/defaulted the locale (see
// internal/http.ResolveLocale), but any other caller — a test, a future
// background job — gets the same "unknown or empty locale falls back to
// English" behavior rather than passing something meaningless to SQL.
func resolveLocale(locale learning.Locale) learning.Locale {
	if learning.IsSupportedLocale(locale) {
		return locale
	}
	return learning.DefaultLocale
}

func mapContentNotFound(err error) error {
	switch {
	case errors.Is(err, learning.ErrLevelNotFound):
		return apperror.NewNotFound("level not found")
	case errors.Is(err, learning.ErrLessonNotFound):
		return apperror.NewNotFound("lesson not found")
	case errors.Is(err, learning.ErrQuizNotFound):
		return apperror.NewNotFound("quiz not found")
	case errors.Is(err, learning.ErrBadgeNotFound):
		return apperror.NewNotFound("badge not found")
	default:
		return apperror.NewInternal("could not load learning content").WithCause(err)
	}
}
