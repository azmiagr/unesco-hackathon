package model

import (
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

const (
	RedeemItemStatusActive   = "active"
	RedeemItemStatusInactive = "inactive"
	RedeemItemStatusRetired  = "retired"

	RedeemClaimPeriodDaily   = "daily"
	RedeemClaimPeriodWeekly  = "weekly"
	RedeemClaimPeriodMonthly = "monthly"

	UserRedeemItemFilterAll   = "all"
	UserRedeemItemFilterOwned = "owned"
)

type GetRedeemTypeParam struct {
	RedeemTypeID uuid.UUID
	Code         string
}

type ListRedeemTypesParam struct {
	Search string
}

type GetRedeemItemParam struct {
	RedeemItemID uuid.UUID
	Name         string
}

type ListRedeemItemsParam struct {
	Search       string
	RedeemTypeID uuid.UUID
	TypeCode     string
	Status       string
	ClaimPeriod  string
	Limit        int
	Offset       int
}

type ListRedeemItemsForUserParam struct {
	UserID       uuid.UUID
	RedeemItemID uuid.UUID
	Search       string
	Filter       string
	Limit        int
	Offset       int
}

type AdminListRedeemTypesRequest struct {
	Search string `form:"search"`
}

type AdminListRedeemItemsRequest struct {
	Search       string `form:"search"`
	RedeemTypeID string `form:"redeem_type_id"`
	TypeCode     string `form:"type_code"`
	Status       string `form:"status"`
	ClaimPeriod  string `form:"claim_period"`
	Page         int    `form:"page"`
	Limit        int    `form:"limit"`
}

type AdminCreateRedeemItemRequest struct {
	Name              string                `form:"name" binding:"required"`
	Type              string                `form:"type" binding:"required"`
	PartnerName       string                `form:"partner_name" binding:"required"`
	Description       string                `form:"description" binding:"required"`
	PriceCoin         int                   `form:"price_coin" binding:"required"`
	MaxClaimPerPeriod int                   `form:"max_claim_per_period" binding:"required"`
	ClaimPeriod       string                `form:"claim_period" binding:"required"`
	MinimumLevel      int                   `form:"minimum_level"`
	IsStockVisible    *bool                 `form:"is_stock_visible"`
	Status            string                `form:"status"`
	Image             *multipart.FileHeader `form:"image"`
}

type AdminUpdateRedeemItemRequest struct {
	Name              string                `form:"name" binding:"required"`
	Type              string                `form:"type" binding:"required"`
	PartnerName       string                `form:"partner_name" binding:"required"`
	Description       string                `form:"description" binding:"required"`
	PriceCoin         int                   `form:"price_coin" binding:"required"`
	MaxClaimPerPeriod int                   `form:"max_claim_per_period" binding:"required"`
	ClaimPeriod       string                `form:"claim_period" binding:"required"`
	MinimumLevel      int                   `form:"minimum_level"`
	IsStockVisible    *bool                 `form:"is_stock_visible"`
	Status            string                `form:"status" binding:"required"`
	Image             *multipart.FileHeader `form:"image"`
}

type UserListRedeemItemsRequest struct {
	Search string `form:"search"`
	Filter string `form:"filter"`
	Page   int    `form:"page"`
	Limit  int    `form:"limit"`
}

type AdminRedeemTypeResponse struct {
	RedeemTypeID uuid.UUID `json:"redeem_type_id"`
	Code         string    `json:"code"`
	Name         string    `json:"name"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type AdminRedeemItemResponse struct {
	RedeemItemID      uuid.UUID               `json:"redeem_item_id"`
	RedeemTypeID      uuid.UUID               `json:"redeem_type_id"`
	Type              AdminRedeemTypeResponse `json:"type"`
	Name              string                  `json:"name"`
	PartnerName       string                  `json:"partner_name"`
	Description       string                  `json:"description"`
	PriceCoin         int                     `json:"price_coin"`
	MaxClaimPerPeriod int                     `json:"max_claim_per_period"`
	ClaimPeriod       string                  `json:"claim_period"`
	MinimumLevel      int                     `json:"minimum_level"`
	ImageURL          string                  `json:"image_url"`
	IsStockVisible    bool                    `json:"is_stock_visible"`
	Status            string                  `json:"status"`
	CreatedAt         time.Time               `json:"created_at"`
	UpdatedAt         time.Time               `json:"updated_at"`
}

type UserRedeemItemRow struct {
	RedeemItemID      uuid.UUID `json:"redeem_item_id"`
	RedeemTypeID      uuid.UUID `json:"redeem_type_id"`
	TypeCode          string    `json:"type_code"`
	TypeName          string    `json:"type_name"`
	Name              string    `json:"name"`
	PartnerName       string    `json:"partner_name"`
	Description       string    `json:"description"`
	PriceCoin         int       `json:"price_coin"`
	MaxClaimPerPeriod int       `json:"max_claim_per_period"`
	ClaimPeriod       string    `json:"claim_period"`
	MinimumLevel      int       `json:"minimum_level"`
	ImageURL          string    `json:"image_url"`
	IsStockVisible    bool      `json:"is_stock_visible"`
	StockRemaining    int       `json:"stock_remaining"`
	UserClaimCount    int       `json:"user_claim_count"`
	OwnedCount        int       `json:"owned_count"`
	CreatedAt         time.Time `json:"created_at"`
}

type UserRedeemItemResponse struct {
	RedeemItemID      uuid.UUID `json:"redeem_item_id"`
	RedeemTypeID      uuid.UUID `json:"redeem_type_id"`
	TypeCode          string    `json:"type_code"`
	TypeName          string    `json:"type_name"`
	Name              string    `json:"name"`
	PartnerName       string    `json:"partner_name"`
	Description       string    `json:"description"`
	PriceCoin         int       `json:"price_coin"`
	MaxClaimPerPeriod int       `json:"max_claim_per_period"`
	ClaimPeriod       string    `json:"claim_period"`
	MinimumLevel      int       `json:"minimum_level"`
	ImageURL          string    `json:"image_url"`
	IsStockVisible    bool      `json:"is_stock_visible"`
	StockRemaining    int       `json:"stock_remaining"`
	UserClaimCount    int       `json:"user_claim_count"`
	OwnedCount        int       `json:"owned_count"`
	IsOwned           bool      `json:"is_owned"`
	CanPurchase       bool      `json:"can_purchase"`
}

type AdminListRedeemTypesResponse struct {
	Types []AdminRedeemTypeResponse `json:"types"`
}

type AdminListRedeemItemsResponse struct {
	Items      []AdminRedeemItemResponse `json:"items"`
	Pagination PaginationResponse        `json:"pagination"`
}

type AdminGetRedeemItemDetailResponse struct {
	Item AdminRedeemItemResponse `json:"item"`
}

type AdminCreateRedeemItemResponse struct {
	Item AdminRedeemItemResponse `json:"item"`
}

type AdminUpdateRedeemItemResponse struct {
	Item AdminRedeemItemResponse `json:"item"`
}

type AdminDeleteRedeemItemResponse struct {
	RedeemItemID uuid.UUID `json:"redeem_item_id"`
}

type UserListRedeemItemsResponse struct {
	Items      []UserRedeemItemResponse `json:"items"`
	Pagination PaginationResponse       `json:"pagination"`
}

type UserPurchaseRedeemItemResponse struct {
	Item        UserRedeemItemResponse `json:"item"`
	Code        string                 `json:"code"`
	CoinBalance int                    `json:"coin_balance"`
}
