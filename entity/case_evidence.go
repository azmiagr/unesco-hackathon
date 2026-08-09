package entity

import (
	"time"

	"github.com/google/uuid"
)

type CaseEvidence struct {
	CaseEvidenceID  uuid.UUID `json:"case_evidence_id" gorm:"type:varchar(36);primaryKey"`
	CaseVersionID   uuid.UUID `json:"case_version_id" gorm:"type:varchar(36);not null;index"`
	TemplateType    string    `json:"template_type" gorm:"type:varchar(50);not null;index"`
	Label           string    `json:"label" gorm:"type:varchar(150);not null"`
	CredibilityTags string    `json:"credibility_tags" gorm:"type:json;not null"`
	IsCritical      bool      `json:"is_critical" gorm:"not null;default:false"`
	SortOrder       int       `json:"sort_order" gorm:"type:int;not null;default:0"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	SocialPost *CaseEvidenceSocialPost `gorm:"foreignKey:CaseEvidenceID;references:CaseEvidenceID;constraint:onDelete:CASCADE"`
}
