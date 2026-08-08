package model

import "github.com/google/uuid"

type GetAvatarParam struct {
	AvatarID uuid.UUID `json:"-"`
}
