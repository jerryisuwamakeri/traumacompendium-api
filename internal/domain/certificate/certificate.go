// Package certificate defines the Certificate aggregate: proof that a
// user completed a learning Level. Unlike SafeReach's share tokens,
// a certificate's serial is not a secret — GET /certificates/verify/:serial
// is a public, unauthenticated endpoint by design, so anyone holding a
// printed or shared certificate can prove it's genuine.
package certificate

import (
	"context"
	"time"
)

type Certificate struct {
	ID       string
	UserID   string
	LevelID  string
	Serial   string
	IssuedAt time.Time
}

type Repository interface {
	Create(ctx context.Context, c *Certificate) error
	// FindByUserAndLevel is used to check whether a certificate has
	// already been issued (CompleteLesson's issuance is idempotent, same
	// posture as AwardBadge) and to regenerate the PDF on demand without
	// minting a new serial each download.
	FindByUserAndLevel(ctx context.Context, userID, levelID string) (Certificate, error)
	FindBySerial(ctx context.Context, serial string) (Certificate, error)
}

// ErrNotFound is returned when no certificate matches.
var ErrNotFound = errNotFound{}

type errNotFound struct{}

func (errNotFound) Error() string { return "certificate: not found" }
