package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rs/zerolog"

	"github.com/traumacompendium/backend/internal/config"
)

func TestNew_SelectsImplementationByMode(t *testing.T) {
	sandbox := New(config.AIConfig{Mode: config.ModeSandbox}, zerolog.Nop())
	if _, ok := sandbox.(*sandboxClient); !ok {
		t.Errorf("New(sandbox mode) = %T, want *sandboxClient", sandbox)
	}

	production := New(config.AIConfig{Mode: config.ModeProduction}, zerolog.Nop())
	if _, ok := production.(*openAIClient); !ok {
		t.Errorf("New(production mode) = %T, want *openAIClient", production)
	}
}

func TestSandboxClient_ChatCompletion_ReturnsTaggedPlaceholder_WithoutCallingOut(t *testing.T) {
	c := New(config.AIConfig{Mode: config.ModeSandbox}, zerolog.Nop())

	resp, err := c.ChatCompletion(context.Background(), ChatRequest{Messages: []Message{
		{Role: RoleSystem, Content: "You are a helpful assistant."},
		{Role: RoleUser, Content: "What is grounding?"},
	}})
	if err != nil {
		t.Fatalf("ChatCompletion() error = %v", err)
	}
	if !strings.Contains(resp.Content, "What is grounding?") {
		t.Errorf("Content = %q, want it to echo the last user message", resp.Content)
	}
}

func TestSandboxClient_Moderate_NeverFlags(t *testing.T) {
	c := New(config.AIConfig{Mode: config.ModeSandbox}, zerolog.Nop())

	result, err := c.Moderate(context.Background(), "anything at all")
	if err != nil {
		t.Fatalf("Moderate() error = %v", err)
	}
	if result.Flagged {
		t.Error("sandbox Moderate() flagged content, want never-flagged")
	}
}

func TestOpenAIClient_ChatCompletion_Success(t *testing.T) {
	var received chatCompletionRequestBody

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("path = %s, want /v1/chat/completions", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			t.Errorf("Authorization = %q, want Bearer test-api-key", r.Header.Get("Authorization"))
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"role": "assistant", "content": "Grounding is a coping technique."}},
			},
		})
	}))
	defer server.Close()

	c := &openAIClient{
		cfg:        config.AIConfig{BaseURL: server.URL, APIKey: "test-api-key", Model: "gpt-4o-mini"},
		httpClient: server.Client(),
	}

	resp, err := c.ChatCompletion(context.Background(), ChatRequest{Messages: []Message{
		{Role: RoleUser, Content: "What is grounding?"},
	}})
	if err != nil {
		t.Fatalf("ChatCompletion() error = %v", err)
	}
	if resp.Content != "Grounding is a coping technique." {
		t.Errorf("Content = %q, want %q", resp.Content, "Grounding is a coping technique.")
	}
	if received.Model != "gpt-4o-mini" {
		t.Errorf("request model = %q, want gpt-4o-mini", received.Model)
	}
	if len(received.Messages) != 1 || received.Messages[0].Content != "What is grounding?" {
		t.Errorf("request messages = %+v, want one message asking about grounding", received.Messages)
	}
}

func TestOpenAIClient_ChatCompletion_NoChoices_ReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"choices": []map[string]any{}})
	}))
	defer server.Close()

	c := &openAIClient{
		cfg:        config.AIConfig{BaseURL: server.URL, APIKey: "test-api-key"},
		httpClient: server.Client(),
	}

	_, err := c.ChatCompletion(context.Background(), ChatRequest{Messages: []Message{{Role: RoleUser, Content: "hi"}}})
	if err == nil {
		t.Fatal("ChatCompletion() error = nil, want error when openai returns no choices")
	}
}

func TestOpenAIClient_ChatCompletion_NonOKStatus_ReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":{"message":"invalid api key"}}`))
	}))
	defer server.Close()

	c := &openAIClient{
		cfg:        config.AIConfig{BaseURL: server.URL, APIKey: "bad-key"},
		httpClient: server.Client(),
	}

	_, err := c.ChatCompletion(context.Background(), ChatRequest{Messages: []Message{{Role: RoleUser, Content: "hi"}}})
	if err == nil {
		t.Fatal("ChatCompletion() error = nil, want error for non-2xx response")
	}
}

func TestOpenAIClient_Moderate_Flagged(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/moderations" {
			t.Errorf("path = %s, want /v1/moderations", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"results": []map[string]any{
				{
					"flagged": true,
					"categories": map[string]bool{
						"self-harm": true,
						"violence":  false,
					},
				},
			},
		})
	}))
	defer server.Close()

	c := &openAIClient{
		cfg:        config.AIConfig{BaseURL: server.URL, APIKey: "test-api-key"},
		httpClient: server.Client(),
	}

	result, err := c.Moderate(context.Background(), "some risky text")
	if err != nil {
		t.Fatalf("Moderate() error = %v", err)
	}
	if !result.Flagged {
		t.Error("Flagged = false, want true")
	}
	if len(result.Categories) != 1 || result.Categories[0] != "self-harm" {
		t.Errorf("Categories = %v, want [self-harm]", result.Categories)
	}
}

func TestOpenAIClient_Moderate_NotFlagged(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"results": []map[string]any{{"flagged": false, "categories": map[string]bool{}}},
		})
	}))
	defer server.Close()

	c := &openAIClient{
		cfg:        config.AIConfig{BaseURL: server.URL, APIKey: "test-api-key"},
		httpClient: server.Client(),
	}

	result, err := c.Moderate(context.Background(), "totally fine text")
	if err != nil {
		t.Fatalf("Moderate() error = %v", err)
	}
	if result.Flagged {
		t.Error("Flagged = true, want false")
	}
}
