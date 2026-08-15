package entity

import (
	"time"

	"github.com/google/uuid"
)

type Title struct {
	TitleID     uuid.UUID `json:"title_id" gorm:"type:varchar(36);primaryKey"`
	Title       string    `json:"title" gorm:"type:varchar(150);not null;uniqueIndex"`
	UnlockLevel int       `json:"unlock_level" gorm:"type:int;not null;index"`
	ImageBorder string    `json:"image_border" gorm:"type:varchar(500);not null"`
	Status      string    `json:"status" gorm:"type:varchar(30);not null;default:'active';index"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
