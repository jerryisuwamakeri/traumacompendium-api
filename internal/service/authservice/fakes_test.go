package authservice

import (
	"context"
	"sync"
	"time"

	"github.com/traumacompendium/backend/internal/domain/auth"
	"github.com/traumacompendium/backend/internal/domain/user"
	"github.com/traumacompendium/backend/internal/notification/email"
	"github.com/traumacompendium/backend/internal/notification/sms"
)

// fakeUserRepo is a minimal in-memory user.Repository used to unit test
// authservice without a database.
type fakeUserRepo struct {
	mu      sync.Mutex
	byID    map[string]*user.User
	nextErr error // if set, the next Create call returns this error
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{byID: make(map[string]*user.User)}
}

func (f *fakeUserRepo) Create(_ context.Context, u *user.User) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.nextErr != nil {
		err := f.nextErr
		f.nextErr = nil
		return err
	}
	for _, existing := range f.byID {
		if existing.Email == u.Email {
			return user.ErrEmailTaken
		}
		if existing.Phone == u.Phone {
			return user.ErrPhoneTaken
		}
	}
	cp := *u
	f.byID[u.ID] = &cp
	return nil
}

func (f *fakeUserRepo) Update(_ context.Context, u *user.User) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.byID[u.ID]; !ok {
		return user.ErrNotFound
	}
	cp := *u
	f.byID[u.ID] = &cp
	return nil
}

func (f *fakeUserRepo) FindByID(_ context.Context, id string) (*user.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.byID[id]
	if !ok {
		return nil, user.ErrNotFound
	}
	cp := *u
	return &cp, nil
}

func (f *fakeUserRepo) FindByEmail(_ context.Context, email string) (*user.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, u := range f.byID {
		if u.Email == email {
			cp := *u
			return &cp, nil
		}
	}
	return nil, user.ErrNotFound
}

func (f *fakeUserRepo) FindByPhone(_ context.Context, phone string) (*user.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, u := range f.byID {
		if u.Phone == phone {
			cp := *u
			return &cp, nil
		}
	}
	return nil, user.ErrNotFound
}

// fakeOTPRepo is a minimal in-memory auth.OTPRepository.
type fakeOTPRepo struct {
	mu    sync.Mutex
	byID  map[string]*auth.OTPCode
}

func newFakeOTPRepo() *fakeOTPRepo {
	return &fakeOTPRepo{byID: make(map[string]*auth.OTPCode)}
}

func (f *fakeOTPRepo) Create(_ context.Context, code *auth.OTPCode) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	cp := *code
	f.byID[code.ID] = &cp
	return nil
}

func (f *fakeOTPRepo) FindActive(_ context.Context, userID string, purpose auth.OTPPurpose) (*auth.OTPCode, error) {
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

func (f *fakeOTPRepo) IncrementAttempts(_ context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	c, ok := f.byID[id]
	if !ok {
		return auth.ErrOTPNotFound
	}
	c.AttemptCount++
	return nil
}

func (f *fakeOTPRepo) MarkConsumed(_ context.Context, id string) error {
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

// fakeRefreshTokenRepo is a minimal in-memory auth.RefreshTokenRepository.
type fakeRefreshTokenRepo struct {
	mu   sync.Mutex
	byID map[string]*auth.RefreshToken
}

func newFakeRefreshTokenRepo() *fakeRefreshTokenRepo {
	return &fakeRefreshTokenRepo{byID: make(map[string]*auth.RefreshToken)}
}

func (f *fakeRefreshTokenRepo) Create(_ context.Context, rt *auth.RefreshToken) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	cp := *rt
	f.byID[rt.ID] = &cp
	return nil
}

func (f *fakeRefreshTokenRepo) FindByID(_ context.Context, id string) (*auth.RefreshToken, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	rt, ok := f.byID[id]
	if !ok {
		return nil, auth.ErrRefreshTokenNotFound
	}
	cp := *rt
	return &cp, nil
}

func (f *fakeRefreshTokenRepo) Revoke(_ context.Context, id string) error {
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

func (f *fakeRefreshTokenRepo) RevokeAllForUser(_ context.Context, userID string) error {
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

// fakeEmailSender and fakeSMSSender record every message they're asked to
// send so tests can assert on delivery without a real transport.
type fakeEmailSender struct {
	mu   sync.Mutex
	sent []email.Message
}

func (f *fakeEmailSender) Send(_ context.Context, msg email.Message) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sent = append(f.sent, msg)
	return nil
}

func (f *fakeEmailSender) last() email.Message {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.sent[len(f.sent)-1]
}

type fakeSMSSender struct {
	mu   sync.Mutex
	sent []sms.Message
}

func (f *fakeSMSSender) Send(_ context.Context, msg sms.Message) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sent = append(f.sent, msg)
	return nil
}

func (f *fakeSMSSender) last() sms.Message {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.sent[len(f.sent)-1]
}

// failingEmailSender always fails — used to assert that OTP delivery
// failures don't abort the flow that triggered them.
type failingEmailSender struct{}

func (failingEmailSender) Send(context.Context, email.Message) error {
	return errSimulatedDeliveryFailure
}

var errSimulatedDeliveryFailure = &simulatedError{"simulated delivery failure"}

type simulatedError struct{ msg string }

func (e *simulatedError) Error() string { return e.msg }
