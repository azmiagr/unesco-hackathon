package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	RegisterStepEmailSubmitted = "email_submitted"
	RegisterStepEmailVerified  = "email_verified"
	RegisterStepAvatarSelected = "avatar_selected"
	RegisterStepCompleted      = "completed"

	DefaultRegisterTitle = "Detektif Baru"
)

type StartRegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type VerifyRegisterOtpRequest struct {
	Code string `json:"code" binding:"required,len=6"`
}

type SelectRegisterAvatarRequest struct {
	AvatarID uuid.UUID `json:"avatar_id" binding:"required"`
}

type CompleteRegisterProfileRequest struct {
	Username string `json:"username" binding:"required,min=3,max=16"`
	Title    string `json:"title" binding:"required"`
}

type RegisterSessionResponse struct {
	Email       string     `json:"email"`
	CurrentStep string     `json:"current_step"`
	AvatarID    *uuid.UUID `json:"avatar_id,omitempty"`
	Title       string     `json:"title,omitempty"`
	ExpiresAt   time.Time  `json:"expires_at"`
}

type RegisterAuthResult struct {
	SessionToken string                  `json:"-"`
	State        RegisterSessionResponse `json:"state"`
}

type CompleteRegisterResult struct {
	Token string `json:"token"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token        string `json:"token,omitempty"`
	RequiresOtp  bool   `json:"requires_otp"`
	Email        string `json:"email,omitempty"`
	SessionToken string `json:"-"`
}

type VerifyAdminLoginOtpRequest struct {
	Code string `json:"code" binding:"required,len=6"`
}
