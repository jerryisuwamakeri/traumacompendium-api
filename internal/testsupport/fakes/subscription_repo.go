package fakes

import (
	"context"
	"sync"

	"github.com/traumacompendium/backend/internal/domain/billing"
)

// SubscriptionRepo is an in-memory billing.Repository.
type SubscriptionRepo struct {
	mu       sync.Mutex
	byUserID map[string]*billing.Subscription
	byRef    map[string]*billing.Subscription
}

func NewSubscriptionRepo() *SubscriptionRepo {
	return &SubscriptionRepo{
		byUserID: make(map[string]*billing.Subscription),
		byRef:    make(map[string]*billing.Subscription),
	}
}

func (f *SubscriptionRepo) Upsert(_ context.Context, s *billing.Subscription) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	cp := *s
	f.byUserID[s.UserID] = &cp
	f.byRef[s.PaystackReference] = &cp
	return nil
}

func (f *SubscriptionRepo) FindByUserID(_ context.Context, userID string) (*billing.Subscription, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	s, ok := f.byUserID[userID]
	if !ok {
		return nil, billing.ErrNotFound
	}
	cp := *s
	return &cp, nil
}

func (f *SubscriptionRepo) FindByReference(_ context.Context, reference string) (*billing.Subscription, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	s, ok := f.byRef[reference]
	if !ok {
		return nil, billing.ErrNotFound
	}
	cp := *s
	return &cp, nil
}
