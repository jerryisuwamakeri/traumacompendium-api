package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/traumacompendium/backend/internal/domain/journal"
	apphttp "github.com/traumacompendium/backend/internal/http"
	"github.com/traumacompendium/backend/internal/http/dto"
	"github.com/traumacompendium/backend/internal/http/middleware"
	"github.com/traumacompendium/backend/internal/pkg/apperror"
	"github.com/traumacompendium/backend/internal/service/journalservice"
)

type JournalHandler struct {
	svc *journalservice.Service
}

func NewJournalHandler(svc *journalservice.Service) *JournalHandler {
	return &JournalHandler{svc: svc}
}

// POST /api/v1/journal/entries
func (h *JournalHandler) Create(c *gin.Context) {
	var req dto.CreateJournalEntryRequest
	if !apphttp.BindJSON(c, &req) {
		return
	}

	e, err := h.svc.Create(c.Request.Context(), middleware.UserID(c), journalservice.CreateInput{
		Title:   req.Title,
		Content: req.Content,
	})
	if err != nil {
		apphttp.Fail(c, err)
		return
	}

	apphttp.Created(c, dto.NewJournalEntryResponse(e))
}

// GET /api/v1/journal/entries?from=&to=&limit=&offset=
func (h *JournalHandler) List(c *gin.Context) {
	filter, err := parseJournalListFilter(c)
	if err != nil {
		apphttp.Fail(c, err)
		return
	}

	result, err := h.svc.List(c.Request.Context(), middleware.UserID(c), filter)
	if err != nil {
		apphttp.Fail(c, err)
		return
	}

	apphttp.OKWithMeta(c, dto.NewJournalEntryResponses(result.Entries), gin.H{
		"total":  result.Total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// GET /api/v1/journal/entries/:id
func (h *JournalHandler) Get(c *gin.Context) {
	e, err := h.svc.Get(c.Request.Context(), middleware.UserID(c), c.Param("id"))
	if err != nil {
		apphttp.Fail(c, err)
		return
	}
	apphttp.OK(c, dto.NewJournalEntryResponse(e))
}

// PATCH /api/v1/journal/entries/:id
func (h *JournalHandler) Update(c *gin.Context) {
	var req dto.UpdateJournalEntryRequest
	if !apphttp.BindJSON(c, &req) {
		return
	}

	e, err := h.svc.Update(c.Request.Context(), middleware.UserID(c), c.Param("id"), journalservice.UpdateInput{
		Title:   req.Title,
		Content: req.Content,
	})
	if err != nil {
		apphttp.Fail(c, err)
		return
	}

	apphttp.OK(c, dto.NewJournalEntryResponse(e))
}

// DELETE /api/v1/journal/entries/:id
func (h *JournalHandler) Delete(c *gin.Context) {
	if err := h.svc.Delete(c.Request.Context(), middleware.UserID(c), c.Param("id")); err != nil {
		apphttp.Fail(c, err)
		return
	}
	apphttp.NoContent(c)
}

// GET /api/v1/journal/export?format=json|csv
func (h *JournalHandler) Export(c *gin.Context) {
	format := journalservice.ExportFormat(c.DefaultQuery("format", string(journalservice.ExportFormatJSON)))

	data, contentType, err := h.svc.Export(c.Request.Context(), middleware.UserID(c), format)
	if err != nil {
		apphttp.Fail(c, err)
		return
	}

	filename := "journal-export." + string(format)
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, contentType, data)
}

func parseJournalListFilter(c *gin.Context) (journal.ListFilter, error) {
	filter := journal.ListFilter{Limit: 20}

	if v := c.Query("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 || n > 100 {
			return journal.ListFilter{}, apperror.NewBadRequest("limit must be an integer between 1 and 100")
		}
		filter.Limit = n
	}

	if v := c.Query("offset"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			return journal.ListFilter{}, apperror.NewBadRequest("offset must be a non-negative integer")
		}
		filter.Offset = n
	}

	if v := c.Query("from"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return journal.ListFilter{}, apperror.NewBadRequest("from must be an RFC3339 timestamp")
		}
		filter.From = &t
	}

	if v := c.Query("to"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return journal.ListFilter{}, apperror.NewBadRequest("to must be an RFC3339 timestamp")
		}
		filter.To = &t
	}

	return filter, nil
}
