package model

import (
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

const (
	RedeemCodeStatusAvailable = "available"
	RedeemCodeStatusClaimed   = "claimed"
	RedeemCodeStatusExpired   = "expired"
)

type GetRedeemCodeParam struct {
	RedeemCodeID uuid.UUID
	RedeemItemID uuid.UUID
	Code         string
}

type ListRedeemCodesParam struct {
	Search       string
	RedeemItemID uuid.UUID
	Status       string
	Limit        int
	Offset       int
}

type AdminRedeemCodeListRow struct {
	RedeemCodeID    uuid.UUID  `json:"redeem_code_id"`
	RedeemItemID    uuid.UUID  `json:"redeem_item_id"`
	RedeemItemName  string     `json:"redeem_item_name"`
	Code            string     `json:"code"`
	Status          string     `json:"status"`
	ClaimedByUserID *uuid.UUID `json:"claimed_by_user_id"`
	ClaimedBy       string     `json:"claimed_by"`
	ClaimedAt       *time.Time `json:"claimed_at"`
	ExpiresAt       time.Time  `json:"expires_at"`
	CreatedAt       time.Time  `json:"created_at"`
}

type AdminCreateRedeemCodeManualRequest struct {
	RedeemItemID string `json:"redeem_item_id" binding:"required"`
	Code         string `json:"code" binding:"required"`
	ExpiresAt    string `json:"expires_at" binding:"required"`
}

type AdminCreateRedeemCodeCSVRequest struct {
	File *multipart.FileHeader `form:"file"`
}

type AdminListRedeemCodesRequest struct {
	Search       string `form:"search"`
	RedeemItemID string `form:"redeem_item_id"`
	Status       string `form:"status"`
	Page         int    `form:"page"`
	Limit        int    `form:"limit"`
}

type AdminRedeemCodeResponse struct {
	RedeemCodeID    uuid.UUID  `json:"redeem_code_id"`
	RedeemItemID    uuid.UUID  `json:"redeem_item_id"`
	RedeemItemName  string     `json:"redeem_item_name"`
	Code            string     `json:"code"`
	Status          string     `json:"status"`
	ClaimedByUserID *uuid.UUID `json:"claimed_by_user_id"`
	ClaimedBy       string     `json:"claimed_by"`
	ClaimedAt       *time.Time `json:"claimed_at"`
	ExpiresAt       time.Time  `json:"expires_at"`
	CreatedAt       time.Time  `json:"created_at"`
}

type AdminCreateRedeemCodeResponse struct {
	RedeemCode AdminRedeemCodeResponse `json:"redeem_code"`
}

type AdminCreateRedeemCodesCSVResponse struct {
	CreatedCount int `json:"created_count"`
}

type AdminListRedeemCodesResponse struct {
	RedeemCodes []AdminRedeemCodeResponse `json:"redeem_codes"`
	Pagination  PaginationResponse        `json:"pagination"`
}

type AdminDeleteRedeemCodeResponse struct {
	RedeemCodeID uuid.UUID `json:"redeem_code_id"`
}
