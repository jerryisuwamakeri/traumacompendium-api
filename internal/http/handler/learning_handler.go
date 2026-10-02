package handler

import (
	"github.com/gin-gonic/gin"

	apphttp "github.com/traumacompendium/backend/internal/http"
	"github.com/traumacompendium/backend/internal/http/dto"
	"github.com/traumacompendium/backend/internal/http/middleware"
	"github.com/traumacompendium/backend/internal/service/learningservice"
)

type LearningHandler struct {
	svc *learningservice.Service
}

func NewLearningHandler(svc *learningservice.Service) *LearningHandler {
	return &LearningHandler{svc: svc}
}

// GET /api/v1/learning/levels
func (h *LearningHandler) ListLevels(c *gin.Context) {
	levels, err := h.svc.ListLevels(c.Request.Context())
	if err != nil {
		apphttp.Fail(c, err)
		return
	}
	apphttp.OK(c, dto.NewLevelResponses(levels))
}

// GET /api/v1/learning/levels/:id/lessons
func (h *LearningHandler) ListLessons(c *gin.Context) {
	lessons, err := h.svc.ListLessons(c.Request.Context(), c.Param("id"))
	if err != nil {
		apphttp.Fail(c, err)
		return
	}
	apphttp.OK(c, dto.NewLessonResponses(lessons))
}

// GET /api/v1/learning/lessons/:id
func (h *LearningHandler) GetLesson(c *gin.Context) {
	lesson, err := h.svc.GetLesson(c.Request.Context(), c.Param("id"))
	if err != nil {
		apphttp.Fail(c, err)
		return
	}
	apphttp.OK(c, dto.NewLessonResponse(lesson))
}

// POST /api/v1/learning/lessons/:id/complete
func (h *LearningHandler) CompleteLesson(c *gin.Context) {
	result, err := h.svc.CompleteLesson(c.Request.Context(), middleware.UserID(c), c.Param("id"))
	if err != nil {
		apphttp.Fail(c, err)
		return
	}
	apphttp.OK(c, dto.NewCompleteLessonResponse(result))
}

// GET /api/v1/learning/lessons/:id/quiz
func (h *LearningHandler) GetQuiz(c *gin.Context) {
	quiz, questions, err := h.svc.GetQuiz(c.Request.Context(), c.Param("id"))
	if err != nil {
		apphttp.Fail(c, err)
		return
	}
	apphttp.OK(c, dto.NewQuizResponse(quiz, questions))
}

// POST /api/v1/learning/quizzes/:id/submit
//
// :id here is the lesson ID, not the quiz ID — consistent with how the
// quiz is fetched (GET .../lessons/:id/quiz), and it avoids the client
// needing to know the quiz's own ID at all.
func (h *LearningHandler) SubmitQuiz(c *gin.Context) {
	var req dto.SubmitQuizRequest
	if !apphttp.BindJSON(c, &req) {
		return
	}

	result, err := h.svc.SubmitQuiz(c.Request.Context(), c.Param("id"), req.Answers)
	if err != nil {
		apphttp.Fail(c, err)
		return
	}
	apphttp.OK(c, dto.NewQuizResultResponse(result))
}

// GET /api/v1/learning/progress
func (h *LearningHandler) GetProgress(c *gin.Context) {
	progress, err := h.svc.GetProgress(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		apphttp.Fail(c, err)
		return
	}
	apphttp.OK(c, dto.NewLevelProgressResponses(progress))
}

// GET /api/v1/learning/badges
func (h *LearningHandler) ListBadges(c *gin.Context) {
	badges, err := h.svc.ListUserBadges(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		apphttp.Fail(c, err)
		return
	}
	apphttp.OK(c, dto.NewBadgeResponses(badges))
}
