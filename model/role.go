package model

import "github.com/google/uuid"

type RoleResponse struct {
	RoleID   uuid.UUID `json:"role_id"`
	RoleName string    `json:"role_name"`
}
