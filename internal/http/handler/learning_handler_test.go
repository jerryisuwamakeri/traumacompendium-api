package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/traumacompendium/backend/internal/config"
	"github.com/traumacompendium/backend/internal/domain/learning"
	"github.com/traumacompendium/backend/internal/http/middleware"
	"github.com/traumacompendium/backend/internal/pkg/jwtutil"
	"github.com/traumacompendium/backend/internal/service/learningservice"
	"github.com/traumacompendium/backend/internal/testsupport/fakes"
)

func newTestLearningRouter() (*gin.Engine, *jwtutil.Manager, string, string) {
	content := fakes.NewLearningContentRepo()
	progress := fakes.NewLearningProgressRepo()

	levelID, lessonID, questionID, correctOptionID := "level-1", "lesson-1", "question-1", "option-correct"
	content.Levels[levelID] = learning.Level{ID: levelID, Number: 1, Name: "Beginner"}
	content.Lessons[lessonID] = learning.Lesson{ID: lessonID, LevelID: levelID, Title: "Lesson 1", OrderIndex: 1}
	content.Quizzes[lessonID] = learning.Quiz{ID: "quiz-1", LessonID: lessonID}
	content.Questions["quiz-1"] = []learning.Question{
		{
			ID: questionID, QuizID: "quiz-1", Text: "2+2?", OrderIndex: 1,
			Options: []learning.Option{
				{ID: correctOptionID, QuestionID: questionID, Text: "4", IsCorrect: true, OrderIndex: 1},
				{ID: "option-wrong", QuestionID: questionID, Text: "5", IsCorrect: false, OrderIndex: 2},
			},
		},
	}
	content.Badges[levelID] = learning.Badge{ID: "badge-1", LevelID: levelID, Name: "Beginner"}

	h := NewLearningHandler(learningservice.New(content, progress, fakes.NewLearningAudioRepo(), &fakes.TTSSynthesizer{}))

	jwt := jwtutil.NewManager(config.JWTConfig{
		AccessSecret:  "test-access-secret",
		RefreshSecret: "test-refresh-secret",
		AccessTTL:     15 * time.Minute,
		RefreshTTL:    30 * 24 * time.Hour,
		Issuer:        "traumacompendium-test",
	})

	r := gin.New()
	group := r.Group("/learning")
	group.Use(middleware.RequireAuth(jwt))
	group.GET("/levels", h.ListLevels)
	group.GET("/levels/:id/lessons", h.ListLessons)
	group.GET("/lessons/:id", h.GetLesson)
	group.POST("/lessons/:id/complete", h.CompleteLesson)
	group.GET("/lessons/:id/quiz", h.GetQuiz)
	group.POST("/quizzes/:id/submit", h.SubmitQuiz)
	group.GET("/progress", h.GetProgress)
	group.GET("/badges", h.ListBadges)

	return r, jwt, lessonID, correctOptionID
}

func TestLearningListLevels(t *testing.T) {
	r, jwt, _, _ := newTestLearningRouter()

	req := authedRequest(t, jwt, "user-1", http.MethodGet, "/learning/levels", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestLearningGetQuiz_DoesNotLeakIsCorrect(t *testing.T) {
	r, jwt, lessonID, _ := newTestLearningRouter()

	req := authedRequest(t, jwt, "user-1", http.MethodGet, "/learning/lessons/"+lessonID+"/quiz", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusOK, w.Body.String())
	}

	// Security property: the wire response must not contain the answer
	// key under any field name, not just the ones our own DTO happens to
	// define — guards against someone later adding Option.IsCorrect to a
	// struct that gets serialized wholesale.
	if strings.Contains(strings.ToLower(w.Body.String()), "iscorrect") || strings.Contains(w.Body.String(), "is_correct") {
		t.Errorf("quiz response leaks the answer key: %s", w.Body.String())
	}
}

func TestLearningCompleteLesson_SingleLessonLevel_AwardsBadge(t *testing.T) {
	r, jwt, lessonID, _ := newTestLearningRouter()

	req := authedRequest(t, jwt, "user-1", http.MethodPost, "/learning/lessons/"+lessonID+"/complete", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp struct {
		Data struct {
			LevelCompleted bool `json:"level_completed"`
			BadgeAwarded   *struct {
				Name string `json:"name"`
			} `json:"badge_awarded"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !resp.Data.LevelCompleted {
		t.Error("LevelCompleted = false, want true (only lesson in the level)")
	}
	if resp.Data.BadgeAwarded == nil || resp.Data.BadgeAwarded.Name != "Beginner" {
		t.Errorf("BadgeAwarded = %+v, want {Beginner}", resp.Data.BadgeAwarded)
	}
}

func TestLearningSubmitQuiz_CorrectAnswer_FullScore(t *testing.T) {
	r, jwt, lessonID, correctOptionID := newTestLearningRouter()

	body := `{"answers":{"question-1":"` + correctOptionID + `"}}`
	req := authedRequest(t, jwt, "user-1", http.MethodPost, "/learning/quizzes/"+lessonID+"/submit", []byte(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp struct {
		Data struct {
			ScorePercent float64 `json:"score_percent"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Data.ScorePercent != 100 {
		t.Errorf("ScorePercent = %v, want 100", resp.Data.ScorePercent)
	}
}

func TestLearningProgress_ReflectsCompletion(t *testing.T) {
	r, jwt, lessonID, _ := newTestLearningRouter()

	completeReq := authedRequest(t, jwt, "user-1", http.MethodPost, "/learning/lessons/"+lessonID+"/complete", nil)
	r.ServeHTTP(httptest.NewRecorder(), completeReq)

	req := authedRequest(t, jwt, "user-1", http.MethodGet, "/learning/progress", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp struct {
		Data []struct {
			CompletedLessons int  `json:"completed_lessons"`
			BadgeEarned      bool `json:"badge_earned"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Data) != 1 || resp.Data[0].CompletedLessons != 1 || !resp.Data[0].BadgeEarned {
		t.Errorf("progress = %+v, want 1 level with 1 completed lesson and badge earned", resp.Data)
	}
}

func TestLearningRoutes_RejectUnauthenticated(t *testing.T) {
	r, _, _, _ := newTestLearningRouter()

	req := httptest.NewRequest(http.MethodGet, "/learning/levels", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}
