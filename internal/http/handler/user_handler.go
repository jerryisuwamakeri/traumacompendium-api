package handler

import (
	"github.com/gin-gonic/gin"

	apphttp "github.com/traumacompendium/backend/internal/http"
	"github.com/traumacompendium/backend/internal/http/dto"
	"github.com/traumacompendium/backend/internal/http/middleware"
	"github.com/traumacompendium/backend/internal/service/userservice"
)

type UserHandler struct {
	svc *userservice.Service
}

func NewUserHandler(svc *userservice.Service) *UserHandler {
	return &UserHandler{svc: svc}
}

// GET /api/v1/users/me — requires middleware.RequireAuth.
func (h *UserHandler) Me(c *gin.Context) {
	u, err := h.svc.GetProfile(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		apphttp.Fail(c, err)
		return
	}
	apphttp.OK(c, dto.NewUserResponse(u))
}

// PATCH /api/v1/users/me — requires middleware.RequireAuth.
func (h *UserHandler) UpdateMe(c *gin.Context) {
	var req dto.UpdateProfileRequest
	if !apphttp.BindJSON(c, &req) {
		return
	}

	u, err := h.svc.UpdateProfile(c.Request.Context(), middleware.UserID(c), userservice.UpdateProfileInput{
		FullName: req.FullName,
	})
	if err != nil {
		apphttp.Fail(c, err)
		return
	}

	apphttp.OK(c, dto.NewUserResponse(u))
}
