package model

import (
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

const (
	CaseStatusDraft     = "draft"
	CaseStatusPublished = "published"
	CaseStatusArchived  = "archived"

	CaseGenerationManual     = "manual"
	CaseGenerationAIAssisted = "ai_assisted"

	CaseThemeMisleadingHealthAdvice = "misleading_health_advice"
	CaseThemeChatbotHallucination   = "chatbot_hallucination"
	CaseThemeClickbaitHeadline      = "clickbait_headline"
	CaseThemeStatisticOutOfContext  = "statistic_out_of_context"
	CaseThemeForumMisinformation    = "forum_misinformation"
	CaseThemeViralConflictContent   = "viral_conflict_content"
	CaseThemeAlgorithmicEchoChamber = "algorithmic_echo_chamber"
	CaseThemeOther                  = "other"

	CaseCompetencyEvidenceEvaluation    = "evidence_evaluation"
	CaseCompetencyClaimAnalysis         = "claim_analysis"
	CaseCompetencyConfidenceCalibration = "confidence_calibration"
	CaseCompetencyReasoning             = "reasoning"
	CaseCompetencySafetyJudgment        = "safety_judgment"

	CaseDifficultyLow    = "low"
	CaseDifficultyMedium = "medium"
	CaseDifficultyHigh   = "high"

	CaseRiskLow    = "low"
	CaseRiskMedium = "medium"
	CaseRiskHigh   = "high"

	InitialCaseVersionNumber = 1
	EmptyJSONArray           = "[]"

	CaseEvidenceTemplateSocialPost         = "social_post"
	CaseEvidenceTemplateArticle            = "article"
	CaseEvidenceTemplateBlog               = "blog"
	CaseEvidenceTemplateForumThread        = "forum_thread"
	CaseEvidenceTemplateChatTranscript     = "chat_transcript"
	CaseEvidenceTemplatePublicAnnouncement = "public_announcement"

	CaseQuestionTypeMCQ              = "mcq"
	CaseQuestionTypeOpenEnded        = "open_ended"
	CaseQuestionTypeConfidenceSlider = "confidence_slider"
)

type GetCaseParam struct {
	CaseID uuid.UUID
	Slug   string
}

type GetCaseVersionParam struct {
	CaseVersionID uuid.UUID
	CaseID        uuid.UUID
	VersionNumber int
	Status        string
}

type ListCasesParam struct {
	Search           string
	Status           string
	Theme            string
	CompetencyFocus  string
	DifficultyLevel  string
	RiskLevel        string
	GenerationSource string
	Limit            int
	Offset           int
}

type AdminListCasesParam struct {
	Search           string
	Status           string
	Theme            string
	CompetencyFocus  string
	DifficultyLevel  string
	RiskLevel        string
	GenerationSource string
	Limit            int
	Offset           int
}

type AdminCreateCaseRequest struct {
	Title                    string                `form:"title" binding:"required"`
	ShortDescription         string                `form:"short_description" binding:"required"`
	Theme                    string                `form:"theme" binding:"required"`
	ThemeOtherText           string                `form:"theme_other_text"`
	CompetencyFocus          string                `form:"competency_focus" binding:"required"`
	DifficultyLevel          string                `form:"difficulty_level" binding:"required"`
	RiskLevel                string                `form:"risk_level" binding:"required"`
	EstimatedDurationMinutes int                   `form:"estimated_duration_minutes" binding:"required"`
	AIModel                  string                `form:"ai_model"`
	MinimumLevel             int                   `form:"minimum_level"`
	MinimumReputation        float64               `form:"minimum_reputation"`
	UnlockRequirement        string                `form:"unlock_requirement"`
	ThumbnailPrompt          string                `form:"thumbnail_prompt"`
	GenerationSource         string                `form:"generation_source"`
	Thumbnail                *multipart.FileHeader `form:"thumbnail"`
}

type AdminCreateCaseResponse struct {
	CaseID           uuid.UUID `json:"case_id"`
	CaseVersionID    uuid.UUID `json:"case_version_id"`
	VersionNumber    int       `json:"version_number"`
	VersionLabel     string    `json:"version_label"`
	Title            string    `json:"title"`
	Slug             string    `json:"slug"`
	Status           string    `json:"status"`
	ThumbnailURL     *string   `json:"thumbnail_url"`
	ThumbnailPrompt  *string   `json:"thumbnail_prompt"`
	AIModel          *string   `json:"ai_model"`
	GenerationSource string    `json:"generation_source"`
	CreatedBy        uuid.UUID `json:"created_by"`
	CreatedAt        time.Time `json:"created_at"`
}

type AdminUpdateCaseRequest = AdminCreateCaseRequest

type AdminUpdateCaseResponse struct {
	CaseID                   uuid.UUID  `json:"case_id"`
	Title                    string     `json:"title"`
	Slug                     string     `json:"slug"`
	ShortDescription         string     `json:"short_description"`
	Theme                    string     `json:"theme"`
	ThemeOtherText           *string    `json:"theme_other_text"`
	CompetencyFocus          string     `json:"competency_focus"`
	DifficultyLevel          string     `json:"difficulty_level"`
	RiskLevel                string     `json:"risk_level"`
	EstimatedDurationMinutes int        `json:"estimated_duration_minutes"`
	AIModel                  *string    `json:"ai_model"`
	MinimumLevel             int        `json:"minimum_level"`
	MinimumReputation        float64    `json:"minimum_reputation"`
	UnlockRequirement        *string    `json:"unlock_requirement"`
	ThumbnailURL             *string    `json:"thumbnail_url"`
	ThumbnailPrompt          *string    `json:"thumbnail_prompt"`
	GenerationSource         string     `json:"generation_source"`
	Status                   string     `json:"status"`
	PublishedAt              *time.Time `json:"published_at"`
	CreatedBy                uuid.UUID  `json:"created_by"`
	CreatedAt                time.Time  `json:"created_at"`
	UpdatedAt                time.Time  `json:"updated_at"`
}

type AdminDeleteCaseResponse struct {
	CaseID uuid.UUID `json:"case_id"`
}

type AdminListCasesRequest struct {
	Search           string `form:"search"`
	Status           string `form:"status"`
	Theme            string `form:"theme"`
	CompetencyFocus  string `form:"competency_focus"`
	DifficultyLevel  string `form:"difficulty_level"`
	RiskLevel        string `form:"risk_level"`
	GenerationSource string `form:"generation_source"`
	Page             int    `form:"page"`
	Limit            int    `form:"limit"`
}

type AdminCaseListRow struct {
	CaseID                   uuid.UUID  `json:"case_id"`
	CurrentCaseVersionID     *uuid.UUID `json:"current_case_version_id"`
	VersionNumber            *int       `json:"version_number"`
	VersionLabel             *string    `json:"version_label"`
	Title                    string     `json:"title"`
	Slug                     string     `json:"slug"`
	ShortDescription         string     `json:"short_description"`
	Theme                    string     `json:"theme"`
	ThemeOtherText           *string    `json:"theme_other_text"`
	CompetencyFocus          string     `json:"competency_focus"`
	DifficultyLevel          string     `json:"difficulty_level"`
	RiskLevel                string     `json:"risk_level"`
	EstimatedDurationMinutes int        `json:"estimated_duration_minutes"`
	AIModel                  *string    `json:"ai_model"`
	MinimumLevel             int        `json:"minimum_level"`
	MinimumReputation        float64    `json:"minimum_reputation"`
	UnlockRequirement        *string    `json:"unlock_requirement"`
	ThumbnailURL             *string    `json:"thumbnail_url"`
	ThumbnailPrompt          *string    `json:"thumbnail_prompt"`
	GenerationSource         string     `json:"generation_source"`
	Status                   string     `json:"status"`
	QuestionCount            int        `json:"question_count"`
	EvidenceCount            int        `json:"evidence_count"`
	PublishedAt              *time.Time `json:"published_at"`
	CreatedBy                uuid.UUID  `json:"created_by"`
	CreatedAt                time.Time  `json:"created_at"`
	UpdatedAt                time.Time  `json:"updated_at"`
}

type AdminListCasesResponse struct {
	Cases      []AdminCaseListRow `json:"cases"`
	Pagination PaginationResponse `json:"pagination"`
}

type AdminCaseEvidenceListRow struct {
	CaseEvidenceID uuid.UUID `json:"case_evidence_id"`
	CaseVersionID  uuid.UUID `json:"case_version_id"`
	TemplateType   string    `json:"template_type"`
	Label          string    `json:"label"`
	IsCritical     bool      `json:"is_critical"`
	HasImage       bool      `json:"has_image"`
	SortOrder      int       `json:"sort_order"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type AdminCaseDetailResponse struct {
	Case      AdminCaseListRow           `json:"case"`
	Evidences []AdminCaseEvidenceListRow `json:"evidences"`
}

type AdminListCaseEvidencesResponse struct {
	CaseID        uuid.UUID                  `json:"case_id"`
	CaseVersionID *uuid.UUID                 `json:"case_version_id"`
	Total         int                        `json:"total"`
	Evidences     []AdminCaseEvidenceListRow `json:"evidences"`
}

type CaseLookupOptionResponse struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type AdminCaseLookupsResponse struct {
	Themes            []CaseLookupOptionResponse `json:"themes"`
	CompetencyFocuses []CaseLookupOptionResponse `json:"competency_focuses"`
	DifficultyLevels  []CaseLookupOptionResponse `json:"difficulty_levels"`
	RiskLevels        []CaseLookupOptionResponse `json:"risk_levels"`
	GenerationSources []CaseLookupOptionResponse `json:"generation_sources"`
}
