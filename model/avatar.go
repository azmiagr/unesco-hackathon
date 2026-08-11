package model

import (
	"time"

	"github.com/google/uuid"
)

type GetAvatarParam struct {
	AvatarID uuid.UUID `json:"-"`
}

type AvatarResponse struct {
	AvatarID    uuid.UUID `json:"avatar_id"`
	ImageURL    string    `json:"image_url"`
	UnlockLevel int       `json:"unlock_level"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ListAvatarsResponse struct {
	Avatars []AvatarResponse `json:"avatars"`
}
