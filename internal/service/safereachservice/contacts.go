package safereachservice

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/traumacompendium/backend/internal/domain/safereach"
	"github.com/traumacompendium/backend/internal/pkg/apperror"
	"github.com/traumacompendium/backend/internal/pkg/sanitize"
)

type RegisterContactInput struct {
	Name         string
	Relationship string
	Phone        string
	Email        string
}

func (s *Service) RegisterContact(ctx context.Context, userID string, in RegisterContactInput) (*safereach.TrustedContact, error) {
	name := strings.TrimSpace(sanitize.StripHTML(in.Name))
	relationship := strings.TrimSpace(sanitize.StripHTML(in.Relationship))
	phone := strings.TrimSpace(in.Phone)
	emailAddr := strings.ToLower(strings.TrimSpace(in.Email))

	if name == "" {
		return nil, apperror.NewValidation("contact name is required")
	}
	if phone == "" && emailAddr == "" {
		return nil, apperror.NewValidation("a trusted contact needs at least a phone number or an email address")
	}

	count, err := s.contacts.CountByUser(ctx, userID)
	if err != nil {
		return nil, apperror.NewInternal("could not register trusted contact").WithCause(err)
	}
	if count >= safereach.MaxTrustedContacts {
		return nil, apperror.NewUnprocessable(fmt.Sprintf("you can register at most %d trusted contacts", safereach.MaxTrustedContacts))
	}

	c := &safereach.TrustedContact{
		ID:           uuid.NewString(),
		UserID:       userID,
		Name:         name,
		Relationship: relationship,
		Phone:        phone,
		Email:        emailAddr,
		CreatedAt:    s.now(),
	}
	if err := s.contacts.Create(ctx, c); err != nil {
		return nil, apperror.NewInternal("could not register trusted contact").WithCause(err)
	}
	return c, nil
}

func (s *Service) ListContacts(ctx context.Context, userID string) ([]*safereach.TrustedContact, error) {
	contacts, err := s.contacts.ListByUser(ctx, userID)
	if err != nil {
		return nil, apperror.NewInternal("could not list trusted contacts").WithCause(err)
	}
	return contacts, nil
}

func (s *Service) DeleteContact(ctx context.Context, userID, contactID string) error {
	if err := s.contacts.Delete(ctx, contactID, userID); err != nil {
		if errors.Is(err, safereach.ErrContactNotFound) {
			return apperror.NewNotFound("trusted contact not found")
		}
		return apperror.NewInternal("could not delete trusted contact").WithCause(err)
	}
	return nil
}
