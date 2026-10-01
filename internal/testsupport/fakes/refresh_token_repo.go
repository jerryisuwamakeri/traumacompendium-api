package fakes

import (
	"context"
	"sync"
	"time"

	"github.com/traumacompendium/backend/internal/domain/auth"
)

// RefreshTokenRepo is an in-memory auth.RefreshTokenRepository.
type RefreshTokenRepo struct {
	mu   sync.Mutex
	byID map[string]*auth.RefreshToken
}

func NewRefreshTokenRepo() *RefreshTokenRepo {
	return &RefreshTokenRepo{byID: make(map[string]*auth.RefreshToken)}
}

func (f *RefreshTokenRepo) Create(_ context.Context, rt *auth.RefreshToken) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	cp := *rt
	f.byID[rt.ID] = &cp
	return nil
}

func (f *RefreshTokenRepo) FindByID(_ context.Context, id string) (*auth.RefreshToken, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	rt, ok := f.byID[id]
	if !ok {
		return nil, auth.ErrRefreshTokenNotFound
	}
	cp := *rt
	return &cp, nil
}

func (f *RefreshTokenRepo) Revoke(_ context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	rt, ok := f.byID[id]
	if !ok {
		return auth.ErrRefreshTokenNotFound
	}
	revokedAt := time.Now()
	rt.RevokedAt = &revokedAt
	return nil
}

func (f *RefreshTokenRepo) RevokeAllForUser(_ context.Context, userID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	revokedAt := time.Now()
	for _, rt := range f.byID {
		if rt.UserID == userID && rt.RevokedAt == nil {
			rt.RevokedAt = &revokedAt
		}
	}
	return nil
}
