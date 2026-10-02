package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/traumacompendium/backend/internal/config"
	"github.com/traumacompendium/backend/internal/http/middleware"
	"github.com/traumacompendium/backend/internal/pkg/jwtutil"
	"github.com/traumacompendium/backend/internal/service/assistantservice"
	"github.com/traumacompendium/backend/internal/testsupport/fakes"
)

func newTestAssistantRouter(premium bool) (*gin.Engine, *jwtutil.Manager, *fakes.AIClient) {
	jwt := jwtutil.NewManager(config.JWTConfig{
		AccessSecret:  "test-access-secret",
		RefreshSecret: "test-refresh-secret",
		AccessTTL:     15 * time.Minute,
		RefreshTTL:    30 * 24 * time.Hour,
		Issuer:        "traumacompendium-test",
	})
	aiClient := fakes.NewAIClient()
	svc := assistantservice.New(fakes.NewAssistantRepo(), aiClient, &fakes.PremiumChecker{Premium: premium}, zerolog.Nop())
	h := NewAssistantHandler(svc)

	r := gin.New()
	group := r.Group("/assistant")
	group.Use(middleware.RequireAuth(jwt))
	group.POST("/messages", h.SendMessage)
	group.GET("/conversations", h.ListConversations)
	group.GET("/conversations/:id/messages", h.ListMessages)

	return r, jwt, aiClient
}

func TestAssistantSendMessage_Premium_Success(t *testing.T) {
	r, jwt, aiClient := newTestAssistantRouter(true)
	aiClient.ChatReply = "Grounding is a coping technique."

	req := authedRequest(t, jwt, "user-1", http.MethodPost, "/assistant/messages", []byte(`{"message":"What is grounding?"}`))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("SendMessage status = %d, want %d, body=%s", w.Code, http.StatusCreated, w.Body.String())
	}

	resp := decodeData[struct {
		ConversationID string `json:"conversation_id"`
		Reply          string `json:"reply"`
	}](t, w.Body.Bytes())
	if resp.ConversationID == "" {
		t.Error("ConversationID is empty, want a newly created conversation")
	}
	if resp.Reply != "Grounding is a coping technique." {
		t.Errorf("Reply = %q, want the model's reply", resp.Reply)
	}
}

func TestAssistantSendMessage_NotPremium_402(t *testing.T) {
	r, jwt, _ := newTestAssistantRouter(false)

	req := authedRequest(t, jwt, "user-1", http.MethodPost, "/assistant/messages", []byte(`{"message":"hello"}`))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusPaymentRequired {
		t.Errorf("status = %d, want %d, body=%s", w.Code, http.StatusPaymentRequired, w.Body.String())
	}
}

func TestAssistantSendMessage_EmptyMessage_400(t *testing.T) {
	r, jwt, _ := newTestAssistantRouter(true)

	req := authedRequest(t, jwt, "user-1", http.MethodPost, "/assistant/messages", []byte(`{"message":""}`))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d, body=%s", w.Code, http.StatusBadRequest, w.Body.String())
	}
}

func TestAssistantConversationsAndMessages_ListFlow(t *testing.T) {
	r, jwt, _ := newTestAssistantRouter(true)

	sendReq := authedRequest(t, jwt, "user-1", http.MethodPost, "/assistant/messages", []byte(`{"message":"hello"}`))
	sendW := httptest.NewRecorder()
	r.ServeHTTP(sendW, sendReq)
	sent := decodeData[struct {
		ConversationID string `json:"conversation_id"`
	}](t, sendW.Body.Bytes())

	listReq := authedRequest(t, jwt, "user-1", http.MethodGet, "/assistant/conversations", nil)
	listW := httptest.NewRecorder()
	r.ServeHTTP(listW, listReq)
	if listW.Code != http.StatusOK {
		t.Fatalf("ListConversations status = %d, want %d", listW.Code, http.StatusOK)
	}

	msgReq := authedRequest(t, jwt, "user-1", http.MethodGet, "/assistant/conversations/"+sent.ConversationID+"/messages", nil)
	msgW := httptest.NewRecorder()
	r.ServeHTTP(msgW, msgReq)
	if msgW.Code != http.StatusOK {
		t.Fatalf("ListMessages status = %d, want %d, body=%s", msgW.Code, http.StatusOK, msgW.Body.String())
	}
}

func TestAssistantMessages_WrongUsersConversation_404(t *testing.T) {
	r, jwt, _ := newTestAssistantRouter(true)

	sendReq := authedRequest(t, jwt, "owner", http.MethodPost, "/assistant/messages", []byte(`{"message":"hello"}`))
	sendW := httptest.NewRecorder()
	r.ServeHTTP(sendW, sendReq)
	sent := decodeData[struct {
		ConversationID string `json:"conversation_id"`
	}](t, sendW.Body.Bytes())

	req := authedRequest(t, jwt, "intruder", http.MethodGet, "/assistant/conversations/"+sent.ConversationID+"/messages", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestAssistantRoutes_RejectUnauthenticated(t *testing.T) {
	r, _, _ := newTestAssistantRouter(true)

	req := httptest.NewRequest(http.MethodGet, "/assistant/conversations", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}
