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
	"github.com/traumacompendium/backend/internal/domain/user"
	"github.com/traumacompendium/backend/internal/http/middleware"
	"github.com/traumacompendium/backend/internal/pkg/jwtutil"
	"github.com/traumacompendium/backend/internal/service/learningservice"
	"github.com/traumacompendium/backend/internal/testsupport/fakes"
)

// testUserID is the subject every authedRequest in this file signs its
// JWT for — seeded into the fake user repo so GetCertificate/
// VerifyCertificate (which both need a FullName to render) have
// someone to look up.
const testUserID = "user-1"

// level2LessonID is a fixture lesson under a level with Number: 2 —
// used by the premium-gating tests. Everything else in this fixture
// (level-1/lesson-1/etc.) stays free regardless of premium, matching
// Beginner's always-free status.
const level2LessonID = "lesson-2"

func newTestLearningRouter(premium bool) (*gin.Engine, *jwtutil.Manager, string, string) {
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

	level2ID := "level-2"
	content.Levels[level2ID] = learning.Level{ID: level2ID, Number: 2, Name: "Self-Awareness"}
	content.Lessons[level2LessonID] = learning.Lesson{ID: level2LessonID, LevelID: level2ID, Title: "Lesson 2", OrderIndex: 1}
	content.Quizzes[level2LessonID] = learning.Quiz{ID: "quiz-2", LessonID: level2LessonID}

	users := fakes.NewUserRepo()
	users.Seed(&user.User{ID: testUserID, FullName: "Ada Lovelace", Email: "ada@example.com", Phone: "+1000000001"})

	h := NewLearningHandler(
		learningservice.New(content, progress, fakes.NewLearningAudioRepo(), fakes.NewCertificateRepo(), users, &fakes.TTSSynthesizer{}),
		&fakes.PremiumChecker{Premium: premium},
	)

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
	group.GET("/levels/:id/certificate", h.GetCertificate)

	r.GET("/certificates/verify/:serial", h.VerifyCertificate)

	return r, jwt, lessonID, correctOptionID
}

func TestLearningListLevels(t *testing.T) {
	r, jwt, _, _ := newTestLearningRouter(false)

	req := authedRequest(t, jwt, "user-1", http.MethodGet, "/learning/levels", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestLearningGetQuiz_DoesNotLeakIsCorrect(t *testing.T) {
	r, jwt, lessonID, _ := newTestLearningRouter(false)

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
	r, jwt, lessonID, _ := newTestLearningRouter(false)

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
	r, jwt, lessonID, correctOptionID := newTestLearningRouter(false)

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
	r, jwt, lessonID, _ := newTestLearningRouter(false)

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
			Level struct {
				ID string `json:"id"`
			} `json:"level"`
			CompletedLessons int  `json:"completed_lessons"`
			BadgeEarned      bool `json:"badge_earned"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	// The fixture now has two levels (level-1 and the level-2 fixture
	// used by the premium-gating tests) — progress covers the whole
	// roadmap, so find level-1's entry specifically rather than assuming
	// it's the only one.
	var level1 *struct {
		Level struct {
			ID string `json:"id"`
		} `json:"level"`
		CompletedLessons int  `json:"completed_lessons"`
		BadgeEarned      bool `json:"badge_earned"`
	}
	for i := range resp.Data {
		if resp.Data[i].Level.ID == "level-1" {
			level1 = &resp.Data[i]
		}
	}
	if level1 == nil || level1.CompletedLessons != 1 || !level1.BadgeEarned {
		t.Errorf("level-1 progress = %+v, want 1 completed lesson and badge earned", level1)
	}
}

func TestLearningRoutes_RejectUnauthenticated(t *testing.T) {
	r, _, _, _ := newTestLearningRouter(false)

	req := httptest.NewRequest(http.MethodGet, "/learning/levels", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestLearningLevel1_AlwaysFree_EvenWithoutPremium(t *testing.T) {
	r, jwt, lessonID, _ := newTestLearningRouter(false)

	req := authedRequest(t, jwt, "user-1", http.MethodGet, "/learning/lessons/"+lessonID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d, body=%s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestLearningLevel2_WithoutPremium_402(t *testing.T) {
	r, jwt, _, _ := newTestLearningRouter(false)

	cases := []struct {
		name   string
		method string
		path   string
		body   []byte
	}{
		{"ListLessons", http.MethodGet, "/learning/levels/level-2/lessons", nil},
		{"GetLesson", http.MethodGet, "/learning/lessons/" + level2LessonID, nil},
		{"GetQuiz", http.MethodGet, "/learning/lessons/" + level2LessonID + "/quiz", nil},
		{"CompleteLesson", http.MethodPost, "/learning/lessons/" + level2LessonID + "/complete", nil},
		{"SubmitQuiz", http.MethodPost, "/learning/quizzes/" + level2LessonID + "/submit", []byte(`{"answers":{}}`)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := authedRequest(t, jwt, "user-1", tc.method, tc.path, tc.body)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != http.StatusPaymentRequired {
				t.Errorf("%s status = %d, want %d, body=%s", tc.name, w.Code, http.StatusPaymentRequired, w.Body.String())
			}
		})
	}
}

func TestLearningLevel2_WithPremium_Allowed(t *testing.T) {
	r, jwt, _, _ := newTestLearningRouter(true)

	req := authedRequest(t, jwt, "user-1", http.MethodGet, "/learning/lessons/"+level2LessonID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d, body=%s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestLearningLevels_MetadataStaysVisible_WithoutPremium(t *testing.T) {
	r, jwt, _, _ := newTestLearningRouter(false)

	// Browsing the roadmap (level names/descriptions) must never be
	// gated — only the lesson content itself.
	req := authedRequest(t, jwt, "user-1", http.MethodGet, "/learning/levels", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusOK, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "Self-Awareness") {
		t.Error("level 2's metadata is missing from the free roadmap listing")
	}
}

func TestLearningCertificate_WithoutPremium_402(t *testing.T) {
	r, jwt, lessonID, _ := newTestLearningRouter(false)

	// Complete the (free) level first so a 402 can't be mistaken for a 404.
	r.ServeHTTP(httptest.NewRecorder(), authedRequest(t, jwt, testUserID, http.MethodPost, "/learning/lessons/"+lessonID+"/complete", nil))

	req := authedRequest(t, jwt, testUserID, http.MethodGet, "/learning/levels/level-1/certificate", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusPaymentRequired {
		t.Errorf("status = %d, want %d, body=%s", w.Code, http.StatusPaymentRequired, w.Body.String())
	}
}

func TestLearningCertificate_NotYetEarned_404(t *testing.T) {
	r, jwt, _, _ := newTestLearningRouter(true)

	req := authedRequest(t, jwt, testUserID, http.MethodGet, "/learning/levels/level-1/certificate", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d, body=%s", w.Code, http.StatusNotFound, w.Body.String())
	}
}

func TestLearningCertificate_EarnedWithPremium_RendersPDFAndVerifies(t *testing.T) {
	r, jwt, lessonID, _ := newTestLearningRouter(true)

	completeReq := authedRequest(t, jwt, testUserID, http.MethodPost, "/learning/lessons/"+lessonID+"/complete", nil)
	r.ServeHTTP(httptest.NewRecorder(), completeReq)

	req := authedRequest(t, jwt, testUserID, http.MethodGet, "/learning/levels/level-1/certificate", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusOK, w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/pdf" {
		t.Errorf("Content-Type = %q, want application/pdf", ct)
	}
	if !strings.HasPrefix(w.Body.String(), "%PDF-") {
		t.Errorf("body does not start with the PDF header: %q", w.Body.String()[:20])
	}
	if !strings.Contains(w.Body.String(), "Ada Lovelace") {
		t.Error("rendered certificate is missing the recipient's name")
	}

	serial := extractCertificateSerial(t, w.Body.String())

	verifyReq := httptest.NewRequest(http.MethodGet, "/certificates/verify/"+serial, nil)
	verifyW := httptest.NewRecorder()
	r.ServeHTTP(verifyW, verifyReq)

	if verifyW.Code != http.StatusOK {
		t.Fatalf("verify status = %d, want %d, body=%s", verifyW.Code, http.StatusOK, verifyW.Body.String())
	}
	var verifyResp struct {
		Data struct {
			Valid     bool   `json:"valid"`
			UserName  string `json:"user_name"`
			LevelName string `json:"level_name"`
		} `json:"data"`
	}
	if err := json.Unmarshal(verifyW.Body.Bytes(), &verifyResp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !verifyResp.Data.Valid || verifyResp.Data.UserName != "Ada Lovelace" || verifyResp.Data.LevelName != "Beginner" {
		t.Errorf("verify response = %+v, want valid=true, user=Ada Lovelace, level=Beginner", verifyResp.Data)
	}
}

func TestLearningCertificate_VerifyUnknownSerial_ReturnsInvalid(t *testing.T) {
	r, _, _, _ := newTestLearningRouter(true)

	req := httptest.NewRequest(http.MethodGet, "/certificates/verify/TC-DOESNOTEXIST", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusOK, w.Body.String())
	}
	var resp struct {
		Data struct {
			Valid bool `json:"valid"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Data.Valid {
		t.Error("Valid = true for an unknown serial, want false")
	}
}

// extractCertificateSerial pulls the "TC-XXXX..." token out of the
// rendered PDF body's "Certificate No. TC-..." line — the PDF response
// is the only place the test can observe the serial the service minted.
func extractCertificateSerial(t *testing.T, pdfBody string) string {
	t.Helper()
	const marker = "Certificate No. "
	i := strings.Index(pdfBody, marker)
	if i == -1 {
		t.Fatalf("could not find certificate serial marker in PDF body")
	}
	rest := pdfBody[i+len(marker):]
	end := strings.IndexAny(rest, ")\\")
	if end == -1 {
		t.Fatalf("could not find end of certificate serial in PDF body")
	}
	return rest[:end]
}
