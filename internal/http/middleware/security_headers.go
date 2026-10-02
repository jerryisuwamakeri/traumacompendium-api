package middleware

import "github.com/gin-gonic/gin"

// SecurityHeaders sets a conservative, fixed set of response headers
// appropriate for a pure JSON/audio API: it never serves HTML, JS, or
// CSS of its own, so the policy can be maximally restrictive rather than
// tuned per page.
//
// Strict-Transport-Security is deliberately NOT set here — that belongs
// on the TLS-terminating reverse proxy/load balancer in front of this
// process in production (this process itself typically only ever speaks
// plain HTTP inside the deployment network), not something to fake at
// the app layer.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Content-Security-Policy", "default-src 'none'")
		h.Set("Permissions-Policy", "geolocation=(), camera=(), microphone=()")
		c.Next()
	}
}
