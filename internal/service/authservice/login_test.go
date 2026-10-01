package authservice

import (
	"context"
	"testing"

	"github.com/traumacompendium/backend/internal/domain/auth"
	"github.com/traumacompendium/backend/internal/pkg/apperror"
)

// registerAndVerifyTestUser registers a user and verifies their email so
// Login tests exercise the "account is ready to sign in" path.
func registerAndVerifyTestUser(t *testing.T, h *testHarness) string {
	t.Helper()
	userID := registerTestUser(t, h)
	code := lastEmailOTPCode(h)
	if err := h.service.VerifyOTP(context.Background(), userID, auth.OTPPurposeEmailVerification, code); err != nil {
		t.Fatalf("VerifyOTP() error = %v", err)
	}
	return userID
}

func TestLogin_Success(t *testing.T) {
	h := newTestHarness()
	registerAndVerifyTestUser(t, h)

	u, pair, err := h.service.Login(context.Background(), LoginInput{
		Identifier: "jane@example.com",
		Password:   "correct horse battery staple",
	})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if u.Email != "jane@example.com" {
		t.Errorf("Email = %q, want %q", u.Email, "jane@example.com")
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Error("Login() returned empty tokens")
	}
}

func TestLogin_ByPhone_Success(t *testing.T) {
	h := newTestHarness()
	registerAndVerifyTestUser(t, h)

	_, pair, err := h.service.Login(context.Background(), LoginInput{
		Identifier: "+2348012345678",
		Password:   "correct horse battery staple",
	})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if pair.AccessToken == "" {
		t.Error("Login() by phone returned empty access token")
	}
}

func TestLogin_WrongPassword_ReturnsUnauthorized(t *testing.T) {
	h := newTestHarness()
	registerAndVerifyTestUser(t, h)

	_, _, err := h.service.Login(context.Background(), LoginInput{
		Identifier: "jane@example.com",
		Password:   "totally-wrong-password",
	})
	appErr, ok := apperror.As(err)
	if !ok || appErr.Code != apperror.CodeUnauthorized {
		t.Errorf("Login() error = %v, want unauthorized", err)
	}
}

func TestLogin_UnknownIdentifier_ReturnsUnauthorizedNotNotFound(t *testing.T) {
	h := newTestHarness()

	// Must not leak account existence: a non-existent identifier gets the
	// exact same error as a wrong password.
	_, _, err := h.service.Login(context.Background(), LoginInput{
		Identifier: "nobody@example.com",
		Password:   "whatever",
	})
	appErr, ok := apperror.As(err)
	if !ok || appErr.Code != apperror.CodeUnauthorized {
		t.Errorf("Login() error = %v, want unauthorized", err)
	}
}

func TestLogin_UnverifiedEmail_ReturnsForbidden(t *testing.T) {
	h := newTestHarness()
	registerTestUser(t, h) // registered but not verified

	_, _, err := h.service.Login(context.Background(), LoginInput{
		Identifier: "jane@example.com",
		Password:   "correct horse battery staple",
	})
	appErr, ok := apperror.As(err)
	if !ok || appErr.Code != apperror.CodeForbidden {
		t.Errorf("Login() error = %v, want forbidden", err)
	}
}
