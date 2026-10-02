package assistantservice

import (
	"context"
	"strings"
	"testing"

	"github.com/rs/zerolog"

	"github.com/traumacompendium/backend/internal/domain/assistant"
	"github.com/traumacompendium/backend/internal/pkg/apperror"
	"github.com/traumacompendium/backend/internal/testsupport/fakes"
)

func newTestService(premium bool) (*Service, *fakes.AssistantRepo, *fakes.AIClient) {
	repo := fakes.NewAssistantRepo()
	aiClient := fakes.NewAIClient()
	svc := New(repo, aiClient, &fakes.PremiumChecker{Premium: premium}, zerolog.Nop())
	return svc, repo, aiClient
}

func TestSendMessage_NotPremium_ReturnsPaymentRequired(t *testing.T) {
	svc, _, _ := newTestService(false)

	_, err := svc.SendMessage(context.Background(), "user-1", SendMessageInput{Message: "hello"})
	appErr, ok := apperror.As(err)
	if !ok || appErr.Code != apperror.CodePaymentRequired {
		t.Errorf("SendMessage() for a non-premium user error = %v, want payment_required", err)
	}
}

func TestSendMessage_EmptyMessage_ReturnsValidationError(t *testing.T) {
	svc, _, _ := newTestService(true)

	_, err := svc.SendMessage(context.Background(), "user-1", SendMessageInput{Message: "   "})
	appErr, ok := apperror.As(err)
	if !ok || appErr.Code != apperror.CodeValidation {
		t.Errorf("SendMessage() with an empty message error = %v, want validation_error", err)
	}
}

func TestSendMessage_NewConversation_PersistsBothTurns(t *testing.T) {
	svc, repo, aiClient := newTestService(true)
	aiClient.ChatReply = "Grounding is a coping technique."
	ctx := context.Background()

	result, err := svc.SendMessage(ctx, "user-1", SendMessageInput{Message: `<b>What</b> is grounding?`})
	if err != nil {
		t.Fatalf("SendMessage() error = %v", err)
	}
	if result.ConversationID == "" {
		t.Fatal("ConversationID is empty, want a newly created conversation")
	}
	if result.Reply != "Grounding is a coping technique." {
		t.Errorf("Reply = %q, want the model's reply", result.Reply)
	}

	messages, _, err := repo.ListMessages(ctx, result.ConversationID, assistant.ListFilter{Limit: 10})
	if err != nil {
		t.Fatalf("ListMessages() error = %v", err)
	}
	if len(messages) != 2 {
		t.Fatalf("len(messages) = %d, want 2 (user + assistant)", len(messages))
	}
	if messages[0].Role != assistant.RoleUser || messages[0].Content != "What is grounding?" {
		t.Errorf("messages[0] = %+v, want the sanitized user message", messages[0])
	}
	if messages[1].Role != assistant.RoleAssistant {
		t.Errorf("messages[1].Role = %q, want assistant", messages[1].Role)
	}
}

func TestSendMessage_ExistingConversation_AppendsToIt(t *testing.T) {
	svc, repo, _ := newTestService(true)
	ctx := context.Background()

	first, err := svc.SendMessage(ctx, "user-1", SendMessageInput{Message: "first message"})
	if err != nil {
		t.Fatalf("SendMessage() first error = %v", err)
	}

	_, err = svc.SendMessage(ctx, "user-1", SendMessageInput{ConversationID: first.ConversationID, Message: "second message"})
	if err != nil {
		t.Fatalf("SendMessage() second error = %v", err)
	}

	messages, total, err := repo.ListMessages(ctx, first.ConversationID, assistant.ListFilter{Limit: 10})
	if err != nil {
		t.Fatalf("ListMessages() error = %v", err)
	}
	if total != 4 {
		t.Errorf("total = %d, want 4 (two user + two assistant turns)", total)
	}
	_ = messages
}

func TestSendMessage_WrongUsersConversation_ReturnsNotFound(t *testing.T) {
	svc, _, _ := newTestService(true)
	ctx := context.Background()

	first, err := svc.SendMessage(ctx, "owner", SendMessageInput{Message: "hello"})
	if err != nil {
		t.Fatalf("SendMessage() error = %v", err)
	}

	_, err = svc.SendMessage(ctx, "intruder", SendMessageInput{ConversationID: first.ConversationID, Message: "hi"})
	appErr, ok := apperror.As(err)
	if !ok || appErr.Code != apperror.CodeNotFound {
		t.Errorf("SendMessage() on another user's conversation error = %v, want not_found", err)
	}
}

func TestSendMessage_CrisisLanguage_SkipsModel_ReturnsFixedResponse(t *testing.T) {
	svc, _, aiClient := newTestService(true)
	ctx := context.Background()

	result, err := svc.SendMessage(ctx, "user-1", SendMessageInput{Message: "I want to kill myself"})
	if err != nil {
		t.Fatalf("SendMessage() error = %v", err)
	}
	if result.Reply != crisisResponse {
		t.Errorf("Reply = %q, want the fixed crisis response", result.Reply)
	}
	if len(aiClient.ChatCalls) != 0 {
		t.Errorf("ChatCompletion was called %d times, want 0 — crisis language must bypass the model entirely", len(aiClient.ChatCalls))
	}
}

func TestSendMessage_SendsSystemPromptAndHistory(t *testing.T) {
	svc, _, aiClient := newTestService(true)
	ctx := context.Background()

	first, err := svc.SendMessage(ctx, "user-1", SendMessageInput{Message: "What is trauma?"})
	if err != nil {
		t.Fatalf("SendMessage() first error = %v", err)
	}
	if _, err := svc.SendMessage(ctx, "user-1", SendMessageInput{ConversationID: first.ConversationID, Message: "Tell me more"}); err != nil {
		t.Fatalf("SendMessage() second error = %v", err)
	}

	if len(aiClient.ChatCalls) != 2 {
		t.Fatalf("len(ChatCalls) = %d, want 2", len(aiClient.ChatCalls))
	}
	lastCall := aiClient.ChatCalls[1]
	if lastCall.Messages[0].Content != systemPrompt {
		t.Errorf("first message role/content = %+v, want the system prompt", lastCall.Messages[0])
	}
	if !strings.Contains(lastCall.Messages[len(lastCall.Messages)-1].Content, "Tell me more") {
		t.Errorf("last message = %+v, want it to be the newest user turn", lastCall.Messages[len(lastCall.Messages)-1])
	}
	// Previous turns (first question + its reply) must be included too.
	found := false
	for _, m := range lastCall.Messages {
		if strings.Contains(m.Content, "What is trauma?") {
			found = true
		}
	}
	if !found {
		t.Error("context window did not include the earlier turn in this conversation")
	}
}

func TestListConversations_OnlyCallersOwn(t *testing.T) {
	svc, _, _ := newTestService(true)
	ctx := context.Background()

	if _, err := svc.SendMessage(ctx, "user-1", SendMessageInput{Message: "hi"}); err != nil {
		t.Fatalf("SendMessage() error = %v", err)
	}
	if _, err := svc.SendMessage(ctx, "user-2", SendMessageInput{Message: "hi"}); err != nil {
		t.Fatalf("SendMessage() error = %v", err)
	}

	result, err := svc.ListConversations(ctx, "user-1", assistant.ListFilter{Limit: 10})
	if err != nil {
		t.Fatalf("ListConversations() error = %v", err)
	}
	if result.Total != 1 {
		t.Errorf("Total = %d, want 1", result.Total)
	}
}

func TestListMessages_WrongUser_ReturnsNotFound(t *testing.T) {
	svc, _, _ := newTestService(true)
	ctx := context.Background()

	first, err := svc.SendMessage(ctx, "owner", SendMessageInput{Message: "hi"})
	if err != nil {
		t.Fatalf("SendMessage() error = %v", err)
	}

	_, err = svc.ListMessages(ctx, "intruder", first.ConversationID, assistant.ListFilter{Limit: 10})
	appErr, ok := apperror.As(err)
	if !ok || appErr.Code != apperror.CodeNotFound {
		t.Errorf("ListMessages() by non-owner error = %v, want not_found", err)
	}
}
