package authservice

import (
	"context"
	"strings"
	"testing"

	"github.com/traumacompendium/backend/internal/pkg/apperror"
	"github.com/traumacompendium/backend/internal/testsupport/fakes"
)

func validRegisterInput() RegisterInput {
	return RegisterInput{
		FullName: "Jane Doe",
		Email:    "Jane@Example.com",
		Phone:    "+2348012345678",
		Password: "correct horse battery staple",
	}
}

func TestRegister_CreatesUnverifiedUserAndSendsBothOTPs(t *testing.T) {
	h := newTestHarness()

	u, err := h.service.Register(context.Background(), validRegisterInput())
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if u.Email != "jane@example.com" {
		t.Errorf("Email = %q, want lower-cased %q", u.Email, "jane@example.com")
	}
	if u.IsEmailVerified() || u.IsPhoneVerified() {
		t.Error("newly registered user should not be verified on either channel")
	}
	if u.PasswordHash == "" || u.PasswordHash == "correct horse battery staple" {
		t.Error("PasswordHash should be set and must not be the plaintext password")
	}

	if len(h.emails.Sent) != 1 {
		t.Fatalf("expected 1 email sent, got %d", len(h.emails.Sent))
	}
	if h.emails.Last().To != "jane@example.com" {
		t.Errorf("email sent to %q, want %q", h.emails.Last().To, "jane@example.com")
	}

	if len(h.smses.Sent) != 1 {
		t.Fatalf("expected 1 sms sent, got %d", len(h.smses.Sent))
	}
	if h.smses.Last().To != "+2348012345678" {
		t.Errorf("sms sent to %q, want %q", h.smses.Last().To, "+2348012345678")
	}
}

func TestRegister_DuplicateEmail_ReturnsConflict(t *testing.T) {
	h := newTestHarness()
	ctx := context.Background()

	if _, err := h.service.Register(ctx, validRegisterInput()); err != nil {
		t.Fatalf("first Register() error = %v", err)
	}

	second := validRegisterInput()
	second.Phone = "+2348000000000" // distinct phone, same email

	_, err := h.service.Register(ctx, second)
	appErr, ok := apperror.As(err)
	if !ok {
		t.Fatalf("Register() error = %v, want *apperror.Error", err)
	}
	if appErr.Code != apperror.CodeConflict {
		t.Errorf("Code = %q, want %q", appErr.Code, apperror.CodeConflict)
	}
}

func TestRegister_DuplicatePhone_ReturnsConflict(t *testing.T) {
	h := newTestHarness()
	ctx := context.Background()

	if _, err := h.service.Register(ctx, validRegisterInput()); err != nil {
		t.Fatalf("first Register() error = %v", err)
	}

	second := validRegisterInput()
	second.Email = "someoneelse@example.com" // distinct email, same phone

	_, err := h.service.Register(ctx, second)
	appErr, ok := apperror.As(err)
	if !ok {
		t.Fatalf("Register() error = %v, want *apperror.Error", err)
	}
	if appErr.Code != apperror.CodeConflict {
		t.Errorf("Code = %q, want %q", appErr.Code, apperror.CodeConflict)
	}
}

func TestRegister_OTPDeliveryFailureDoesNotFailRegistration(t *testing.T) {
	h := newTestHarness()
	h.service.emailSender = fakes.FailingEmailSender{}

	u, err := h.service.Register(context.Background(), validRegisterInput())
	if err != nil {
		t.Fatalf("Register() error = %v, want registration to succeed despite delivery failure", err)
	}
	if u == nil {
		t.Fatal("Register() returned nil user")
	}
}

func TestRegister_EmailIsTrimmedAndLowercased(t *testing.T) {
	h := newTestHarness()
	in := validRegisterInput()
	in.Email = "  Jane@Example.COM  "

	u, err := h.service.Register(context.Background(), in)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if strings.ContainsAny(u.Email, " ") || u.Email != "jane@example.com" {
		t.Errorf("Email = %q, want normalized %q", u.Email, "jane@example.com")
	}
}
