package model

import (
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

const (
	ItemCategoryAvatar = "avatar"

	ItemStatusActive   = "active"
	ItemStatusInactive = "inactive"
	ItemStatusRetired  = "retired"
)

type GetItemCategoryParam struct {
	ItemCategoryID uuid.UUID
	Code           string
}

type ListItemCategoriesParam struct {
	Search string
}

type GetItemParam struct {
	ItemID uuid.UUID
}

type ListItemsParam struct {
	Search         string
	ItemCategoryID uuid.UUID
	CategoryCode   string
	Status         string
	IsVisible      *bool
	IsFeatured     *bool
	Limit          int
	Offset         int
}

type AdminListItemsRequest struct {
	Search         string `form:"search"`
	ItemCategoryID string `form:"item_category_id"`
	CategoryCode   string `form:"category_code"`
	Status         string `form:"status"`
	IsVisible      *bool  `form:"is_visible"`
	IsFeatured     *bool  `form:"is_featured"`
	Page           int    `form:"page"`
	Limit          int    `form:"limit"`
}

type AdminCreateItemRequest struct {
	Name           string                `form:"name" binding:"required"`
	ItemCategoryID string                `form:"item_category_id" binding:"required"`
	Description    string                `form:"description" binding:"required"`
	PriceCoin      int                   `form:"price_coin" binding:"required"`
	IsVisible      *bool                 `form:"is_visible"`
	IsFeatured     *bool                 `form:"is_featured"`
	Status         string                `form:"status"`
	Image          *multipart.FileHeader `form:"image"`
}

type AdminUpdateItemRequest struct {
	Name           string                `form:"name" binding:"required"`
	ItemCategoryID string                `form:"item_category_id" binding:"required"`
	Description    string                `form:"description" binding:"required"`
	PriceCoin      int                   `form:"price_coin" binding:"required"`
	IsVisible      *bool                 `form:"is_visible"`
	IsFeatured     *bool                 `form:"is_featured"`
	Status         string                `form:"status" binding:"required"`
	Image          *multipart.FileHeader `form:"image"`
}

type AdminItemCategoryResponse struct {
	ItemCategoryID uuid.UUID `json:"item_category_id"`
	Code           string    `json:"code"`
	Name           string    `json:"name"`
	Description    *string   `json:"description"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type AdminItemResponse struct {
	ItemID         uuid.UUID                 `json:"item_id"`
	ItemCategoryID uuid.UUID                 `json:"item_category_id"`
	Category       AdminItemCategoryResponse `json:"category"`
	Name           string                    `json:"name"`
	Description    string                    `json:"description"`
	PriceCoin      int                       `json:"price_coin"`
	ImageURL       string                    `json:"image_url"`
	IsVisible      bool                      `json:"is_visible"`
	IsFeatured     bool                      `json:"is_featured"`
	Status         string                    `json:"status"`
	CreatedAt      time.Time                 `json:"created_at"`
	UpdatedAt      time.Time                 `json:"updated_at"`
}

type AdminListItemsResponse struct {
	Items      []AdminItemResponse `json:"items"`
	Pagination PaginationResponse  `json:"pagination"`
}

type AdminGetItemDetailResponse struct {
	Item AdminItemResponse `json:"item"`
}

type AdminCreateItemResponse struct {
	Item AdminItemResponse `json:"item"`
}

type AdminUpdateItemResponse struct {
	Item AdminItemResponse `json:"item"`
}

type AdminDeleteItemResponse struct {
	ItemID uuid.UUID `json:"item_id"`
}

type AdminListItemCategoriesRequest struct {
	Search string `form:"search"`
}

type AdminListItemCategoriesResponse struct {
	Categories []AdminItemCategoryResponse `json:"categories"`
}
