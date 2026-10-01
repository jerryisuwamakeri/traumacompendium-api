// Package dto defines the HTTP request/response shapes, decoupled from
// both the wire-level handler code and the domain/service types — a
// domain field can be renamed or restructured without silently changing
// the public API contract, and vice versa.
package dto

import (
	"time"

	"github.com/traumacompendium/backend/internal/domain/user"
)

// Password length is bounded 8-72 on every request that sets one: 8 is a
// reasonable minimum for a consumer app, and 72 is bcrypt's hard input
// limit — anything longer is silently truncated by the hashing library,
// which would be a subtle correctness bug, so it's rejected explicitly
// here instead.
type RegisterRequest struct {
	FullName string `json:"full_name" binding:"required,min=2,max=150"`
	Email    string `json:"email" binding:"required,email,max=255"`
	Phone    string `json:"phone" binding:"required,e164"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

type UserResponse struct {
	ID            string    `json:"id"`
	FullName      string    `json:"full_name"`
	Email         string    `json:"email"`
	EmailVerified bool      `json:"email_verified"`
	Phone         string    `json:"phone"`
	PhoneVerified bool      `json:"phone_verified"`
	Role          string    `json:"role"`
	CreatedAt     time.Time `json:"created_at"`
}

func NewUserResponse(u *user.User) UserResponse {
	return UserResponse{
		ID:            u.ID,
		FullName:      u.FullName,
		Email:         u.Email,
		EmailVerified: u.IsEmailVerified(),
		Phone:         u.Phone,
		PhoneVerified: u.IsPhoneVerified(),
		Role:          string(u.Role),
		CreatedAt:     u.CreatedAt,
	}
}

type VerifyOTPRequest struct {
	UserID string `json:"user_id" binding:"required,uuid"`
	Code   string `json:"code" binding:"required,len=6,numeric"`
}

type ResendOTPRequest struct {
	UserID  string `json:"user_id" binding:"required,uuid"`
	Purpose string `json:"purpose" binding:"required,oneof=email_verification phone_verification"`
}

type LoginRequest struct {
	Identifier string `json:"identifier" binding:"required"`
	Password   string `json:"password" binding:"required"`
}

type TokenResponse struct {
	AccessToken      string    `json:"access_token"`
	AccessExpiresAt  time.Time `json:"access_expires_at"`
	RefreshToken     string    `json:"refresh_token"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
}

type LoginResponse struct {
	User   UserResponse  `json:"user"`
	Tokens TokenResponse `json:"tokens"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Code        string `json:"code" binding:"required,len=6,numeric"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=72"`
}
