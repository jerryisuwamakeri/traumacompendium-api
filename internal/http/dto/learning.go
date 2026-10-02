package dto

import (
	"github.com/traumacompendium/backend/internal/domain/learning"
	"github.com/traumacompendium/backend/internal/service/learningservice"
)

type LevelResponse struct {
	ID          string `json:"id"`
	Number      int    `json:"number"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func NewLevelResponse(l learning.Level) LevelResponse {
	return LevelResponse{ID: l.ID, Number: l.Number, Name: l.Name, Description: l.Description}
}

func NewLevelResponses(levels []learning.Level) []LevelResponse {
	out := make([]LevelResponse, len(levels))
	for i, l := range levels {
		out[i] = NewLevelResponse(l)
	}
	return out
}

type LessonResponse struct {
	ID         string `json:"id"`
	LevelID    string `json:"level_id"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	OrderIndex int    `json:"order_index"`
}

func NewLessonResponse(l learning.Lesson) LessonResponse {
	return LessonResponse{ID: l.ID, LevelID: l.LevelID, Title: l.Title, Content: l.Content, OrderIndex: l.OrderIndex}
}

func NewLessonResponses(lessons []learning.Lesson) []LessonResponse {
	out := make([]LessonResponse, len(lessons))
	for i, l := range lessons {
		out[i] = NewLessonResponse(l)
	}
	return out
}

// QuizOptionResponse deliberately omits IsCorrect — see
// learning.Option's doc comment. A client taking the quiz must not be
// able to read the answer key out of the response body.
type QuizOptionResponse struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type QuizQuestionResponse struct {
	ID      string               `json:"id"`
	Text    string               `json:"text"`
	Options []QuizOptionResponse `json:"options"`
}

type QuizResponse struct {
	ID        string                 `json:"id"`
	LessonID  string                 `json:"lesson_id"`
	Questions []QuizQuestionResponse `json:"questions"`
}

func NewQuizResponse(quiz learning.Quiz, questions []learning.Question) QuizResponse {
	qs := make([]QuizQuestionResponse, len(questions))
	for i, q := range questions {
		opts := make([]QuizOptionResponse, len(q.Options))
		for j, o := range q.Options {
			opts[j] = QuizOptionResponse{ID: o.ID, Text: o.Text}
		}
		qs[i] = QuizQuestionResponse{ID: q.ID, Text: q.Text, Options: opts}
	}
	return QuizResponse{ID: quiz.ID, LessonID: quiz.LessonID, Questions: qs}
}

type SubmitQuizRequest struct {
	// Answers maps question ID -> chosen option ID. Omitted questions
	// are scored as incorrect (see SubmitQuiz's doc comment).
	Answers map[string]string `json:"answers" binding:"required"`
}

type QuizResultResponse struct {
	TotalQuestions int     `json:"total_questions"`
	CorrectAnswers int     `json:"correct_answers"`
	ScorePercent   float64 `json:"score_percent"`
}

func NewQuizResultResponse(r learningservice.QuizResult) QuizResultResponse {
	return QuizResultResponse{
		TotalQuestions: r.TotalQuestions,
		CorrectAnswers: r.CorrectAnswers,
		ScorePercent:   r.ScorePercent,
	}
}

type BadgeResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IconKey     string `json:"icon_key"`
}

func NewBadgeResponse(b learning.Badge) BadgeResponse {
	return BadgeResponse{ID: b.ID, Name: b.Name, Description: b.Description, IconKey: b.IconKey}
}

func NewBadgeResponses(badges []learning.Badge) []BadgeResponse {
	out := make([]BadgeResponse, len(badges))
	for i, b := range badges {
		out[i] = NewBadgeResponse(b)
	}
	return out
}

type CompleteLessonResponse struct {
	Lesson         LessonResponse `json:"lesson"`
	LevelCompleted bool           `json:"level_completed"`
	BadgeAwarded   *BadgeResponse `json:"badge_awarded,omitempty"`
}

func NewCompleteLessonResponse(r learningservice.CompleteLessonResult) CompleteLessonResponse {
	resp := CompleteLessonResponse{
		Lesson:         NewLessonResponse(r.Lesson),
		LevelCompleted: r.LevelCompleted,
	}
	if r.BadgeAwarded != nil {
		b := NewBadgeResponse(*r.BadgeAwarded)
		resp.BadgeAwarded = &b
	}
	return resp
}

type LevelProgressResponse struct {
	Level            LevelResponse `json:"level"`
	TotalLessons     int           `json:"total_lessons"`
	CompletedLessons int           `json:"completed_lessons"`
	BadgeEarned      bool          `json:"badge_earned"`
}

func NewLevelProgressResponses(progress []learningservice.LevelProgress) []LevelProgressResponse {
	out := make([]LevelProgressResponse, len(progress))
	for i, p := range progress {
		out[i] = LevelProgressResponse{
			Level:            NewLevelResponse(p.Level),
			TotalLessons:     p.TotalLessons,
			CompletedLessons: p.CompletedLessons,
			BadgeEarned:      p.BadgeEarned,
		}
	}
	return out
}
