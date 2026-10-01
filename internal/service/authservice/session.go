package authservice

import (
	"context"
	"errors"

	authdomain "github.com/traumacompendium/backend/internal/domain/auth"
	"github.com/traumacompendium/backend/internal/pkg/apperror"
	"github.com/traumacompendium/backend/internal/pkg/hash"
)

// RefreshSession validates a refresh token and rotates it: the presented
// token is revoked and a brand-new access/refresh pair is issued. Rotation
// means a stolen-and-replayed refresh token stops working the moment the
// legitimate client refreshes again, surfacing the theft instead of
// granting the attacker an indefinitely renewable session.
func (s *Service) RefreshSession(ctx context.Context, rawRefreshToken string, meta SessionMetadata) (TokenPair, error) {
	claims, err := s.jwt.ParseRefreshToken(rawRefreshToken)
	if err != nil {
		return TokenPair{}, apperror.NewUnauthorized("invalid or expired refresh token")
	}

	record, err := s.refreshTokens.FindByID(ctx, claims.ID)
	if err != nil {
		if errors.Is(err, authdomain.ErrRefreshTokenNotFound) {
			return TokenPair{}, apperror.NewUnauthorized("invalid or expired refresh token")
		}
		return TokenPair{}, apperror.NewInternal("could not refresh session").WithCause(err)
	}

	if !hash.TokenMatches(record.TokenHash, rawRefreshToken) || !record.IsUsable(s.now()) {
		return TokenPair{}, apperror.NewUnauthorized("invalid or expired refresh token")
	}

	if err := s.refreshTokens.Revoke(ctx, record.ID); err != nil {
		return TokenPair{}, apperror.NewInternal("could not refresh session").WithCause(err)
	}

	u, err := s.users.FindByID(ctx, record.UserID)
	if err != nil {
		return TokenPair{}, apperror.NewInternal("could not refresh session").WithCause(err)
	}

	pair, err := s.issueTokenPair(ctx, u, meta)
	if err != nil {
		return TokenPair{}, apperror.NewInternal("could not refresh session").WithCause(err)
	}

	return pair, nil
}

// Logout revokes a single session. It is intentionally idempotent — an
// already-invalid or already-revoked token is not an error, since the
// caller's goal ("make sure this token can't be used again") is already
// satisfied.
func (s *Service) Logout(ctx context.Context, rawRefreshToken string) error {
	claims, err := s.jwt.ParseRefreshToken(rawRefreshToken)
	if err != nil {
		return nil
	}

	if err := s.refreshTokens.Revoke(ctx, claims.ID); err != nil && !errors.Is(err, authdomain.ErrRefreshTokenNotFound) {
		return apperror.NewInternal("could not sign out").WithCause(err)
	}

	return nil
}
