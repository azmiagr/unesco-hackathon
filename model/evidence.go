package model

import (
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

type GetCaseEvidenceParam struct {
	CaseEvidenceID uuid.UUID
	CaseVersionID  uuid.UUID
	TemplateType   string
}

type ListCaseEvidencesParam struct {
	CaseVersionID uuid.UUID
	TemplateType  string
}

type ReorderCaseEvidenceParam struct {
	CaseEvidenceID uuid.UUID
	SortOrder      int
}

type AdminCreateSocialPostEvidenceRequest struct {
	Label             string                `form:"label" binding:"required"`
	CredibilityTags   string                `form:"credibility_tags" binding:"required"`
	IsCritical        bool                  `form:"is_critical"`
	SortOrder         int                   `form:"sort_order"`
	AuthorName        string                `form:"author_name" binding:"required"`
	AuthorHandle      string                `form:"author_handle" binding:"required"`
	Platform          string                `form:"platform" binding:"required"`
	PostText          string                `form:"post_text" binding:"required"`
	Timestamp         string                `form:"timestamp" binding:"required"`
	LikesCount        int                   `form:"likes_count"`
	SharesCount       int                   `form:"shares_count"`
	CommentsCount     int                   `form:"comments_count"`
	IsVerifiedAccount bool                  `form:"is_verified_account"`
	ImagePrompt       string                `form:"image_prompt"`
	Image             *multipart.FileHeader `form:"image"`
}

type AdminSocialPostEvidenceResponse struct {
	CaseEvidenceID    uuid.UUID `json:"case_evidence_id"`
	CaseVersionID     uuid.UUID `json:"case_version_id"`
	TemplateType      string    `json:"template_type"`
	Label             string    `json:"label"`
	CredibilityTags   []string  `json:"credibility_tags"`
	IsCritical        bool      `json:"is_critical"`
	SortOrder         int       `json:"sort_order"`
	AuthorName        string    `json:"author_name"`
	AuthorHandle      string    `json:"author_handle"`
	Platform          string    `json:"platform"`
	PostText          string    `json:"post_text"`
	Timestamp         time.Time `json:"timestamp"`
	LikesCount        int       `json:"likes_count"`
	SharesCount       int       `json:"shares_count"`
	CommentsCount     int       `json:"comments_count"`
	IsVerifiedAccount bool      `json:"is_verified_account"`
	ImagePrompt       *string   `json:"image_prompt"`
	ImageURL          *string   `json:"image_url"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type AdminCreateSocialPostEvidenceResponse struct {
	Evidence AdminSocialPostEvidenceResponse `json:"evidence"`
}
