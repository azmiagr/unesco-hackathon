package entity

import (
	"time"

	"github.com/google/uuid"
)

type CaseEvidence struct {
	CaseEvidenceID  uuid.UUID `json:"case_evidence_id" gorm:"type:varchar(36);primaryKey"`
	CaseVersionID   uuid.UUID `json:"case_version_id" gorm:"type:varchar(36);not null;index;index:idx_case_evidences_version_sort_created,priority:1"`
	TemplateType    string    `json:"template_type" gorm:"type:varchar(50);not null;index"`
	Label           string    `json:"label" gorm:"type:varchar(150);not null"`
	CredibilityTags string    `json:"credibility_tags" gorm:"type:json;not null"`
	IsCritical      bool      `json:"is_critical" gorm:"not null;default:false"`
	SortOrder       int       `json:"sort_order" gorm:"type:int;not null;default:0;index:idx_case_evidences_version_sort_created,priority:2"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime;index:idx_case_evidences_version_sort_created,priority:3"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	SocialPost         *CaseEvidenceSocialPost         `gorm:"foreignKey:CaseEvidenceID;references:CaseEvidenceID;constraint:onDelete:CASCADE"`
	Article            *CaseEvidenceArticle            `gorm:"foreignKey:CaseEvidenceID;references:CaseEvidenceID;constraint:onDelete:CASCADE"`
	Blog               *CaseEvidenceBlog               `gorm:"foreignKey:CaseEvidenceID;references:CaseEvidenceID;constraint:onDelete:CASCADE"`
	ForumThread        *CaseEvidenceForumThread        `gorm:"foreignKey:CaseEvidenceID;references:CaseEvidenceID;constraint:onDelete:CASCADE"`
	ChatTranscript     *CaseEvidenceChatTranscript     `gorm:"foreignKey:CaseEvidenceID;references:CaseEvidenceID;constraint:onDelete:CASCADE"`
	PublicAnnouncement *CaseEvidencePublicAnnouncement `gorm:"foreignKey:CaseEvidenceID;references:CaseEvidenceID;constraint:onDelete:CASCADE"`
	QuestionReferences []CaseQuestionEvidenceReference `gorm:"foreignKey:CaseEvidenceID;references:CaseEvidenceID;constraint:onDelete:CASCADE"`
}
