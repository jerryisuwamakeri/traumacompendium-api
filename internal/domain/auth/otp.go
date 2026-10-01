// Package auth defines the entities that support authentication but are
// not the User aggregate itself: one-time verification codes and refresh
// token sessions.
package auth

import (
	"context"
	"time"
)

// OTPPurpose scopes a code to exactly one use so a code issued for email
// verification can never be replayed to reset a password, etc.
type OTPPurpose string

const (
	OTPPurposeEmailVerification OTPPurpose = "email_verification"
	OTPPurposePhoneVerification OTPPurpose = "phone_verification"
	OTPPurposePasswordReset     OTPPurpose = "password_reset"
)

// OTPCode is a hashed, time-limited one-time code. The raw code is never
// persisted — only CodeHash (see internal/pkg/hash.Token) — so a database
// leak does not hand out usable codes.
type OTPCode struct {
	ID           string
	UserID       string
	Purpose      OTPPurpose
	CodeHash     string
	ExpiresAt    time.Time
	ConsumedAt   *time.Time
	AttemptCount int
	CreatedAt    time.Time
}

func (o *OTPCode) IsExpired(now time.Time) bool { return now.After(o.ExpiresAt) }
func (o *OTPCode) IsConsumed() bool             { return o.ConsumedAt != nil }

// OTPRepository persists OTP codes. FindActive returns the most recent,
// not-yet-consumed code for a user+purpose so services can enforce
// resend cooldowns and attempt limits against a single row.
type OTPRepository interface {
	Create(ctx context.Context, code *OTPCode) error
	FindActive(ctx context.Context, userID string, purpose OTPPurpose) (*OTPCode, error)
	IncrementAttempts(ctx context.Context, id string) error
	MarkConsumed(ctx context.Context, id string) error
}

var ErrOTPNotFound = errOTPNotFound{}

type errOTPNotFound struct{}

func (errOTPNotFound) Error() string { return "auth: otp code not found" }
