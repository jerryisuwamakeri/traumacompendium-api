// Package learningservice implements the structured-learning pillar:
// browsing levels/lessons, taking a lesson's quiz, marking a lesson
// complete, and the progress/badges that fall out of completion.
package learningservice

import (
	"context"
	"errors"
	"time"

	"github.com/traumacompendium/backend/internal/domain/learning"
	"github.com/traumacompendium/backend/internal/pkg/apperror"
)

type Service struct {
	content  learning.ContentRepository
	progress learning.ProgressRepository
	now      func() time.Time
}

func New(content learning.ContentRepository, progress learning.ProgressRepository) *Service {
	return &Service{content: content, progress: progress, now: time.Now}
}

func (s *Service) ListLevels(ctx context.Context) ([]learning.Level, error) {
	levels, err := s.content.ListLevels(ctx)
	if err != nil {
		return nil, apperror.NewInternal("could not list levels").WithCause(err)
	}
	return levels, nil
}

func (s *Service) ListLessons(ctx context.Context, levelID string) ([]learning.Lesson, error) {
	if _, err := s.content.FindLevel(ctx, levelID); err != nil {
		return nil, mapContentNotFound(err)
	}

	lessons, err := s.content.ListLessonsByLevel(ctx, levelID)
	if err != nil {
		return nil, apperror.NewInternal("could not list lessons").WithCause(err)
	}
	return lessons, nil
}

func (s *Service) GetLesson(ctx context.Context, lessonID string) (learning.Lesson, error) {
	lesson, err := s.content.FindLesson(ctx, lessonID)
	if err != nil {
		return learning.Lesson{}, mapContentNotFound(err)
	}
	return lesson, nil
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
