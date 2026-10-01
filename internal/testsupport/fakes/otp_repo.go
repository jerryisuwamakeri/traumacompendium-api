package fakes

import (
	"context"
	"sync"
	"time"

	"github.com/traumacompendium/backend/internal/domain/auth"
)

// OTPRepo is an in-memory auth.OTPRepository.
type OTPRepo struct {
	mu   sync.Mutex
	byID map[string]*auth.OTPCode
}

func NewOTPRepo() *OTPRepo {
	return &OTPRepo{byID: make(map[string]*auth.OTPCode)}
}

func (f *OTPRepo) Create(_ context.Context, code *auth.OTPCode) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	cp := *code
	f.byID[code.ID] = &cp
	return nil
}

func (f *OTPRepo) FindActive(_ context.Context, userID string, purpose auth.OTPPurpose) (*auth.OTPCode, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var latest *auth.OTPCode
	for _, c := range f.byID {
		if c.UserID != userID || c.Purpose != purpose || c.ConsumedAt != nil {
			continue
		}
		if latest == nil || c.CreatedAt.After(latest.CreatedAt) {
			latest = c
		}
	}
	if latest == nil {
		return nil, auth.ErrOTPNotFound
	}
	cp := *latest
	return &cp, nil
}

func (f *OTPRepo) IncrementAttempts(_ context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	c, ok := f.byID[id]
	if !ok {
		return auth.ErrOTPNotFound
	}
	c.AttemptCount++
	return nil
}

func (f *OTPRepo) MarkConsumed(_ context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	c, ok := f.byID[id]
	if !ok {
		return auth.ErrOTPNotFound
	}
	consumedAt := time.Now()
	c.ConsumedAt = &consumedAt
	return nil
}
