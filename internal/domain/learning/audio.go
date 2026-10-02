package learning

import (
	"context"
	"time"
)

// Audio is a cached text-to-speech rendering of one lesson's content in
// one locale.
type Audio struct {
	LessonID    string
	Locale      Locale
	Data        []byte
	ContentType string
	GeneratedAt time.Time
}

// AudioRepository persists the lesson-audio cache. There is no Delete in
// Phase 1 — a lesson's text never changes after it ships, so a cached
// rendering never goes stale.
type AudioRepository interface {
	Get(ctx context.Context, lessonID string, locale Locale) (Audio, error)
	Save(ctx context.Context, audio Audio) error
}

var ErrAudioNotFound = errAudioNotFound{}

type errAudioNotFound struct{}

func (errAudioNotFound) Error() string { return "learning: audio not found" }
