package dto

import (
	"time"

	"github.com/traumacompendium/backend/internal/domain/safereach"
	"github.com/traumacompendium/backend/internal/service/safereachservice"
)

type RegisterTrustedContactRequest struct {
	Name         string `json:"name" binding:"required,max=200"`
	Relationship string `json:"relationship" binding:"max=100"`
	Phone        string `json:"phone" binding:"omitempty,e164"`
	Email        string `json:"email" binding:"omitempty,email"`
}

type TrustedContactResponse struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Relationship string    `json:"relationship"`
	Phone        string    `json:"phone,omitempty"`
	Email        string    `json:"email,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

func NewTrustedContactResponse(c *safereach.TrustedContact) TrustedContactResponse {
	return TrustedContactResponse{
		ID: c.ID, Name: c.Name, Relationship: c.Relationship,
		Phone: c.Phone, Email: c.Email, CreatedAt: c.CreatedAt,
	}
}

func NewTrustedContactResponses(contacts []*safereach.TrustedContact) []TrustedContactResponse {
	out := make([]TrustedContactResponse, len(contacts))
	for i, c := range contacts {
		out[i] = NewTrustedContactResponse(c)
	}
	return out
}

type TriggerAlertRequest struct {
	Level   string   `json:"level" binding:"required,oneof=check_in urgent critical"`
	Message string   `json:"message" binding:"max=500"`
	Lat     *float64 `json:"lat" binding:"omitempty,min=-90,max=90"`
	Lng     *float64 `json:"lng" binding:"omitempty,min=-180,max=180"`
}

type RecordLocationRequest struct {
	Lat float64 `json:"lat" binding:"required,min=-90,max=90"`
	Lng float64 `json:"lng" binding:"required,min=-180,max=180"`
}

type RegisterDeviceTokenRequest struct {
	Token    string `json:"token" binding:"required"`
	Platform string `json:"platform" binding:"omitempty,oneof=ios android web"`
}

type AlertResponse struct {
	ID          string     `json:"id"`
	Level       string     `json:"level"`
	Status      string     `json:"status"`
	Message     string     `json:"message,omitempty"`
	TriggeredAt time.Time  `json:"triggered_at"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
}

func NewAlertResponse(a *safereach.Alert) AlertResponse {
	return AlertResponse{
		ID:          a.ID,
		Level:       string(a.Level),
		Status:      string(a.Status),
		Message:     a.Message,
		TriggeredAt: a.TriggeredAt,
		ResolvedAt:  a.ResolvedAt,
	}
}

func NewAlertResponses(alerts []*safereach.Alert) []AlertResponse {
	out := make([]AlertResponse, len(alerts))
	for i, a := range alerts {
		out[i] = NewAlertResponse(a)
	}
	return out
}

// PublicAlertViewResponse is what the unauthenticated, token-gated share
// link returns — deliberately smaller than AlertResponse (no internal
// alert ID echoed back, for instance): just enough for a trusted contact
// to see what's going on.
type PublicAlertViewResponse struct {
	Level     string     `json:"level"`
	Status    string     `json:"status"`
	Message   string     `json:"message,omitempty"`
	Lat       *float64   `json:"lat,omitempty"`
	Lng       *float64   `json:"lng,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

func NewPublicAlertViewResponse(v safereachservice.PublicAlertView) PublicAlertViewResponse {
	return PublicAlertViewResponse{
		Level:     string(v.Level),
		Status:    string(v.Status),
		Message:   v.Message,
		Lat:       v.Lat,
		Lng:       v.Lng,
		UpdatedAt: v.UpdatedAt,
	}
}
