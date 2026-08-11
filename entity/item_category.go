package entity

import (
	"time"

	"github.com/google/uuid"
)

type ItemCategory struct {
	ItemCategoryID uuid.UUID `json:"item_category_id" gorm:"type:varchar(36);primaryKey"`
	Code           string    `json:"code" gorm:"type:varchar(80);not null;uniqueIndex"`
	Name           string    `json:"name" gorm:"type:varchar(120);not null"`
	Description    *string   `json:"description" gorm:"type:text"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	Items []Item `gorm:"foreignKey:ItemCategoryID;references:ItemCategoryID;constraint:onDelete:RESTRICT"`
}
