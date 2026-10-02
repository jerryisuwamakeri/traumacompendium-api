// Package apperror defines the single error type used across service,
// repository, and HTTP layers so handlers can translate any error into a
// consistent API response without layer-specific error sniffing.
package apperror

import (
	"errors"
	"fmt"
	"net/http"
)

// Code is a short, machine-readable identifier for an error condition.
// Clients (web, mobile) should branch on Code, never on Message — Message
// is for humans and may change wording over time.
type Code string

const (
	CodeValidation      Code = "validation_error"
	CodeBadRequest      Code = "bad_request"
	CodeUnauthorized    Code = "unauthorized"
	CodeForbidden       Code = "forbidden"
	CodeNotFound        Code = "not_found"
	CodeConflict        Code = "conflict"
	CodeTooManyRequests Code = "too_many_requests"
	CodeInternal        Code = "internal_error"
	CodeUnprocessable   Code = "unprocessable_entity"
	CodePaymentRequired Code = "payment_required"
)

// Error is an application-level error carrying everything the HTTP layer
// needs to render a response: a stable code, a user-facing message, the
// HTTP status to use, and optionally the underlying cause (logged, never
// exposed to the client).
type Error struct {
	Code    Code
	Message string
	Status  int
	Fields  map[string]string // optional per-field validation messages
	cause   error
}

func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error { return e.cause }

// WithCause attaches an underlying error for logging purposes. The cause is
// never serialized into the HTTP response.
func (e *Error) WithCause(err error) *Error {
	clone := *e
	clone.cause = err
	return &clone
}

// WithFields attaches field-level validation messages.
func (e *Error) WithFields(fields map[string]string) *Error {
	clone := *e
	clone.Fields = fields
	return &clone
}

func new(status int, code Code, message string) *Error {
	return &Error{Status: status, Code: code, Message: message}
}

func NewValidation(message string) *Error {
	return new(http.StatusBadRequest, CodeValidation, message)
}

func NewBadRequest(message string) *Error {
	return new(http.StatusBadRequest, CodeBadRequest, message)
}

func NewUnauthorized(message string) *Error {
	return new(http.StatusUnauthorized, CodeUnauthorized, message)
}

func NewForbidden(message string) *Error {
	return new(http.StatusForbidden, CodeForbidden, message)
}

func NewNotFound(message string) *Error {
	return new(http.StatusNotFound, CodeNotFound, message)
}

func NewConflict(message string) *Error {
	return new(http.StatusConflict, CodeConflict, message)
}

func NewUnprocessable(message string) *Error {
	return new(http.StatusUnprocessableEntity, CodeUnprocessable, message)
}

func NewTooManyRequests(message string) *Error {
	return new(http.StatusTooManyRequests, CodeTooManyRequests, message)
}

// NewPaymentRequired signals a premium-gated feature the caller's
// current plan doesn't cover (see billingservice.IsPremium).
func NewPaymentRequired(message string) *Error {
	return new(http.StatusPaymentRequired, CodePaymentRequired, message)
}

func NewInternal(message string) *Error {
	return new(http.StatusInternalServerError, CodeInternal, message)
}

// As extracts an *Error from any error chain, so the HTTP layer can map
// known application errors to the right status code while falling back to
// a generic 500 for unexpected errors.
func As(err error) (*Error, bool) {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}
