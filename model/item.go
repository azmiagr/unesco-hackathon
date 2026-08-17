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

	UserItemPurchaseTypeShop   = "shop"
	UserItemPurchaseTypeRedeem = "redeem"
	UserItemPurchaseTypeGrant  = "grant"

	UserShopItemOwnershipNotOwned = "not_owned"
	UserShopItemOwnershipOwned    = "owned"
	UserShopItemOwnershipEquipped = "equipped"
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

type GetUserItemParam struct {
	UserID  uuid.UUID
	ItemID  uuid.UUID
	TitleID uuid.UUID
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

type ListVisibleShopItemsParam struct {
	UserID        uuid.UUID
	ItemID        uuid.UUID
	ExcludeItemID uuid.UUID
	Search        string
	CategoryCode  string
	Limit         int
	Offset        int
	Random        bool
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

type UserListShopItemsRequest struct {
	Search       string `form:"search"`
	CategoryCode string `form:"category_code"`
	Page         int    `form:"page"`
	Limit        int    `form:"limit"`
}

type AdminCreateItemRequest struct {
	Name           string                `form:"name" binding:"required"`
	ItemCategoryID string                `form:"item_category_id" binding:"required"`
	AvatarID       string                `form:"avatar_id"`
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
	AvatarID       string                `form:"avatar_id"`
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

type UserShopItemRow struct {
	ItemID          uuid.UUID  `json:"item_id"`
	ItemCategoryID  uuid.UUID  `json:"item_category_id"`
	CategoryCode    string     `json:"category_code"`
	CategoryName    string     `json:"category_name"`
	AvatarID        *uuid.UUID `json:"avatar_id"`
	Name            string     `json:"name"`
	Description     string     `json:"description"`
	PriceCoin       int        `json:"price_coin"`
	ImageURL        string     `json:"image_url"`
	UserItemID      *uuid.UUID `json:"user_item_id"`
	EquippedAt      *time.Time `json:"equipped_at"`
	CurrentAvatarID *uuid.UUID `json:"current_avatar_id"`
	CreatedAt       time.Time  `json:"created_at"`
}

type UserShopItemResponse struct {
	ItemID          uuid.UUID  `json:"item_id"`
	ItemCategoryID  uuid.UUID  `json:"item_category_id"`
	CategoryCode    string     `json:"category_code"`
	CategoryName    string     `json:"category_name"`
	AvatarID        *uuid.UUID `json:"avatar_id"`
	Name            string     `json:"name"`
	Description     string     `json:"description"`
	PriceCoin       int        `json:"price_coin"`
	ImageURL        string     `json:"image_url"`
	OwnershipStatus string     `json:"ownership_status"`
	IsOwned         bool       `json:"is_owned"`
	IsEquipped      bool       `json:"is_equipped"`
	CanPurchase     bool       `json:"can_purchase"`
	CanEquip        bool       `json:"can_equip"`
}

type UserItemCategoryResponse struct {
	ItemCategoryID uuid.UUID `json:"item_category_id"`
	Code           string    `json:"code"`
	Name           string    `json:"name"`
	Description    *string   `json:"description"`
}

type AdminItemResponse struct {
	ItemID         uuid.UUID                 `json:"item_id"`
	ItemCategoryID uuid.UUID                 `json:"item_category_id"`
	AvatarID       *uuid.UUID                `json:"avatar_id"`
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

type UserListShopItemsResponse struct {
	Items      []UserShopItemResponse `json:"items"`
	Pagination PaginationResponse     `json:"pagination"`
}

type UserGetShopItemDetailResponse struct {
	Item         UserShopItemResponse   `json:"item"`
	RelatedItems []UserShopItemResponse `json:"related_items"`
	CoinBalance  int                    `json:"coin_balance"`
}

type UserPurchaseShopItemResponse struct {
	Item        UserShopItemResponse `json:"item"`
	CoinBalance int                  `json:"coin_balance"`
}

type UserEquipShopItemResponse struct {
	Item     UserShopItemResponse `json:"item"`
	AvatarID uuid.UUID            `json:"avatar_id"`
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

type UserListItemCategoriesRequest struct {
	Search string `form:"search"`
}

type UserListItemCategoriesResponse struct {
	Categories []UserItemCategoryResponse `json:"categories"`
}
