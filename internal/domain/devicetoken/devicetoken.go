// Package devicetoken registers the push-notification tokens
// (internal/notification/push's FCM device tokens) a user's devices have
// opted in with, so SafeReach (and later, any other push-worthy event)
// knows where to deliver.
package devicetoken

import "context"

// DeviceToken is one device's FCM registration token.
type DeviceToken struct {
	ID       string
	UserID   string
	Token    string
	Platform string // "ios" | "android" | "web", informational only
}

// Repository persists device tokens. Register is an upsert — a device
// token can be re-registered (common on every app launch) or reassigned
// to a different user (someone signs into a different account on the
// same device), so re-registering the same token simply repoints it
// rather than erroring on the unique constraint.
type Repository interface {
	Register(ctx context.Context, d *DeviceToken) error
	ListByUser(ctx context.Context, userID string) ([]*DeviceToken, error)
}
