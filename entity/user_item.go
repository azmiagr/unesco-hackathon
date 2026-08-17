package entity

import (
	"time"

	"github.com/google/uuid"
)

type UserItem struct {
	UserItemID   uuid.UUID  `json:"user_item_id" gorm:"type:varchar(36);primaryKey"`
	UserID       uuid.UUID  `json:"user_id" gorm:"type:varchar(36);not null;uniqueIndex:uq_user_items_user_item;uniqueIndex:uq_user_items_user_title;index:idx_user_items_user_redeem_item;index:idx_user_items_user_redeem_period,priority:1"`
	ItemID       *uuid.UUID `json:"item_id" gorm:"type:varchar(36);uniqueIndex:uq_user_items_user_item;index"`
	TitleID      *uuid.UUID `json:"title_id" gorm:"type:varchar(36);uniqueIndex:uq_user_items_user_title;index"`
	RedeemItemID *uuid.UUID `json:"redeem_item_id" gorm:"type:varchar(36);index:idx_user_items_user_redeem_item;index:idx_user_items_user_redeem_period,priority:3"`
	RedeemCodeID *uuid.UUID `json:"redeem_code_id" gorm:"type:varchar(36);uniqueIndex:uq_user_items_redeem_code"`
	PurchaseType string     `json:"purchase_type" gorm:"type:enum('shop','redeem','grant');not null;default:'shop';index;index:idx_user_items_user_redeem_period,priority:2"`
	CoinSpent    int        `json:"coin_spent" gorm:"type:int;not null;default:0"`
	PurchasedAt  time.Time  `json:"purchased_at" gorm:"not null;index;index:idx_user_items_user_redeem_period,priority:4"`
	EquippedAt   *time.Time `json:"equipped_at" gorm:"index"`
	CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}
