package http

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/traumacompendium/backend/internal/pkg/apperror"
)

// ParseLimit reads the "limit" query parameter, defaulting to def and
// rejecting anything outside [1, max]. Shared by every paginated list
// endpoint so the bounds-checking behavior (and error message) stays
// identical across modules.
func ParseLimit(c *gin.Context, def, max int) (int, error) {
	v := c.Query("limit")
	if v == "" {
		return def, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 || n > max {
		return 0, apperror.NewBadRequest("limit must be an integer between 1 and " + strconv.Itoa(max))
	}
	return n, nil
}

// ParseOffset reads the "offset" query parameter, defaulting to 0.
func ParseOffset(c *gin.Context) (int, error) {
	v := c.Query("offset")
	if v == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		return 0, apperror.NewBadRequest("offset must be a non-negative integer")
	}
	return n, nil
}

// ParseOptionalTime reads an RFC3339 timestamp from the named query
// parameter, returning nil if it's absent.
func ParseOptionalTime(c *gin.Context, param string) (*time.Time, error) {
	v := c.Query(param)
	if v == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return nil, apperror.NewBadRequest(param + " must be an RFC3339 timestamp")
	}
	return &t, nil
}
