package assistantservice

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	"github.com/traumacompendium/backend/internal/domain/assistant"
	"github.com/traumacompendium/backend/internal/notification/ai"
	"github.com/traumacompendium/backend/internal/pkg/apperror"
	"github.com/traumacompendium/backend/internal/pkg/sanitize"
)

type SendMessageInput struct {
	// ConversationID is optional — empty starts a new conversation.
	ConversationID string
	Message        string
}

type SendMessageResult struct {
	ConversationID string
	Reply          string
}

// SendMessage is premium-gated: a non-premium caller gets 402 before
// anything else happens (sanitizing input, touching the conversation,
// calling the model).
func (s *Service) SendMessage(ctx context.Context, userID string, in SendMessageInput) (SendMessageResult, error) {
	premium, err := s.premium.IsPremium(ctx, userID)
	if err != nil {
		return SendMessageResult{}, apperror.NewInternal("could not check subscription").WithCause(err)
	}
	if !premium {
		return SendMessageResult{}, apperror.NewPaymentRequired("the AI assistant is a Premium feature")
	}

	message := strings.TrimSpace(sanitize.StripHTML(in.Message))
	if message == "" {
		return SendMessageResult{}, apperror.NewValidation("message is required")
	}

	conversationID, err := s.resolveConversation(ctx, userID, in.ConversationID)
	if err != nil {
		return SendMessageResult{}, err
	}

	userMsg := &assistant.Message{ID: uuid.NewString(), ConversationID: conversationID, Role: assistant.RoleUser, Content: message, CreatedAt: s.now()}
	if err := s.repo.AddMessage(ctx, userMsg); err != nil {
		return SendMessageResult{}, apperror.NewInternal("could not send message").WithCause(err)
	}

	var reply string
	if containsCrisisLanguage(message) {
		reply = crisisResponse
	} else {
		reply, err = s.generateReply(ctx, conversationID)
		if err != nil {
			return SendMessageResult{}, err
		}
	}

	assistantMsg := &assistant.Message{ID: uuid.NewString(), ConversationID: conversationID, Role: assistant.RoleAssistant, Content: reply, CreatedAt: s.now()}
	if err := s.repo.AddMessage(ctx, assistantMsg); err != nil {
		return SendMessageResult{}, apperror.NewInternal("could not send message").WithCause(err)
	}

	return SendMessageResult{ConversationID: conversationID, Reply: reply}, nil
}

// resolveConversation returns an existing, owner-verified conversation
// ID, or creates a new one when conversationID is empty.
func (s *Service) resolveConversation(ctx context.Context, userID, conversationID string) (string, error) {
	if conversationID == "" {
		conv := &assistant.Conversation{ID: uuid.NewString(), UserID: userID, CreatedAt: s.now()}
		if err := s.repo.CreateConversation(ctx, conv); err != nil {
			return "", apperror.NewInternal("could not start conversation").WithCause(err)
		}
		return conv.ID, nil
	}

	if _, err := s.repo.FindConversation(ctx, conversationID, userID); err != nil {
		if errors.Is(err, assistant.ErrConversationNotFound) {
			return "", apperror.NewNotFound("conversation not found")
		}
		return "", apperror.NewInternal("could not load conversation").WithCause(err)
	}
	return conversationID, nil
}

// generateReply builds the bounded context window and asks the model
// for a reply — only reached for messages containsCrisisLanguage did
// not flag.
func (s *Service) generateReply(ctx context.Context, conversationID string) (string, error) {
	recent, err := s.repo.RecentMessages(ctx, conversationID, contextWindowSize)
	if err != nil {
		return "", apperror.NewInternal("could not load conversation history").WithCause(err)
	}

	chatMessages := make([]ai.Message, 0, len(recent)+1)
	chatMessages = append(chatMessages, ai.Message{Role: ai.RoleSystem, Content: systemPrompt})
	for _, m := range recent {
		role := ai.RoleUser
		if m.Role == assistant.RoleAssistant {
			role = ai.RoleAssistant
		}
		chatMessages = append(chatMessages, ai.Message{Role: role, Content: m.Content})
	}

	resp, err := s.ai.ChatCompletion(ctx, ai.ChatRequest{Messages: chatMessages})
	if err != nil {
		return "", apperror.NewInternal("could not reach the assistant").WithCause(err)
	}
	return resp.Content, nil
}

type ListConversationsResult struct {
	Conversations []*assistant.Conversation
	Total         int
}

func (s *Service) ListConversations(ctx context.Context, userID string, filter assistant.ListFilter) (ListConversationsResult, error) {
	conversations, total, err := s.repo.ListConversations(ctx, userID, filter)
	if err != nil {
		return ListConversationsResult{}, apperror.NewInternal("could not list conversations").WithCause(err)
	}
	return ListConversationsResult{Conversations: conversations, Total: total}, nil
}

type ListMessagesResult struct {
	Messages []*assistant.Message
	Total    int
}

func (s *Service) ListMessages(ctx context.Context, userID, conversationID string, filter assistant.ListFilter) (ListMessagesResult, error) {
	if _, err := s.repo.FindConversation(ctx, conversationID, userID); err != nil {
		if errors.Is(err, assistant.ErrConversationNotFound) {
			return ListMessagesResult{}, apperror.NewNotFound("conversation not found")
		}
		return ListMessagesResult{}, apperror.NewInternal("could not load conversation").WithCause(err)
	}

	messages, total, err := s.repo.ListMessages(ctx, conversationID, filter)
	if err != nil {
		return ListMessagesResult{}, apperror.NewInternal("could not list messages").WithCause(err)
	}
	return ListMessagesResult{Messages: messages, Total: total}, nil
}
