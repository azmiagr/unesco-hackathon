package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Item struct {
	ItemID         uuid.UUID      `json:"item_id" gorm:"type:varchar(36);primaryKey"`
	ItemCategoryID uuid.UUID      `json:"item_category_id" gorm:"type:varchar(36);not null;index"`
	Name           string         `json:"name" gorm:"type:varchar(150);not null"`
	Description    string         `json:"description" gorm:"type:text;not null"`
	PriceCoin      int            `json:"price_coin" gorm:"type:int;not null;default:0"`
	ImageURL       string         `json:"image_url" gorm:"type:varchar(500);not null"`
	IsVisible      bool           `json:"is_visible" gorm:"type:boolean;not null;default:true;index"`
	IsFeatured     bool           `json:"is_featured" gorm:"type:boolean;not null;default:false;index"`
	Status         string         `json:"status" gorm:"type:enum('active','inactive','retired');not null;default:'active';index"`
	CreatedAt      time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}
