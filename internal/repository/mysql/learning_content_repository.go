package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/traumacompendium/backend/internal/domain/learning"
)

type LearningContentRepository struct {
	db *sql.DB
}

func NewLearningContentRepository(db *sql.DB) *LearningContentRepository {
	return &LearningContentRepository{db: db}
}

func (r *LearningContentRepository) ListLevels(ctx context.Context) ([]learning.Level, error) {
	const q = `SELECT id, number, name, description, created_at FROM levels ORDER BY number ASC`
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("mysql: list levels: %w", err)
	}
	defer rows.Close() //nolint:errcheck // read-only cursor, nothing actionable on close failure

	levels := make([]learning.Level, 0)
	for rows.Next() {
		var l learning.Level
		if err := rows.Scan(&l.ID, &l.Number, &l.Name, &l.Description, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("mysql: scan level: %w", err)
		}
		levels = append(levels, l)
	}
	return levels, rows.Err()
}

func (r *LearningContentRepository) FindLevel(ctx context.Context, id string) (learning.Level, error) {
	const q = `SELECT id, number, name, description, created_at FROM levels WHERE id = ?`
	var l learning.Level
	err := r.db.QueryRowContext(ctx, q, id).Scan(&l.ID, &l.Number, &l.Name, &l.Description, &l.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return learning.Level{}, learning.ErrLevelNotFound
		}
		return learning.Level{}, fmt.Errorf("mysql: find level: %w", err)
	}
	return l, nil
}

func (r *LearningContentRepository) ListLessonsByLevel(ctx context.Context, levelID string) ([]learning.Lesson, error) {
	const q = `
		SELECT id, level_id, title, content, order_index, created_at
		FROM lessons WHERE level_id = ? ORDER BY order_index ASC
	`
	rows, err := r.db.QueryContext(ctx, q, levelID)
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

func (r *LearningContentRepository) FindLesson(ctx context.Context, id string) (learning.Lesson, error) {
	const q = `
		SELECT id, level_id, title, content, order_index, created_at
		FROM lessons WHERE id = ?
	`
	l, err := scanLesson(r.db.QueryRowContext(ctx, q, id))
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
	err := row.Scan(&l.ID, &l.LevelID, &l.Title, &l.Content, &l.OrderIndex, &l.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return learning.Lesson{}, err // caller maps to ErrLessonNotFound
		}
		return learning.Lesson{}, fmt.Errorf("mysql: scan lesson: %w", err)
	}
	return l, nil
}

func (r *LearningContentRepository) FindQuizByLesson(ctx context.Context, lessonID string) (learning.Quiz, []learning.Question, error) {
	const quizQ = `SELECT id, lesson_id FROM quizzes WHERE lesson_id = ?`
	var quiz learning.Quiz
	if err := r.db.QueryRowContext(ctx, quizQ, lessonID).Scan(&quiz.ID, &quiz.LessonID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return learning.Quiz{}, nil, learning.ErrQuizNotFound
		}
		return learning.Quiz{}, nil, fmt.Errorf("mysql: find quiz: %w", err)
	}

	const questionsQ = `
		SELECT id, quiz_id, question, order_index
		FROM quiz_questions WHERE quiz_id = ? ORDER BY order_index ASC
	`
	qRows, err := r.db.QueryContext(ctx, questionsQ, quiz.ID)
	if err != nil {
		return learning.Quiz{}, nil, fmt.Errorf("mysql: list quiz questions: %w", err)
	}
	defer qRows.Close() //nolint:errcheck // read-only cursor, nothing actionable on close failure

	var questions []learning.Question
	questionIndex := make(map[string]int)
	for qRows.Next() {
		var q learning.Question
		if err := qRows.Scan(&q.ID, &q.QuizID, &q.Text, &q.OrderIndex); err != nil {
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
	args := make([]any, len(questions))
	for i, q := range questions {
		placeholders[i] = "?"
		args[i] = q.ID
	}
	optionsQ := `
		SELECT id, question_id, text, is_correct, order_index
		FROM quiz_options WHERE question_id IN (` + strings.Join(placeholders, ",") + `)
		ORDER BY question_id, order_index ASC
	`
	oRows, err := r.db.QueryContext(ctx, optionsQ, args...)
	if err != nil {
		return learning.Quiz{}, nil, fmt.Errorf("mysql: list quiz options: %w", err)
	}
	defer oRows.Close() //nolint:errcheck // read-only cursor, nothing actionable on close failure

	for oRows.Next() {
		var o learning.Option
		if err := oRows.Scan(&o.ID, &o.QuestionID, &o.Text, &o.IsCorrect, &o.OrderIndex); err != nil {
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

func (r *LearningContentRepository) FindBadgeByLevel(ctx context.Context, levelID string) (learning.Badge, error) {
	const q = `SELECT id, level_id, name, description, icon_key FROM badges WHERE level_id = ?`
	var b learning.Badge
	err := r.db.QueryRowContext(ctx, q, levelID).Scan(&b.ID, &b.LevelID, &b.Name, &b.Description, &b.IconKey)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return learning.Badge{}, learning.ErrBadgeNotFound
		}
		return learning.Badge{}, fmt.Errorf("mysql: find badge by level: %w", err)
	}
	return b, nil
}

func (r *LearningContentRepository) ListBadgesByIDs(ctx context.Context, ids []string) ([]learning.Badge, error) {
	if len(ids) == 0 {
		return []learning.Badge{}, nil
	}

	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	q := `SELECT id, level_id, name, description, icon_key FROM badges WHERE id IN (` +
		strings.Join(placeholders, ",") + `)`

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("mysql: list badges by ids: %w", err)
	}
	defer rows.Close() //nolint:errcheck // read-only cursor, nothing actionable on close failure

	badges := make([]learning.Badge, 0, len(ids))
	for rows.Next() {
		var b learning.Badge
		if err := rows.Scan(&b.ID, &b.LevelID, &b.Name, &b.Description, &b.IconKey); err != nil {
			return nil, fmt.Errorf("mysql: scan badge: %w", err)
		}
		badges = append(badges, b)
	}
	return badges, rows.Err()
}
