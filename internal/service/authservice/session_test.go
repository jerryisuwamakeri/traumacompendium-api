package authservice

import (
	"context"
	"testing"

	"github.com/traumacompendium/backend/internal/pkg/apperror"
)

func loginTestUser(t *testing.T, h *testHarness) TokenPair {
	t.Helper()
	registerAndVerifyTestUser(t, h)
	_, pair, err := h.service.Login(context.Background(), LoginInput{
		Identifier: "jane@example.com",
		Password:   "correct horse battery staple",
	})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	return pair
}

func TestRefreshSession_Success_RotatesToken(t *testing.T) {
	h := newTestHarness()
	original := loginTestUser(t, h)

	rotated, err := h.service.RefreshSession(context.Background(), original.RefreshToken, SessionMetadata{})
	if err != nil {
		t.Fatalf("RefreshSession() error = %v", err)
	}

	if rotated.RefreshToken == original.RefreshToken {
		t.Error("RefreshSession() returned the same refresh token; expected rotation")
	}
	if rotated.AccessToken == "" {
		t.Error("RefreshSession() returned an empty access token")
	}
}

func TestRefreshSession_ReplayOfRotatedToken_Fails(t *testing.T) {
	h := newTestHarness()
	original := loginTestUser(t, h)

	if _, err := h.service.RefreshSession(context.Background(), original.RefreshToken, SessionMetadata{}); err != nil {
		t.Fatalf("first RefreshSession() error = %v", err)
	}

	// Replaying the now-rotated-away original refresh token must fail —
	// this is what makes rotation useful as theft detection.
	_, err := h.service.RefreshSession(context.Background(), original.RefreshToken, SessionMetadata{})
	appErr, ok := apperror.As(err)
	if !ok || appErr.Code != apperror.CodeUnauthorized {
		t.Errorf("replayed RefreshSession() error = %v, want unauthorized", err)
	}
}

func TestRefreshSession_GarbageToken_ReturnsUnauthorized(t *testing.T) {
	h := newTestHarness()

	_, err := h.service.RefreshSession(context.Background(), "not-a-real-token", SessionMetadata{})
	appErr, ok := apperror.As(err)
	if !ok || appErr.Code != apperror.CodeUnauthorized {
		t.Errorf("RefreshSession() error = %v, want unauthorized", err)
	}
}

func TestLogout_RevokesToken(t *testing.T) {
	h := newTestHarness()
	pair := loginTestUser(t, h)

	if err := h.service.Logout(context.Background(), pair.RefreshToken); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}

	_, err := h.service.RefreshSession(context.Background(), pair.RefreshToken, SessionMetadata{})
	appErr, ok := apperror.As(err)
	if !ok || appErr.Code != apperror.CodeUnauthorized {
		t.Errorf("RefreshSession() after Logout() error = %v, want unauthorized", err)
	}
}

func TestLogout_IsIdempotent(t *testing.T) {
	h := newTestHarness()
	pair := loginTestUser(t, h)

	if err := h.service.Logout(context.Background(), pair.RefreshToken); err != nil {
		t.Fatalf("first Logout() error = %v", err)
	}
	if err := h.service.Logout(context.Background(), pair.RefreshToken); err != nil {
		t.Errorf("second Logout() error = %v, want nil (idempotent)", err)
	}
}

func TestLogout_GarbageToken_DoesNotError(t *testing.T) {
	h := newTestHarness()

	if err := h.service.Logout(context.Background(), "not-a-real-token"); err != nil {
		t.Errorf("Logout(garbage) error = %v, want nil", err)
	}
}
