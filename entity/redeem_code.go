package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RedeemCode struct {
	RedeemCodeID     uuid.UUID      `json:"redeem_code_id" gorm:"type:varchar(36);primaryKey"`
	RedeemItemID     uuid.UUID      `json:"redeem_item_id" gorm:"type:varchar(36);not null;index"`
	Code             string         `json:"code" gorm:"type:varchar(120);not null;uniqueIndex"`
	ExpiresAt        time.Time      `json:"expires_at" gorm:"not null;index"`
	ClaimedByUserID  *uuid.UUID     `json:"claimed_by_user_id" gorm:"type:varchar(36);index"`
	ClaimedAt        *time.Time     `json:"claimed_at" gorm:"index"`
	CreatedByAdminID *uuid.UUID     `json:"created_by_admin_id" gorm:"type:varchar(36);index"`
	CreatedAt        time.Time      `json:"created_at" gorm:"autoCreateTime"`
	DeletedAt        gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}
