// Package userservice implements profile read/update — deliberately kept
// separate from authservice, which owns identity and credentials. Profile
// concerns will grow in later phases (avatar, notification preferences,
// SafeReach trusted contacts) without bloating the auth service.
package userservice

import (
	"context"
	"errors"
	"strings"

	"github.com/traumacompendium/backend/internal/domain/user"
	"github.com/traumacompendium/backend/internal/pkg/apperror"
)

type Service struct {
	users user.Repository
}

func New(users user.Repository) *Service {
	return &Service{users: users}
}

func (s *Service) GetProfile(ctx context.Context, userID string) (*user.User, error) {
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return nil, apperror.NewNotFound("account not found")
		}
		return nil, apperror.NewInternal("could not load profile").WithCause(err)
	}
	return u, nil
}

type UpdateProfileInput struct {
	FullName string
}

// UpdateProfile currently only allows changing the display name. Email
// and phone are intentionally not editable here — changing either would
// need to restart the verification flow (and, for phone, use up Termii
// credit), which belongs in its own deliberate flow rather than a plain
// profile PATCH.
func (s *Service) UpdateProfile(ctx context.Context, userID string, in UpdateProfileInput) (*user.User, error) {
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return nil, apperror.NewNotFound("account not found")
		}
		return nil, apperror.NewInternal("could not update profile").WithCause(err)
	}

	u.FullName = strings.TrimSpace(in.FullName)

	if err := s.users.Update(ctx, u); err != nil {
		return nil, apperror.NewInternal("could not update profile").WithCause(err)
	}

	return u, nil
}
