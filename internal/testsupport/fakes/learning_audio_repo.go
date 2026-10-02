package fakes

import (
	"context"
	"sync"

	"github.com/traumacompendium/backend/internal/domain/learning"
)

// LearningAudioRepo is an in-memory learning.AudioRepository.
type LearningAudioRepo struct {
	mu   sync.Mutex
	data map[string]learning.Audio // keyed by lessonID+"|"+locale
}

func NewLearningAudioRepo() *LearningAudioRepo {
	return &LearningAudioRepo{data: make(map[string]learning.Audio)}
}

func audioKey(lessonID, locale string) string { return lessonID + "|" + locale }

func (f *LearningAudioRepo) Get(_ context.Context, lessonID string, locale learning.Locale) (learning.Audio, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	a, ok := f.data[audioKey(lessonID, locale)]
	if !ok {
		return learning.Audio{}, learning.ErrAudioNotFound
	}
	return a, nil
}

func (f *LearningAudioRepo) Save(_ context.Context, audio learning.Audio) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.data[audioKey(audio.LessonID, audio.Locale)] = audio
	return nil
}

// Count reports how many entries are cached — used by tests to assert a
// cache hit didn't re-synthesize.
func (f *LearningAudioRepo) Count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.data)
}
