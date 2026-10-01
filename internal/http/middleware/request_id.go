// Package middleware holds cross-cutting Gin middleware: request ID
// propagation, structured access logging, panic recovery, CORS, JWT
// authentication/RBAC, and Redis-backed rate limiting.
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// HeaderRequestID is both the inbound header this middleware honors
	// (so a client or upstream proxy can supply its own trace ID) and the
	// outbound header it always sets.
	HeaderRequestID = "X-Request-ID"
	contextKeyRequestID = "request_id"
)

// RequestID assigns every request a unique ID — reusing one supplied by
// the caller/proxy if present — so a single request can be traced across
// logs, error responses, and (in a later phase) downstream service calls.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(HeaderRequestID)
		if id == "" {
			id = uuid.NewString()
		}
		c.Set(contextKeyRequestID, id)
		c.Writer.Header().Set(HeaderRequestID, id)
		c.Next()
	}
}

// GetRequestID retrieves the ID set by RequestID, or "" if the middleware
// was not installed (should not happen outside of tests).
func GetRequestID(c *gin.Context) string {
	id, _ := c.Get(contextKeyRequestID)
	s, _ := id.(string)
	return s
}
