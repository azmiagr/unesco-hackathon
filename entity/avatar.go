package entity

import (
	"time"

	"github.com/google/uuid"
)

type Avatar struct {
	AvatarID    uuid.UUID `json:"avatar_id" gorm:"type:varchar(36);primaryKey"`
	ImageURL    string    `json:"image_url" gorm:"type:varchar(255);not null"`
	UnlockLevel int       `json:"unlock_level" gorm:"type:int;default:0"`
	Status      string    `json:"status" gorm:"type:enum('active', 'inactive');default:'active'"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	UserProfiles []UserProfile `gorm:"foreignKey:AvatarID;references:AvatarID;constraint:onDelete:CASCADE"`
}
