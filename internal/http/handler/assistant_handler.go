package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/traumacompendium/backend/internal/domain/assistant"
	apphttp "github.com/traumacompendium/backend/internal/http"
	"github.com/traumacompendium/backend/internal/http/dto"
	"github.com/traumacompendium/backend/internal/http/middleware"
	"github.com/traumacompendium/backend/internal/service/assistantservice"
)

type AssistantHandler struct {
	svc *assistantservice.Service
}

func NewAssistantHandler(svc *assistantservice.Service) *AssistantHandler {
	return &AssistantHandler{svc: svc}
}

// POST /api/v1/assistant/messages
func (h *AssistantHandler) SendMessage(c *gin.Context) {
	var req dto.SendMessageRequest
	if !apphttp.BindJSON(c, &req) {
		return
	}

	result, err := h.svc.SendMessage(c.Request.Context(), middleware.UserID(c), assistantservice.SendMessageInput{
		ConversationID: req.ConversationID,
		Message:        req.Message,
	})
	if err != nil {
		apphttp.Fail(c, err)
		return
	}
	apphttp.Created(c, dto.NewSendMessageResponse(result))
}

// GET /api/v1/assistant/conversations?limit=&offset=
func (h *AssistantHandler) ListConversations(c *gin.Context) {
	limit, err := apphttp.ParseLimit(c, 20, 100)
	if err != nil {
		apphttp.Fail(c, err)
		return
	}
	offset, err := apphttp.ParseOffset(c)
	if err != nil {
		apphttp.Fail(c, err)
		return
	}

	result, err := h.svc.ListConversations(c.Request.Context(), middleware.UserID(c), assistant.ListFilter{Limit: limit, Offset: offset})
	if err != nil {
		apphttp.Fail(c, err)
		return
	}
	apphttp.OKWithMeta(c, dto.NewConversationResponses(result.Conversations), gin.H{
		"total": result.Total, "limit": limit, "offset": offset,
	})
}

// GET /api/v1/assistant/conversations/:id/messages?limit=&offset=
func (h *AssistantHandler) ListMessages(c *gin.Context) {
	limit, err := apphttp.ParseLimit(c, 20, 100)
	if err != nil {
		apphttp.Fail(c, err)
		return
	}
	offset, err := apphttp.ParseOffset(c)
	if err != nil {
		apphttp.Fail(c, err)
		return
	}

	result, err := h.svc.ListMessages(c.Request.Context(), middleware.UserID(c), c.Param("id"), assistant.ListFilter{Limit: limit, Offset: offset})
	if err != nil {
		apphttp.Fail(c, err)
		return
	}
	apphttp.OKWithMeta(c, dto.NewAssistantMessageResponses(result.Messages), gin.H{
		"total": result.Total, "limit": limit, "offset": offset,
	})
}
