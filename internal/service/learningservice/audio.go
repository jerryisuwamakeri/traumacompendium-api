package learningservice

import (
	"context"
	"errors"

	"github.com/traumacompendium/backend/internal/domain/learning"
	"github.com/traumacompendium/backend/internal/notification/tts"
	"github.com/traumacompendium/backend/internal/pkg/apperror"
)

// GetLessonAudio returns a lesson's text-to-speech audio in the given
// locale, synthesizing and caching it on the first request and serving
// the cached clip on every request after. Audio is never pre-generated
// at seed time — see migration 0020's doc comment — so the very first
// request for a given (lesson, locale) pays the synthesis cost; every
// request after is a cache hit.
func (s *Service) GetLessonAudio(ctx context.Context, lessonID string, locale learning.Locale) (learning.Audio, error) {
	locale = resolveLocale(locale)

	cached, err := s.audio.Get(ctx, lessonID, locale)
	if err == nil {
		return cached, nil
	}
	if !errors.Is(err, learning.ErrAudioNotFound) {
		return learning.Audio{}, apperror.NewInternal("could not load lesson audio").WithCause(err)
	}

	lesson, err := s.content.FindLesson(ctx, lessonID, locale)
	if err != nil {
		return learning.Audio{}, mapContentNotFound(err)
	}

	result, err := s.synth.Synthesize(ctx, tts.Request{
		Text:   lesson.Title + ". " + lesson.Content,
		Locale: locale,
	})
	if err != nil {
		return learning.Audio{}, apperror.NewUnprocessable("audio is not available for this lesson in this language").WithCause(err)
	}

	audio := learning.Audio{
		LessonID:    lessonID,
		Locale:      locale,
		Data:        result.Audio,
		ContentType: result.ContentType,
		GeneratedAt: s.now(),
	}
	if err := s.audio.Save(ctx, audio); err != nil {
		// The synthesis itself succeeded — serve it even if caching
		// failed, rather than erroring out and making the caller
		// re-synthesize. The next request will just pay the synthesis
		// cost again too.
		return audio, nil //nolint:nilerr // intentional: cache-write failure must not fail a successful synthesis
	}

	return audio, nil
}
