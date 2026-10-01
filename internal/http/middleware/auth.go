package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	apphttp "github.com/traumacompendium/backend/internal/http"
	"github.com/traumacompendium/backend/internal/pkg/apperror"
	"github.com/traumacompendium/backend/internal/pkg/jwtutil"
)

const (
	contextKeyUserID = "auth_user_id"
	contextKeyRole   = "auth_role"
)

// RequireAuth validates the Bearer access token on every request it
// guards and, on success, makes the authenticated user's ID and role
// available to handlers via UserID/Role. It must run before any
// RequireRole in the middleware chain.
func RequireAuth(jwt *jwtutil.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := bearerToken(c)
		if !ok {
			apphttp.Fail(c, apperror.NewUnauthorized("missing or malformed Authorization header"))
			c.Abort()
			return
		}

		claims, err := jwt.ParseAccessToken(token)
		if err != nil {
			apphttp.Fail(c, apperror.NewUnauthorized("invalid or expired access token"))
			c.Abort()
			return
		}

		c.Set(contextKeyUserID, claims.UserID)
		c.Set(contextKeyRole, claims.Role)
		c.Next()
	}
}

// RequireRole restricts a route to users whose role (set by RequireAuth)
// is one of the given roles. Used for the admin-only surface added
// alongside the learning content model.
func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}

	return func(c *gin.Context) {
		if !allowed[Role(c)] {
			apphttp.Fail(c, apperror.NewForbidden("you do not have permission to perform this action"))
			c.Abort()
			return
		}
		c.Next()
	}
}

func bearerToken(c *gin.Context) (string, bool) {
	header := c.GetHeader("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}
	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	return token, token != ""
}

// UserID returns the authenticated user's ID, set by RequireAuth. Empty
// if called on a route not guarded by RequireAuth.
func UserID(c *gin.Context) string {
	v, _ := c.Get(contextKeyUserID)
	s, _ := v.(string)
	return s
}

// Role returns the authenticated user's role, set by RequireAuth.
func Role(c *gin.Context) string {
	v, _ := c.Get(contextKeyRole)
	s, _ := v.(string)
	return s
}
