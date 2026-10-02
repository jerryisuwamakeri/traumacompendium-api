package fakes

import (
	"context"
	"sync"

	"github.com/traumacompendium/backend/internal/notification/ai"
)

// AIClient is a controllable fake ai.Client: tests choose which exact
// text bodies Moderate flags, and can force either call to error.
type AIClient struct {
	mu sync.Mutex

	FlaggedBodies map[string]bool
	Categories    []string
	ModerateErr   error

	ChatReply string
	ChatErr   error
	ChatCalls []ai.ChatRequest
}

func NewAIClient() *AIClient {
	return &AIClient{FlaggedBodies: make(map[string]bool)}
}

func (f *AIClient) ChatCompletion(_ context.Context, req ai.ChatRequest) (ai.ChatResponse, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.ChatCalls = append(f.ChatCalls, req)
	if f.ChatErr != nil {
		return ai.ChatResponse{}, f.ChatErr
	}
	return ai.ChatResponse{Content: f.ChatReply}, nil
}

func (f *AIClient) Moderate(_ context.Context, text string) (ai.ModerationResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.ModerateErr != nil {
		return ai.ModerationResult{}, f.ModerateErr
	}
	return ai.ModerationResult{Flagged: f.FlaggedBodies[text], Categories: f.Categories}, nil
}
