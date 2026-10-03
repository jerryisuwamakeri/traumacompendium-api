package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	apphttp "github.com/traumacompendium/backend/internal/http"
	"github.com/traumacompendium/backend/internal/http/dto"
	"github.com/traumacompendium/backend/internal/http/middleware"
	"github.com/traumacompendium/backend/internal/pkg/apperror"
	"github.com/traumacompendium/backend/internal/service/learningservice"
)

// PremiumChecker is the one thing this handler needs from billing — a
// narrow interface (billingservice.Service satisfies it with no
// adapter) so this package doesn't depend on billing's full surface
// just to gate Levels 2-6.
type PremiumChecker interface {
	IsPremium(ctx context.Context, userID string) (bool, error)
}

type LearningHandler struct {
	svc     *learningservice.Service
	billing PremiumChecker
}

func NewLearningHandler(svc *learningservice.Service, billing PremiumChecker) *LearningHandler {
	return &LearningHandler{svc: svc, billing: billing}
}

// GET /api/v1/learning/levels?locale=yo
func (h *LearningHandler) ListLevels(c *gin.Context) {
	levels, err := h.svc.ListLevels(c.Request.Context(), apphttp.ResolveLocale(c))
	if err != nil {
		apphttp.Fail(c, err)
		return
	}
	apphttp.OK(c, dto.NewLevelResponses(levels))
}

// GET /api/v1/learning/levels/:id/lessons?locale=yo
func (h *LearningHandler) ListLessons(c *gin.Context) {
	locale := apphttp.ResolveLocale(c)
	levelID := c.Param("id")

	if err := h.checkLevelAccess(c, levelID, locale); err != nil {
		apphttp.Fail(c, err)
		return
	}

	lessons, err := h.svc.ListLessons(c.Request.Context(), levelID, locale)
	if err != nil {
		apphttp.Fail(c, err)
		return
	}
	apphttp.OK(c, dto.NewLessonResponses(lessons))
}

// GET /api/v1/learning/lessons/:id?locale=yo
func (h *LearningHandler) GetLesson(c *gin.Context) {
	locale := apphttp.ResolveLocale(c)

	lesson, err := h.svc.GetLesson(c.Request.Context(), c.Param("id"), locale)
	if err != nil {
		apphttp.Fail(c, err)
		return
	}
	if err := h.checkLevelAccess(c, lesson.LevelID, locale); err != nil {
		apphttp.Fail(c, err)
		return
	}
	apphttp.OK(c, dto.NewLessonResponse(lesson))
}

// POST /api/v1/learning/lessons/:id/complete?locale=yo
func (h *LearningHandler) CompleteLesson(c *gin.Context) {
	locale := apphttp.ResolveLocale(c)

	if err := h.checkLessonAccess(c, c.Param("id"), locale); err != nil {
		apphttp.Fail(c, err)
		return
	}

	result, err := h.svc.CompleteLesson(c.Request.Context(), middleware.UserID(c), c.Param("id"), locale)
	if err != nil {
		apphttp.Fail(c, err)
		return
	}
	apphttp.OK(c, dto.NewCompleteLessonResponse(result))
}

// GET /api/v1/learning/lessons/:id/quiz?locale=yo
func (h *LearningHandler) GetQuiz(c *gin.Context) {
	locale := apphttp.ResolveLocale(c)

	if err := h.checkLessonAccess(c, c.Param("id"), locale); err != nil {
		apphttp.Fail(c, err)
		return
	}

	quiz, questions, err := h.svc.GetQuiz(c.Request.Context(), c.Param("id"), locale)
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

	locale := apphttp.ResolveLocale(c)
	if err := h.checkLessonAccess(c, c.Param("id"), locale); err != nil {
		apphttp.Fail(c, err)
		return
	}

	result, err := h.svc.SubmitQuiz(c.Request.Context(), c.Param("id"), req.Answers)
	if err != nil {
		apphttp.Fail(c, err)
		return
	}
	apphttp.OK(c, dto.NewQuizResultResponse(result))
}

// GET /api/v1/learning/progress?locale=yo
func (h *LearningHandler) GetProgress(c *gin.Context) {
	progress, err := h.svc.GetProgress(c.Request.Context(), middleware.UserID(c), apphttp.ResolveLocale(c))
	if err != nil {
		apphttp.Fail(c, err)
		return
	}
	apphttp.OK(c, dto.NewLevelProgressResponses(progress))
}

// GET /api/v1/learning/badges?locale=yo
func (h *LearningHandler) ListBadges(c *gin.Context) {
	badges, err := h.svc.ListUserBadges(c.Request.Context(), middleware.UserID(c), apphttp.ResolveLocale(c))
	if err != nil {
		apphttp.Fail(c, err)
		return
	}
	apphttp.OK(c, dto.NewBadgeResponses(badges))
}

// GET /api/v1/learning/lessons/:id/audio?locale=yo
func (h *LearningHandler) GetLessonAudio(c *gin.Context) {
	locale := apphttp.ResolveLocale(c)

	if err := h.checkLessonAccess(c, c.Param("id"), locale); err != nil {
		apphttp.Fail(c, err)
		return
	}

	audio, err := h.svc.GetLessonAudio(c.Request.Context(), c.Param("id"), locale)
	if err != nil {
		apphttp.Fail(c, err)
		return
	}
	c.Data(http.StatusOK, audio.ContentType, audio.Data)
}

// GET /api/v1/learning/levels/:id/certificate
//
// Unlike lesson content, certificate download is Premium-only for
// every level, Beginner included — completing a level is free, but the
// credential itself is a paid perk, so this does not reuse
// checkLevelAccess's level-1 exemption.
func (h *LearningHandler) GetCertificate(c *gin.Context) {
	locale := apphttp.ResolveLocale(c)
	userID := middleware.UserID(c)

	premium, err := h.billing.IsPremium(c.Request.Context(), userID)
	if err != nil {
		apphttp.Fail(c, apperror.NewInternal("could not check subscription").WithCause(err))
		return
	}
	if !premium {
		apphttp.Fail(c, apperror.NewPaymentRequired("certificates are a Premium feature"))
		return
	}

	pdfBytes, err := h.svc.GetCertificate(c.Request.Context(), userID, c.Param("id"), locale)
	if err != nil {
		apphttp.Fail(c, err)
		return
	}
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

// GET /api/v1/certificates/verify/:serial — public, unauthenticated.
func (h *LearningHandler) VerifyCertificate(c *gin.Context) {
	result, err := h.svc.VerifyCertificate(c.Request.Context(), c.Param("serial"))
	if err != nil {
		apphttp.Fail(c, err)
		return
	}
	apphttp.OK(c, dto.NewCertificateVerificationResponse(result))
}

// checkLevelAccess is the premium gate: Beginner (level 1) stays free so
// the brief's "roadmap is browsable" holds, but every level beyond it
// requires an active Premium subscription (see the brief's Premium tier:
// "Full learning paths"). Level/badge metadata (ListLevels, GetProgress,
// ListBadges) is never gated — only the lesson text/quiz content itself.
func (h *LearningHandler) checkLevelAccess(c *gin.Context, levelID, locale string) error {
	level, err := h.svc.GetLevel(c.Request.Context(), levelID, locale)
	if err != nil {
		return err
	}
	if level.Number <= 1 {
		return nil
	}

	premium, err := h.billing.IsPremium(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		return apperror.NewInternal("could not check subscription").WithCause(err)
	}
	if !premium {
		return apperror.NewPaymentRequired("this level's content is a Premium feature")
	}
	return nil
}

// checkLessonAccess resolves a lesson to its level and applies the same
// gate — used by every endpoint that only has a lesson ID, not a level
// ID, to work with (quiz, completion, audio).
func (h *LearningHandler) checkLessonAccess(c *gin.Context, lessonID, locale string) error {
	lesson, err := h.svc.GetLesson(c.Request.Context(), lessonID, locale)
	if err != nil {
		return err
	}
	return h.checkLevelAccess(c, lesson.LevelID, locale)
}
