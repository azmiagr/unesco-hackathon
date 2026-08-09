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

type AdminCreateCaseRequest struct {
	Title                    string                `form:"title" binding:"required"`
	ShortDescription         string                `form:"short_description" binding:"required"`
	Theme                    string                `form:"theme" binding:"required"`
	ThemeOtherText           string                `form:"theme_other_text"`
	CompetencyFocus          string                `form:"competency_focus" binding:"required"`
	DifficultyLevel          string                `form:"difficulty_level" binding:"required"`
	RiskLevel                string                `form:"risk_level" binding:"required"`
	EstimatedDurationMinutes int                   `form:"estimated_duration_minutes" binding:"required"`
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
	GenerationSource string    `json:"generation_source"`
	CreatedBy        uuid.UUID `json:"created_by"`
	CreatedAt        time.Time `json:"created_at"`
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
