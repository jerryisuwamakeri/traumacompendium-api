// Package http wires the HTTP transport: the Gin router, middleware,
// request/response DTOs, and handlers. It is the only layer allowed to
// know about Gin, HTTP status codes, or JSON wire shapes — services speak
// only Go types and apperror.Error.
package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/traumacompendium/backend/internal/pkg/apperror"
	"github.com/traumacompendium/backend/internal/pkg/i18n"
)

// envelope is the single JSON shape every response on this API uses,
// success or failure, so clients can write one response-parsing code
// path instead of branching per endpoint.
type envelope struct {
	Data  any            `json:"data,omitempty"`
	Error *errorBody     `json:"error,omitempty"`
	Meta  map[string]any `json:"meta,omitempty"`
}

type errorBody struct {
	Code    apperror.Code     `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// OK writes a 200 with the given payload as data.
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, envelope{Data: data})
}

// Created writes a 201 with the given payload as data.
func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, envelope{Data: data})
}

// NoContent writes a 204 with no body.
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// OKWithMeta writes a 200 with both data and pagination/aggregate meta.
func OKWithMeta(c *gin.Context, data any, meta map[string]any) {
	c.JSON(http.StatusOK, envelope{Data: data, Meta: meta})
}

// Fail translates any error into the right HTTP status and JSON error
// body. Known *apperror.Error values map directly; everything else is
// logged as an unexpected error and reduced to an opaque 500 so internal
// details never reach the client.
func Fail(c *gin.Context, err error) {
	if appErr, ok := apperror.As(err); ok {
		c.JSON(appErr.Status, envelope{Error: &errorBody{
			Code:    appErr.Code,
			Message: appErr.Message,
			Fields:  appErr.Fields,
		}})
		return
	}

	_ = c.Error(err) // recorded for the logging middleware; never serialized
	c.JSON(http.StatusInternalServerError, envelope{Error: &errorBody{
		Code:    apperror.CodeInternal,
		Message: "an unexpected error occurred",
	}})
}

// BindJSON decodes and validates a request body into dst (via Gin's
// binding, which runs go-playground/validator against the struct's
// `binding` tags). On failure it writes a 400 with per-field messages —
// localized per ResolveLocale(c), the same locale resolution the
// learning endpoints use — and returns false so the caller can simply
// `if !BindJSON(...) { return }`.
func BindJSON(c *gin.Context, dst any) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		Fail(c, validationError(err, ResolveLocale(c)))
		return false
	}
	return true
}

func validationError(err error, locale string) *apperror.Error {
	var verrs validator.ValidationErrors
	if errors.As(err, &verrs) {
		fields := make(map[string]string, len(verrs))
		for _, fe := range verrs {
			fields[fe.Field()] = fieldErrorMessage(fe, locale)
		}
		return apperror.NewValidation("validation failed").WithFields(fields)
	}
	return apperror.NewBadRequest("could not parse request body")
}

func fieldErrorMessage(fe validator.FieldError, locale string) string {
	switch fe.Tag() {
	case "required":
		return i18n.T(locale, i18n.KeyRequired)
	case "email":
		return i18n.T(locale, i18n.KeyEmail)
	case "min":
		return i18n.T(locale, i18n.KeyTooShort)
	case "max":
		return i18n.T(locale, i18n.KeyTooLong)
	case "e164":
		return i18n.T(locale, i18n.KeyInvalidPhone)
	default:
		return i18n.T(locale, i18n.KeyInvalid)
	}
}
