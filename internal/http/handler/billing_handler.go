package handler

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	apphttp "github.com/traumacompendium/backend/internal/http"
	"github.com/traumacompendium/backend/internal/http/dto"
	"github.com/traumacompendium/backend/internal/http/middleware"
	"github.com/traumacompendium/backend/internal/pkg/apperror"
	"github.com/traumacompendium/backend/internal/service/billingservice"
)

type BillingHandler struct {
	svc *billingservice.Service
}

func NewBillingHandler(svc *billingservice.Service) *BillingHandler {
	return &BillingHandler{svc: svc}
}

// POST /api/v1/billing/subscribe
func (h *BillingHandler) Subscribe(c *gin.Context) {
	result, err := h.svc.Subscribe(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		apphttp.Fail(c, err)
		return
	}
	apphttp.OK(c, dto.NewSubscribeResponse(result))
}

// GET /api/v1/billing/subscription
func (h *BillingHandler) GetSubscription(c *gin.Context) {
	sub, err := h.svc.GetSubscription(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		apphttp.Fail(c, err)
		return
	}
	apphttp.OK(c, dto.NewSubscriptionResponse(sub))
}

// POST /api/v1/billing/webhook — unauthenticated (Paystack calls this
// directly); authenticity comes entirely from the signature HandleWebhook
// verifies, not from any session. The raw body must be read here, before
// anything else touches the request, since the signature is computed
// over the exact bytes Paystack sent — any JSON-decode-and-re-encode
// round trip would invalidate it.
func (h *BillingHandler) Webhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		apphttp.Fail(c, apperror.NewBadRequest("could not read webhook body"))
		return
	}

	signature := c.GetHeader("X-Paystack-Signature")
	if err := h.svc.HandleWebhook(c.Request.Context(), body, signature); err != nil {
		apphttp.Fail(c, err)
		return
	}

	// Paystack expects a plain 200 acknowledging receipt — not the
	// standard envelope every other endpoint uses, since this response
	// is read by Paystack's webhook delivery system, not this API's own
	// clients.
	c.Status(http.StatusOK)
}
