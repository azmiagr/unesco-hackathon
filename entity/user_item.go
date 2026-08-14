package entity

import (
	"time"

	"github.com/google/uuid"
)

type UserItem struct {
	UserItemID   uuid.UUID  `json:"user_item_id" gorm:"type:varchar(36);primaryKey"`
	UserID       uuid.UUID  `json:"user_id" gorm:"type:varchar(36);not null;uniqueIndex:uq_user_items_user_item;index:idx_user_items_user_redeem_item"`
	ItemID       *uuid.UUID `json:"item_id" gorm:"type:varchar(36);uniqueIndex:uq_user_items_user_item;index"`
	RedeemItemID *uuid.UUID `json:"redeem_item_id" gorm:"type:varchar(36);index:idx_user_items_user_redeem_item"`
	RedeemCodeID *uuid.UUID `json:"redeem_code_id" gorm:"type:varchar(36);uniqueIndex:uq_user_items_redeem_code"`
	PurchaseType string     `json:"purchase_type" gorm:"type:enum('shop','redeem');not null;default:'shop';index"`
	CoinSpent    int        `json:"coin_spent" gorm:"type:int;not null;default:0"`
	PurchasedAt  time.Time  `json:"purchased_at" gorm:"not null;index"`
	EquippedAt   *time.Time `json:"equipped_at" gorm:"index"`
	CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}
