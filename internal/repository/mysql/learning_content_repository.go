package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/traumacompendium/backend/internal/domain/learning"
)

// LearningContentRepository reads the learning catalog, resolving every
// text field through a LEFT JOIN against that entity's *_translations
// table. Passing locale "en" (learning.DefaultLocale) simply never
// matches a translation row (translations only ever hold non-English
// rows — see migration 0021), so the COALESCE falls through to the
// canonical column for every query uniformly; there is no separate
// "English" code path.
type LearningContentRepository struct {
	db *sql.DB
}

func NewLearningContentRepository(db *sql.DB) *LearningContentRepository {
	return &LearningContentRepository{db: db}
}

func (r *LearningContentRepository) ListLevels(ctx context.Context, locale learning.Locale) ([]learning.Level, error) {
	const q = `
		SELECT l.id, l.number,
		       COALESCE(t.name, l.name), COALESCE(t.description, l.description),
		       l.created_at,
		       COALESCE(t.locale, ?), COALESCE(t.needs_review, FALSE)
		FROM levels l
		LEFT JOIN level_translations t ON t.level_id = l.id AND t.locale = ?
		ORDER BY l.number ASC
	`
	rows, err := r.db.QueryContext(ctx, q, learning.DefaultLocale, locale)
	if err != nil {
		return nil, fmt.Errorf("mysql: list levels: %w", err)
	}
	defer rows.Close() //nolint:errcheck // read-only cursor, nothing actionable on close failure

	levels := make([]learning.Level, 0)
	for rows.Next() {
		var l learning.Level
		if err := rows.Scan(&l.ID, &l.Number, &l.Name, &l.Description, &l.CreatedAt, &l.Locale, &l.NeedsReview); err != nil {
			return nil, fmt.Errorf("mysql: scan level: %w", err)
		}
		levels = append(levels, l)
	}
	return levels, rows.Err()
}

func (r *LearningContentRepository) FindLevel(ctx context.Context, id string, locale learning.Locale) (learning.Level, error) {
	const q = `
		SELECT l.id, l.number,
		       COALESCE(t.name, l.name), COALESCE(t.description, l.description),
		       l.created_at,
		       COALESCE(t.locale, ?), COALESCE(t.needs_review, FALSE)
		FROM levels l
		LEFT JOIN level_translations t ON t.level_id = l.id AND t.locale = ?
		WHERE l.id = ?
	`
	var l learning.Level
	err := r.db.QueryRowContext(ctx, q, learning.DefaultLocale, locale, id).
		Scan(&l.ID, &l.Number, &l.Name, &l.Description, &l.CreatedAt, &l.Locale, &l.NeedsReview)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return learning.Level{}, learning.ErrLevelNotFound
		}
		return learning.Level{}, fmt.Errorf("mysql: find level: %w", err)
	}
	return l, nil
}

func (r *LearningContentRepository) ListLessonsByLevel(ctx context.Context, levelID string, locale learning.Locale) ([]learning.Lesson, error) {
	const q = `
		SELECT le.id, le.level_id,
		       COALESCE(t.title, le.title), COALESCE(t.content, le.content),
		       le.order_index, le.created_at,
		       COALESCE(t.locale, ?), COALESCE(t.needs_review, FALSE)
		FROM lessons le
		LEFT JOIN lesson_translations t ON t.lesson_id = le.id AND t.locale = ?
		WHERE le.level_id = ?
		ORDER BY le.order_index ASC
	`
	rows, err := r.db.QueryContext(ctx, q, learning.DefaultLocale, locale, levelID)
	if err != nil {
		return nil, fmt.Errorf("mysql: list lessons by level: %w", err)
	}
	defer rows.Close() //nolint:errcheck // read-only cursor, nothing actionable on close failure

	lessons := make([]learning.Lesson, 0)
	for rows.Next() {
		l, err := scanLesson(rows)
		if err != nil {
			return nil, err
		}
		lessons = append(lessons, l)
	}
	return lessons, rows.Err()
}

func (r *LearningContentRepository) FindLesson(ctx context.Context, id string, locale learning.Locale) (learning.Lesson, error) {
	const q = `
		SELECT le.id, le.level_id,
		       COALESCE(t.title, le.title), COALESCE(t.content, le.content),
		       le.order_index, le.created_at,
		       COALESCE(t.locale, ?), COALESCE(t.needs_review, FALSE)
		FROM lessons le
		LEFT JOIN lesson_translations t ON t.lesson_id = le.id AND t.locale = ?
		WHERE le.id = ?
	`
	l, err := scanLesson(r.db.QueryRowContext(ctx, q, learning.DefaultLocale, locale, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return learning.Lesson{}, learning.ErrLessonNotFound
		}
		return learning.Lesson{}, err
	}
	return l, nil
}

func scanLesson(row interface{ Scan(dest ...any) error }) (learning.Lesson, error) {
	var l learning.Lesson
	err := row.Scan(&l.ID, &l.LevelID, &l.Title, &l.Content, &l.OrderIndex, &l.CreatedAt, &l.Locale, &l.NeedsReview)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return learning.Lesson{}, err // caller maps to ErrLessonNotFound
		}
		return learning.Lesson{}, fmt.Errorf("mysql: scan lesson: %w", err)
	}
	return l, nil
}

func (r *LearningContentRepository) FindQuizByLesson(ctx context.Context, lessonID string, locale learning.Locale) (learning.Quiz, []learning.Question, error) {
	const quizQ = `SELECT id, lesson_id FROM quizzes WHERE lesson_id = ?`
	var quiz learning.Quiz
	if err := r.db.QueryRowContext(ctx, quizQ, lessonID).Scan(&quiz.ID, &quiz.LessonID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return learning.Quiz{}, nil, learning.ErrQuizNotFound
		}
		return learning.Quiz{}, nil, fmt.Errorf("mysql: find quiz: %w", err)
	}

	const questionsQ = `
		SELECT q.id, q.quiz_id, COALESCE(t.text, q.question), q.order_index,
		       COALESCE(t.locale, ?), COALESCE(t.needs_review, FALSE)
		FROM quiz_questions q
		LEFT JOIN quiz_question_translations t ON t.question_id = q.id AND t.locale = ?
		WHERE q.quiz_id = ? ORDER BY q.order_index ASC
	`
	qRows, err := r.db.QueryContext(ctx, questionsQ, learning.DefaultLocale, locale, quiz.ID)
	if err != nil {
		return learning.Quiz{}, nil, fmt.Errorf("mysql: list quiz questions: %w", err)
	}
	defer qRows.Close() //nolint:errcheck // read-only cursor, nothing actionable on close failure

	var questions []learning.Question
	questionIndex := make(map[string]int)
	for qRows.Next() {
		var q learning.Question
		if err := qRows.Scan(&q.ID, &q.QuizID, &q.Text, &q.OrderIndex, &q.Locale, &q.NeedsReview); err != nil {
			return learning.Quiz{}, nil, fmt.Errorf("mysql: scan quiz question: %w", err)
		}
		questionIndex[q.ID] = len(questions)
		questions = append(questions, q)
	}
	if err := qRows.Err(); err != nil {
		return learning.Quiz{}, nil, fmt.Errorf("mysql: iterate quiz questions: %w", err)
	}
	if len(questions) == 0 {
		return quiz, questions, nil
	}

	placeholders := make([]string, len(questions))
	args := make([]any, 0, len(questions)+2)
	args = append(args, learning.DefaultLocale, locale)
	for i, q := range questions {
		placeholders[i] = "?"
		args = append(args, q.ID)
	}
	optionsQ := `
		SELECT o.id, o.question_id, COALESCE(t.text, o.text), o.is_correct, o.order_index,
		       COALESCE(t.locale, ?), COALESCE(t.needs_review, FALSE)
		FROM quiz_options o
		LEFT JOIN quiz_option_translations t ON t.option_id = o.id AND t.locale = ?
		WHERE o.question_id IN (` + strings.Join(placeholders, ",") + `)
		ORDER BY o.question_id, o.order_index ASC
	`
	oRows, err := r.db.QueryContext(ctx, optionsQ, args...)
	if err != nil {
		return learning.Quiz{}, nil, fmt.Errorf("mysql: list quiz options: %w", err)
	}
	defer oRows.Close() //nolint:errcheck // read-only cursor, nothing actionable on close failure

	for oRows.Next() {
		var o learning.Option
		if err := oRows.Scan(&o.ID, &o.QuestionID, &o.Text, &o.IsCorrect, &o.OrderIndex, &o.Locale, &o.NeedsReview); err != nil {
			return learning.Quiz{}, nil, fmt.Errorf("mysql: scan quiz option: %w", err)
		}
		idx, ok := questionIndex[o.QuestionID]
		if !ok {
			continue
		}
		questions[idx].Options = append(questions[idx].Options, o)
	}
	if err := oRows.Err(); err != nil {
		return learning.Quiz{}, nil, fmt.Errorf("mysql: iterate quiz options: %w", err)
	}

	return quiz, questions, nil
}

func (r *LearningContentRepository) FindBadgeByLevel(ctx context.Context, levelID string, locale learning.Locale) (learning.Badge, error) {
	const q = `
		SELECT b.id, b.level_id,
		       COALESCE(t.name, b.name), COALESCE(t.description, b.description), b.icon_key,
		       COALESCE(t.locale, ?), COALESCE(t.needs_review, FALSE)
		FROM badges b
		LEFT JOIN badge_translations t ON t.badge_id = b.id AND t.locale = ?
		WHERE b.level_id = ?
	`
	var b learning.Badge
	err := r.db.QueryRowContext(ctx, q, learning.DefaultLocale, locale, levelID).
		Scan(&b.ID, &b.LevelID, &b.Name, &b.Description, &b.IconKey, &b.Locale, &b.NeedsReview)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return learning.Badge{}, learning.ErrBadgeNotFound
		}
		return learning.Badge{}, fmt.Errorf("mysql: find badge by level: %w", err)
	}
	return b, nil
}

func (r *LearningContentRepository) ListBadgesByIDs(ctx context.Context, ids []string, locale learning.Locale) ([]learning.Badge, error) {
	if len(ids) == 0 {
		return []learning.Badge{}, nil
	}

	placeholders := make([]string, len(ids))
	args := make([]any, 0, len(ids)+2)
	args = append(args, learning.DefaultLocale, locale)
	for i, id := range ids {
		placeholders[i] = "?"
		args = append(args, id)
	}

	q := `
		SELECT b.id, b.level_id,
		       COALESCE(t.name, b.name), COALESCE(t.description, b.description), b.icon_key,
		       COALESCE(t.locale, ?), COALESCE(t.needs_review, FALSE)
		FROM badges b
		LEFT JOIN badge_translations t ON t.badge_id = b.id AND t.locale = ?
		WHERE b.id IN (` + strings.Join(placeholders, ",") + `)
	`

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("mysql: list badges by ids: %w", err)
	}
	defer rows.Close() //nolint:errcheck // read-only cursor, nothing actionable on close failure

	badges := make([]learning.Badge, 0, len(ids))
	for rows.Next() {
		var b learning.Badge
		if err := rows.Scan(&b.ID, &b.LevelID, &b.Name, &b.Description, &b.IconKey, &b.Locale, &b.NeedsReview); err != nil {
			return nil, fmt.Errorf("mysql: scan badge: %w", err)
		}
		badges = append(badges, b)
	}
	return badges, rows.Err()
}
