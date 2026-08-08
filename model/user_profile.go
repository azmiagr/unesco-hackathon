package model

import "github.com/google/uuid"

type GetUserProfileParam struct {
	UserProfileID uuid.UUID `json:"-"`
	UserID        uuid.UUID `json:"-"`
}
