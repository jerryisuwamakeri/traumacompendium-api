package mysql

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/traumacompendium/backend/internal/domain/billing"
)

func TestWebhookEventRepository_MarkProcessed_FirstTimeSucceeds(t *testing.T) {
	db := openTestDB(t)
	repo := NewWebhookEventRepository(db)

	if err := repo.MarkProcessed(context.Background(), uuid.NewString()); err != nil {
		t.Fatalf("MarkProcessed() first delivery error = %v", err)
	}
}

func TestWebhookEventRepository_MarkProcessed_Replay_ReturnsAlreadyProcessed(t *testing.T) {
	db := openTestDB(t)
	repo := NewWebhookEventRepository(db)
	key := uuid.NewString()
	ctx := context.Background()

	if err := repo.MarkProcessed(ctx, key); err != nil {
		t.Fatalf("MarkProcessed() first delivery error = %v", err)
	}

	err := repo.MarkProcessed(ctx, key)
	if !errors.Is(err, billing.ErrAlreadyProcessed) {
		t.Errorf("MarkProcessed() replay error = %v, want %v", err, billing.ErrAlreadyProcessed)
	}
}
