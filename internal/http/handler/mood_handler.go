package handler

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/traumacompendium/backend/internal/domain/mood"
	apphttp "github.com/traumacompendium/backend/internal/http"
	"github.com/traumacompendium/backend/internal/http/dto"
	"github.com/traumacompendium/backend/internal/http/middleware"
	"github.com/traumacompendium/backend/internal/service/moodservice"
)

type MoodHandler struct {
	svc *moodservice.Service
}

func NewMoodHandler(svc *moodservice.Service) *MoodHandler {
	return &MoodHandler{svc: svc}
}

// POST /api/v1/mood/entries
func (h *MoodHandler) Create(c *gin.Context) {
	var req dto.CreateMoodEntryRequest
	if !apphttp.BindJSON(c, &req) {
		return
	}

	e, err := h.svc.Create(c.Request.Context(), middleware.UserID(c), moodservice.CreateInput{
		Score: req.Score,
		Tags:  req.Tags,
		Note:  req.Note,
	})
	if err != nil {
		apphttp.Fail(c, err)
		return
	}

	apphttp.Created(c, dto.NewMoodEntryResponse(e))
}

// GET /api/v1/mood/entries?from=&to=&limit=&offset=
func (h *MoodHandler) List(c *gin.Context) {
	filter, err := parseMoodListFilter(c)
	if err != nil {
		apphttp.Fail(c, err)
		return
	}

	result, err := h.svc.List(c.Request.Context(), middleware.UserID(c), filter)
	if err != nil {
		apphttp.Fail(c, err)
		return
	}

	apphttp.OKWithMeta(c, dto.NewMoodEntryResponses(result.Entries), gin.H{
		"total":  result.Total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// GET /api/v1/mood/entries/:id
func (h *MoodHandler) Get(c *gin.Context) {
	e, err := h.svc.Get(c.Request.Context(), middleware.UserID(c), c.Param("id"))
	if err != nil {
		apphttp.Fail(c, err)
		return
	}
	apphttp.OK(c, dto.NewMoodEntryResponse(e))
}

// DELETE /api/v1/mood/entries/:id
func (h *MoodHandler) Delete(c *gin.Context) {
	if err := h.svc.Delete(c.Request.Context(), middleware.UserID(c), c.Param("id")); err != nil {
		apphttp.Fail(c, err)
		return
	}
	apphttp.NoContent(c)
}

// GET /api/v1/mood/summary?from=&to=
func (h *MoodHandler) Summary(c *gin.Context) {
	from, err := apphttp.ParseOptionalTime(c, "from")
	if err != nil {
		apphttp.Fail(c, err)
		return
	}
	to, err := apphttp.ParseOptionalTime(c, "to")
	if err != nil {
		apphttp.Fail(c, err)
		return
	}

	var fromVal, toVal time.Time
	if from != nil {
		fromVal = *from
	}
	if to != nil {
		toVal = *to
	}

	summary, err := h.svc.Summarize(c.Request.Context(), middleware.UserID(c), fromVal, toVal)
	if err != nil {
		apphttp.Fail(c, err)
		return
	}

	apphttp.OK(c, dto.NewMoodSummaryResponse(summary))
}

func parseMoodListFilter(c *gin.Context) (mood.ListFilter, error) {
	limit, err := apphttp.ParseLimit(c, 20, 100)
	if err != nil {
		return mood.ListFilter{}, err
	}
	offset, err := apphttp.ParseOffset(c)
	if err != nil {
		return mood.ListFilter{}, err
	}
	from, err := apphttp.ParseOptionalTime(c, "from")
	if err != nil {
		return mood.ListFilter{}, err
	}
	to, err := apphttp.ParseOptionalTime(c, "to")
	if err != nil {
		return mood.ListFilter{}, err
	}

	return mood.ListFilter{Limit: limit, Offset: offset, From: from, To: to}, nil
}
