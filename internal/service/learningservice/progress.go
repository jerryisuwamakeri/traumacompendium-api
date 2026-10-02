package learningservice

import (
	"context"
	"errors"

	"github.com/traumacompendium/backend/internal/domain/learning"
	"github.com/traumacompendium/backend/internal/pkg/apperror"
)

// CompleteLessonResult reports whether finishing this lesson also
// completed its level, so the client can show a "level complete!"
// celebration exactly once — on the request that actually earns it, not
// on every subsequent re-completion.
type CompleteLessonResult struct {
	Lesson         learning.Lesson
	LevelCompleted bool
	BadgeAwarded   *learning.Badge
}

func (s *Service) CompleteLesson(ctx context.Context, userID, lessonID string, locale learning.Locale) (CompleteLessonResult, error) {
	locale = resolveLocale(locale)

	lesson, err := s.content.FindLesson(ctx, lessonID, locale)
	if err != nil {
		return CompleteLessonResult{}, mapContentNotFound(err)
	}

	if err := s.progress.MarkLessonComplete(ctx, userID, lessonID, s.now()); err != nil {
		return CompleteLessonResult{}, apperror.NewInternal("could not record lesson completion").WithCause(err)
	}

	result := CompleteLessonResult{Lesson: lesson}

	// Only lesson IDs are needed to compute completion, which is
	// locale-independent — DefaultLocale avoids an unnecessary join for
	// text this call never uses.
	levelLessons, err := s.content.ListLessonsByLevel(ctx, lesson.LevelID, learning.DefaultLocale)
	if err != nil {
		return CompleteLessonResult{}, apperror.NewInternal("could not check level completion").WithCause(err)
	}
	levelLessonIDs := make([]string, len(levelLessons))
	for i, l := range levelLessons {
		levelLessonIDs[i] = l.ID
	}

	completedIDs, err := s.progress.ListCompletedLessonIDs(ctx, userID, levelLessonIDs)
	if err != nil {
		return CompleteLessonResult{}, apperror.NewInternal("could not check level completion").WithCause(err)
	}
	result.LevelCompleted = len(completedIDs) == len(levelLessonIDs)

	if !result.LevelCompleted {
		return result, nil
	}

	badge, err := s.content.FindBadgeByLevel(ctx, lesson.LevelID, locale)
	if err != nil {
		if errors.Is(err, learning.ErrBadgeNotFound) {
			// A level is allowed to have no badge configured; completing
			// it is still valid.
			return result, nil
		}
		return CompleteLessonResult{}, apperror.NewInternal("could not check badge").WithCause(err)
	}

	alreadyOwned, err := userOwnsBadge(ctx, s.progress, userID, badge.ID)
	if err != nil {
		return CompleteLessonResult{}, apperror.NewInternal("could not check badge").WithCause(err)
	}
	if alreadyOwned {
		return result, nil
	}

	if err := s.progress.AwardBadge(ctx, userID, badge.ID, s.now()); err != nil {
		return CompleteLessonResult{}, apperror.NewInternal("could not award badge").WithCause(err)
	}
	result.BadgeAwarded = &badge

	return result, nil
}

func userOwnsBadge(ctx context.Context, progress learning.ProgressRepository, userID, badgeID string) (bool, error) {
	owned, err := progress.ListUserBadgeIDs(ctx, userID)
	if err != nil {
		return false, err
	}
	for _, id := range owned {
		if id == badgeID {
			return true, nil
		}
	}
	return false, nil
}

// LevelProgress is one level's completion state for a user.
type LevelProgress struct {
	Level            learning.Level
	TotalLessons     int
	CompletedLessons int
	BadgeEarned      bool
}

func (s *Service) GetProgress(ctx context.Context, userID string, locale learning.Locale) ([]LevelProgress, error) {
	locale = resolveLocale(locale)

	levels, err := s.content.ListLevels(ctx, locale)
	if err != nil {
		return nil, apperror.NewInternal("could not load progress").WithCause(err)
	}

	ownedBadges, err := s.progress.ListUserBadgeIDs(ctx, userID)
	if err != nil {
		return nil, apperror.NewInternal("could not load progress").WithCause(err)
	}
	ownedSet := make(map[string]bool, len(ownedBadges))
	for _, id := range ownedBadges {
		ownedSet[id] = true
	}

	result := make([]LevelProgress, len(levels))
	for i, level := range levels {
		lessons, err := s.content.ListLessonsByLevel(ctx, level.ID, learning.DefaultLocale)
		if err != nil {
			return nil, apperror.NewInternal("could not load progress").WithCause(err)
		}
		lessonIDs := make([]string, len(lessons))
		for j, l := range lessons {
			lessonIDs[j] = l.ID
		}

		completed, err := s.progress.ListCompletedLessonIDs(ctx, userID, lessonIDs)
		if err != nil {
			return nil, apperror.NewInternal("could not load progress").WithCause(err)
		}

		badgeEarned := false
		if badge, err := s.content.FindBadgeByLevel(ctx, level.ID, learning.DefaultLocale); err == nil {
			badgeEarned = ownedSet[badge.ID]
		}

		result[i] = LevelProgress{
			Level:            level,
			TotalLessons:     len(lessons),
			CompletedLessons: len(completed),
			BadgeEarned:      badgeEarned,
		}
	}

	return result, nil
}

func (s *Service) ListUserBadges(ctx context.Context, userID string, locale learning.Locale) ([]learning.Badge, error) {
	ids, err := s.progress.ListUserBadgeIDs(ctx, userID)
	if err != nil {
		return nil, apperror.NewInternal("could not list badges").WithCause(err)
	}

	badges, err := s.content.ListBadgesByIDs(ctx, ids, resolveLocale(locale))
	if err != nil {
		return nil, apperror.NewInternal("could not list badges").WithCause(err)
	}
	return badges, nil
}
