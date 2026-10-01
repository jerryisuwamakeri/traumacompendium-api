// Package handler adapts HTTP requests to service calls: decode + validate
// the request DTO, call exactly one service method, map the result (or
// error) to a response. Handlers deliberately contain no business logic.
package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/traumacompendium/backend/internal/domain/auth"
	apphttp "github.com/traumacompendium/backend/internal/http"
	"github.com/traumacompendium/backend/internal/http/dto"
	"github.com/traumacompendium/backend/internal/service/authservice"
)

type AuthHandler struct {
	svc *authservice.Service
}

func NewAuthHandler(svc *authservice.Service) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func sessionMetadata(c *gin.Context) authservice.SessionMetadata {
	return authservice.SessionMetadata{
		UserAgent: c.Request.UserAgent(),
		IPAddress: c.ClientIP(),
	}
}

// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if !apphttp.BindJSON(c, &req) {
		return
	}

	u, err := h.svc.Register(c.Request.Context(), authservice.RegisterInput{
		FullName: req.FullName,
		Email:    req.Email,
		Phone:    req.Phone,
		Password: req.Password,
	})
	if err != nil {
		apphttp.Fail(c, err)
		return
	}

	apphttp.Created(c, dto.NewUserResponse(u))
}

// POST /api/v1/auth/verify-email
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	h.verify(c, auth.OTPPurposeEmailVerification)
}

// POST /api/v1/auth/verify-phone
func (h *AuthHandler) VerifyPhone(c *gin.Context) {
	h.verify(c, auth.OTPPurposePhoneVerification)
}

func (h *AuthHandler) verify(c *gin.Context, purpose auth.OTPPurpose) {
	var req dto.VerifyOTPRequest
	if !apphttp.BindJSON(c, &req) {
		return
	}

	if err := h.svc.VerifyOTP(c.Request.Context(), req.UserID, purpose, req.Code); err != nil {
		apphttp.Fail(c, err)
		return
	}

	apphttp.NoContent(c)
}

// POST /api/v1/auth/resend-otp
func (h *AuthHandler) ResendOTP(c *gin.Context) {
	var req dto.ResendOTPRequest
	if !apphttp.BindJSON(c, &req) {
		return
	}

	if err := h.svc.ResendOTP(c.Request.Context(), req.UserID, auth.OTPPurpose(req.Purpose)); err != nil {
		apphttp.Fail(c, err)
		return
	}

	apphttp.NoContent(c)
}

// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if !apphttp.BindJSON(c, &req) {
		return
	}

	u, pair, err := h.svc.Login(c.Request.Context(), authservice.LoginInput{
		Identifier: req.Identifier,
		Password:   req.Password,
		Meta:       sessionMetadata(c),
	})
	if err != nil {
		apphttp.Fail(c, err)
		return
	}

	apphttp.OK(c, dto.LoginResponse{
		User:   dto.NewUserResponse(u),
		Tokens: tokenResponse(pair),
	})
}

// POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if !apphttp.BindJSON(c, &req) {
		return
	}

	pair, err := h.svc.RefreshSession(c.Request.Context(), req.RefreshToken, sessionMetadata(c))
	if err != nil {
		apphttp.Fail(c, err)
		return
	}

	apphttp.OK(c, tokenResponse(pair))
}

// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	var req dto.LogoutRequest
	if !apphttp.BindJSON(c, &req) {
		return
	}

	if err := h.svc.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		apphttp.Fail(c, err)
		return
	}

	apphttp.NoContent(c)
}

// POST /api/v1/auth/forgot-password
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req dto.ForgotPasswordRequest
	if !apphttp.BindJSON(c, &req) {
		return
	}

	if err := h.svc.ForgotPassword(c.Request.Context(), req.Email); err != nil {
		apphttp.Fail(c, err)
		return
	}

	apphttp.NoContent(c)
}

// POST /api/v1/auth/reset-password
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if !apphttp.BindJSON(c, &req) {
		return
	}

	err := h.svc.ResetPassword(c.Request.Context(), authservice.ResetPasswordInput{
		Email:       req.Email,
		Code:        req.Code,
		NewPassword: req.NewPassword,
	})
	if err != nil {
		apphttp.Fail(c, err)
		return
	}

	apphttp.NoContent(c)
}

func tokenResponse(pair authservice.TokenPair) dto.TokenResponse {
	return dto.TokenResponse{
		AccessToken:      pair.AccessToken,
		AccessExpiresAt:  pair.AccessExpiresAt,
		RefreshToken:     pair.RefreshToken,
		RefreshExpiresAt: pair.RefreshExpiresAt,
	}
}
