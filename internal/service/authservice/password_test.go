package authservice

import (
	"context"
	"testing"

	"github.com/traumacompendium/backend/internal/pkg/apperror"
	"github.com/traumacompendium/backend/internal/pkg/hash"
)

func lastPasswordResetCode(h *testHarness) string {
	body := h.emails.last().Body
	const marker = "code is "
	i := indexOf(body, marker)
	return body[i+len(marker) : i+len(marker)+6]
}

func TestForgotPassword_ExistingAccount_SendsEmail(t *testing.T) {
	h := newTestHarness()
	registerTestUser(t, h)

	if err := h.service.ForgotPassword(context.Background(), "jane@example.com"); err != nil {
		t.Fatalf("ForgotPassword() error = %v", err)
	}

	// 1 from registration + 1 from the reset request
	if len(h.emails.sent) != 2 {
		t.Fatalf("expected 2 emails sent, got %d", len(h.emails.sent))
	}
}

func TestForgotPassword_UnknownEmail_StillReturnsNil(t *testing.T) {
	h := newTestHarness()

	err := h.service.ForgotPassword(context.Background(), "nobody@example.com")
	if err != nil {
		t.Errorf("ForgotPassword(unknown email) error = %v, want nil (must not leak existence)", err)
	}
	if len(h.emails.sent) != 0 {
		t.Errorf("expected no email sent for unknown address, got %d", len(h.emails.sent))
	}
}

func TestResetPassword_Success_AllowsLoginWithNewPassword(t *testing.T) {
	h := newTestHarness()
	registerAndVerifyTestUser(t, h)

	if err := h.service.ForgotPassword(context.Background(), "jane@example.com"); err != nil {
		t.Fatalf("ForgotPassword() error = %v", err)
	}
	code := lastPasswordResetCode(h)

	err := h.service.ResetPassword(context.Background(), ResetPasswordInput{
		Email:       "jane@example.com",
		Code:        code,
		NewPassword: "a brand new strong password",
	})
	if err != nil {
		t.Fatalf("ResetPassword() error = %v", err)
	}

	_, _, loginErr := h.service.Login(context.Background(), LoginInput{
		Identifier: "jane@example.com",
		Password:   "a brand new strong password",
	})
	if loginErr != nil {
		t.Errorf("Login() with new password error = %v", loginErr)
	}

	_, _, oldLoginErr := h.service.Login(context.Background(), LoginInput{
		Identifier: "jane@example.com",
		Password:   "correct horse battery staple",
	})
	if oldLoginErr == nil {
		t.Error("Login() with old password succeeded, want failure after reset")
	}
}

func TestResetPassword_RevokesExistingSessions(t *testing.T) {
	h := newTestHarness()
	pair := loginTestUser(t, h)

	if err := h.service.ForgotPassword(context.Background(), "jane@example.com"); err != nil {
		t.Fatalf("ForgotPassword() error = %v", err)
	}
	code := lastPasswordResetCode(h)

	err := h.service.ResetPassword(context.Background(), ResetPasswordInput{
		Email:       "jane@example.com",
		Code:        code,
		NewPassword: "a brand new strong password",
	})
	if err != nil {
		t.Fatalf("ResetPassword() error = %v", err)
	}

	_, refreshErr := h.service.RefreshSession(context.Background(), pair.RefreshToken, SessionMetadata{})
	appErr, ok := apperror.As(refreshErr)
	if !ok || appErr.Code != apperror.CodeUnauthorized {
		t.Errorf("RefreshSession() after password reset error = %v, want unauthorized", refreshErr)
	}
}

func TestResetPassword_WrongCode_DoesNotChangePassword(t *testing.T) {
	h := newTestHarness()
	registerAndVerifyTestUser(t, h)

	if err := h.service.ForgotPassword(context.Background(), "jane@example.com"); err != nil {
		t.Fatalf("ForgotPassword() error = %v", err)
	}

	err := h.service.ResetPassword(context.Background(), ResetPasswordInput{
		Email:       "jane@example.com",
		Code:        "000000",
		NewPassword: "a brand new strong password",
	})
	if err == nil {
		t.Fatal("ResetPassword() with wrong code error = nil, want error")
	}

	u, findErr := h.users.FindByEmail(context.Background(), "jane@example.com")
	if findErr != nil {
		t.Fatalf("FindByEmail() error = %v", findErr)
	}
	if !hash.PasswordMatches(u.PasswordHash, "correct horse battery staple") {
		t.Error("password was changed despite an incorrect reset code")
	}
}

func TestResetPassword_UnknownEmail_ReturnsBadRequestNotNotFound(t *testing.T) {
	h := newTestHarness()

	err := h.service.ResetPassword(context.Background(), ResetPasswordInput{
		Email:       "nobody@example.com",
		Code:        "123456",
		NewPassword: "a brand new strong password",
	})
	appErr, ok := apperror.As(err)
	if !ok || appErr.Code != apperror.CodeBadRequest {
		t.Errorf("ResetPassword(unknown email) error = %v, want bad_request (must not leak existence via not_found)", err)
	}
}

// Sanity check that consumeOTP is purpose-scoped: a code issued for email
// verification must not be usable to reset a password, even for the same
// user.
func TestResetPassword_CannotReuseEmailVerificationCode(t *testing.T) {
	h := newTestHarness()
	registerTestUser(t, h)
	emailVerificationCode := lastEmailOTPCode(h)

	err := h.service.ResetPassword(context.Background(), ResetPasswordInput{
		Email:       "jane@example.com",
		Code:        emailVerificationCode,
		NewPassword: "a brand new strong password",
	})
	if err == nil {
		t.Fatal("ResetPassword() using an email-verification code succeeded, want it to be rejected")
	}
}
