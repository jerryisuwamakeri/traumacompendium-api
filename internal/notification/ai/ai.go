// Package ai talks to an LLM provider for two purposes elsewhere in this
// codebase: the AI assistant's chat replies (educational trauma content,
// never clinical advice — see assistantservice) and community content
// moderation (flagging risky posts/comments into a review queue — see
// communityservice).
//
// Mirrors internal/notification/tts exactly: a Client interface, a
// sandboxClient that never calls out (used in dev/CI/tests), and an
// openAIClient for config.AIConfig.Mode == production, calling OpenAI's
// REST API over plain net/http.
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog"

	"github.com/traumacompendium/backend/internal/config"
)

// Role mirrors OpenAI's chat message roles.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

// Message is a single turn in a chat conversation.
type Message struct {
	Role    Role
	Content string
}

// ChatRequest is a full conversation (system prompt included) to send for
// completion.
type ChatRequest struct {
	Messages []Message
}

// ChatResponse is the model's reply.
type ChatResponse struct {
	Content string
}

// ModerationResult reports whether a piece of text was flagged as
// violating the provider's content policy, and which categories (if any)
// triggered the flag — used only for the moderation queue's context, not
// for any automated decision beyond "flagged or not."
type ModerationResult struct {
	Flagged    bool
	Categories []string
}

// Client is what this codebase needs from an LLM provider.
type Client interface {
	ChatCompletion(ctx context.Context, req ChatRequest) (ChatResponse, error)
	Moderate(ctx context.Context, text string) (ModerationResult, error)
}

// New returns the Client appropriate for cfg.Mode.
func New(cfg config.AIConfig, log zerolog.Logger) Client {
	if cfg.Mode == config.ModeProduction {
		return &openAIClient{
			cfg:        cfg,
			httpClient: &http.Client{Timeout: 30 * time.Second},
		}
	}
	return &sandboxClient{log: log}
}

// sandboxClient never calls a real provider. ChatCompletion returns a
// tagged placeholder reply (so callers exercising the full
// "ask, persist, respond" pipeline in dev/CI get a real, distinguishable
// string end to end); Moderate always reports not-flagged, so sandbox
// content always publishes — there is nothing to moderate without a real
// provider to ask.
type sandboxClient struct {
	log zerolog.Logger
}

func (s *sandboxClient) ChatCompletion(_ context.Context, req ChatRequest) (ChatResponse, error) {
	var lastUser string
	for i := len(req.Messages) - 1; i >= 0; i-- {
		if req.Messages[i].Role == RoleUser {
			lastUser = req.Messages[i].Content
			break
		}
	}

	s.log.Info().
		Str("channel", "ai").
		Str("provider", "openai").
		Str("mode", "sandbox").
		Int("message_count", len(req.Messages)).
		Msg("notification: ai chat completion suppressed in sandbox mode (would have been sent)")

	return ChatResponse{Content: fmt.Sprintf("[sandbox reply] %s", lastUser)}, nil
}

func (s *sandboxClient) Moderate(_ context.Context, text string) (ModerationResult, error) {
	s.log.Info().
		Str("channel", "ai").
		Str("provider", "openai").
		Str("mode", "sandbox").
		Int("text_length", len(text)).
		Msg("notification: ai moderation suppressed in sandbox mode (reporting not-flagged)")
	return ModerationResult{Flagged: false}, nil
}

// openAIClient calls OpenAI's REST API directly.
// https://platform.openai.com/docs/api-reference/chat
// https://platform.openai.com/docs/api-reference/moderations
type openAIClient struct {
	cfg        config.AIConfig
	httpClient *http.Client
}

type chatCompletionRequestBody struct {
	Model    string            `json:"model"`
	Messages []chatMessageBody `json:"messages"`
}

type chatMessageBody struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

type chatCompletionResponseBody struct {
	Choices []struct {
		Message chatMessageBody `json:"message"`
	} `json:"choices"`
}

func (c *openAIClient) ChatCompletion(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	messages := make([]chatMessageBody, len(req.Messages))
	for i, m := range req.Messages {
		messages[i] = chatMessageBody(m)
	}

	body, err := c.post(ctx, "/v1/chat/completions", chatCompletionRequestBody{
		Model:    c.cfg.Model,
		Messages: messages,
	})
	if err != nil {
		return ChatResponse{}, err
	}

	var parsed chatCompletionResponseBody
	if err := json.Unmarshal(body, &parsed); err != nil {
		return ChatResponse{}, fmt.Errorf("ai: decode chat completion response: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return ChatResponse{}, fmt.Errorf("ai: openai returned no choices")
	}

	return ChatResponse{Content: parsed.Choices[0].Message.Content}, nil
}

type moderationRequestBody struct {
	Input string `json:"input"`
}

type moderationResponseBody struct {
	Results []struct {
		Flagged    bool            `json:"flagged"`
		Categories map[string]bool `json:"categories"`
	} `json:"results"`
}

func (c *openAIClient) Moderate(ctx context.Context, text string) (ModerationResult, error) {
	body, err := c.post(ctx, "/v1/moderations", moderationRequestBody{Input: text})
	if err != nil {
		return ModerationResult{}, err
	}

	var parsed moderationResponseBody
	if err := json.Unmarshal(body, &parsed); err != nil {
		return ModerationResult{}, fmt.Errorf("ai: decode moderation response: %w", err)
	}
	if len(parsed.Results) == 0 {
		return ModerationResult{}, fmt.Errorf("ai: openai returned no moderation results")
	}

	result := parsed.Results[0]
	var categories []string
	for name, flagged := range result.Categories {
		if flagged {
			categories = append(categories, name)
		}
	}

	return ModerationResult{Flagged: result.Flagged, Categories: categories}, nil
}

func (c *openAIClient) post(ctx context.Context, path string, payload any) ([]byte, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("ai: encode request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.BaseURL+path, bytes.NewReader(encoded))
	if err != nil {
		return nil, fmt.Errorf("ai: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ai: request openai: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck // response already fully read below

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ai: read openai response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("ai: openai returned %d for %s: %s", resp.StatusCode, path, string(body))
	}

	return body, nil
}
