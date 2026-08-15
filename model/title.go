package model

import (
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

const (
	TitleStatusActive   = "active"
	TitleStatusInactive = "inactive"
)

type GetTitleParam struct {
	TitleID uuid.UUID
}

type UserTitleRow struct {
	TitleID        uuid.UUID  `json:"title_id"`
	Title          string     `json:"title"`
	UnlockLevel    int        `json:"unlock_level"`
	ImageBorder    string     `json:"image_border"`
	Status         string     `json:"status"`
	UserItemID     *uuid.UUID `json:"user_item_id"`
	EquippedAt     *time.Time `json:"equipped_at"`
	CurrentTitleID *uuid.UUID `json:"current_title_id"`
}

type UserTitleResponse struct {
	TitleID     uuid.UUID `json:"title_id"`
	Title       string    `json:"title"`
	UnlockLevel int       `json:"unlock_level"`
	ImageBorder string    `json:"image_border"`
	IsOwned     bool      `json:"is_owned"`
	IsEquipped  bool      `json:"is_equipped"`
	CanEquip    bool      `json:"can_equip"`
}

type ListUserTitlesResponse struct {
	Titles []UserTitleResponse `json:"titles"`
}

type EquipTitleResponse struct {
	Title UserTitleResponse `json:"title"`
}

type AdminListTitlesRequest struct {
	Search string `form:"search"`
	Page   int    `form:"page"`
	Limit  int    `form:"limit"`
}

type AdminCreateTitleRequest struct {
	Title       string                `form:"title" binding:"required"`
	UnlockLevel int                   `form:"unlock_level" binding:"required"`
	Status      string                `form:"status"`
	Image       *multipart.FileHeader `form:"image"`
}

type AdminUpdateTitleRequest struct {
	Title       string                `form:"title" binding:"required"`
	UnlockLevel int                   `form:"unlock_level" binding:"required"`
	Status      string                `form:"status" binding:"required"`
	Image       *multipart.FileHeader `form:"image"`
}

type AdminTitleResponse struct {
	TitleID     uuid.UUID `json:"title_id"`
	Title       string    `json:"title"`
	UnlockLevel int       `json:"unlock_level"`
	ImageBorder string    `json:"image_border"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type AdminListTitlesResponse struct {
	Titles     []AdminTitleResponse `json:"titles"`
	Pagination PaginationResponse   `json:"pagination"`
}

type AdminGetTitleResponse struct {
	Title AdminTitleResponse `json:"title"`
}

type AdminCreateTitleResponse struct {
	Title AdminTitleResponse `json:"title"`
}

type AdminUpdateTitleResponse struct {
	Title AdminTitleResponse `json:"title"`
}

type AdminDeleteTitleResponse struct {
	TitleID uuid.UUID `json:"title_id"`
}
