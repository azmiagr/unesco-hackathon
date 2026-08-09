package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Case struct {
	CaseID                   uuid.UUID      `json:"case_id" gorm:"type:varchar(36);primaryKey"`
	CreatedBy                uuid.UUID      `json:"created_by" gorm:"type:varchar(36);not null;index"`
	Title                    string         `json:"title" gorm:"type:varchar(200);not null"`
	Slug                     string         `json:"slug" gorm:"type:varchar(220);not null;uniqueIndex"`
	ShortDescription         string         `json:"short_description" gorm:"type:text;not null"`
	Theme                    string         `json:"theme" gorm:"type:varchar(100);not null;index"`
	ThemeOtherText           *string        `json:"theme_other_text" gorm:"type:varchar(255)"`
	CompetencyFocus          string         `json:"competency_focus" gorm:"type:varchar(50);not null;index"`
	DifficultyLevel          string         `json:"difficulty_level" gorm:"type:enum('low','medium','high');not null;index"`
	RiskLevel                string         `json:"risk_level" gorm:"type:enum('low','medium','high');not null;index"`
	EstimatedDurationMinutes int            `json:"estimated_duration_minutes" gorm:"type:int;not null"`
	MinimumLevel             int            `json:"minimum_level" gorm:"type:int;not null;default:1"`
	MinimumReputation        float64        `json:"minimum_reputation" gorm:"type:decimal(8,2);not null;default:0"`
	UnlockRequirement        *string        `json:"unlock_requirement" gorm:"type:json"`
	ThumbnailURL             *string        `json:"thumbnail_url" gorm:"type:varchar(500)"`
	ThumbnailPrompt          *string        `json:"thumbnail_prompt" gorm:"type:text"`
	GenerationSource         string         `json:"generation_source" gorm:"type:enum('manual','ai_assisted');not null;default:'manual';index"`
	Status                   string         `json:"status" gorm:"type:enum('draft','published','archived');not null;default:'draft';index"`
	PublishedAt              *time.Time     `json:"published_at"`
	CreatedAt                time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt                time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt                gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	Versions []CaseVersion `gorm:"foreignKey:CaseID;references:CaseID;constraint:onDelete:CASCADE"`
}
