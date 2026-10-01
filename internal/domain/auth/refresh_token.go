package auth

import (
	"context"
	"time"
)

// RefreshToken is a server-side record of an issued refresh token,
// keyed by the JWT's jti (see internal/pkg/jwtutil). Storing a hash of
// the token (never the raw value) lets us revoke a single session
// (logout) or every session for a user (e.g. after a password reset)
// without needing a shared revocation list for every access token too —
// access tokens are short-lived and simply expire on their own.
type RefreshToken struct {
	ID        string // the JWT's jti claim
	UserID    string
	TokenHash string
	UserAgent string
	IPAddress string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

func (r *RefreshToken) IsRevoked() bool           { return r.RevokedAt != nil }
func (r *RefreshToken) IsExpired(now time.Time) bool { return now.After(r.ExpiresAt) }
func (r *RefreshToken) IsUsable(now time.Time) bool  { return !r.IsRevoked() && !r.IsExpired(now) }

// RefreshTokenRepository persists refresh token sessions.
type RefreshTokenRepository interface {
	Create(ctx context.Context, rt *RefreshToken) error
	FindByID(ctx context.Context, id string) (*RefreshToken, error)
	Revoke(ctx context.Context, id string) error
	RevokeAllForUser(ctx context.Context, userID string) error
}

var ErrRefreshTokenNotFound = errRefreshTokenNotFound{}

type errRefreshTokenNotFound struct{}

func (errRefreshTokenNotFound) Error() string { return "auth: refresh token not found" }
