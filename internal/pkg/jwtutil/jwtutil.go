// Package jwtutil issues and verifies the two JWTs used by the API:
// short-lived access tokens (authorize requests) and longer-lived refresh
// tokens (mint new access tokens). They are signed with different secrets
// so a leaked access-token secret alone can't be used to forge a refresh
// token, and vice versa.
package jwtutil

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/traumacompendium/backend/internal/config"
)

var (
	ErrInvalidToken = errors.New("jwtutil: invalid or expired token")
	ErrWrongTokenUse = errors.New("jwtutil: token used for the wrong purpose")
)

// TokenUse distinguishes an access token from a refresh token so one can
// never be presented in place of the other, even though both are valid
// JWTs signed by this service.
type TokenUse string

const (
	UseAccess  TokenUse = "access"
	UseRefresh TokenUse = "refresh"
)

// Claims is the JWT payload for both token kinds. Role is only meaningful
// (and only trusted) on access tokens; refresh tokens carry just enough to
// be looked up and rotated.
type Claims struct {
	UserID string   `json:"uid"`
	Role   string   `json:"role,omitempty"`
	Use    TokenUse `json:"use"`
	jwt.RegisteredClaims
}

// Manager issues and verifies tokens using the configured secrets and
// TTLs. Holding config once at construction keeps call sites simple.
type Manager struct {
	cfg config.JWTConfig
}

func NewManager(cfg config.JWTConfig) *Manager {
	return &Manager{cfg: cfg}
}

// IssuedAccessToken is the result of minting an access token: the signed
// JWT plus its absolute expiry, so callers can relay the expiry to clients
// without re-parsing the token.
type IssuedToken struct {
	Token     string
	ExpiresAt time.Time
	// ID is only set for refresh tokens — it is the JWT's jti claim and is
	// what gets hashed and stored server-side for revocation tracking.
	ID string
}

func (m *Manager) GenerateAccessToken(userID, role string) (IssuedToken, error) {
	now := time.Now()
	expiresAt := now.Add(m.cfg.AccessTTL)

	claims := Claims{
		UserID: userID,
		Role:   role,
		Use:    UseAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.cfg.Issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(m.cfg.AccessSecret))
	if err != nil {
		return IssuedToken{}, fmt.Errorf("jwtutil: sign access token: %w", err)
	}

	return IssuedToken{Token: signed, ExpiresAt: expiresAt}, nil
}

func (m *Manager) GenerateRefreshToken(userID string) (IssuedToken, error) {
	now := time.Now()
	expiresAt := now.Add(m.cfg.RefreshTTL)
	jti := uuid.NewString()

	claims := Claims{
		UserID: userID,
		Use:    UseRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			Issuer:    m.cfg.Issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(m.cfg.RefreshSecret))
	if err != nil {
		return IssuedToken{}, fmt.Errorf("jwtutil: sign refresh token: %w", err)
	}

	return IssuedToken{Token: signed, ExpiresAt: expiresAt, ID: jti}, nil
}

func (m *Manager) ParseAccessToken(tokenString string) (*Claims, error) {
	return m.parse(tokenString, m.cfg.AccessSecret, UseAccess)
}

func (m *Manager) ParseRefreshToken(tokenString string) (*Claims, error) {
	return m.parse(tokenString, m.cfg.RefreshSecret, UseRefresh)
}

func (m *Manager) parse(tokenString, secret string, want TokenUse) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	}, jwt.WithIssuer(m.cfg.Issuer))
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims.Use != want {
		return nil, ErrWrongTokenUse
	}

	return claims, nil
}
