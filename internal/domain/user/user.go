// Package user defines the User aggregate and the repository contract the
// rest of the application programs against — no package outside
// internal/repository/mysql knows how a user is actually persisted.
package user

import (
	"context"
	"time"
)

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

type Status string

const (
	StatusActive    Status = "active"
	StatusSuspended Status = "suspended"
)

// User is a registered account. Email and Phone are both required and
// unique; an account only becomes fully usable once both are verified
// (see IsFullyVerified), matching the brief's requirement to reach users
// reliably for both learning-progress notifications and (in a later
// phase) SafeReach alerts.
type User struct {
	ID               string
	FullName         string
	Email            string
	EmailVerifiedAt  *time.Time
	Phone            string
	PhoneVerifiedAt  *time.Time
	PasswordHash     string
	Role             Role
	Status           Status
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
}

func (u *User) IsEmailVerified() bool { return u.EmailVerifiedAt != nil }
func (u *User) IsPhoneVerified() bool { return u.PhoneVerifiedAt != nil }
func (u *User) IsFullyVerified() bool { return u.IsEmailVerified() && u.IsPhoneVerified() }
func (u *User) IsActive() bool        { return u.Status == StatusActive && u.DeletedAt == nil }

// Repository persists and retrieves User aggregates. Implementations must
// treat Email and Phone as case-insensitively unique and return
// ErrNotFound (via the caller mapping to apperror) when no row matches.
type Repository interface {
	Create(ctx context.Context, u *User) error
	Update(ctx context.Context, u *User) error
	FindByID(ctx context.Context, id string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByPhone(ctx context.Context, phone string) (*User, error)
}

// ErrNotFound is returned by Repository implementations when no matching
// row exists. Services translate it into an apperror.NewNotFound (or, for
// auth flows where existence must not be leaked, a generic unauthorized).
var ErrNotFound = errNotFound{}

type errNotFound struct{}

func (errNotFound) Error() string { return "user: not found" }

// ErrEmailTaken and ErrPhoneTaken are returned by Repository.Create when
// the unique constraint on that column is violated. Keeping these in the
// domain package (rather than the mysql repository package) lets the
// service layer react to them without importing an infrastructure
// package.
var (
	ErrEmailTaken = errEmailTaken{}
	ErrPhoneTaken = errPhoneTaken{}
)

type errEmailTaken struct{}

func (errEmailTaken) Error() string { return "user: email already registered" }

type errPhoneTaken struct{}

func (errPhoneTaken) Error() string { return "user: phone already registered" }
