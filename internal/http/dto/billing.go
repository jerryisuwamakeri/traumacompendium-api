package dto

import (
	"time"

	"github.com/traumacompendium/backend/internal/domain/billing"
	"github.com/traumacompendium/backend/internal/service/billingservice"
)

type SubscribeResponse struct {
	AuthorizationURL string `json:"authorization_url"`
	Reference        string `json:"reference"`
}

func NewSubscribeResponse(r billingservice.SubscribeResult) SubscribeResponse {
	return SubscribeResponse{AuthorizationURL: r.AuthorizationURL, Reference: r.Reference}
}

type SubscriptionResponse struct {
	Plan             string     `json:"plan"`
	Status           string     `json:"status"`
	CurrentPeriodEnd *time.Time `json:"current_period_end,omitempty"`
}

// NewSubscriptionResponse handles the "never subscribed" case (sub ==
// nil) by reporting the implicit Free plan — the client never needs to
// special-case a missing subscription.
func NewSubscriptionResponse(sub *billing.Subscription) SubscriptionResponse {
	if sub == nil {
		return SubscriptionResponse{Plan: string(billing.PlanFree), Status: string(billing.StatusActive)}
	}
	return SubscriptionResponse{
		Plan:             string(sub.Plan),
		Status:           string(sub.Status),
		CurrentPeriodEnd: &sub.CurrentPeriodEnd,
	}
}
