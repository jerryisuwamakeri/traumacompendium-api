package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/traumacompendium/backend/internal/domain/learning"
)

type LearningAudioRepository struct {
	db *sql.DB
}

func NewLearningAudioRepository(db *sql.DB) *LearningAudioRepository {
	return &LearningAudioRepository{db: db}
}

func (r *LearningAudioRepository) Get(ctx context.Context, lessonID string, locale learning.Locale) (learning.Audio, error) {
	const q = `
		SELECT lesson_id, locale, audio_data, content_type, generated_at
		FROM lesson_audio WHERE lesson_id = ? AND locale = ?
	`
	var a learning.Audio
	err := r.db.QueryRowContext(ctx, q, lessonID, locale).
		Scan(&a.LessonID, &a.Locale, &a.Data, &a.ContentType, &a.GeneratedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return learning.Audio{}, learning.ErrAudioNotFound
		}
		return learning.Audio{}, fmt.Errorf("mysql: get lesson audio: %w", err)
	}
	return a, nil
}

// Save is idempotent for the same reason the progress repository's
// writes are (see LearningProgressRepository): a second synthesis of the
// same (lesson_id, locale) — e.g. a race between two concurrent
// cache-miss requests — just leaves the first result in place rather
// than erroring or double-writing.
func (r *LearningAudioRepository) Save(ctx context.Context, audio learning.Audio) error {
	const q = `
		INSERT INTO lesson_audio (lesson_id, locale, audio_data, content_type, generated_at)
		VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE lesson_id = lesson_id
	`
	_, err := r.db.ExecContext(ctx, q, audio.LessonID, audio.Locale, audio.Data, audio.ContentType, audio.GeneratedAt)
	if err != nil {
		return fmt.Errorf("mysql: save lesson audio: %w", err)
	}
	return nil
}
