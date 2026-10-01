package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// Logging emits one structured log line per request, correlated by
// request ID, after the handler completes — so latency and final status
// are always included (unlike logging at request-start).
func Logging(log zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		status := c.Writer.Status()
		event := log.Info()
		if status >= 500 {
			event = log.Error()
		} else if status >= 400 {
			event = log.Warn()
		}

		event.
			Str("request_id", GetRequestID(c)).
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", status).
			Dur("latency", time.Since(start)).
			Str("client_ip", c.ClientIP()).
			Int("response_bytes", c.Writer.Size())

		if len(c.Errors) > 0 {
			event.Str("errors", c.Errors.String())
		}

		event.Msg("http: request handled")
	}
}
