package dto

import (
	"time"

	"github.com/traumacompendium/backend/internal/domain/assistant"
	"github.com/traumacompendium/backend/internal/service/assistantservice"
)

type SendMessageRequest struct {
	ConversationID string `json:"conversation_id" binding:"omitempty,uuid"`
	Message        string `json:"message" binding:"required,max=4000"`
}

type SendMessageResponse struct {
	ConversationID string `json:"conversation_id"`
	Reply          string `json:"reply"`
}

func NewSendMessageResponse(r assistantservice.SendMessageResult) SendMessageResponse {
	return SendMessageResponse{ConversationID: r.ConversationID, Reply: r.Reply}
}

type ConversationResponse struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

func NewConversationResponse(c *assistant.Conversation) ConversationResponse {
	return ConversationResponse{ID: c.ID, CreatedAt: c.CreatedAt}
}

func NewConversationResponses(conversations []*assistant.Conversation) []ConversationResponse {
	out := make([]ConversationResponse, len(conversations))
	for i, c := range conversations {
		out[i] = NewConversationResponse(c)
	}
	return out
}

type AssistantMessageResponse struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

func NewAssistantMessageResponse(m *assistant.Message) AssistantMessageResponse {
	return AssistantMessageResponse{ID: m.ID, Role: string(m.Role), Content: m.Content, CreatedAt: m.CreatedAt}
}

func NewAssistantMessageResponses(messages []*assistant.Message) []AssistantMessageResponse {
	out := make([]AssistantMessageResponse, len(messages))
	for i, m := range messages {
		out[i] = NewAssistantMessageResponse(m)
	}
	return out
}
