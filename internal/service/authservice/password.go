package authservice

import (
	"context"
	"errors"
	"strings"

	"github.com/traumacompendium/backend/internal/domain/auth"
	"github.com/traumacompendium/backend/internal/domain/user"
	"github.com/traumacompendium/backend/internal/pkg/apperror"
	"github.com/traumacompendium/backend/internal/pkg/hash"
)

// ForgotPassword sends a password-reset code if, and only if, an account
// with the given email exists — but always returns nil either way, so a
// caller (and therefore an attacker probing for valid emails) cannot
// distinguish "sent" from "no such account" from the response alone.
func (s *Service) ForgotPassword(ctx context.Context, emailAddr string) error {
	u, err := s.users.FindByEmail(ctx, strings.ToLower(strings.TrimSpace(emailAddr)))
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return nil
		}
		return apperror.NewInternal("could not process request").WithCause(err)
	}

	if err := s.sendOTP(ctx, u, auth.OTPPurposePasswordReset); err != nil {
		s.log.Error().Err(err).Str("user_id", u.ID).Msg("authservice: failed to send password reset OTP")
	}

	return nil
}

type ResetPasswordInput struct {
	Email       string
	Code        string
	NewPassword string
}

// ResetPassword consumes a password_reset OTP and sets a new password.
// Every existing session (refresh token) for the account is revoked, so a
// credential that leaked before the reset can't still ride an old
// session afterward.
func (s *Service) ResetPassword(ctx context.Context, in ResetPasswordInput) error {
	u, err := s.users.FindByEmail(ctx, strings.ToLower(strings.TrimSpace(in.Email)))
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			// Same error a wrong code would produce — existence must not
			// leak here either.
			return apperror.NewBadRequest("invalid or expired code")
		}
		return apperror.NewInternal("could not reset password").WithCause(err)
	}

	if err := s.consumeOTP(ctx, u.ID, auth.OTPPurposePasswordReset, in.Code); err != nil {
		return err
	}

	newHash, err := hash.Password(in.NewPassword)
	if err != nil {
		return apperror.NewInternal("could not reset password").WithCause(err)
	}
	u.PasswordHash = newHash
	u.UpdatedAt = s.now()

	if err := s.users.Update(ctx, u); err != nil {
		return apperror.NewInternal("could not reset password").WithCause(err)
	}

	if err := s.refreshTokens.RevokeAllForUser(ctx, u.ID); err != nil {
		s.log.Error().Err(err).Str("user_id", u.ID).Msg("authservice: failed to revoke sessions after password reset")
	}

	return nil
}
