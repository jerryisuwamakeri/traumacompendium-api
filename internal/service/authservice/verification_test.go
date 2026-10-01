package authservice

import (
	"context"
	"testing"
	"time"

	"github.com/traumacompendium/backend/internal/domain/auth"
	"github.com/traumacompendium/backend/internal/pkg/apperror"
	"github.com/traumacompendium/backend/internal/pkg/hash"
)

func registerTestUser(t *testing.T, h *testHarness) string {
	t.Helper()
	u, err := h.service.Register(context.Background(), validRegisterInput())
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	return u.ID
}

// lastOTPCode extracts the raw code embedded in the last message sent to
// the given channel — the fakes capture exactly what the real sender
// would receive, so this mirrors what a user would read off their phone
// or inbox in sandbox mode.
func lastEmailOTPCode(h *testHarness) string {
	body := h.emails.Last().Body
	// "Your verification code is 123456. It expires in ..."
	const marker = "code is "
	i := indexOf(body, marker)
	return body[i+len(marker) : i+len(marker)+6]
}

func indexOf(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func TestVerifyOTP_CorrectCode_MarksEmailVerified(t *testing.T) {
	h := newTestHarness()
	userID := registerTestUser(t, h)
	code := lastEmailOTPCode(h)

	err := h.service.VerifyOTP(context.Background(), userID, auth.OTPPurposeEmailVerification, code)
	if err != nil {
		t.Fatalf("VerifyOTP() error = %v", err)
	}

	u, err := h.users.FindByID(context.Background(), userID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if !u.IsEmailVerified() {
		t.Error("expected user's email to be marked verified")
	}
	if u.IsPhoneVerified() {
		t.Error("phone should remain unverified after verifying email only")
	}
}

func TestVerifyOTP_WrongCode_ReturnsUnauthorizedAndIncrementsAttempts(t *testing.T) {
	h := newTestHarness()
	userID := registerTestUser(t, h)

	err := h.service.VerifyOTP(context.Background(), userID, auth.OTPPurposeEmailVerification, "000000")
	appErr, ok := apperror.As(err)
	if !ok {
		t.Fatalf("VerifyOTP() error = %v, want *apperror.Error", err)
	}
	if appErr.Code != apperror.CodeUnauthorized {
		t.Errorf("Code = %q, want %q", appErr.Code, apperror.CodeUnauthorized)
	}

	record, err := h.otps.FindActive(context.Background(), userID, auth.OTPPurposeEmailVerification)
	if err != nil {
		t.Fatalf("FindActive() error = %v", err)
	}
	if record.AttemptCount != 1 {
		t.Errorf("AttemptCount = %d, want 1", record.AttemptCount)
	}
}

func TestVerifyOTP_TooManyAttempts_ReturnsTooManyRequests(t *testing.T) {
	h := newTestHarness()
	h.service.otpCfg.MaxAttempts = 2
	userID := registerTestUser(t, h)

	for i := 0; i < 2; i++ {
		_ = h.service.VerifyOTP(context.Background(), userID, auth.OTPPurposeEmailVerification, "000000")
	}

	err := h.service.VerifyOTP(context.Background(), userID, auth.OTPPurposeEmailVerification, "000000")
	appErr, ok := apperror.As(err)
	if !ok {
		t.Fatalf("VerifyOTP() error = %v, want *apperror.Error", err)
	}
	if appErr.Code != apperror.CodeTooManyRequests {
		t.Errorf("Code = %q, want %q", appErr.Code, apperror.CodeTooManyRequests)
	}
}

func TestVerifyOTP_ExpiredCode_ReturnsBadRequest(t *testing.T) {
	h := newTestHarness()
	userID := registerTestUser(t, h)
	code := lastEmailOTPCode(h)

	// Fast-forward past the code's TTL.
	h.service.now = func() time.Time { return time.Now().Add(1 * time.Hour) }

	err := h.service.VerifyOTP(context.Background(), userID, auth.OTPPurposeEmailVerification, code)
	appErr, ok := apperror.As(err)
	if !ok {
		t.Fatalf("VerifyOTP() error = %v, want *apperror.Error", err)
	}
	if appErr.Code != apperror.CodeValidation && appErr.Code != apperror.CodeBadRequest {
		t.Errorf("Code = %q, want bad_request", appErr.Code)
	}
}

func TestResendOTP_WithinCooldown_ReturnsTooManyRequests(t *testing.T) {
	h := newTestHarness()
	userID := registerTestUser(t, h)

	err := h.service.ResendOTP(context.Background(), userID, auth.OTPPurposeEmailVerification)
	appErr, ok := apperror.As(err)
	if !ok {
		t.Fatalf("ResendOTP() error = %v, want *apperror.Error", err)
	}
	if appErr.Code != apperror.CodeTooManyRequests {
		t.Errorf("Code = %q, want %q", appErr.Code, apperror.CodeTooManyRequests)
	}
}

func TestResendOTP_AfterCooldown_SendsNewCode(t *testing.T) {
	h := newTestHarness()
	userID := registerTestUser(t, h)

	h.service.now = func() time.Time { return time.Now().Add(2 * time.Minute) }

	if err := h.service.ResendOTP(context.Background(), userID, auth.OTPPurposeEmailVerification); err != nil {
		t.Fatalf("ResendOTP() error = %v", err)
	}

	if len(h.emails.Sent) != 2 {
		t.Fatalf("expected 2 emails sent (register + resend), got %d", len(h.emails.Sent))
	}
}

func TestVerifyOTP_CodeHashedAtRest(t *testing.T) {
	h := newTestHarness()
	userID := registerTestUser(t, h)
	code := lastEmailOTPCode(h)

	record, err := h.otps.FindActive(context.Background(), userID, auth.OTPPurposeEmailVerification)
	if err != nil {
		t.Fatalf("FindActive() error = %v", err)
	}
	if record.CodeHash == code {
		t.Error("CodeHash should not equal the raw code")
	}
	if !hash.TokenMatches(record.CodeHash, code) {
		t.Error("stored hash should match the raw code via TokenMatches")
	}
}
