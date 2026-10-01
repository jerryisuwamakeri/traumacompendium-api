package middleware

import (
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	apphttp "github.com/traumacompendium/backend/internal/http"
	"github.com/traumacompendium/backend/internal/pkg/apperror"
)

// Recovery converts a panic in any downstream handler into a logged error
// (with stack trace) and a standard 500 envelope, instead of crashing the
// process or leaking Gin's default plain-text panic output.
func Recovery(log zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Error().
					Str("request_id", GetRequestID(c)).
					Interface("panic", r).
					Str("stack", string(debug.Stack())).
					Msg("http: recovered from panic")

				if !c.Writer.Written() {
					apphttp.Fail(c, apperror.NewInternal("an unexpected error occurred"))
				}
				c.Abort()
			}
		}()
		c.Next()
	}
}
