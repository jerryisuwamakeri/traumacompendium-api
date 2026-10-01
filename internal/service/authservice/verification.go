package authservice

import (
	"context"
	"errors"

	"github.com/traumacompendium/backend/internal/domain/auth"
	"github.com/traumacompendium/backend/internal/domain/user"
	"github.com/traumacompendium/backend/internal/pkg/apperror"
	"github.com/traumacompendium/backend/internal/pkg/hash"
)

// consumeOTP validates code against the active OTP for userID+purpose —
// checking the attempt lockout, expiry, and the code itself — and marks it
// consumed on success. It is the shared core behind VerifyOTP (which
// additionally stamps the user's verified_at) and ResetPassword (which
// additionally rotates the password), so the expiry/attempt-limit rules
// live in exactly one place.
func (s *Service) consumeOTP(ctx context.Context, userID string, purpose auth.OTPPurpose, code string) error {
	record, err := s.otps.FindActive(ctx, userID, purpose)
	if err != nil {
		if errors.Is(err, auth.ErrOTPNotFound) {
			return apperror.NewBadRequest("no active verification code found; request a new one")
		}
		return apperror.NewInternal("could not verify code").WithCause(err)
	}

	if record.AttemptCount >= s.otpCfg.MaxAttempts {
		return apperror.NewTooManyRequests("too many incorrect attempts; request a new code")
	}

	if record.IsExpired(s.now()) {
		return apperror.NewBadRequest("verification code has expired; request a new one")
	}

	if !hash.TokenMatches(record.CodeHash, code) {
		if err := s.otps.IncrementAttempts(ctx, record.ID); err != nil {
			s.log.Error().Err(err).Str("otp_id", record.ID).Msg("authservice: failed to record otp attempt")
		}
		return apperror.NewUnauthorized("incorrect verification code")
	}

	if err := s.otps.MarkConsumed(ctx, record.ID); err != nil {
		return apperror.NewInternal("could not complete verification").WithCause(err)
	}

	return nil
}

// VerifyOTP checks a code against the active OTP for userID+purpose and,
// on success, stamps the corresponding verified_at timestamp on the user.
func (s *Service) VerifyOTP(ctx context.Context, userID string, purpose auth.OTPPurpose, code string) error {
	if err := s.consumeOTP(ctx, userID, purpose, code); err != nil {
		return err
	}

	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return apperror.NewInternal("could not complete verification").WithCause(err)
	}

	now := s.now()
	switch purpose {
	case auth.OTPPurposeEmailVerification:
		u.EmailVerifiedAt = &now
	case auth.OTPPurposePhoneVerification:
		u.PhoneVerifiedAt = &now
	}
	u.UpdatedAt = now

	if err := s.users.Update(ctx, u); err != nil {
		return apperror.NewInternal("could not complete verification").WithCause(err)
	}

	return nil
}

// ResendOTP issues a fresh code for userID+purpose, subject to a cooldown
// since the last code was requested (curbs SMS/email cost abuse).
func (s *Service) ResendOTP(ctx context.Context, userID string, purpose auth.OTPPurpose) error {
	existing, err := s.otps.FindActive(ctx, userID, purpose)
	if err != nil && !errors.Is(err, auth.ErrOTPNotFound) {
		return apperror.NewInternal("could not resend code").WithCause(err)
	}
	if err == nil && s.now().Before(existing.CreatedAt.Add(resendCooldown)) {
		return apperror.NewTooManyRequests("please wait before requesting another code")
	}

	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return apperror.NewNotFound("account not found")
		}
		return apperror.NewInternal("could not resend code").WithCause(err)
	}

	if err := s.sendOTP(ctx, u, purpose); err != nil {
		return apperror.NewInternal("could not deliver verification code").WithCause(err)
	}

	return nil
}
