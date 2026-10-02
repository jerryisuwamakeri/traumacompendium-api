package fakes

import (
	"context"
	"sync"

	"github.com/traumacompendium/backend/internal/domain/devicetoken"
)

// DeviceTokenRepo is an in-memory devicetoken.Repository.
type DeviceTokenRepo struct {
	mu      sync.Mutex
	byToken map[string]*devicetoken.DeviceToken
}

func NewDeviceTokenRepo() *DeviceTokenRepo {
	return &DeviceTokenRepo{byToken: make(map[string]*devicetoken.DeviceToken)}
}

// Register mirrors the real repository's upsert semantics: re-registering
// the same token repoints it to whichever user registered it last.
func (f *DeviceTokenRepo) Register(_ context.Context, d *devicetoken.DeviceToken) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	cp := *d
	f.byToken[d.Token] = &cp
	return nil
}

func (f *DeviceTokenRepo) ListByUser(_ context.Context, userID string) ([]*devicetoken.DeviceToken, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []*devicetoken.DeviceToken
	for _, d := range f.byToken {
		if d.UserID == userID {
			cp := *d
			out = append(out, &cp)
		}
	}
	return out, nil
}
