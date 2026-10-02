package http

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// ResolveLocale reads the caller's requested language: the explicit
// "?locale=" query param takes priority, then the primary tag of the
// Accept-Language header (e.g. "yo-NG,yo;q=0.9,en;q=0.5" -> "yo"), else
// "" — an empty result means "no preference stated," which every
// learning service method treats as the default locale (see
// learningservice.resolveLocale). An unsupported or malformed value is
// never an error here; validation/fallback is the service layer's job,
// so a locale typo degrades to English instead of failing the request.
func ResolveLocale(c *gin.Context) string {
	if v := c.Query("locale"); v != "" {
		return v
	}

	header := c.GetHeader("Accept-Language")
	if header == "" {
		return ""
	}

	// Take the first comma-separated entry, then strip any ";q=" weight
	// and region subtag, e.g. "yo-NG;q=0.9" -> "yo".
	first := strings.TrimSpace(strings.Split(header, ",")[0])
	first = strings.Split(first, ";")[0]
	first = strings.Split(first, "-")[0]
	return strings.ToLower(first)
}
