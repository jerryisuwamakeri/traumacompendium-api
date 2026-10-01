package handler

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	apphttp "github.com/traumacompendium/backend/internal/http"
)

type HealthHandler struct {
	db    *sql.DB
	redis *redis.Client
}

func NewHealthHandler(db *sql.DB, redis *redis.Client) *HealthHandler {
	return &HealthHandler{db: db, redis: redis}
}

// GET /healthz — liveness: the process is up and able to handle requests
// at all. Used by orchestrators to decide whether to restart the
// container; it deliberately does not check dependencies, so a database
// blip doesn't trigger unnecessary restarts.
func (h *HealthHandler) Liveness(c *gin.Context) {
	apphttp.OK(c, gin.H{"status": "ok"})
}

// GET /readyz — readiness: the process is additionally able to serve real
// traffic, because its dependencies (database, cache) are reachable. Used
// by load balancers/orchestrators to decide whether to route traffic to
// this instance.
func (h *HealthHandler) Readiness(c *gin.Context) {
	checks := gin.H{}
	ready := true

	if err := h.db.PingContext(c.Request.Context()); err != nil {
		checks["database"] = "unreachable"
		ready = false
	} else {
		checks["database"] = "ok"
	}

	if err := h.redis.Ping(c.Request.Context()).Err(); err != nil {
		checks["redis"] = "unreachable"
		ready = false
	} else {
		checks["redis"] = "ok"
	}

	if !ready {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready", "checks": checks})
		return
	}

	apphttp.OK(c, gin.H{"status": "ready", "checks": checks})
}
