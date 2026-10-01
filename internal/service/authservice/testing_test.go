package authservice

import (
	"time"

	"github.com/rs/zerolog"

	"github.com/traumacompendium/backend/internal/config"
	"github.com/traumacompendium/backend/internal/pkg/jwtutil"
	"github.com/traumacompendium/backend/internal/testsupport/fakes"
)

// testHarness bundles a Service with its fakes so each test can assert
// against the fakes directly (e.g. "was an email actually sent?").
type testHarness struct {
	service *Service
	users   *fakes.UserRepo
	otps    *fakes.OTPRepo
	tokens  *fakes.RefreshTokenRepo
	emails  *fakes.EmailSender
	smses   *fakes.SMSSender
}

func newTestHarness() *testHarness {
	users := fakes.NewUserRepo()
	otps := fakes.NewOTPRepo()
	tokens := fakes.NewRefreshTokenRepo()
	emails := &fakes.EmailSender{}
	smses := &fakes.SMSSender{}

	jwtManager := jwtutil.NewManager(config.JWTConfig{
		AccessSecret:  "test-access-secret",
		RefreshSecret: "test-refresh-secret",
		AccessTTL:     15 * time.Minute,
		RefreshTTL:    30 * 24 * time.Hour,
		Issuer:        "traumacompendium-test",
	})

	otpCfg := config.OTPConfig{
		Length:      6,
		TTL:         10 * time.Minute,
		MaxAttempts: 5,
	}

	svc := New(users, otps, tokens, jwtManager, emails, smses, otpCfg, zerolog.Nop())

	return &testHarness{service: svc, users: users, otps: otps, tokens: tokens, emails: emails, smses: smses}
}
