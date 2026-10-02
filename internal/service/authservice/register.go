package authservice

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/traumacompendium/backend/internal/domain/auth"
	"github.com/traumacompendium/backend/internal/domain/user"
	"github.com/traumacompendium/backend/internal/notification/email"
	"github.com/traumacompendium/backend/internal/notification/sms"
	"github.com/traumacompendium/backend/internal/pkg/apperror"
	"github.com/traumacompendium/backend/internal/pkg/hash"
	"github.com/traumacompendium/backend/internal/pkg/otp"
	"github.com/traumacompendium/backend/internal/pkg/sanitize"
)

type RegisterInput struct {
	FullName string
	Email    string
	Phone    string
	Password string
}

// Register creates a new, unverified account and dispatches an OTP to
// both the email address and the phone number. Both channels are sent
// immediately so the user can complete whichever verification step is
// faster for them (SMS delivery can be unreliable); each channel's
// verified_at is tracked independently.
func (s *Service) Register(ctx context.Context, in RegisterInput) (*user.User, error) {
	passwordHash, err := hash.Password(in.Password)
	if err != nil {
		return nil, apperror.NewInternal("could not process password").WithCause(err)
	}

	now := s.now()
	u := &user.User{
		ID:           uuid.NewString(),
		FullName:     strings.TrimSpace(sanitize.StripHTML(in.FullName)),
		Email:        strings.ToLower(strings.TrimSpace(in.Email)),
		Phone:        strings.TrimSpace(in.Phone),
		PasswordHash: passwordHash,
		Role:         user.RoleUser,
		Status:       user.StatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.users.Create(ctx, u); err != nil {
		switch {
		case errors.Is(err, user.ErrEmailTaken):
			return nil, apperror.NewConflict("an account with this email already exists")
		case errors.Is(err, user.ErrPhoneTaken):
			return nil, apperror.NewConflict("an account with this phone number already exists")
		default:
			return nil, apperror.NewInternal("could not create account").WithCause(err)
		}
	}

	if err := s.sendOTP(ctx, u, auth.OTPPurposeEmailVerification); err != nil {
		s.log.Error().Err(err).Str("user_id", u.ID).Msg("authservice: failed to send email verification OTP")
	}
	if err := s.sendOTP(ctx, u, auth.OTPPurposePhoneVerification); err != nil {
		s.log.Error().Err(err).Str("user_id", u.ID).Msg("authservice: failed to send phone verification OTP")
	}

	return u, nil
}

// sendOTP generates a fresh code, persists its hash, and dispatches it
// over the channel matching purpose. Delivery failures are returned to
// the caller to log/handle — registration itself still succeeds, since
// ResendOTP gives the user a way to retry delivery.
func (s *Service) sendOTP(ctx context.Context, u *user.User, purpose auth.OTPPurpose) error {
	code, err := otp.Generate(s.otpCfg.Length)
	if err != nil {
		return fmt.Errorf("authservice: generate otp: %w", err)
	}

	now := s.now()
	record := &auth.OTPCode{
		ID:        uuid.NewString(),
		UserID:    u.ID,
		Purpose:   purpose,
		CodeHash:  hash.Token(code),
		CreatedAt: now,
		ExpiresAt: now.Add(s.otpCfg.TTL),
	}
	if err := s.otps.Create(ctx, record); err != nil {
		return fmt.Errorf("authservice: store otp: %w", err)
	}

	switch purpose {
	case auth.OTPPurposeEmailVerification:
		return s.emailSender.Send(ctx, email.Message{
			To:      u.Email,
			Subject: "Verify your TraumaCompendium email",
			Body:    fmt.Sprintf("Your verification code is %s. It expires in %s.", code, s.otpCfg.TTL),
		})
	case auth.OTPPurposePhoneVerification:
		return s.smsSender.Send(ctx, sms.Message{
			To:   u.Phone,
			Body: fmt.Sprintf("Your TraumaCompendium verification code is %s. It expires in %s.", code, s.otpCfg.TTL),
		})
	case auth.OTPPurposePasswordReset:
		return s.emailSender.Send(ctx, email.Message{
			To:      u.Email,
			Subject: "Reset your TraumaCompendium password",
			Body:    fmt.Sprintf("Your password reset code is %s. It expires in %s. If you did not request this, you can ignore this email.", code, s.otpCfg.TTL),
		})
	default:
		return fmt.Errorf("authservice: unknown otp purpose %q", purpose)
	}
}
