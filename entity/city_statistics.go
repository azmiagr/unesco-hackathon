package entity

import (
	"time"

	"github.com/google/uuid"
)

type CityStatistics struct {
	CityStatisticsID  uuid.UUID `json:"city_statistics_id" gorm:"type:varchar(36);primaryKey"`
	StatKey           string    `json:"stat_key" gorm:"type:varchar(80);not null;uniqueIndex"`
	InformationHealth int       `json:"information_health" gorm:"type:int;not null;default:70"`
	PublicTrust       int       `json:"public_trust" gorm:"type:int;not null;default:70"`
	SocialStability   int       `json:"social_stability" gorm:"type:int;not null;default:70"`
	PublicWellbeing   int       `json:"public_wellbeing" gorm:"type:int;not null;default:70"`
	VisualState       string    `json:"visual_state" gorm:"type:varchar(50);not null;default:'normal'"`
	CreatedAt         time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
