package model

import (
	"time"

	"github.com/google/uuid"
)

type UserInventoryItemRow struct {
	UserItemID          uuid.UUID  `json:"user_item_id"`
	PurchaseType        string     `json:"purchase_type"`
	CoinSpent           int        `json:"coin_spent"`
	PurchasedAt         time.Time  `json:"purchased_at"`
	EquippedAt          *time.Time `json:"equipped_at"`
	ItemID              *uuid.UUID `json:"item_id"`
	TitleID             *uuid.UUID `json:"title_id"`
	ItemCategoryID      *uuid.UUID `json:"item_category_id"`
	CategoryCode        string     `json:"category_code"`
	CategoryName        string     `json:"category_name"`
	AvatarID            *uuid.UUID `json:"avatar_id"`
	ItemName            string     `json:"item_name"`
	ItemDescription     string     `json:"item_description"`
	ItemImageURL        string     `json:"item_image_url"`
	ItemStatus          string     `json:"item_status"`
	TitleName           string     `json:"title_name"`
	TitleUnlockLevel    int        `json:"title_unlock_level"`
	TitleImageBorder    string     `json:"title_image_border"`
	RedeemItemID        *uuid.UUID `json:"redeem_item_id"`
	RedeemCodeID        *uuid.UUID `json:"redeem_code_id"`
	RedeemTypeID        *uuid.UUID `json:"redeem_type_id"`
	RedeemTypeCode      string     `json:"redeem_type_code"`
	RedeemTypeName      string     `json:"redeem_type_name"`
	RedeemName          string     `json:"redeem_name"`
	PartnerName         string     `json:"partner_name"`
	RedeemDescription   string     `json:"redeem_description"`
	RedeemImageURL      string     `json:"redeem_image_url"`
	RedeemCode          string     `json:"redeem_code"`
	RedeemCodeExpiresAt *time.Time `json:"redeem_code_expires_at"`
	RedeemCodeClaimedAt *time.Time `json:"redeem_code_claimed_at"`
}

type UserInventoryResponse struct {
	Groups  []UserInventoryGroupResponse `json:"groups"`
	Summary UserInventorySummaryResponse `json:"summary"`
}

type UserInventorySummaryResponse struct {
	TotalItems  int `json:"total_items"`
	ShopCount   int `json:"shop_count"`
	RedeemCount int `json:"redeem_count"`
	GrantCount  int `json:"grant_count"`
}

type UserInventoryGroupResponse struct {
	Type  string                      `json:"type"`
	Label string                      `json:"label"`
	Count int                         `json:"count"`
	Items []UserInventoryItemResponse `json:"items"`
}

type UserInventoryItemResponse struct {
	UserItemID   uuid.UUID                    `json:"user_item_id"`
	PurchaseType string                       `json:"purchase_type"`
	CoinSpent    int                          `json:"coin_spent"`
	PurchasedAt  time.Time                    `json:"purchased_at"`
	EquippedAt   *time.Time                   `json:"equipped_at"`
	Shop         *UserInventoryShopResponse   `json:"shop,omitempty"`
	Redeem       *UserInventoryRedeemResponse `json:"redeem,omitempty"`
	Title        *UserInventoryTitleResponse  `json:"title,omitempty"`
}

type UserInventoryTitleResponse struct {
	TitleID     uuid.UUID `json:"title_id"`
	Title       string    `json:"title"`
	UnlockLevel int       `json:"unlock_level"`
	ImageBorder string    `json:"image_border"`
	IsEquipped  bool      `json:"is_equipped"`
}

type UserInventoryShopResponse struct {
	ItemID         uuid.UUID  `json:"item_id"`
	ItemCategoryID uuid.UUID  `json:"item_category_id"`
	CategoryCode   string     `json:"category_code"`
	CategoryName   string     `json:"category_name"`
	AvatarID       *uuid.UUID `json:"avatar_id"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	ImageURL       string     `json:"image_url"`
	Status         string     `json:"status"`
	IsEquipped     bool       `json:"is_equipped"`
}

type UserInventoryRedeemResponse struct {
	RedeemItemID  uuid.UUID  `json:"redeem_item_id"`
	RedeemCodeID  *uuid.UUID `json:"redeem_code_id"`
	RedeemTypeID  uuid.UUID  `json:"redeem_type_id"`
	TypeCode      string     `json:"type_code"`
	TypeName      string     `json:"type_name"`
	Name          string     `json:"name"`
	PartnerName   string     `json:"partner_name"`
	Description   string     `json:"description"`
	ImageURL      string     `json:"image_url"`
	Code          string     `json:"code"`
	CodeStatus    string     `json:"code_status"`
	CodeExpiresAt *time.Time `json:"code_expires_at"`
	CodeClaimedAt *time.Time `json:"code_claimed_at"`
}
