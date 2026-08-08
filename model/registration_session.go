package model

import (
	"time"

	"github.com/google/uuid"
)

type GetRegistrationSessionParam struct {
	RegistrationSessionID uuid.UUID
	SessionTokenHash      string
	Email                 string
	CurrentStep           string
}

type RegistrationSessionStateResponse struct {
	Email       string     `json:"email"`
	CurrentStep string     `json:"current_step"`
	AvatarID    *uuid.UUID `json:"avatar_id,omitempty"`
	Title       string     `json:"title,omitempty"`
	ExpiresAt   time.Time  `json:"expires_at"`
}
