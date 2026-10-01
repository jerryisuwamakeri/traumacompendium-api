package jwtutil

import (
	"testing"
	"time"

	"github.com/traumacompendium/backend/internal/config"
)

func testManager() *Manager {
	return NewManager(config.JWTConfig{
		AccessSecret:  "access-secret-for-tests",
		RefreshSecret: "refresh-secret-for-tests",
		AccessTTL:     15 * time.Minute,
		RefreshTTL:    30 * 24 * time.Hour,
		Issuer:        "traumacompendium-test",
	})
}

func TestAccessToken_RoundTrip(t *testing.T) {
	m := testManager()

	issued, err := m.GenerateAccessToken("user-123", "user")
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	claims, err := m.ParseAccessToken(issued.Token)
	if err != nil {
		t.Fatalf("ParseAccessToken() error = %v", err)
	}

	if claims.UserID != "user-123" {
		t.Errorf("UserID = %q, want %q", claims.UserID, "user-123")
	}
	if claims.Role != "user" {
		t.Errorf("Role = %q, want %q", claims.Role, "user")
	}
	if claims.Use != UseAccess {
		t.Errorf("Use = %q, want %q", claims.Use, UseAccess)
	}
}

func TestRefreshToken_RoundTrip(t *testing.T) {
	m := testManager()

	issued, err := m.GenerateRefreshToken("user-123")
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}
	if issued.ID == "" {
		t.Fatal("expected a non-empty refresh token ID (jti)")
	}

	claims, err := m.ParseRefreshToken(issued.Token)
	if err != nil {
		t.Fatalf("ParseRefreshToken() error = %v", err)
	}
	if claims.ID != issued.ID {
		t.Errorf("claims.ID = %q, want %q", claims.ID, issued.ID)
	}
}

func TestParse_RejectsWrongTokenUse_CrossSecret(t *testing.T) {
	m := testManager()

	// Access and refresh tokens are signed with different secrets, so an
	// access token presented as a refresh token (or vice versa) already
	// fails at signature verification — the first, strongest line of
	// defense against replaying one token type as the other.
	access, err := m.GenerateAccessToken("user-123", "user")
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}
	if _, err := m.ParseRefreshToken(access.Token); err != ErrInvalidToken {
		t.Errorf("ParseRefreshToken(access token) error = %v, want %v", err, ErrInvalidToken)
	}

	refresh, err := m.GenerateRefreshToken("user-123")
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}
	if _, err := m.ParseAccessToken(refresh.Token); err != ErrInvalidToken {
		t.Errorf("ParseAccessToken(refresh token) error = %v, want %v", err, ErrInvalidToken)
	}
}

func TestParse_RejectsWrongTokenUse_SameSecret(t *testing.T) {
	// Defense in depth: even if access and refresh secrets were ever
	// misconfigured to the same value, the embedded Use claim must still
	// stop a token minted for one purpose from being accepted for the
	// other.
	sharedCfg := config.JWTConfig{
		AccessSecret:  "shared-secret",
		RefreshSecret: "shared-secret",
		AccessTTL:     15 * time.Minute,
		RefreshTTL:    30 * 24 * time.Hour,
		Issuer:        "traumacompendium-test",
	}
	m := NewManager(sharedCfg)

	access, err := m.GenerateAccessToken("user-123", "user")
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}
	if _, err := m.ParseRefreshToken(access.Token); err != ErrWrongTokenUse {
		t.Errorf("ParseRefreshToken(access token, shared secret) error = %v, want %v", err, ErrWrongTokenUse)
	}
}

func TestParse_RejectsCrossSecretForgery(t *testing.T) {
	legit := testManager()
	attacker := NewManager(config.JWTConfig{
		AccessSecret:  "attacker-controlled-secret",
		RefreshSecret: "attacker-controlled-secret-2",
		AccessTTL:     15 * time.Minute,
		RefreshTTL:    30 * 24 * time.Hour,
		Issuer:        "traumacompendium-test",
	})

	forged, err := attacker.GenerateAccessToken("victim-user", "admin")
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	if _, err := legit.ParseAccessToken(forged.Token); err != ErrInvalidToken {
		t.Errorf("ParseAccessToken(forged) error = %v, want %v", err, ErrInvalidToken)
	}
}

func TestParse_RejectsGarbageToken(t *testing.T) {
	m := testManager()

	if _, err := m.ParseAccessToken("not-a-jwt"); err != ErrInvalidToken {
		t.Errorf("ParseAccessToken(garbage) error = %v, want %v", err, ErrInvalidToken)
	}
}

func TestParse_RejectsExpiredToken(t *testing.T) {
	m := NewManager(config.JWTConfig{
		AccessSecret:  "access-secret-for-tests",
		RefreshSecret: "refresh-secret-for-tests",
		AccessTTL:     -1 * time.Minute, // already expired
		RefreshTTL:    30 * 24 * time.Hour,
		Issuer:        "traumacompendium-test",
	})

	issued, err := m.GenerateAccessToken("user-123", "user")
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	if _, err := m.ParseAccessToken(issued.Token); err != ErrInvalidToken {
		t.Errorf("ParseAccessToken(expired) error = %v, want %v", err, ErrInvalidToken)
	}
}
