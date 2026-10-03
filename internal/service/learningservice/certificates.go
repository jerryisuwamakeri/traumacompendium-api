package learningservice

import (
	"context"
	"errors"
	"time"

	"github.com/traumacompendium/backend/internal/domain/certificate"
	"github.com/traumacompendium/backend/internal/domain/learning"
	"github.com/traumacompendium/backend/internal/pkg/apperror"
	"github.com/traumacompendium/backend/internal/pkg/pdf"
)

// GetCertificate renders the PDF for a level the user has already
// earned. Premium gating (402) is the HTTP handler's job, same as the
// rest of Levels 2-6 — this only returns 404 when no certificate has
// been issued yet, regardless of plan.
func (s *Service) GetCertificate(ctx context.Context, userID, levelID string, locale learning.Locale) ([]byte, error) {
	locale = resolveLocale(locale)

	level, err := s.content.FindLevel(ctx, levelID, locale)
	if err != nil {
		return nil, mapContentNotFound(err)
	}

	cert, err := s.certificates.FindByUserAndLevel(ctx, userID, levelID)
	if err != nil {
		if errors.Is(err, certificate.ErrNotFound) {
			return nil, apperror.NewNotFound("certificate not yet earned")
		}
		return nil, apperror.NewInternal("could not load certificate").WithCause(err)
	}

	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, apperror.NewInternal("could not load certificate").WithCause(err)
	}

	out, err := pdf.NewCertificate(u.FullName, level.Name, cert.IssuedAt, cert.Serial)
	if err != nil {
		return nil, apperror.NewInternal("could not render certificate").WithCause(err)
	}
	return out, nil
}

// CertificateVerification is the public-facing result of
// GET /certificates/verify/:serial — the response shape with Valid:
// false deliberately carries no other fields, since an invalid serial
// shouldn't let a caller distinguish "wrong serial" from "serial
// belongs to someone else."
type CertificateVerification struct {
	Valid     bool
	UserName  string
	LevelName string
	IssuedAt  time.Time
}

func (s *Service) VerifyCertificate(ctx context.Context, serial string) (CertificateVerification, error) {
	cert, err := s.certificates.FindBySerial(ctx, serial)
	if err != nil {
		if errors.Is(err, certificate.ErrNotFound) {
			return CertificateVerification{Valid: false}, nil
		}
		return CertificateVerification{}, apperror.NewInternal("could not verify certificate").WithCause(err)
	}

	u, err := s.users.FindByID(ctx, cert.UserID)
	if err != nil {
		return CertificateVerification{}, apperror.NewInternal("could not verify certificate").WithCause(err)
	}

	level, err := s.content.FindLevel(ctx, cert.LevelID, learning.DefaultLocale)
	if err != nil {
		return CertificateVerification{}, apperror.NewInternal("could not verify certificate").WithCause(err)
	}

	return CertificateVerification{
		Valid:     true,
		UserName:  u.FullName,
		LevelName: level.Name,
		IssuedAt:  cert.IssuedAt,
	}, nil
}
