package entity

import (
	"time"

	"github.com/google/uuid"
)

type CaseEvidenceSocialPost struct {
	CaseEvidenceID    uuid.UUID `json:"case_evidence_id" gorm:"type:varchar(36);primaryKey"`
	AuthorName        string    `json:"author_name" gorm:"type:varchar(150);not null"`
	AuthorHandle      string    `json:"author_handle" gorm:"type:varchar(100);not null"`
	Platform          string    `json:"platform" gorm:"type:varchar(100);not null"`
	PostText          string    `json:"post_text" gorm:"type:text;not null"`
	Timestamp         time.Time `json:"timestamp" gorm:"not null"`
	LikesCount        int       `json:"likes_count" gorm:"type:int;not null;default:0"`
	SharesCount       int       `json:"shares_count" gorm:"type:int;not null;default:0"`
	CommentsCount     int       `json:"comments_count" gorm:"type:int;not null;default:0"`
	IsVerifiedAccount bool      `json:"is_verified_account" gorm:"not null;default:false"`
	ImagePrompt       *string   `json:"image_prompt" gorm:"type:text"`
	ImageURL          *string   `json:"image_url" gorm:"type:varchar(500)"`
	CreatedAt         time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
