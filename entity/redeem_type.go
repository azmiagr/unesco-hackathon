package entity

import (
	"time"

	"github.com/google/uuid"
)

type RedeemType struct {
	RedeemTypeID uuid.UUID `json:"redeem_type_id" gorm:"type:varchar(36);primaryKey"`
	Code         string    `json:"code" gorm:"type:varchar(80);not null;uniqueIndex"`
	Name         string    `json:"name" gorm:"type:varchar(120);not null"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	RedeemItems []RedeemItem `gorm:"foreignKey:RedeemTypeID;references:RedeemTypeID;constraint:onDelete:RESTRICT"`
}
