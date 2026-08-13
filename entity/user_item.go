package entity

import (
	"time"

	"github.com/google/uuid"
)

type UserItem struct {
	UserItemID  uuid.UUID  `json:"user_item_id" gorm:"type:varchar(36);primaryKey"`
	UserID      uuid.UUID  `json:"user_id" gorm:"type:varchar(36);not null;uniqueIndex:idx_user_item_owner"`
	ItemID      uuid.UUID  `json:"item_id" gorm:"type:varchar(36);not null;uniqueIndex:idx_user_item_owner;index"`
	PurchasedAt time.Time  `json:"purchased_at" gorm:"not null;index"`
	EquippedAt  *time.Time `json:"equipped_at" gorm:"index"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}
