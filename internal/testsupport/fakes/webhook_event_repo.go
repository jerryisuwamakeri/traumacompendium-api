package fakes

import (
	"context"
	"sync"

	"github.com/traumacompendium/backend/internal/domain/billing"
)

// WebhookEventRepo is an in-memory billing.WebhookEventRepository.
type WebhookEventRepo struct {
	mu   sync.Mutex
	seen map[string]bool
}

func NewWebhookEventRepo() *WebhookEventRepo {
	return &WebhookEventRepo{seen: make(map[string]bool)}
}

func (f *WebhookEventRepo) MarkProcessed(_ context.Context, dedupeKey string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.seen[dedupeKey] {
		return billing.ErrAlreadyProcessed
	}
	f.seen[dedupeKey] = true
	return nil
}
