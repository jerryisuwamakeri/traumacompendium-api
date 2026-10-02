// Package safereach implements the SafeReach emergency-alert system: a
// user's trusted contacts, and alerts that dispatch to them (SMS/email/
// push, see internal/service/safereachservice) with a time-limited share
// link for live status/location. This is rapid trusted-contact alerting,
// not a 911/112 replacement — the brief is explicit about that, and
// nothing here talks to real emergency dispatch.
package safereach

import (
	"context"
	"time"
)

// Level is the urgency of an alert, matching the brief's three tiers.
type Level string

const (
	LevelCheckIn  Level = "check_in"
	LevelUrgent   Level = "urgent"
	LevelCritical Level = "critical"
)

func IsValidLevel(l Level) bool {
	switch l {
	case LevelCheckIn, LevelUrgent, LevelCritical:
		return true
	default:
		return false
	}
}

// Status is an alert's lifecycle state.
type Status string

const (
	StatusActive    Status = "active"
	StatusCancelled Status = "cancelled"
	StatusResolved  Status = "resolved"
)

// IsOpen reports whether the alert is still live — accepting new
// location updates and reachable via its share link.
func (s Status) IsOpen() bool { return s == StatusActive }

// MaxTrustedContacts bounds how many contacts one user can register.
// Enforced by the service (not the schema) so the limit can change
// without a migration.
const MaxTrustedContacts = 5

// TrustedContact is someone a user wants notified during an alert. A
// contact is just contact info the owning user manages — not
// necessarily a platform account of their own.
type TrustedContact struct {
	ID           string
	UserID       string
	Name         string
	Relationship string
	Phone        string
	Email        string
	CreatedAt    time.Time
}

// Alert is one SafeReach emergency event. ShareTokenHash gates the
// unauthenticated public view endpoint for as long as the alert stays
// open; it stops granting access once the alert is cancelled or
// resolved (see AlertRepository.Update).
//
// Only the hash is ever persisted — same posture as OTP codes and
// refresh tokens (see internal/pkg/hash) — because the raw token is a
// bearer credential for live location access; a database leak must not
// hand that out. The raw value is generated once at alert-creation time,
// embedded in the SMS/email/push message, and never stored.
type Alert struct {
	ID             string
	UserID         string
	Level          Level
	Status         Status
	Message        string
	ShareTokenHash string
	TriggeredAt    time.Time
	ResolvedAt     *time.Time
}

func (a Alert) IsOpen() bool { return a.Status.IsOpen() }

// LocationUpdate is one GPS fix recorded during an open alert.
type LocationUpdate struct {
	ID         string
	AlertID    string
	Lat        float64
	Lng        float64
	RecordedAt time.Time
}

// ListFilter paginates the alert event log.
type ListFilter struct {
	Limit  int
	Offset int
}

// TrustedContactRepository persists a user's trusted contacts.
type TrustedContactRepository interface {
	Create(ctx context.Context, c *TrustedContact) error
	Delete(ctx context.Context, id, userID string) error
	FindByID(ctx context.Context, id, userID string) (*TrustedContact, error)
	ListByUser(ctx context.Context, userID string) ([]*TrustedContact, error)
	CountByUser(ctx context.Context, userID string) (int, error)
}

// AlertRepository persists alerts and their location-update history.
type AlertRepository interface {
	Create(ctx context.Context, a *Alert) error
	Update(ctx context.Context, a *Alert) error
	// FindByID is unscoped — used by the public, token-gated view
	// endpoint, which has no authenticated user to scope by and instead
	// verifies the caller-supplied token against Alert.ShareTokenHash.
	FindByID(ctx context.Context, id string) (*Alert, error)
	// FindOwnedByID scopes the lookup to userID, for endpoints only the
	// alert's owner may use (cancel, resolve, location updates).
	FindOwnedByID(ctx context.Context, id, userID string) (*Alert, error)
	List(ctx context.Context, userID string, filter ListFilter) ([]*Alert, int, error)
	AddLocationUpdate(ctx context.Context, u *LocationUpdate) error
	LatestLocation(ctx context.Context, alertID string) (*LocationUpdate, error)
}

var (
	ErrAlertNotFound   = errAlertNotFound{}
	ErrContactNotFound = errContactNotFound{}
)

type errAlertNotFound struct{}

func (errAlertNotFound) Error() string { return "safereach: alert not found" }

type errContactNotFound struct{}

func (errContactNotFound) Error() string { return "safereach: trusted contact not found" }
