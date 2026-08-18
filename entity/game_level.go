package entity

import (
	"time"

	"github.com/google/uuid"
)

type GameLevel struct {
	GameLevelID uuid.UUID `json:"game_level_id" gorm:"type:varchar(36);primaryKey"`
	Level       int       `json:"level" gorm:"type:int;not null;uniqueIndex"`
	XPRequired  int       `json:"xp_required" gorm:"type:int;not null;default:0;index"`
	Title       string    `json:"title" gorm:"type:varchar(150);not null"`
	RewardCoin  int       `json:"reward_coin" gorm:"type:int;not null;default:0"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
