package handler

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"github.com/traumacompendium/backend/internal/domain/safereach"
	apphttp "github.com/traumacompendium/backend/internal/http"
	"github.com/traumacompendium/backend/internal/http/dto"
	"github.com/traumacompendium/backend/internal/http/middleware"
	"github.com/traumacompendium/backend/internal/pkg/apperror"
	"github.com/traumacompendium/backend/internal/service/safereachservice"
)

type SafeReachHandler struct {
	svc            *safereachservice.Service
	redis          *redis.Client
	allowedOrigins []string
	log            zerolog.Logger
}

func NewSafeReachHandler(svc *safereachservice.Service, redisClient *redis.Client, allowedOrigins []string, log zerolog.Logger) *SafeReachHandler {
	return &SafeReachHandler{svc: svc, redis: redisClient, allowedOrigins: allowedOrigins, log: log}
}

// POST /api/v1/safereach/contacts
func (h *SafeReachHandler) RegisterContact(c *gin.Context) {
	var req dto.RegisterTrustedContactRequest
	if !apphttp.BindJSON(c, &req) {
		return
	}

	contact, err := h.svc.RegisterContact(c.Request.Context(), middleware.UserID(c), safereachservice.RegisterContactInput{
		Name:         req.Name,
		Relationship: req.Relationship,
		Phone:        req.Phone,
		Email:        req.Email,
	})
	if err != nil {
		apphttp.Fail(c, err)
		return
	}
	apphttp.Created(c, dto.NewTrustedContactResponse(contact))
}

// GET /api/v1/safereach/contacts
func (h *SafeReachHandler) ListContacts(c *gin.Context) {
	contacts, err := h.svc.ListContacts(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		apphttp.Fail(c, err)
		return
	}
	apphttp.OK(c, dto.NewTrustedContactResponses(contacts))
}

// DELETE /api/v1/safereach/contacts/:id
func (h *SafeReachHandler) DeleteContact(c *gin.Context) {
	if err := h.svc.DeleteContact(c.Request.Context(), middleware.UserID(c), c.Param("id")); err != nil {
		apphttp.Fail(c, err)
		return
	}
	apphttp.NoContent(c)
}

// POST /api/v1/safereach/alerts
func (h *SafeReachHandler) TriggerAlert(c *gin.Context) {
	var req dto.TriggerAlertRequest
	if !apphttp.BindJSON(c, &req) {
		return
	}

	a, err := h.svc.TriggerAlert(c.Request.Context(), middleware.UserID(c), safereachservice.TriggerAlertInput{
		Level:   safereach.Level(req.Level),
		Message: req.Message,
		Lat:     req.Lat,
		Lng:     req.Lng,
	})
	if err != nil {
		apphttp.Fail(c, err)
		return
	}
	apphttp.Created(c, dto.NewAlertResponse(a))
}

// GET /api/v1/safereach/alerts?limit=&offset=
func (h *SafeReachHandler) ListAlerts(c *gin.Context) {
	limit, err := apphttp.ParseLimit(c, 20, 100)
	if err != nil {
		apphttp.Fail(c, err)
		return
	}
	offset, err := apphttp.ParseOffset(c)
	if err != nil {
		apphttp.Fail(c, err)
		return
	}

	result, err := h.svc.ListAlerts(c.Request.Context(), middleware.UserID(c), safereach.ListFilter{Limit: limit, Offset: offset})
	if err != nil {
		apphttp.Fail(c, err)
		return
	}

	apphttp.OKWithMeta(c, dto.NewAlertResponses(result.Alerts), gin.H{
		"total": result.Total, "limit": limit, "offset": offset,
	})
}

// GET /api/v1/safereach/alerts/:id
func (h *SafeReachHandler) GetAlert(c *gin.Context) {
	a, err := h.svc.GetAlert(c.Request.Context(), middleware.UserID(c), c.Param("id"))
	if err != nil {
		apphttp.Fail(c, err)
		return
	}
	apphttp.OK(c, dto.NewAlertResponse(a))
}

// POST /api/v1/safereach/alerts/:id/cancel
func (h *SafeReachHandler) CancelAlert(c *gin.Context) {
	a, err := h.svc.CancelAlert(c.Request.Context(), middleware.UserID(c), c.Param("id"))
	if err != nil {
		apphttp.Fail(c, err)
		return
	}
	apphttp.OK(c, dto.NewAlertResponse(a))
}

// POST /api/v1/safereach/alerts/:id/resolve
func (h *SafeReachHandler) ResolveAlert(c *gin.Context) {
	a, err := h.svc.ResolveAlert(c.Request.Context(), middleware.UserID(c), c.Param("id"))
	if err != nil {
		apphttp.Fail(c, err)
		return
	}
	apphttp.OK(c, dto.NewAlertResponse(a))
}

// POST /api/v1/safereach/alerts/:id/location
func (h *SafeReachHandler) RecordLocation(c *gin.Context) {
	var req dto.RecordLocationRequest
	if !apphttp.BindJSON(c, &req) {
		return
	}

	if err := h.svc.RecordLocation(c.Request.Context(), middleware.UserID(c), c.Param("id"), req.Lat, req.Lng); err != nil {
		apphttp.Fail(c, err)
		return
	}
	apphttp.NoContent(c)
}

// POST /api/v1/safereach/device-tokens
func (h *SafeReachHandler) RegisterDeviceToken(c *gin.Context) {
	var req dto.RegisterDeviceTokenRequest
	if !apphttp.BindJSON(c, &req) {
		return
	}

	if err := h.svc.RegisterDeviceToken(c.Request.Context(), middleware.UserID(c), req.Token, req.Platform); err != nil {
		apphttp.Fail(c, err)
		return
	}
	apphttp.NoContent(c)
}

// GET /api/v1/safereach/public/alerts/:id?token=... — unauthenticated,
// token-gated. See safereachservice.GetPublicView for why a wrong or
// missing token is indistinguishable from a nonexistent alert.
func (h *SafeReachHandler) PublicView(c *gin.Context) {
	view, err := h.svc.GetPublicView(c.Request.Context(), c.Param("id"), c.Query("token"))
	if err != nil {
		apphttp.Fail(c, err)
		return
	}
	apphttp.OK(c, dto.NewPublicAlertViewResponse(view))
}

// GET /api/v1/safereach/alerts/:id/stream — authenticated WebSocket,
// owner-only: forwards live location updates (published by
// RecordLocation to safereachservice.LocationChannel) for as long as the
// connection and the alert both stay open.
func (h *SafeReachHandler) Stream(c *gin.Context) {
	userID := middleware.UserID(c)
	alertID := c.Param("id")

	// Ownership must be checked before the WebSocket upgrade — once
	// upgraded, there's no way to send back a normal JSON error response.
	if _, err := h.svc.GetAlert(c.Request.Context(), userID, alertID); err != nil {
		apphttp.Fail(c, err)
		return
	}

	ctx := c.Request.Context()
	pubsub := h.redis.Subscribe(ctx, safereachservice.LocationChannel(alertID))

	// Subscribe before upgrading, and wait for Redis to confirm it, so a
	// location update published the instant the client connects is
	// never lost to a subscribe/publish race — go-redis's Subscribe
	// sends the command but doesn't itself guarantee the server has
	// processed it; Receive blocks for that confirmation.
	if _, err := pubsub.Receive(ctx); err != nil {
		_ = pubsub.Close()
		h.log.Warn().Err(err).Str("alert_id", alertID).Msg("safereach: failed to confirm redis subscription")
		apphttp.Fail(c, apperror.NewInternal("could not open live stream"))
		return
	}
	defer pubsub.Close() //nolint:errcheck // best-effort on stream end

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			if origin == "" {
				return true // non-browser clients (mobile apps) send no Origin header
			}
			return slices.Contains(h.allowedOrigins, "*") || slices.Contains(h.allowedOrigins, origin)
		},
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.log.Warn().Err(err).Str("alert_id", alertID).Msg("safereach: websocket upgrade failed")
		return
	}
	defer conn.Close() //nolint:errcheck // best-effort on stream end

	// This is a one-way (server -> client) stream, but we still need to
	// notice the client closing the connection — otherwise this
	// goroutine only exits on the next failed write, which may be a
	// long time after an idle client has disconnected.
	closed := make(chan struct{})
	go func() {
		defer close(closed)
		for {
			if _, _, err := conn.NextReader(); err != nil {
				return
			}
		}
	}()

	messages := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case <-closed:
			return
		case msg, ok := <-messages:
			if !ok {
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload)); err != nil {
				return
			}
		}
	}
}
