package learningservice

import (
	"context"
	"testing"

	"github.com/traumacompendium/backend/internal/pkg/apperror"
)

func TestGetQuiz_IncludesIsCorrect(t *testing.T) {
	// Confirms the service layer itself does carry IsCorrect through (it
	// is the DTO layer's job to strip it before this reaches a client
	// taking the quiz — see http/dto.NewQuizResponse).
	svc, content, _ := newTestService()
	_, _, lesson2ID, _, _, correctOptionID, _ := seedTwoLessonLevel(content)

	_, questions, err := svc.GetQuiz(context.Background(), lesson2ID)
	if err != nil {
		t.Fatalf("GetQuiz() error = %v", err)
	}

	found := false
	for _, opt := range questions[0].Options {
		if opt.ID == correctOptionID {
			found = true
			if !opt.IsCorrect {
				t.Error("expected the correct option's IsCorrect to be true")
			}
		}
	}
	if !found {
		t.Fatal("correct option not found in quiz response")
	}
}

func TestSubmitQuiz_AllCorrect_100Percent(t *testing.T) {
	svc, content, _ := newTestService()
	_, _, lesson2ID, _, questionID, correctOptionID, _ := seedTwoLessonLevel(content)

	result, err := svc.SubmitQuiz(context.Background(), lesson2ID, map[string]string{questionID: correctOptionID})
	if err != nil {
		t.Fatalf("SubmitQuiz() error = %v", err)
	}
	if result.CorrectAnswers != 1 || result.TotalQuestions != 1 {
		t.Errorf("result = %+v, want 1/1", result)
	}
	if result.ScorePercent != 100 {
		t.Errorf("ScorePercent = %v, want 100", result.ScorePercent)
	}
}

func TestSubmitQuiz_WrongAnswer_0Percent(t *testing.T) {
	svc, content, _ := newTestService()
	_, _, lesson2ID, _, questionID, _, wrongOptionID := seedTwoLessonLevel(content)

	result, err := svc.SubmitQuiz(context.Background(), lesson2ID, map[string]string{questionID: wrongOptionID})
	if err != nil {
		t.Fatalf("SubmitQuiz() error = %v", err)
	}
	if result.CorrectAnswers != 0 {
		t.Errorf("CorrectAnswers = %d, want 0", result.CorrectAnswers)
	}
}

func TestSubmitQuiz_UnansweredQuestion_CountsAsIncorrect(t *testing.T) {
	svc, content, _ := newTestService()
	_, _, lesson2ID, _, _, _, _ := seedTwoLessonLevel(content)

	result, err := svc.SubmitQuiz(context.Background(), lesson2ID, map[string]string{})
	if err != nil {
		t.Fatalf("SubmitQuiz() error = %v", err)
	}
	if result.CorrectAnswers != 0 || result.TotalQuestions != 1 {
		t.Errorf("result = %+v, want 0/1", result)
	}
}

func TestSubmitQuiz_NoQuizForLesson_ReturnsNotFound(t *testing.T) {
	svc, content, _ := newTestService()
	_, lesson1ID, _, _, _, _, _ := seedTwoLessonLevel(content) // lesson1 has no quiz

	_, err := svc.SubmitQuiz(context.Background(), lesson1ID, map[string]string{})
	appErr, ok := apperror.As(err)
	if !ok || appErr.Code != apperror.CodeNotFound {
		t.Errorf("SubmitQuiz() error = %v, want not_found", err)
	}
}
