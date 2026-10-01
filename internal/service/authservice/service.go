// Package authservice implements the business logic for registration,
// verification, login, session (refresh token) management, and password
// reset. It depends only on domain interfaces and the notification
// Sender interfaces — never on database/sql or a concrete provider SDK —
// so it is fully unit-testable with in-memory fakes.
package authservice

import (
	"context"
	"time"

	"github.com/rs/zerolog"

	"github.com/traumacompendium/backend/internal/config"
	authdomain "github.com/traumacompendium/backend/internal/domain/auth"
	"github.com/traumacompendium/backend/internal/domain/user"
	"github.com/traumacompendium/backend/internal/notification/email"
	"github.com/traumacompendium/backend/internal/notification/sms"
	"github.com/traumacompendium/backend/internal/pkg/hash"
	"github.com/traumacompendium/backend/internal/pkg/jwtutil"
)

// resendCooldown is the minimum time a client must wait between
// consecutive OTP resend requests for the same purpose, to curb SMS/email
// cost abuse. It is intentionally a constant rather than a config knob —
// changing anti-abuse tuning should be a code change, not a runtime flag.
const resendCooldown = 60 * time.Second

type Service struct {
	users         user.Repository
	otps          authdomain.OTPRepository
	refreshTokens authdomain.RefreshTokenRepository
	jwt           *jwtutil.Manager
	emailSender   email.Sender
	smsSender     sms.Sender
	otpCfg        config.OTPConfig
	log           zerolog.Logger

	// now is swappable in tests to exercise expiry/cooldown logic
	// deterministically.
	now func() time.Time
}

func New(
	users user.Repository,
	otps authdomain.OTPRepository,
	refreshTokens authdomain.RefreshTokenRepository,
	jwt *jwtutil.Manager,
	emailSender email.Sender,
	smsSender sms.Sender,
	otpCfg config.OTPConfig,
	log zerolog.Logger,
) *Service {
	return &Service{
		users:         users,
		otps:          otps,
		refreshTokens: refreshTokens,
		jwt:           jwt,
		emailSender:   emailSender,
		smsSender:     smsSender,
		otpCfg:        otpCfg,
		log:           log,
		now:           time.Now,
	}
}

// TokenPair is what every flow that establishes or refreshes a session
// (login, refresh) returns to the caller.
type TokenPair struct {
	AccessToken      string
	AccessExpiresAt  time.Time
	RefreshToken     string
	RefreshExpiresAt time.Time
}

// SessionMetadata identifies the client issuing a login/refresh request,
// recorded against the refresh token for auditability and future
// "active sessions" UI.
type SessionMetadata struct {
	UserAgent string
	IPAddress string
}

func (s *Service) issueTokenPair(ctx context.Context, u *user.User, meta SessionMetadata) (TokenPair, error) {
	access, err := s.jwt.GenerateAccessToken(u.ID, string(u.Role))
	if err != nil {
		return TokenPair{}, err
	}

	refresh, err := s.jwt.GenerateRefreshToken(u.ID)
	if err != nil {
		return TokenPair{}, err
	}

	record := &authdomain.RefreshToken{
		ID:        refresh.ID,
		UserID:    u.ID,
		TokenHash: hash.Token(refresh.Token),
		UserAgent: meta.UserAgent,
		IPAddress: meta.IPAddress,
		ExpiresAt: refresh.ExpiresAt,
	}
	if err := s.refreshTokens.Create(ctx, record); err != nil {
		return TokenPair{}, err
	}

	return TokenPair{
		AccessToken:      access.Token,
		AccessExpiresAt:  access.ExpiresAt,
		RefreshToken:     refresh.Token,
		RefreshExpiresAt: refresh.ExpiresAt,
	}, nil
}
