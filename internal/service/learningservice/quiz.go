package learningservice

import (
	"context"

	"github.com/traumacompendium/backend/internal/domain/learning"
	"github.com/traumacompendium/backend/internal/pkg/apperror"
)

// GetQuiz returns the quiz for a lesson, options included (with
// IsCorrect populated) — it is the HTTP DTO layer's responsibility to
// strip IsCorrect before this reaches a client taking the quiz. See
// learning.Option's doc comment.
func (s *Service) GetQuiz(ctx context.Context, lessonID string) (learning.Quiz, []learning.Question, error) {
	quiz, questions, err := s.content.FindQuizByLesson(ctx, lessonID)
	if err != nil {
		return learning.Quiz{}, nil, mapContentNotFound(err)
	}
	return quiz, questions, nil
}

// QuizResult is the outcome of scoring a submission. It is not
// persisted — see the package-level design note in service.go — so
// re-submitting the same quiz simply recomputes a fresh result.
type QuizResult struct {
	TotalQuestions int
	CorrectAnswers int
	ScorePercent   float64
}

// SubmitQuiz scores answers (question ID -> chosen option ID) against
// the correct options on file. An unanswered or invalid question ID in
// answers simply counts as incorrect, rather than erroring — a partial
// submission is still meaningfully scorable.
func (s *Service) SubmitQuiz(ctx context.Context, lessonID string, answers map[string]string) (QuizResult, error) {
	_, questions, err := s.content.FindQuizByLesson(ctx, lessonID)
	if err != nil {
		return QuizResult{}, mapContentNotFound(err)
	}
	if len(questions) == 0 {
		return QuizResult{}, apperror.NewUnprocessable("this quiz has no questions")
	}

	correct := 0
	for _, q := range questions {
		chosen, ok := answers[q.ID]
		if !ok {
			continue
		}
		for _, opt := range q.Options {
			if opt.ID == chosen && opt.IsCorrect {
				correct++
				break
			}
		}
	}

	return QuizResult{
		TotalQuestions: len(questions),
		CorrectAnswers: correct,
		ScorePercent:   roundToTwoDecimals(100 * float64(correct) / float64(len(questions))),
	}, nil
}

func roundToTwoDecimals(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}
