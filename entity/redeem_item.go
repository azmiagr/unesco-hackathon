package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RedeemItem struct {
	RedeemItemID      uuid.UUID      `json:"redeem_item_id" gorm:"type:varchar(36);primaryKey"`
	RedeemTypeID      uuid.UUID      `json:"redeem_type_id" gorm:"type:varchar(36);not null;index"`
	Name              string         `json:"name" gorm:"type:varchar(150);not null"`
	PartnerName       string         `json:"partner_name" gorm:"type:varchar(150);not null"`
	Description       string         `json:"description" gorm:"type:longtext;not null"`
	PriceCoin         int            `json:"price_coin" gorm:"type:int;not null;default:0"`
	MaxClaimPerPeriod int            `json:"max_claim_per_period" gorm:"type:int;not null;default:1"`
	ClaimPeriod       string         `json:"claim_period" gorm:"type:enum('daily','weekly','monthly');not null;default:'weekly';index"`
	MinimumLevel      int            `json:"minimum_level" gorm:"type:int;not null;default:1"`
	ImageURL          string         `json:"image_url" gorm:"type:varchar(500);not null"`
	IsStockVisible    bool           `json:"is_stock_visible" gorm:"type:boolean;not null;default:true;index"`
	Status            string         `json:"status" gorm:"type:enum('active','inactive','retired');not null;default:'active';index"`
	CreatedAt         time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt         gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	RedeemCodes []RedeemCode `json:"redeem_codes" gorm:"foreignKey:RedeemItemID;references:RedeemItemID"`
}
