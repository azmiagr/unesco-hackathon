package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CaseVersion struct {
	CaseVersionID uuid.UUID      `json:"case_version_id" gorm:"type:varchar(36);primaryKey"`
	CaseID        uuid.UUID      `json:"case_id" gorm:"type:varchar(36);not null;index"`
	VersionNumber int            `json:"version_number" gorm:"type:int;not null;index"`
	Status        string         `json:"status" gorm:"type:enum('draft','published','archived');not null;default:'draft';index"`
	Briefing      *string        `json:"briefing" gorm:"type:json"`
	Evidence      string         `json:"evidence" gorm:"type:json;not null"`
	Questions     string         `json:"questions" gorm:"type:json;not null"`
	ChatbotConfig *string        `json:"chatbot_config" gorm:"type:json"`
	ScoringRule   *string        `json:"scoring_rule" gorm:"type:json"`
	OutcomeRules  *string        `json:"outcome_rules" gorm:"type:json"`
	CreatedBy     uuid.UUID      `json:"created_by" gorm:"type:varchar(36);not null;index"`
	PublishedAt   *time.Time     `json:"published_at"`
	CreatedAt     time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}
