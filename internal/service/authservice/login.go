package authservice

import (
	"context"
	"errors"
	"strings"

	"github.com/traumacompendium/backend/internal/domain/user"
	"github.com/traumacompendium/backend/internal/pkg/apperror"
	"github.com/traumacompendium/backend/internal/pkg/hash"
)

type LoginInput struct {
	// Identifier is either an email address or a phone number — whichever
	// the user finds easier to type in on their device.
	Identifier string
	Password   string
	Meta       SessionMetadata
}

// Login authenticates a user and establishes a new session. Email
// verification is required (SMS delivery can be unreliable and should
// not be a hard login gate in Phase 1); phone verification is not.
func (s *Service) Login(ctx context.Context, in LoginInput) (*user.User, TokenPair, error) {
	u, err := s.findByIdentifier(ctx, in.Identifier)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			// Deliberately the same message as a wrong password: existence
			// of an account must not be leaked through a different error.
			return nil, TokenPair{}, apperror.NewUnauthorized("invalid credentials")
		}
		return nil, TokenPair{}, apperror.NewInternal("could not sign in").WithCause(err)
	}

	if !hash.PasswordMatches(u.PasswordHash, in.Password) {
		return nil, TokenPair{}, apperror.NewUnauthorized("invalid credentials")
	}

	if !u.IsActive() {
		return nil, TokenPair{}, apperror.NewForbidden("this account has been suspended")
	}

	if !u.IsEmailVerified() {
		return nil, TokenPair{}, apperror.NewForbidden("please verify your email before signing in")
	}

	pair, err := s.issueTokenPair(ctx, u, in.Meta)
	if err != nil {
		return nil, TokenPair{}, apperror.NewInternal("could not establish session").WithCause(err)
	}

	return u, pair, nil
}

func (s *Service) findByIdentifier(ctx context.Context, identifier string) (*user.User, error) {
	identifier = strings.TrimSpace(identifier)
	if strings.Contains(identifier, "@") {
		return s.users.FindByEmail(ctx, strings.ToLower(identifier))
	}
	return s.users.FindByPhone(ctx, identifier)
}
