package service

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"regexp"
	"strings"

	"github.com/azmiagr/unesco-hackathon/entity"
	"github.com/azmiagr/unesco-hackathon/internal/repository"
	"github.com/azmiagr/unesco-hackathon/model"
	"github.com/azmiagr/unesco-hackathon/pkg/database/mariadb"
	appErrors "github.com/azmiagr/unesco-hackathon/pkg/errors"
	"github.com/azmiagr/unesco-hackathon/pkg/supabase"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const maxCaseThumbnailSize = 5 * 1024 * 1024

var allowedCaseThemes = map[string]bool{
	model.CaseThemeMisleadingHealthAdvice: true,
	model.CaseThemeChatbotHallucination:   true,
	model.CaseThemeClickbaitHeadline:      true,
	model.CaseThemeStatisticOutOfContext:  true,
	model.CaseThemeForumMisinformation:    true,
	model.CaseThemeViralConflictContent:   true,
	model.CaseThemeAlgorithmicEchoChamber: true,
	model.CaseThemeOther:                  true,
}

var allowedCaseCompetencyFocuses = map[string]bool{
	model.CaseCompetencyEvidenceEvaluation:    true,
	model.CaseCompetencyClaimAnalysis:         true,
	model.CaseCompetencyConfidenceCalibration: true,
	model.CaseCompetencyReasoning:             true,
	model.CaseCompetencySafetyJudgment:        true,
}

var allowedCaseDifficultyLevels = map[string]bool{
	model.CaseDifficultyLow:    true,
	model.CaseDifficultyMedium: true,
	model.CaseDifficultyHigh:   true,
}

var allowedCaseRiskLevels = map[string]bool{
	model.CaseRiskLow:    true,
	model.CaseRiskMedium: true,
	model.CaseRiskHigh:   true,
}

var allowedCaseGenerationSources = map[string]bool{
	model.CaseGenerationManual:     true,
	model.CaseGenerationAIAssisted: true,
}

type ICaseService interface {
	CreateCaseByAdmin(adminUserID uuid.UUID, req model.AdminCreateCaseRequest) (*model.AdminCreateCaseResponse, error)
	GetCaseLookups() (*model.AdminCaseLookupsResponse, error)
}

type CaseService struct {
	db              *gorm.DB
	caseRepo        repository.ICaseRepository
	caseVersionRepo repository.ICaseVersionRepository
	storage         supabase.Interface
}

func NewCaseService(
	caseRepo repository.ICaseRepository,
	caseVersionRepo repository.ICaseVersionRepository,
	storage supabase.Interface,
) ICaseService {
	return &CaseService{
		db:              mariadb.Connection,
		caseRepo:        caseRepo,
		caseVersionRepo: caseVersionRepo,
		storage:         storage,
	}
}

func (s *CaseService) CreateCaseByAdmin(adminUserID uuid.UUID, req model.AdminCreateCaseRequest) (*model.AdminCreateCaseResponse, error) {
	if adminUserID == uuid.Nil {
		return nil, appErrors.Unauthorized("unauthorized")
	}

	title := strings.TrimSpace(req.Title)
	shortDescription := strings.TrimSpace(req.ShortDescription)
	theme := strings.ToLower(strings.TrimSpace(req.Theme))
	themeOtherText := strings.TrimSpace(req.ThemeOtherText)
	competencyFocus := strings.ToLower(strings.TrimSpace(req.CompetencyFocus))
	difficultyLevel := strings.ToLower(strings.TrimSpace(req.DifficultyLevel))
	riskLevel := strings.ToLower(strings.TrimSpace(req.RiskLevel))
	generationSource := strings.ToLower(strings.TrimSpace(req.GenerationSource))
	thumbnailPrompt := strings.TrimSpace(req.ThumbnailPrompt)
	unlockRequirement := strings.TrimSpace(req.UnlockRequirement)

	if title == "" {
		return nil, appErrors.BadRequest("title is required")
	}

	if shortDescription == "" {
		return nil, appErrors.BadRequest("short description is required")
	}

	if !allowedCaseThemes[theme] {
		return nil, appErrors.BadRequest("invalid theme")
	}

	if theme == model.CaseThemeOther && themeOtherText == "" {
		return nil, appErrors.BadRequest("theme other text is required")
	}

	if theme != model.CaseThemeOther {
		themeOtherText = ""
	}

	if !allowedCaseCompetencyFocuses[competencyFocus] {
		return nil, appErrors.BadRequest("invalid competency focus")
	}

	if !allowedCaseDifficultyLevels[difficultyLevel] {
		return nil, appErrors.BadRequest("invalid difficulty level")
	}

	if !allowedCaseRiskLevels[riskLevel] {
		return nil, appErrors.BadRequest("invalid risk level")
	}

	if req.EstimatedDurationMinutes < 1 {
		return nil, appErrors.BadRequest("estimated duration minutes must be greater than 0")
	}

	minimumLevel := max(req.MinimumLevel, 1)

	minimumReputation := req.MinimumReputation
	if minimumReputation < 0 {
		return nil, appErrors.BadRequest("minimum reputation cannot be negative")
	}

	if generationSource == "" {
		generationSource = model.CaseGenerationManual
	}

	if !allowedCaseGenerationSources[generationSource] {
		return nil, appErrors.BadRequest("invalid generation source")
	}

	unlockRequirementValue, err := normalizeOptionalJSONString(unlockRequirement, "unlock requirement")
	if err != nil {
		return nil, err
	}

	thumbnailPromptValue := optionalString(thumbnailPrompt)
	themeOtherTextValue := optionalString(themeOtherText)

	thumbnailURL, err := s.uploadCaseThumbnail(req.Thumbnail)
	if err != nil {
		return nil, err
	}

	shouldDeleteThumbnail := thumbnailURL != nil
	defer func() {
		if shouldDeleteThumbnail && thumbnailURL != nil {
			_ = supabase.DeleteFileIfPresent(s.storage, *thumbnailURL)
		}
	}()

	caseID := uuid.New()
	slug, err := s.buildUniqueSlug(caseID, title)
	if err != nil {
		return nil, err
	}

	tx := s.db.Begin()
	defer tx.Rollback()

	caseEntity := &entity.Case{
		CaseID:                   caseID,
		CreatedBy:                adminUserID,
		Title:                    title,
		Slug:                     slug,
		ShortDescription:         shortDescription,
		Theme:                    theme,
		ThemeOtherText:           themeOtherTextValue,
		CompetencyFocus:          competencyFocus,
		DifficultyLevel:          difficultyLevel,
		RiskLevel:                riskLevel,
		EstimatedDurationMinutes: req.EstimatedDurationMinutes,
		MinimumLevel:             minimumLevel,
		MinimumReputation:        minimumReputation,
		UnlockRequirement:        unlockRequirementValue,
		ThumbnailURL:             thumbnailURL,
		ThumbnailPrompt:          thumbnailPromptValue,
		GenerationSource:         generationSource,
		Status:                   model.CaseStatusDraft,
	}

	err = s.caseRepo.CreateCase(tx, caseEntity)
	if err != nil {
		return nil, appErrors.InternalServer("failed to create case")
	}

	caseVersion := &entity.CaseVersion{
		CaseVersionID: uuid.New(),
		CaseID:        caseEntity.CaseID,
		VersionNumber: model.InitialCaseVersionNumber,
		Status:        model.CaseStatusDraft,
		Evidence:      model.EmptyJSONArray,
		Questions:     model.EmptyJSONArray,
		CreatedBy:     adminUserID,
	}

	err = s.caseVersionRepo.CreateCaseVersion(tx, caseVersion)
	if err != nil {
		return nil, appErrors.InternalServer("failed to create case version")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	shouldDeleteThumbnail = false

	return &model.AdminCreateCaseResponse{
		CaseID:           caseEntity.CaseID,
		CaseVersionID:    caseVersion.CaseVersionID,
		VersionNumber:    caseVersion.VersionNumber,
		VersionLabel:     formatCaseVersionLabel(caseVersion.VersionNumber),
		Title:            caseEntity.Title,
		Slug:             caseEntity.Slug,
		Status:           caseEntity.Status,
		ThumbnailURL:     caseEntity.ThumbnailURL,
		ThumbnailPrompt:  caseEntity.ThumbnailPrompt,
		GenerationSource: caseEntity.GenerationSource,
		CreatedBy:        caseEntity.CreatedBy,
		CreatedAt:        caseEntity.CreatedAt,
	}, nil
}

func (s *CaseService) GetCaseLookups() (*model.AdminCaseLookupsResponse, error) {
	return &model.AdminCaseLookupsResponse{
		Themes: []model.CaseLookupOptionResponse{
			{Value: model.CaseThemeMisleadingHealthAdvice, Label: "Saran kesehatan menyesatkan"},
			{Value: model.CaseThemeChatbotHallucination, Label: "Halusinasi chatbot"},
			{Value: model.CaseThemeClickbaitHeadline, Label: "Judul artikel manipulatif"},
			{Value: model.CaseThemeStatisticOutOfContext, Label: "Statistik di luar konteks"},
			{Value: model.CaseThemeForumMisinformation, Label: "Validasi informasi keliru di forum"},
			{Value: model.CaseThemeViralConflictContent, Label: "Konten viral yang memperkuat konflik"},
			{Value: model.CaseThemeAlgorithmicEchoChamber, Label: "Sistem rekomendasi/ruang gema"},
			{Value: model.CaseThemeOther, Label: "Lainnya"},
		},
		CompetencyFocuses: []model.CaseLookupOptionResponse{
			{Value: model.CaseCompetencyEvidenceEvaluation, Label: "Evaluasi bukti"},
			{Value: model.CaseCompetencyClaimAnalysis, Label: "Analisis klaim"},
			{Value: model.CaseCompetencyConfidenceCalibration, Label: "Kalibrasi keyakinan"},
			{Value: model.CaseCompetencyReasoning, Label: "Penalaran"},
			{Value: model.CaseCompetencySafetyJudgment, Label: "Penilaian keamanan/keputusan"},
		},
		DifficultyLevels: []model.CaseLookupOptionResponse{
			{Value: model.CaseDifficultyLow, Label: "Easy"},
			{Value: model.CaseDifficultyMedium, Label: "Medium"},
			{Value: model.CaseDifficultyHigh, Label: "Hard"},
		},
		RiskLevels: []model.CaseLookupOptionResponse{
			{Value: model.CaseRiskLow, Label: "Low"},
			{Value: model.CaseRiskMedium, Label: "Medium"},
			{Value: model.CaseRiskHigh, Label: "High"},
		},
		GenerationSources: []model.CaseLookupOptionResponse{
			{Value: model.CaseGenerationManual, Label: "Manual"},
			{Value: model.CaseGenerationAIAssisted, Label: "AI-Assisted"},
		},
	}, nil
}

func (s *CaseService) uploadCaseThumbnail(file *multipart.FileHeader) (*string, error) {
	if file == nil {
		return nil, nil
	}

	url, err := supabase.UploadOptionalImage(
		s.storage,
		file,
		maxCaseThumbnailSize,
		"thumbnail size exceeds 5MB limit",
	)
	if err != nil {
		return nil, appErrors.BadRequest("failed to upload thumbnail")
	}

	return &url, nil
}

func (s *CaseService) buildUniqueSlug(caseID uuid.UUID, title string) (string, error) {
	baseSlug := slugify(title)
	if baseSlug == "" {
		baseSlug = "case"
	}

	slug := baseSlug
	exists, err := s.caseRepo.CaseExists(s.db, model.GetCaseParam{
		Slug: slug,
	})
	if err != nil {
		return "", appErrors.InternalServer("failed to check case slug")
	}

	if exists {
		slug = fmt.Sprintf("%s-%s", baseSlug, caseID.String()[:8])
	}

	return slug, nil
}

func normalizeOptionalJSONString(raw string, fieldName string) (*string, error) {
	if raw == "" {
		return nil, nil
	}

	var payload any
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil, appErrors.BadRequest(fieldName + " must be valid json")
	}

	return &raw, nil
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}

	return &value
}

func slugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))

	nonAlphaNumeric := regexp.MustCompile(`[^a-z0-9]+`)
	value = nonAlphaNumeric.ReplaceAllString(value, "-")

	multipleDashes := regexp.MustCompile(`-+`)
	value = multipleDashes.ReplaceAllString(value, "-")

	return strings.Trim(value, "-")
}

func formatCaseVersionLabel(versionNumber int) string {
	return fmt.Sprintf("%d.0", versionNumber)
}
