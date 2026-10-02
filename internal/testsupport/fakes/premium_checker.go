package fakes

import "context"

// PremiumChecker is a controllable fake implementing the narrow
// PremiumChecker interfaces that assistantservice (and anything else
// gating on subscription status) depend on.
type PremiumChecker struct {
	Premium bool
	Err     error
}

func (f *PremiumChecker) IsPremium(_ context.Context, _ string) (bool, error) {
	return f.Premium, f.Err
}
