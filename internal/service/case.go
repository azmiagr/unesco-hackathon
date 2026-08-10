package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"mime/multipart"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/azmiagr/unesco-hackathon/entity"
	"github.com/azmiagr/unesco-hackathon/internal/repository"
	"github.com/azmiagr/unesco-hackathon/model"
	"github.com/azmiagr/unesco-hackathon/pkg/database/mariadb"
	appErrors "github.com/azmiagr/unesco-hackathon/pkg/errors"
	"github.com/azmiagr/unesco-hackathon/pkg/helper"
	"github.com/azmiagr/unesco-hackathon/pkg/supabase"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	maxCaseThumbnailSize     = 5 * 1024 * 1024
	maxEvidenceImageSize     = 5 * 1024 * 1024
	maxSocialPostLabelLength = 150
	defaultAdminCaseLimit    = 10
	maxAdminCaseLimit        = 100
	maxEvidenceLabelLength   = 150
)

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

var allowedCaseStatuses = map[string]bool{
	model.CaseStatusDraft:     true,
	model.CaseStatusPublished: true,
	model.CaseStatusArchived:  true,
}

type ICaseService interface {
	CreateCaseByAdmin(adminUserID uuid.UUID, req model.AdminCreateCaseRequest) (*model.AdminCreateCaseResponse, error)
	UpdateCaseByAdmin(adminUserID uuid.UUID, caseID uuid.UUID, req model.AdminUpdateCaseRequest) (*model.AdminUpdateCaseResponse, error)
	HardDeleteCaseByAdmin(adminUserID uuid.UUID, caseID uuid.UUID) (*model.AdminDeleteCaseResponse, error)
	ListCasesByAdmin(req model.AdminListCasesRequest) (*model.AdminListCasesResponse, error)
	GetCaseDetailByAdmin(caseID uuid.UUID) (*model.AdminCaseDetailResponse, error)
	ListCaseEvidencesByAdmin(caseID uuid.UUID) (*model.AdminListCaseEvidencesResponse, error)
	GetCaseLookups() (*model.AdminCaseLookupsResponse, error)
	CreateSocialPostEvidenceByAdmin(adminUserID uuid.UUID, caseID uuid.UUID, caseVersionID uuid.UUID, req model.AdminCreateSocialPostEvidenceRequest) (*model.AdminCreateSocialPostEvidenceResponse, error)
	CreateArticleEvidenceByAdmin(adminUserID uuid.UUID, caseID uuid.UUID, caseVersionID uuid.UUID, req model.AdminCreateArticleEvidenceRequest) (*model.AdminCreateArticleEvidenceResponse, error)
	CreateBlogEvidenceByAdmin(adminUserID uuid.UUID, caseID uuid.UUID, caseVersionID uuid.UUID, req model.AdminCreateBlogEvidenceRequest) (*model.AdminCreateBlogEvidenceResponse, error)
	CreateForumThreadEvidenceByAdmin(adminUserID uuid.UUID, caseID uuid.UUID, caseVersionID uuid.UUID, req model.AdminCreateForumThreadEvidenceRequest) (*model.AdminCreateForumThreadEvidenceResponse, error)
	CreateChatTranscriptEvidenceByAdmin(adminUserID uuid.UUID, caseID uuid.UUID, caseVersionID uuid.UUID, req model.AdminCreateChatTranscriptEvidenceRequest) (*model.AdminCreateChatTranscriptEvidenceResponse, error)
	CreatePublicAnnouncementEvidenceByAdmin(adminUserID uuid.UUID, caseID uuid.UUID, caseVersionID uuid.UUID, req model.AdminCreatePublicAnnouncementEvidenceRequest) (*model.AdminCreatePublicAnnouncementEvidenceResponse, error)
	UpdateSocialPostEvidenceByAdmin(adminUserID uuid.UUID, caseID uuid.UUID, caseVersionID uuid.UUID, caseEvidenceID uuid.UUID, req model.AdminUpdateSocialPostEvidenceRequest) (*model.AdminUpdateSocialPostEvidenceResponse, error)
	UpdateArticleEvidenceByAdmin(adminUserID uuid.UUID, caseID uuid.UUID, caseVersionID uuid.UUID, caseEvidenceID uuid.UUID, req model.AdminUpdateArticleEvidenceRequest) (*model.AdminUpdateArticleEvidenceResponse, error)
	UpdateBlogEvidenceByAdmin(adminUserID uuid.UUID, caseID uuid.UUID, caseVersionID uuid.UUID, caseEvidenceID uuid.UUID, req model.AdminUpdateBlogEvidenceRequest) (*model.AdminUpdateBlogEvidenceResponse, error)
	UpdateForumThreadEvidenceByAdmin(adminUserID uuid.UUID, caseID uuid.UUID, caseVersionID uuid.UUID, caseEvidenceID uuid.UUID, req model.AdminUpdateForumThreadEvidenceRequest) (*model.AdminUpdateForumThreadEvidenceResponse, error)
	UpdateChatTranscriptEvidenceByAdmin(adminUserID uuid.UUID, caseID uuid.UUID, caseVersionID uuid.UUID, caseEvidenceID uuid.UUID, req model.AdminUpdateChatTranscriptEvidenceRequest) (*model.AdminUpdateChatTranscriptEvidenceResponse, error)
	UpdatePublicAnnouncementEvidenceByAdmin(adminUserID uuid.UUID, caseID uuid.UUID, caseVersionID uuid.UUID, caseEvidenceID uuid.UUID, req model.AdminUpdatePublicAnnouncementEvidenceRequest) (*model.AdminUpdatePublicAnnouncementEvidenceResponse, error)
	DeleteCaseEvidenceByAdmin(adminUserID uuid.UUID, caseID uuid.UUID, caseVersionID uuid.UUID, caseEvidenceID uuid.UUID) (*model.AdminDeleteCaseEvidenceResponse, error)
}

type CaseService struct {
	db               *gorm.DB
	caseRepo         repository.ICaseRepository
	caseVersionRepo  repository.ICaseVersionRepository
	caseEvidenceRepo repository.ICaseEvidenceRepository
	storage          supabase.Interface
}

func NewCaseService(
	caseRepo repository.ICaseRepository,
	caseVersionRepo repository.ICaseVersionRepository,
	caseEvidenceRepo repository.ICaseEvidenceRepository,
	storage supabase.Interface,
) ICaseService {
	return &CaseService{
		db:               mariadb.Connection,
		caseRepo:         caseRepo,
		caseVersionRepo:  caseVersionRepo,
		caseEvidenceRepo: caseEvidenceRepo,
		storage:          storage,
	}
}

func (s *CaseService) CreateCaseByAdmin(adminUserID uuid.UUID, req model.AdminCreateCaseRequest) (*model.AdminCreateCaseResponse, error) {
	if adminUserID == uuid.Nil {
		return nil, appErrors.Unauthorized("unauthorized")
	}

	theme := strings.ToLower(strings.TrimSpace(req.Theme))
	themeOtherText := strings.TrimSpace(req.ThemeOtherText)
	competencyFocus := strings.ToLower(strings.TrimSpace(req.CompetencyFocus))
	difficultyLevel := strings.ToLower(strings.TrimSpace(req.DifficultyLevel))
	riskLevel := strings.ToLower(strings.TrimSpace(req.RiskLevel))
	generationSource := strings.ToLower(strings.TrimSpace(req.GenerationSource))
	thumbnailPrompt := strings.TrimSpace(req.ThumbnailPrompt)
	unlockRequirement := strings.TrimSpace(req.UnlockRequirement)
	aiModel := strings.TrimSpace(req.AIModel)

	title, err := helper.RequireTrimmedString(req.Title, "title is required")
	if err != nil {
		return nil, err
	}

	shortDescription, err := helper.RequireTrimmedString(req.ShortDescription, "short description is required")
	if err != nil {
		return nil, err
	}

	if !allowedCaseThemes[theme] {
		return nil, appErrors.BadRequest("invalid theme")
	}

	if theme == model.CaseThemeOther && themeOtherText == "" {
		themeOtherText, err = helper.RequireTrimmedString(req.ThemeOtherText, "theme other text is required")
		if err != nil {
			return nil, err
		}
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
		AIModel:                  optionalString(aiModel),
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
		AIModel:          caseEntity.AIModel,
		GenerationSource: caseEntity.GenerationSource,
		CreatedBy:        caseEntity.CreatedBy,
		CreatedAt:        caseEntity.CreatedAt,
	}, nil
}

func (s *CaseService) UpdateCaseByAdmin(
	adminUserID uuid.UUID,
	caseID uuid.UUID,
	req model.AdminUpdateCaseRequest,
) (*model.AdminUpdateCaseResponse, error) {
	if adminUserID == uuid.Nil {
		return nil, appErrors.Unauthorized("unauthorized")
	}

	if caseID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid case id")
	}

	theme := strings.ToLower(strings.TrimSpace(req.Theme))
	themeOtherText := strings.TrimSpace(req.ThemeOtherText)
	competencyFocus := strings.ToLower(strings.TrimSpace(req.CompetencyFocus))
	difficultyLevel := strings.ToLower(strings.TrimSpace(req.DifficultyLevel))
	riskLevel := strings.ToLower(strings.TrimSpace(req.RiskLevel))
	generationSource := strings.ToLower(strings.TrimSpace(req.GenerationSource))
	thumbnailPrompt := strings.TrimSpace(req.ThumbnailPrompt)
	unlockRequirement := strings.TrimSpace(req.UnlockRequirement)
	aiModel := strings.TrimSpace(req.AIModel)

	title, err := helper.RequireTrimmedString(req.Title, "title is required")
	if err != nil {
		return nil, err
	}

	shortDescription, err := helper.RequireTrimmedString(req.ShortDescription, "short description is required")
	if err != nil {
		return nil, err
	}

	if !allowedCaseThemes[theme] {
		return nil, appErrors.BadRequest("invalid theme")
	}

	if theme == model.CaseThemeOther && themeOtherText == "" {
		themeOtherText, err = helper.RequireTrimmedString(req.ThemeOtherText, "theme other text is required")
		if err != nil {
			return nil, err
		}
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

	thumbnailURL, err := s.uploadCaseThumbnail(req.Thumbnail)
	if err != nil {
		return nil, err
	}

	shouldDeleteNewThumbnail := thumbnailURL != nil
	defer func() {
		if shouldDeleteNewThumbnail && thumbnailURL != nil {
			_ = supabase.DeleteFileIfPresent(s.storage, *thumbnailURL)
		}
	}()

	tx := s.db.Begin()
	defer tx.Rollback()

	caseEntity, err := s.caseRepo.GetCaseForUpdate(tx, caseID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("case not found")
		}
		return nil, appErrors.InternalServer("failed to get case")
	}

	slug, err := s.buildUniqueSlugForUpdate(caseEntity.CaseID, title)
	if err != nil {
		return nil, err
	}

	oldThumbnailURL := caseEntity.ThumbnailURL
	caseEntity.Title = title
	caseEntity.Slug = slug
	caseEntity.ShortDescription = shortDescription
	caseEntity.Theme = theme
	caseEntity.ThemeOtherText = optionalString(themeOtherText)
	caseEntity.CompetencyFocus = competencyFocus
	caseEntity.DifficultyLevel = difficultyLevel
	caseEntity.RiskLevel = riskLevel
	caseEntity.EstimatedDurationMinutes = req.EstimatedDurationMinutes
	caseEntity.AIModel = optionalString(aiModel)
	caseEntity.MinimumLevel = minimumLevel
	caseEntity.MinimumReputation = minimumReputation
	caseEntity.UnlockRequirement = unlockRequirementValue
	caseEntity.ThumbnailPrompt = optionalString(thumbnailPrompt)
	caseEntity.GenerationSource = generationSource
	if thumbnailURL != nil {
		caseEntity.ThumbnailURL = thumbnailURL
	}

	err = s.caseRepo.UpdateCase(tx, caseEntity)
	if err != nil {
		return nil, appErrors.InternalServer("failed to update case")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	shouldDeleteNewThumbnail = false
	if thumbnailURL != nil && oldThumbnailURL != nil && *oldThumbnailURL != *thumbnailURL {
		_ = supabase.DeleteFileIfPresent(s.storage, *oldThumbnailURL)
	}

	return mapAdminUpdateCaseResponse(caseEntity), nil
}

func (s *CaseService) HardDeleteCaseByAdmin(adminUserID uuid.UUID, caseID uuid.UUID) (*model.AdminDeleteCaseResponse, error) {
	if adminUserID == uuid.Nil {
		return nil, appErrors.Unauthorized("unauthorized")
	}

	if caseID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid case id")
	}

	tx := s.db.Begin()
	defer tx.Rollback()

	caseEntity, err := s.caseRepo.GetCaseForUpdate(tx, caseID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("case not found")
		}
		return nil, appErrors.InternalServer("failed to get case")
	}

	thumbnailURL := caseEntity.ThumbnailURL
	err = s.caseRepo.HardDeleteCase(tx, caseEntity.CaseID)
	if err != nil {
		return nil, appErrors.InternalServer("failed to delete case")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	if thumbnailURL != nil {
		_ = supabase.DeleteFileIfPresent(s.storage, *thumbnailURL)
	}

	return &model.AdminDeleteCaseResponse{
		CaseID: caseEntity.CaseID,
	}, nil
}

func (s *CaseService) ListCasesByAdmin(req model.AdminListCasesRequest) (*model.AdminListCasesResponse, error) {
	page := req.Page
	if page < 1 {
		page = 1
	}

	limit := req.Limit
	if limit < 1 {
		limit = defaultAdminCaseLimit
	}
	if limit > maxAdminCaseLimit {
		limit = maxAdminCaseLimit
	}

	status := strings.ToLower(strings.TrimSpace(req.Status))
	if status != "" && !allowedCaseStatuses[status] {
		return nil, appErrors.BadRequest("invalid status filter")
	}

	theme := strings.ToLower(strings.TrimSpace(req.Theme))
	if theme != "" && !allowedCaseThemes[theme] {
		return nil, appErrors.BadRequest("invalid theme filter")
	}

	competencyFocus := strings.ToLower(strings.TrimSpace(req.CompetencyFocus))
	if competencyFocus != "" && !allowedCaseCompetencyFocuses[competencyFocus] {
		return nil, appErrors.BadRequest("invalid competency focus filter")
	}

	difficultyLevel := strings.ToLower(strings.TrimSpace(req.DifficultyLevel))
	if difficultyLevel != "" && !allowedCaseDifficultyLevels[difficultyLevel] {
		return nil, appErrors.BadRequest("invalid difficulty level filter")
	}

	riskLevel := strings.ToLower(strings.TrimSpace(req.RiskLevel))
	if riskLevel != "" && !allowedCaseRiskLevels[riskLevel] {
		return nil, appErrors.BadRequest("invalid risk level filter")
	}

	generationSource := strings.ToLower(strings.TrimSpace(req.GenerationSource))
	if generationSource != "" && !allowedCaseGenerationSources[generationSource] {
		return nil, appErrors.BadRequest("invalid generation source filter")
	}

	param := model.AdminListCasesParam{
		Search:           strings.TrimSpace(req.Search),
		Status:           status,
		Theme:            theme,
		CompetencyFocus:  competencyFocus,
		DifficultyLevel:  difficultyLevel,
		RiskLevel:        riskLevel,
		GenerationSource: generationSource,
		Limit:            limit,
		Offset:           (page - 1) * limit,
	}

	cases, total, err := s.caseRepo.ListAdminCases(s.db, param)
	if err != nil {
		return nil, appErrors.InternalServer("failed to list cases")
	}

	for i := range cases {
		cases[i].VersionLabel = optionalVersionLabel(cases[i].VersionNumber)
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}

	return &model.AdminListCasesResponse{
		Cases: cases,
		Pagination: model.PaginationResponse{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

func (s *CaseService) GetCaseDetailByAdmin(caseID uuid.UUID) (*model.AdminCaseDetailResponse, error) {
	if caseID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid case id")
	}

	caseDetail, err := s.caseRepo.GetAdminCaseDetail(s.db, caseID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("case not found")
		}
		return nil, appErrors.InternalServer("failed to get case detail")
	}

	caseDetail.VersionLabel = optionalVersionLabel(caseDetail.VersionNumber)

	evidences := []model.AdminCaseEvidenceListRow{}
	if caseDetail.CurrentCaseVersionID != nil {
		evidences, err = s.caseEvidenceRepo.ListAdminCaseEvidenceRows(s.db, *caseDetail.CurrentCaseVersionID)
		if err != nil {
			return nil, appErrors.InternalServer("failed to list case evidences")
		}
	}

	return &model.AdminCaseDetailResponse{
		Case:      *caseDetail,
		Evidences: evidences,
	}, nil
}

func (s *CaseService) ListCaseEvidencesByAdmin(caseID uuid.UUID) (*model.AdminListCaseEvidencesResponse, error) {
	if caseID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid case id")
	}

	caseDetail, err := s.caseRepo.GetAdminCaseDetail(s.db, caseID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("case not found")
		}
		return nil, appErrors.InternalServer("failed to get case detail")
	}

	evidences := []model.AdminCaseEvidenceListRow{}
	if caseDetail.CurrentCaseVersionID != nil {
		evidences, err = s.caseEvidenceRepo.ListAdminCaseEvidenceRows(s.db, *caseDetail.CurrentCaseVersionID)
		if err != nil {
			return nil, appErrors.InternalServer("failed to list case evidences")
		}
	}

	return &model.AdminListCaseEvidencesResponse{
		CaseID:        caseDetail.CaseID,
		CaseVersionID: caseDetail.CurrentCaseVersionID,
		Total:         len(evidences),
		Evidences:     evidences,
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

func (s *CaseService) CreateSocialPostEvidenceByAdmin(
	adminUserID uuid.UUID,
	caseID uuid.UUID,
	caseVersionID uuid.UUID,
	req model.AdminCreateSocialPostEvidenceRequest,
) (*model.AdminCreateSocialPostEvidenceResponse, error) {
	if adminUserID == uuid.Nil {
		return nil, appErrors.Unauthorized("unauthorized")
	}

	if caseID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid case id")
	}

	if caseVersionID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid case version id")
	}

	label, err := helper.RequireTrimmedString(req.Label, "label is required")
	if err != nil {
		return nil, err
	}

	authorName, err := helper.RequireTrimmedString(req.AuthorName, "author name is required")
	if err != nil {
		return nil, err
	}

	authorHandle, err := helper.RequireTrimmedString(req.AuthorHandle, "author handle is required")
	if err != nil {
		return nil, err
	}

	platform, err := helper.RequireTrimmedString(req.Platform, "platform is required")
	if err != nil {
		return nil, err
	}

	postText, err := helper.RequireTrimmedString(req.PostText, "post text is required")
	if err != nil {
		return nil, err
	}

	imagePrompt := strings.TrimSpace(req.ImagePrompt)

	if len(label) > maxSocialPostLabelLength {
		return nil, appErrors.BadRequest("label is too long")
	}

	credibilityTags, credibilityTagsJSON, err := normalizeCredibilityTags(req.CredibilityTags)
	if err != nil {
		return nil, err
	}

	timestamp, err := parseEvidenceTimestamp(req.Timestamp)
	if err != nil {
		return nil, err
	}

	if req.LikesCount < 0 || req.SharesCount < 0 || req.CommentsCount < 0 {
		return nil, appErrors.BadRequest("engagement counts cannot be negative")
	}

	imageURL, err := s.uploadEvidenceImage(req.Image)
	if err != nil {
		return nil, err
	}

	shouldDeleteImage := imageURL != nil
	defer func() {
		if shouldDeleteImage && imageURL != nil {
			_ = supabase.DeleteFileIfPresent(s.storage, *imageURL)
		}
	}()

	tx := s.db.Begin()
	defer tx.Rollback()

	caseVersion, err := s.caseVersionRepo.GetCaseVersionForUpdate(tx, caseVersionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("case version not found")
		}
		return nil, appErrors.InternalServer("failed to get case version")
	}

	if caseVersion.CaseID != caseID {
		return nil, appErrors.NotFound("case version not found")
	}

	if caseVersion.Status != model.CaseStatusDraft {
		return nil, appErrors.Conflict("case version is not editable")
	}

	evidence := &entity.CaseEvidence{
		CaseEvidenceID:  uuid.New(),
		CaseVersionID:   caseVersion.CaseVersionID,
		TemplateType:    model.CaseEvidenceTemplateSocialPost,
		Label:           label,
		CredibilityTags: credibilityTagsJSON,
		IsCritical:      req.IsCritical,
		SortOrder:       req.SortOrder,
	}

	socialPost := &entity.CaseEvidenceSocialPost{
		CaseEvidenceID:    evidence.CaseEvidenceID,
		AuthorName:        authorName,
		AuthorHandle:      authorHandle,
		Platform:          platform,
		PostText:          postText,
		Timestamp:         timestamp,
		LikesCount:        req.LikesCount,
		SharesCount:       req.SharesCount,
		CommentsCount:     req.CommentsCount,
		IsVerifiedAccount: req.IsVerifiedAccount,
		ImagePrompt:       optionalString(imagePrompt),
		ImageURL:          imageURL,
	}

	err = s.caseEvidenceRepo.CreateSocialPostEvidence(tx, evidence, socialPost)
	if err != nil {
		return nil, appErrors.InternalServer("failed to create social post evidence")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	shouldDeleteImage = false

	return &model.AdminCreateSocialPostEvidenceResponse{
		Evidence: mapSocialPostEvidenceResponse(evidence, socialPost, credibilityTags),
	}, nil
}

func (s *CaseService) CreateArticleEvidenceByAdmin(
	adminUserID uuid.UUID,
	caseID uuid.UUID,
	caseVersionID uuid.UUID,
	req model.AdminCreateArticleEvidenceRequest,
) (*model.AdminCreateArticleEvidenceResponse, error) {
	if adminUserID == uuid.Nil {
		return nil, appErrors.Unauthorized("unauthorized")
	}

	if caseID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid case id")
	}

	if caseVersionID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid case version id")
	}

	label, err := helper.RequireTrimmedString(req.Label, "label is required")
	if err != nil {
		return nil, err
	}

	headline, err := helper.RequireTrimmedString(req.Headline, "headline is required")
	if err != nil {
		return nil, err
	}

	sourceName, err := helper.RequireTrimmedString(req.SourceName, "source name is required")
	if err != nil {
		return nil, err
	}

	authorName, err := helper.RequireTrimmedString(req.AuthorName, "author name is required")
	if err != nil {
		return nil, err
	}

	bodyText, err := helper.RequireTrimmedString(req.BodyText, "body text is required")
	if err != nil {
		return nil, err
	}

	imagePrompt := strings.TrimSpace(req.ImagePrompt)

	if len(label) > maxEvidenceLabelLength {
		return nil, appErrors.BadRequest("label is too long")
	}

	credibilityTags, credibilityTagsJSON, err := normalizeCredibilityTags(req.CredibilityTags)
	if err != nil {
		return nil, err
	}

	publishDate, err := parseEvidencePublishDate(req.PublishDate)
	if err != nil {
		return nil, err
	}

	articleURL, err := normalizeOptionalURL(req.URL)
	if err != nil {
		return nil, err
	}

	imageURL, err := s.uploadEvidenceImage(req.Image)
	if err != nil {
		return nil, err
	}

	shouldDeleteImage := imageURL != nil
	defer func() {
		if shouldDeleteImage && imageURL != nil {
			_ = supabase.DeleteFileIfPresent(s.storage, *imageURL)
		}
	}()

	tx := s.db.Begin()
	defer tx.Rollback()

	caseVersion, err := s.caseVersionRepo.GetCaseVersionForUpdate(tx, caseVersionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("case version not found")
		}
		return nil, appErrors.InternalServer("failed to get case version")
	}

	if caseVersion.CaseID != caseID {
		return nil, appErrors.NotFound("case version not found")
	}

	if caseVersion.Status != model.CaseStatusDraft {
		return nil, appErrors.Conflict("case version is not editable")
	}

	evidence := &entity.CaseEvidence{
		CaseEvidenceID:  uuid.New(),
		CaseVersionID:   caseVersion.CaseVersionID,
		TemplateType:    model.CaseEvidenceTemplateArticle,
		Label:           label,
		CredibilityTags: credibilityTagsJSON,
		IsCritical:      req.IsCritical,
		SortOrder:       req.SortOrder,
	}

	article := &entity.CaseEvidenceArticle{
		CaseEvidenceID: evidence.CaseEvidenceID,
		Headline:       headline,
		SourceName:     sourceName,
		AuthorName:     authorName,
		PublishDate:    publishDate,
		URL:            articleURL,
		BodyText:       bodyText,
		ImagePrompt:    optionalString(imagePrompt),
		ImageURL:       imageURL,
	}

	err = s.caseEvidenceRepo.CreateArticleEvidence(tx, evidence, article)
	if err != nil {
		return nil, appErrors.InternalServer("failed to create article evidence")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	shouldDeleteImage = false

	return &model.AdminCreateArticleEvidenceResponse{
		Evidence: mapArticleEvidenceResponse(evidence, article, credibilityTags),
	}, nil
}

func (s *CaseService) CreateBlogEvidenceByAdmin(
	adminUserID uuid.UUID,
	caseID uuid.UUID,
	caseVersionID uuid.UUID,
	req model.AdminCreateBlogEvidenceRequest,
) (*model.AdminCreateBlogEvidenceResponse, error) {
	if adminUserID == uuid.Nil {
		return nil, appErrors.Unauthorized("unauthorized")
	}

	if caseID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid case id")
	}

	if caseVersionID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid case version id")
	}

	label, err := helper.RequireTrimmedString(req.Label, "label is required")
	if err != nil {
		return nil, err
	}

	title, err := helper.RequireTrimmedString(req.Title, "title is required")
	if err != nil {
		return nil, err
	}

	authorName, err := helper.RequireTrimmedString(req.AuthorName, "author name is required")
	if err != nil {
		return nil, err
	}

	blogName, err := helper.RequireTrimmedString(req.BlogName, "blog name is required")
	if err != nil {
		return nil, err
	}

	bodyText, err := helper.RequireTrimmedString(req.BodyText, "body text is required")
	if err != nil {
		return nil, err
	}

	if len(label) > maxEvidenceLabelLength {
		return nil, appErrors.BadRequest("label is too long")
	}

	credibilityTags, credibilityTagsJSON, err := normalizeCredibilityTags(req.CredibilityTags)
	if err != nil {
		return nil, err
	}

	publishDate, err := parseEvidencePublishDate(req.PublishDate)
	if err != nil {
		return nil, err
	}

	tx := s.db.Begin()
	defer tx.Rollback()

	caseVersion, err := s.caseVersionRepo.GetCaseVersionForUpdate(tx, caseVersionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("case version not found")
		}
		return nil, appErrors.InternalServer("failed to get case version")
	}

	if caseVersion.CaseID != caseID {
		return nil, appErrors.NotFound("case version not found")
	}

	if caseVersion.Status != model.CaseStatusDraft {
		return nil, appErrors.Conflict("case version is not editable")
	}

	evidence := &entity.CaseEvidence{
		CaseEvidenceID:  uuid.New(),
		CaseVersionID:   caseVersion.CaseVersionID,
		TemplateType:    model.CaseEvidenceTemplateBlog,
		Label:           label,
		CredibilityTags: credibilityTagsJSON,
		IsCritical:      req.IsCritical,
		SortOrder:       req.SortOrder,
	}

	blog := &entity.CaseEvidenceBlog{
		CaseEvidenceID: evidence.CaseEvidenceID,
		Title:          title,
		AuthorName:     authorName,
		BlogName:       blogName,
		PublishDate:    publishDate,
		BodyText:       bodyText,
	}

	err = s.caseEvidenceRepo.CreateBlogEvidence(tx, evidence, blog)
	if err != nil {
		return nil, appErrors.InternalServer("failed to create blog evidence")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	return &model.AdminCreateBlogEvidenceResponse{
		Evidence: mapBlogEvidenceResponse(evidence, blog, credibilityTags),
	}, nil
}

func (s *CaseService) CreateForumThreadEvidenceByAdmin(
	adminUserID uuid.UUID,
	caseID uuid.UUID,
	caseVersionID uuid.UUID,
	req model.AdminCreateForumThreadEvidenceRequest,
) (*model.AdminCreateForumThreadEvidenceResponse, error) {
	if adminUserID == uuid.Nil {
		return nil, appErrors.Unauthorized("unauthorized")
	}

	if caseID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid case id")
	}

	if caseVersionID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid case version id")
	}

	label, err := helper.RequireTrimmedString(req.Label, "label is required")
	if err != nil {
		return nil, err
	}

	threadTitle, err := helper.RequireTrimmedString(req.ThreadTitle, "thread title is required")
	if err != nil {
		return nil, err
	}

	forumName, err := helper.RequireTrimmedString(req.ForumName, "forum name is required")
	if err != nil {
		return nil, err
	}

	if len(label) > maxEvidenceLabelLength {
		return nil, appErrors.BadRequest("label is too long")
	}

	credibilityTags, credibilityTagsJSON, err := normalizeCredibilityTagItems(req.CredibilityTags)
	if err != nil {
		return nil, err
	}

	if len(req.Posts) == 0 {
		return nil, appErrors.BadRequest("posts are required")
	}

	evidenceID := uuid.New()
	posts := make([]entity.CaseEvidenceForumThreadPost, 0, len(req.Posts))

	for i, postReq := range req.Posts {
		authorName, err := helper.RequireTrimmedString(postReq.AuthorName, "post author name is required")
		if err != nil {
			return nil, err
		}

		text, err := helper.RequireTrimmedString(postReq.Text, "post text is required")
		if err != nil {
			return nil, err
		}

		timestamp, err := parseEvidenceTimestamp(postReq.Timestamp)
		if err != nil {
			return nil, err
		}

		if postReq.UpvoteCount < 0 {
			return nil, appErrors.BadRequest("upvote count cannot be negative")
		}

		posts = append(posts, entity.CaseEvidenceForumThreadPost{
			CaseEvidenceForumThreadPostID: uuid.New(),
			CaseEvidenceID:                evidenceID,
			AuthorName:                    authorName,
			Text:                          text,
			Timestamp:                     timestamp,
			UpvoteCount:                   postReq.UpvoteCount,
			SortOrder:                     i + 1,
		})
	}

	tx := s.db.Begin()
	defer tx.Rollback()

	caseVersion, err := s.caseVersionRepo.GetCaseVersionForUpdate(tx, caseVersionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("case version not found")
		}
		return nil, appErrors.InternalServer("failed to get case version")
	}

	if caseVersion.CaseID != caseID {
		return nil, appErrors.NotFound("case version not found")
	}

	if caseVersion.Status != model.CaseStatusDraft {
		return nil, appErrors.Conflict("case version is not editable")
	}

	evidence := &entity.CaseEvidence{
		CaseEvidenceID:  evidenceID,
		CaseVersionID:   caseVersion.CaseVersionID,
		TemplateType:    model.CaseEvidenceTemplateForumThread,
		Label:           label,
		CredibilityTags: credibilityTagsJSON,
		IsCritical:      req.IsCritical,
		SortOrder:       req.SortOrder,
	}

	forumThread := &entity.CaseEvidenceForumThread{
		CaseEvidenceID: evidence.CaseEvidenceID,
		ThreadTitle:    threadTitle,
		ForumName:      forumName,
	}

	err = s.caseEvidenceRepo.CreateForumThreadEvidence(tx, evidence, forumThread, posts)
	if err != nil {
		return nil, appErrors.InternalServer("failed to create forum thread evidence")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	return &model.AdminCreateForumThreadEvidenceResponse{
		Evidence: mapForumThreadEvidenceResponse(evidence, forumThread, posts, credibilityTags),
	}, nil
}

func (s *CaseService) CreateChatTranscriptEvidenceByAdmin(
	adminUserID uuid.UUID,
	caseID uuid.UUID,
	caseVersionID uuid.UUID,
	req model.AdminCreateChatTranscriptEvidenceRequest,
) (*model.AdminCreateChatTranscriptEvidenceResponse, error) {
	if adminUserID == uuid.Nil {
		return nil, appErrors.Unauthorized("unauthorized")
	}

	if caseID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid case id")
	}

	if caseVersionID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid case version id")
	}

	label, err := helper.RequireTrimmedString(req.Label, "label is required")
	if err != nil {
		return nil, err
	}

	if len(label) > maxEvidenceLabelLength {
		return nil, appErrors.BadRequest("label is too long")
	}

	credibilityTags, credibilityTagsJSON, err := normalizeCredibilityTagItems(req.CredibilityTags)
	if err != nil {
		return nil, err
	}

	if len(req.Participants) == 0 {
		return nil, appErrors.BadRequest("participants are required")
	}

	if len(req.Messages) == 0 {
		return nil, appErrors.BadRequest("messages are required")
	}

	evidenceID := uuid.New()

	participantEntities := make([]entity.CaseEvidenceChatTranscriptParticipant, 0, len(req.Participants))
	participantSet := map[string]bool{}

	for i, participant := range req.Participants {
		name, err := helper.RequireTrimmedString(participant, "participant name is required")
		if err != nil {
			return nil, err
		}

		if participantSet[name] {
			return nil, appErrors.BadRequest("participants cannot contain duplicates")
		}

		participantSet[name] = true
		participantEntities = append(participantEntities, entity.CaseEvidenceChatTranscriptParticipant{
			CaseEvidenceChatTranscriptParticipantID: uuid.New(),
			CaseEvidenceID:                          evidenceID,
			Name:                                    name,
			SortOrder:                               i + 1,
		})
	}

	messageEntities := make([]entity.CaseEvidenceChatTranscriptMessage, 0, len(req.Messages))

	for i, messageReq := range req.Messages {
		sender, err := helper.RequireTrimmedString(messageReq.Sender, "message sender is required")
		if err != nil {
			return nil, err
		}

		if !participantSet[sender] {
			return nil, appErrors.BadRequest("message sender must be one of participants")
		}

		text, err := helper.RequireTrimmedString(messageReq.Text, "message text is required")
		if err != nil {
			return nil, err
		}

		timestamp, err := parseEvidenceTimestamp(messageReq.Timestamp)
		if err != nil {
			return nil, err
		}

		messageEntities = append(messageEntities, entity.CaseEvidenceChatTranscriptMessage{
			CaseEvidenceChatTranscriptMessageID: uuid.New(),
			CaseEvidenceID:                      evidenceID,
			Sender:                              sender,
			Text:                                text,
			Timestamp:                           timestamp,
			SortOrder:                           i + 1,
		})
	}

	tx := s.db.Begin()
	defer tx.Rollback()

	caseVersion, err := s.caseVersionRepo.GetCaseVersionForUpdate(tx, caseVersionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("case version not found")
		}
		return nil, appErrors.InternalServer("failed to get case version")
	}

	if caseVersion.CaseID != caseID {
		return nil, appErrors.NotFound("case version not found")
	}

	if caseVersion.Status != model.CaseStatusDraft {
		return nil, appErrors.Conflict("case version is not editable")
	}

	evidence := &entity.CaseEvidence{
		CaseEvidenceID:  evidenceID,
		CaseVersionID:   caseVersion.CaseVersionID,
		TemplateType:    model.CaseEvidenceTemplateChatTranscript,
		Label:           label,
		CredibilityTags: credibilityTagsJSON,
		IsCritical:      req.IsCritical,
		SortOrder:       req.SortOrder,
	}

	chatTranscript := &entity.CaseEvidenceChatTranscript{
		CaseEvidenceID: evidence.CaseEvidenceID,
	}

	err = s.caseEvidenceRepo.CreateChatTranscriptEvidence(tx, evidence, chatTranscript, participantEntities, messageEntities)
	if err != nil {
		return nil, appErrors.InternalServer("failed to create chat transcript evidence")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	return &model.AdminCreateChatTranscriptEvidenceResponse{
		Evidence: mapChatTranscriptEvidenceResponse(evidence, participantEntities, messageEntities, credibilityTags),
	}, nil
}

func (s *CaseService) CreatePublicAnnouncementEvidenceByAdmin(
	adminUserID uuid.UUID,
	caseID uuid.UUID,
	caseVersionID uuid.UUID,
	req model.AdminCreatePublicAnnouncementEvidenceRequest,
) (*model.AdminCreatePublicAnnouncementEvidenceResponse, error) {
	if adminUserID == uuid.Nil {
		return nil, appErrors.Unauthorized("unauthorized")
	}

	if caseID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid case id")
	}

	if caseVersionID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid case version id")
	}

	label, err := helper.RequireTrimmedString(req.Label, "label is required")
	if err != nil {
		return nil, err
	}

	issuingBody, err := helper.RequireTrimmedString(req.IssuingBody, "issuing body is required")
	if err != nil {
		return nil, err
	}

	title, err := helper.RequireTrimmedString(req.Title, "title is required")
	if err != nil {
		return nil, err
	}

	bodyText, err := helper.RequireTrimmedString(req.BodyText, "body text is required")
	if err != nil {
		return nil, err
	}

	if len(label) > maxEvidenceLabelLength {
		return nil, appErrors.BadRequest("label is too long")
	}

	credibilityTags, credibilityTagsJSON, err := normalizeCredibilityTagItems(req.CredibilityTags)
	if err != nil {
		return nil, err
	}

	announcementDate, err := parseEvidencePublishDate(req.Date)
	if err != nil {
		return nil, err
	}

	tx := s.db.Begin()
	defer tx.Rollback()

	caseVersion, err := s.caseVersionRepo.GetCaseVersionForUpdate(tx, caseVersionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("case version not found")
		}
		return nil, appErrors.InternalServer("failed to get case version")
	}

	if caseVersion.CaseID != caseID {
		return nil, appErrors.NotFound("case version not found")
	}

	if caseVersion.Status != model.CaseStatusDraft {
		return nil, appErrors.Conflict("case version is not editable")
	}

	evidence := &entity.CaseEvidence{
		CaseEvidenceID:  uuid.New(),
		CaseVersionID:   caseVersion.CaseVersionID,
		TemplateType:    model.CaseEvidenceTemplatePublicAnnouncement,
		Label:           label,
		CredibilityTags: credibilityTagsJSON,
		IsCritical:      req.IsCritical,
		SortOrder:       req.SortOrder,
	}

	publicAnnouncement := &entity.CaseEvidencePublicAnnouncement{
		CaseEvidenceID: evidence.CaseEvidenceID,
		IssuingBody:    issuingBody,
		Title:          title,
		Date:           announcementDate,
		BodyText:       bodyText,
	}

	err = s.caseEvidenceRepo.CreatePublicAnnouncementEvidence(tx, evidence, publicAnnouncement)
	if err != nil {
		return nil, appErrors.InternalServer("failed to create public announcement evidence")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	return &model.AdminCreatePublicAnnouncementEvidenceResponse{
		Evidence: mapPublicAnnouncementEvidenceResponse(evidence, publicAnnouncement, credibilityTags),
	}, nil
}

func (s *CaseService) UpdateSocialPostEvidenceByAdmin(
	adminUserID uuid.UUID,
	caseID uuid.UUID,
	caseVersionID uuid.UUID,
	caseEvidenceID uuid.UUID,
	req model.AdminUpdateSocialPostEvidenceRequest,
) (*model.AdminUpdateSocialPostEvidenceResponse, error) {
	err := validateAdminEvidenceIDs(adminUserID, caseID, caseVersionID, caseEvidenceID)
	if err != nil {
		return nil, err
	}

	label, err := helper.RequireTrimmedString(req.Label, "label is required")
	if err != nil {
		return nil, err
	}

	authorName, err := helper.RequireTrimmedString(req.AuthorName, "author name is required")
	if err != nil {
		return nil, err
	}

	authorHandle, err := helper.RequireTrimmedString(req.AuthorHandle, "author handle is required")
	if err != nil {
		return nil, err
	}

	platform, err := helper.RequireTrimmedString(req.Platform, "platform is required")
	if err != nil {
		return nil, err
	}

	postText, err := helper.RequireTrimmedString(req.PostText, "post text is required")
	if err != nil {
		return nil, err
	}

	if len(label) > maxEvidenceLabelLength {
		return nil, appErrors.BadRequest("label is too long")
	}

	credibilityTags, credibilityTagsJSON, err := normalizeCredibilityTags(req.CredibilityTags)
	if err != nil {
		return nil, err
	}

	timestamp, err := parseEvidenceTimestamp(req.Timestamp)
	if err != nil {
		return nil, err
	}

	imagePrompt := strings.TrimSpace(req.ImagePrompt)
	newImageURL, err := s.uploadEvidenceImage(req.Image)
	if err != nil {
		return nil, err
	}

	shouldDeleteNewImage := newImageURL != nil
	defer func() {
		if shouldDeleteNewImage && newImageURL != nil {
			_ = supabase.DeleteFileIfPresent(s.storage, *newImageURL)
		}
	}()

	tx := s.db.Begin()
	defer tx.Rollback()

	evidence, err := s.getEditableEvidenceByAdmin(tx, caseID, caseVersionID, caseEvidenceID, model.CaseEvidenceTemplateSocialPost)
	if err != nil {
		return nil, err
	}

	socialPost := evidence.SocialPost
	if socialPost == nil {
		socialPost = &entity.CaseEvidenceSocialPost{
			CaseEvidenceID: evidence.CaseEvidenceID,
		}
	}

	oldImageURL := socialPost.ImageURL
	evidence.Label = label
	evidence.CredibilityTags = credibilityTagsJSON
	evidence.IsCritical = req.IsCritical
	evidence.SortOrder = req.SortOrder

	socialPost.AuthorName = authorName
	socialPost.AuthorHandle = authorHandle
	socialPost.Platform = platform
	socialPost.PostText = postText
	socialPost.Timestamp = timestamp
	socialPost.LikesCount = req.LikesCount
	socialPost.SharesCount = req.SharesCount
	socialPost.CommentsCount = req.CommentsCount
	socialPost.IsVerifiedAccount = req.IsVerifiedAccount
	socialPost.ImagePrompt = optionalString(imagePrompt)
	if newImageURL != nil {
		socialPost.ImageURL = newImageURL
	}

	err = s.caseEvidenceRepo.UpdateCaseEvidence(tx, evidence)
	if err != nil {
		return nil, appErrors.InternalServer("failed to update case evidence")
	}

	err = s.caseEvidenceRepo.UpdateSocialPostEvidence(tx, socialPost)
	if err != nil {
		return nil, appErrors.InternalServer("failed to update social post evidence")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	shouldDeleteNewImage = false
	if newImageURL != nil && oldImageURL != nil && *oldImageURL != *newImageURL {
		_ = supabase.DeleteFileIfPresent(s.storage, *oldImageURL)
	}

	return &model.AdminUpdateSocialPostEvidenceResponse{
		Evidence: mapSocialPostEvidenceResponse(evidence, socialPost, credibilityTags),
	}, nil
}

func (s *CaseService) UpdateArticleEvidenceByAdmin(
	adminUserID uuid.UUID,
	caseID uuid.UUID,
	caseVersionID uuid.UUID,
	caseEvidenceID uuid.UUID,
	req model.AdminUpdateArticleEvidenceRequest,
) (*model.AdminUpdateArticleEvidenceResponse, error) {
	err := validateAdminEvidenceIDs(adminUserID, caseID, caseVersionID, caseEvidenceID)
	if err != nil {
		return nil, err
	}

	label, err := helper.RequireTrimmedString(req.Label, "label is required")
	if err != nil {
		return nil, err
	}

	headline, err := helper.RequireTrimmedString(req.Headline, "headline is required")
	if err != nil {
		return nil, err
	}

	sourceName, err := helper.RequireTrimmedString(req.SourceName, "source name is required")
	if err != nil {
		return nil, err
	}

	authorName, err := helper.RequireTrimmedString(req.AuthorName, "author name is required")
	if err != nil {
		return nil, err
	}

	bodyText, err := helper.RequireTrimmedString(req.BodyText, "body text is required")
	if err != nil {
		return nil, err
	}

	if len(label) > maxEvidenceLabelLength {
		return nil, appErrors.BadRequest("label is too long")
	}

	credibilityTags, credibilityTagsJSON, err := normalizeCredibilityTags(req.CredibilityTags)
	if err != nil {
		return nil, err
	}

	publishDate, err := parseEvidencePublishDate(req.PublishDate)
	if err != nil {
		return nil, err
	}

	urlValue, err := normalizeOptionalURL(req.URL)
	if err != nil {
		return nil, err
	}

	imagePrompt := strings.TrimSpace(req.ImagePrompt)
	newImageURL, err := s.uploadEvidenceImage(req.Image)
	if err != nil {
		return nil, err
	}

	shouldDeleteNewImage := newImageURL != nil
	defer func() {
		if shouldDeleteNewImage && newImageURL != nil {
			_ = supabase.DeleteFileIfPresent(s.storage, *newImageURL)
		}
	}()

	tx := s.db.Begin()
	defer tx.Rollback()

	evidence, err := s.getEditableEvidenceByAdmin(tx, caseID, caseVersionID, caseEvidenceID, model.CaseEvidenceTemplateArticle)
	if err != nil {
		return nil, err
	}

	article := evidence.Article
	if article == nil {
		article = &entity.CaseEvidenceArticle{CaseEvidenceID: evidence.CaseEvidenceID}
	}

	oldImageURL := article.ImageURL
	evidence.Label = label
	evidence.CredibilityTags = credibilityTagsJSON
	evidence.IsCritical = req.IsCritical
	evidence.SortOrder = req.SortOrder

	article.Headline = headline
	article.SourceName = sourceName
	article.AuthorName = authorName
	article.PublishDate = publishDate
	article.URL = urlValue
	article.BodyText = bodyText
	article.ImagePrompt = optionalString(imagePrompt)
	if newImageURL != nil {
		article.ImageURL = newImageURL
	}

	err = s.caseEvidenceRepo.UpdateCaseEvidence(tx, evidence)
	if err != nil {
		return nil, appErrors.InternalServer("failed to update case evidence")
	}

	err = s.caseEvidenceRepo.UpdateArticleEvidence(tx, article)
	if err != nil {
		return nil, appErrors.InternalServer("failed to update article evidence")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	shouldDeleteNewImage = false
	if newImageURL != nil && oldImageURL != nil && *oldImageURL != *newImageURL {
		_ = supabase.DeleteFileIfPresent(s.storage, *oldImageURL)
	}

	return &model.AdminUpdateArticleEvidenceResponse{
		Evidence: mapArticleEvidenceResponse(evidence, article, credibilityTags),
	}, nil
}

func (s *CaseService) UpdateBlogEvidenceByAdmin(
	adminUserID uuid.UUID,
	caseID uuid.UUID,
	caseVersionID uuid.UUID,
	caseEvidenceID uuid.UUID,
	req model.AdminUpdateBlogEvidenceRequest,
) (*model.AdminUpdateBlogEvidenceResponse, error) {
	if err := validateAdminEvidenceIDs(adminUserID, caseID, caseVersionID, caseEvidenceID); err != nil {
		return nil, err
	}

	label, err := helper.RequireTrimmedString(req.Label, "label is required")
	if err != nil {
		return nil, err
	}

	title, err := helper.RequireTrimmedString(req.Title, "title is required")
	if err != nil {
		return nil, err
	}

	authorName, err := helper.RequireTrimmedString(req.AuthorName, "author name is required")
	if err != nil {
		return nil, err
	}

	blogName, err := helper.RequireTrimmedString(req.BlogName, "blog name is required")
	if err != nil {
		return nil, err
	}

	bodyText, err := helper.RequireTrimmedString(req.BodyText, "body text is required")
	if err != nil {
		return nil, err
	}

	if len(label) > maxEvidenceLabelLength {
		return nil, appErrors.BadRequest("label is too long")
	}

	credibilityTags, credibilityTagsJSON, err := normalizeCredibilityTags(req.CredibilityTags)
	if err != nil {
		return nil, err
	}

	publishDate, err := parseEvidencePublishDate(req.PublishDate)
	if err != nil {
		return nil, err
	}

	tx := s.db.Begin()
	defer tx.Rollback()

	evidence, err := s.getEditableEvidenceByAdmin(tx, caseID, caseVersionID, caseEvidenceID, model.CaseEvidenceTemplateBlog)
	if err != nil {
		return nil, err
	}

	blog := evidence.Blog
	if blog == nil {
		blog = &entity.CaseEvidenceBlog{CaseEvidenceID: evidence.CaseEvidenceID}
	}

	evidence.Label = label
	evidence.CredibilityTags = credibilityTagsJSON
	evidence.IsCritical = req.IsCritical
	evidence.SortOrder = req.SortOrder

	blog.Title = title
	blog.AuthorName = authorName
	blog.BlogName = blogName
	blog.PublishDate = publishDate
	blog.BodyText = bodyText

	err = s.caseEvidenceRepo.UpdateCaseEvidence(tx, evidence)
	if err != nil {
		return nil, appErrors.InternalServer("failed to update case evidence")
	}

	err = s.caseEvidenceRepo.UpdateBlogEvidence(tx, blog)
	if err != nil {
		return nil, appErrors.InternalServer("failed to update blog evidence")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	return &model.AdminUpdateBlogEvidenceResponse{
		Evidence: mapBlogEvidenceResponse(evidence, blog, credibilityTags),
	}, nil
}

func (s *CaseService) UpdateForumThreadEvidenceByAdmin(
	adminUserID uuid.UUID,
	caseID uuid.UUID,
	caseVersionID uuid.UUID,
	caseEvidenceID uuid.UUID,
	req model.AdminUpdateForumThreadEvidenceRequest,
) (*model.AdminUpdateForumThreadEvidenceResponse, error) {
	err := validateAdminEvidenceIDs(adminUserID, caseID, caseVersionID, caseEvidenceID)
	if err != nil {
		return nil, err
	}

	label, err := helper.RequireTrimmedString(req.Label, "label is required")
	if err != nil {
		return nil, err
	}

	threadTitle, err := helper.RequireTrimmedString(req.ThreadTitle, "thread title is required")
	if err != nil {
		return nil, err
	}

	forumName, err := helper.RequireTrimmedString(req.ForumName, "forum name is required")
	if err != nil {
		return nil, err
	}

	if len(label) > maxEvidenceLabelLength {
		return nil, appErrors.BadRequest("label is too long")
	}

	credibilityTags, credibilityTagsJSON, err := normalizeCredibilityTagItems(req.CredibilityTags)
	if err != nil {
		return nil, err
	}

	if len(req.Posts) == 0 {
		return nil, appErrors.BadRequest("posts are required")
	}

	posts := make([]entity.CaseEvidenceForumThreadPost, 0, len(req.Posts))
	for i, postReq := range req.Posts {
		authorName, err := helper.RequireTrimmedString(postReq.AuthorName, "post author name is required")
		if err != nil {
			return nil, err
		}

		text, err := helper.RequireTrimmedString(postReq.Text, "post text is required")
		if err != nil {
			return nil, err
		}

		timestamp, err := parseEvidenceTimestamp(postReq.Timestamp)
		if err != nil {
			return nil, err
		}

		if postReq.UpvoteCount < 0 {
			return nil, appErrors.BadRequest("upvote count cannot be negative")
		}

		posts = append(posts, entity.CaseEvidenceForumThreadPost{
			CaseEvidenceForumThreadPostID: uuid.New(),
			CaseEvidenceID:                caseEvidenceID,
			AuthorName:                    authorName,
			Text:                          text,
			Timestamp:                     timestamp,
			UpvoteCount:                   postReq.UpvoteCount,
			SortOrder:                     i + 1,
		})
	}

	tx := s.db.Begin()
	defer tx.Rollback()

	evidence, err := s.getEditableEvidenceByAdmin(tx, caseID, caseVersionID, caseEvidenceID, model.CaseEvidenceTemplateForumThread)
	if err != nil {
		return nil, err
	}

	forumThread := evidence.ForumThread
	if forumThread == nil {
		forumThread = &entity.CaseEvidenceForumThread{CaseEvidenceID: evidence.CaseEvidenceID}
	}

	evidence.Label = label
	evidence.CredibilityTags = credibilityTagsJSON
	evidence.IsCritical = req.IsCritical
	evidence.SortOrder = req.SortOrder

	forumThread.ThreadTitle = threadTitle
	forumThread.ForumName = forumName

	err = s.caseEvidenceRepo.UpdateCaseEvidence(tx, evidence)
	if err != nil {
		return nil, appErrors.InternalServer("failed to update case evidence")
	}

	err = s.caseEvidenceRepo.UpdateForumThreadEvidence(tx, forumThread)
	if err != nil {
		return nil, appErrors.InternalServer("failed to update forum thread evidence")
	}

	err = s.caseEvidenceRepo.ReplaceForumThreadPosts(tx, evidence.CaseEvidenceID, posts)
	if err != nil {
		return nil, appErrors.InternalServer("failed to replace forum thread posts")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	return &model.AdminUpdateForumThreadEvidenceResponse{
		Evidence: mapForumThreadEvidenceResponse(evidence, forumThread, posts, credibilityTags),
	}, nil
}

func (s *CaseService) UpdateChatTranscriptEvidenceByAdmin(
	adminUserID uuid.UUID,
	caseID uuid.UUID,
	caseVersionID uuid.UUID,
	caseEvidenceID uuid.UUID,
	req model.AdminUpdateChatTranscriptEvidenceRequest,
) (*model.AdminUpdateChatTranscriptEvidenceResponse, error) {
	err := validateAdminEvidenceIDs(adminUserID, caseID, caseVersionID, caseEvidenceID)
	if err != nil {
		return nil, err
	}

	label, err := helper.RequireTrimmedString(req.Label, "label is required")
	if err != nil {
		return nil, err
	}

	if len(label) > maxEvidenceLabelLength {
		return nil, appErrors.BadRequest("label is too long")
	}

	credibilityTags, credibilityTagsJSON, err := normalizeCredibilityTagItems(req.CredibilityTags)
	if err != nil {
		return nil, err
	}

	if len(req.Participants) == 0 {
		return nil, appErrors.BadRequest("participants are required")
	}

	if len(req.Messages) == 0 {
		return nil, appErrors.BadRequest("messages are required")
	}

	participantEntities := make([]entity.CaseEvidenceChatTranscriptParticipant, 0, len(req.Participants))
	participantSet := map[string]bool{}
	for i, participant := range req.Participants {
		name, err := helper.RequireTrimmedString(participant, "participant name is required")
		if err != nil {
			return nil, err
		}

		if participantSet[name] {
			return nil, appErrors.BadRequest("participants cannot contain duplicates")
		}

		participantSet[name] = true
		participantEntities = append(participantEntities, entity.CaseEvidenceChatTranscriptParticipant{
			CaseEvidenceChatTranscriptParticipantID: uuid.New(),
			CaseEvidenceID:                          caseEvidenceID,
			Name:                                    name,
			SortOrder:                               i + 1,
		})
	}

	messageEntities := make([]entity.CaseEvidenceChatTranscriptMessage, 0, len(req.Messages))
	for i, messageReq := range req.Messages {
		sender, err := helper.RequireTrimmedString(messageReq.Sender, "message sender is required")
		if err != nil {
			return nil, err
		}

		if !participantSet[sender] {
			return nil, appErrors.BadRequest("message sender must be one of participants")
		}

		text, err := helper.RequireTrimmedString(messageReq.Text, "message text is required")
		if err != nil {
			return nil, err
		}

		timestamp, err := parseEvidenceTimestamp(messageReq.Timestamp)
		if err != nil {
			return nil, err
		}

		messageEntities = append(messageEntities, entity.CaseEvidenceChatTranscriptMessage{
			CaseEvidenceChatTranscriptMessageID: uuid.New(),
			CaseEvidenceID:                      caseEvidenceID,
			Sender:                              sender,
			Text:                                text,
			Timestamp:                           timestamp,
			SortOrder:                           i + 1,
		})
	}

	tx := s.db.Begin()
	defer tx.Rollback()

	evidence, err := s.getEditableEvidenceByAdmin(tx, caseID, caseVersionID, caseEvidenceID, model.CaseEvidenceTemplateChatTranscript)
	if err != nil {
		return nil, err
	}

	chatTranscript := evidence.ChatTranscript
	if chatTranscript == nil {
		chatTranscript = &entity.CaseEvidenceChatTranscript{CaseEvidenceID: evidence.CaseEvidenceID}
	}

	evidence.Label = label
	evidence.CredibilityTags = credibilityTagsJSON
	evidence.IsCritical = req.IsCritical
	evidence.SortOrder = req.SortOrder

	err = s.caseEvidenceRepo.UpdateCaseEvidence(tx, evidence)
	if err != nil {
		return nil, appErrors.InternalServer("failed to update case evidence")
	}

	err = s.caseEvidenceRepo.UpdateChatTranscriptEvidence(tx, chatTranscript)
	if err != nil {
		return nil, appErrors.InternalServer("failed to update chat transcript evidence")
	}

	err = s.caseEvidenceRepo.ReplaceChatTranscriptParticipants(tx, evidence.CaseEvidenceID, participantEntities)
	if err != nil {
		return nil, appErrors.InternalServer("failed to replace chat transcript participants")
	}

	err = s.caseEvidenceRepo.ReplaceChatTranscriptMessages(tx, evidence.CaseEvidenceID, messageEntities)
	if err != nil {
		return nil, appErrors.InternalServer("failed to replace chat transcript messages")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	return &model.AdminUpdateChatTranscriptEvidenceResponse{
		Evidence: mapChatTranscriptEvidenceResponse(evidence, participantEntities, messageEntities, credibilityTags),
	}, nil
}

func (s *CaseService) UpdatePublicAnnouncementEvidenceByAdmin(
	adminUserID uuid.UUID,
	caseID uuid.UUID,
	caseVersionID uuid.UUID,
	caseEvidenceID uuid.UUID,
	req model.AdminUpdatePublicAnnouncementEvidenceRequest,
) (*model.AdminUpdatePublicAnnouncementEvidenceResponse, error) {
	err := validateAdminEvidenceIDs(adminUserID, caseID, caseVersionID, caseEvidenceID)
	if err != nil {
		return nil, err
	}

	label, err := helper.RequireTrimmedString(req.Label, "label is required")
	if err != nil {
		return nil, err
	}

	issuingBody, err := helper.RequireTrimmedString(req.IssuingBody, "issuing body is required")
	if err != nil {
		return nil, err
	}

	title, err := helper.RequireTrimmedString(req.Title, "title is required")
	if err != nil {
		return nil, err
	}

	bodyText, err := helper.RequireTrimmedString(req.BodyText, "body text is required")
	if err != nil {
		return nil, err
	}

	if len(label) > maxEvidenceLabelLength {
		return nil, appErrors.BadRequest("label is too long")
	}

	credibilityTags, credibilityTagsJSON, err := normalizeCredibilityTagItems(req.CredibilityTags)
	if err != nil {
		return nil, err
	}

	announcementDate, err := parseEvidencePublishDate(req.Date)
	if err != nil {
		return nil, err
	}

	tx := s.db.Begin()
	defer tx.Rollback()

	evidence, err := s.getEditableEvidenceByAdmin(tx, caseID, caseVersionID, caseEvidenceID, model.CaseEvidenceTemplatePublicAnnouncement)
	if err != nil {
		return nil, err
	}

	publicAnnouncement := evidence.PublicAnnouncement
	if publicAnnouncement == nil {
		publicAnnouncement = &entity.CaseEvidencePublicAnnouncement{CaseEvidenceID: evidence.CaseEvidenceID}
	}

	evidence.Label = label
	evidence.CredibilityTags = credibilityTagsJSON
	evidence.IsCritical = req.IsCritical
	evidence.SortOrder = req.SortOrder

	publicAnnouncement.IssuingBody = issuingBody
	publicAnnouncement.Title = title
	publicAnnouncement.Date = announcementDate
	publicAnnouncement.BodyText = bodyText

	err = s.caseEvidenceRepo.UpdateCaseEvidence(tx, evidence)
	if err != nil {
		return nil, appErrors.InternalServer("failed to update case evidence")
	}

	err = s.caseEvidenceRepo.UpdatePublicAnnouncementEvidence(tx, publicAnnouncement)
	if err != nil {
		return nil, appErrors.InternalServer("failed to update public announcement evidence")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	return &model.AdminUpdatePublicAnnouncementEvidenceResponse{
		Evidence: mapPublicAnnouncementEvidenceResponse(evidence, publicAnnouncement, credibilityTags),
	}, nil
}

func (s *CaseService) DeleteCaseEvidenceByAdmin(
	adminUserID uuid.UUID,
	caseID uuid.UUID,
	caseVersionID uuid.UUID,
	caseEvidenceID uuid.UUID,
) (*model.AdminDeleteCaseEvidenceResponse, error) {
	if err := validateAdminEvidenceIDs(adminUserID, caseID, caseVersionID, caseEvidenceID); err != nil {
		return nil, err
	}

	tx := s.db.Begin()
	defer tx.Rollback()

	evidence, err := s.getEditableEvidenceByAdmin(tx, caseID, caseVersionID, caseEvidenceID, "")
	if err != nil {
		return nil, err
	}

	var imageURL *string
	if evidence.SocialPost != nil {
		imageURL = evidence.SocialPost.ImageURL
	}
	if evidence.Article != nil {
		imageURL = evidence.Article.ImageURL
	}

	err = s.caseEvidenceRepo.DeleteCaseEvidence(tx, evidence.CaseEvidenceID)
	if err != nil {
		return nil, appErrors.InternalServer("failed to delete case evidence")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	if imageURL != nil {
		_ = supabase.DeleteFileIfPresent(s.storage, *imageURL)
	}

	return &model.AdminDeleteCaseEvidenceResponse{
		CaseEvidenceID: evidence.CaseEvidenceID,
	}, nil
}

func validateAdminEvidenceIDs(adminUserID uuid.UUID, caseID uuid.UUID, caseVersionID uuid.UUID, caseEvidenceID uuid.UUID) error {
	if adminUserID == uuid.Nil {
		return appErrors.Unauthorized("unauthorized")
	}

	if caseID == uuid.Nil {
		return appErrors.BadRequest("invalid case id")
	}

	if caseVersionID == uuid.Nil {
		return appErrors.BadRequest("invalid case version id")
	}

	if caseEvidenceID == uuid.Nil {
		return appErrors.BadRequest("invalid case evidence id")
	}

	return nil
}

func (s *CaseService) getEditableEvidenceByAdmin(
	tx *gorm.DB,
	caseID uuid.UUID,
	caseVersionID uuid.UUID,
	caseEvidenceID uuid.UUID,
	templateType string,
) (*entity.CaseEvidence, error) {
	caseVersion, err := s.caseVersionRepo.GetCaseVersionForUpdate(tx, caseVersionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("case version not found")
		}
		return nil, appErrors.InternalServer("failed to get case version")
	}

	if caseVersion.CaseID != caseID {
		return nil, appErrors.NotFound("case version not found")
	}

	if caseVersion.Status != model.CaseStatusDraft {
		return nil, appErrors.Conflict("case version is not editable")
	}

	evidence, err := s.caseEvidenceRepo.GetCaseEvidence(tx, model.GetCaseEvidenceParam{
		CaseEvidenceID: caseEvidenceID,
		CaseVersionID:  caseVersionID,
		TemplateType:   templateType,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("case evidence not found")
		}
		return nil, appErrors.InternalServer("failed to get case evidence")
	}

	return evidence, nil
}

func parseEvidencePublishDate(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, appErrors.BadRequest("publish date is required")
	}

	layouts := []string{
		"2006-01-02",
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
	}

	for _, layout := range layouts {
		value, err := time.Parse(layout, raw)
		if err == nil {
			return value.UTC(), nil
		}
	}

	return time.Time{}, appErrors.BadRequest("invalid publish date")
}

func normalizeOptionalURL(raw string) (*string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	parsedURL, err := url.ParseRequestURI(raw)
	if err != nil {
		return nil, appErrors.BadRequest("invalid url")
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, appErrors.BadRequest("url must use http or https")
	}

	return &raw, nil
}

func mapArticleEvidenceResponse(
	evidence *entity.CaseEvidence,
	article *entity.CaseEvidenceArticle,
	credibilityTags []string,
) model.AdminArticleEvidenceResponse {
	return model.AdminArticleEvidenceResponse{
		CaseEvidenceID:  evidence.CaseEvidenceID,
		CaseVersionID:   evidence.CaseVersionID,
		TemplateType:    evidence.TemplateType,
		Label:           evidence.Label,
		CredibilityTags: credibilityTags,
		IsCritical:      evidence.IsCritical,
		SortOrder:       evidence.SortOrder,
		Headline:        article.Headline,
		SourceName:      article.SourceName,
		AuthorName:      article.AuthorName,
		PublishDate:     article.PublishDate,
		URL:             article.URL,
		BodyText:        article.BodyText,
		ImagePrompt:     article.ImagePrompt,
		ImageURL:        article.ImageURL,
		CreatedAt:       evidence.CreatedAt,
		UpdatedAt:       evidence.UpdatedAt,
	}
}

func mapBlogEvidenceResponse(
	evidence *entity.CaseEvidence,
	blog *entity.CaseEvidenceBlog,
	credibilityTags []string,
) model.AdminBlogEvidenceResponse {
	return model.AdminBlogEvidenceResponse{
		CaseEvidenceID:  evidence.CaseEvidenceID,
		CaseVersionID:   evidence.CaseVersionID,
		TemplateType:    evidence.TemplateType,
		Label:           evidence.Label,
		CredibilityTags: credibilityTags,
		IsCritical:      evidence.IsCritical,
		SortOrder:       evidence.SortOrder,
		Title:           blog.Title,
		AuthorName:      blog.AuthorName,
		BlogName:        blog.BlogName,
		PublishDate:     blog.PublishDate,
		BodyText:        blog.BodyText,
		CreatedAt:       evidence.CreatedAt,
		UpdatedAt:       evidence.UpdatedAt,
	}
}

func mapForumThreadEvidenceResponse(
	evidence *entity.CaseEvidence,
	forumThread *entity.CaseEvidenceForumThread,
	posts []entity.CaseEvidenceForumThreadPost,
	credibilityTags []string,
) model.AdminForumThreadEvidenceResponse {
	postResponses := make([]model.AdminForumThreadPostResponse, 0, len(posts))
	for _, post := range posts {
		postResponses = append(postResponses, model.AdminForumThreadPostResponse{
			CaseEvidenceForumThreadPostID: post.CaseEvidenceForumThreadPostID,
			CaseEvidenceID:                post.CaseEvidenceID,
			AuthorName:                    post.AuthorName,
			Text:                          post.Text,
			Timestamp:                     post.Timestamp,
			UpvoteCount:                   post.UpvoteCount,
			SortOrder:                     post.SortOrder,
			CreatedAt:                     post.CreatedAt,
			UpdatedAt:                     post.UpdatedAt,
		})
	}

	return model.AdminForumThreadEvidenceResponse{
		CaseEvidenceID:  evidence.CaseEvidenceID,
		CaseVersionID:   evidence.CaseVersionID,
		TemplateType:    evidence.TemplateType,
		Label:           evidence.Label,
		CredibilityTags: credibilityTags,
		IsCritical:      evidence.IsCritical,
		SortOrder:       evidence.SortOrder,
		ThreadTitle:     forumThread.ThreadTitle,
		ForumName:       forumThread.ForumName,
		Posts:           postResponses,
		CreatedAt:       evidence.CreatedAt,
		UpdatedAt:       evidence.UpdatedAt,
	}
}

func mapChatTranscriptEvidenceResponse(
	evidence *entity.CaseEvidence,
	participants []entity.CaseEvidenceChatTranscriptParticipant,
	messages []entity.CaseEvidenceChatTranscriptMessage,
	credibilityTags []string,
) model.AdminChatTranscriptEvidenceResponse {
	participantResponses := make([]model.AdminChatTranscriptParticipantResponse, 0, len(participants))
	for _, participant := range participants {
		participantResponses = append(participantResponses, model.AdminChatTranscriptParticipantResponse{
			CaseEvidenceChatTranscriptParticipantID: participant.CaseEvidenceChatTranscriptParticipantID,
			CaseEvidenceID:                          participant.CaseEvidenceID,
			Name:                                    participant.Name,
			SortOrder:                               participant.SortOrder,
			CreatedAt:                               participant.CreatedAt,
			UpdatedAt:                               participant.UpdatedAt,
		})
	}

	messageResponses := make([]model.AdminChatTranscriptMessageResponse, 0, len(messages))
	for _, message := range messages {
		messageResponses = append(messageResponses, model.AdminChatTranscriptMessageResponse{
			CaseEvidenceChatTranscriptMessageID: message.CaseEvidenceChatTranscriptMessageID,
			CaseEvidenceID:                      message.CaseEvidenceID,
			Sender:                              message.Sender,
			Text:                                message.Text,
			Timestamp:                           message.Timestamp,
			SortOrder:                           message.SortOrder,
			CreatedAt:                           message.CreatedAt,
			UpdatedAt:                           message.UpdatedAt,
		})
	}

	return model.AdminChatTranscriptEvidenceResponse{
		CaseEvidenceID:  evidence.CaseEvidenceID,
		CaseVersionID:   evidence.CaseVersionID,
		TemplateType:    evidence.TemplateType,
		Label:           evidence.Label,
		CredibilityTags: credibilityTags,
		IsCritical:      evidence.IsCritical,
		SortOrder:       evidence.SortOrder,
		Participants:    participantResponses,
		Messages:        messageResponses,
		CreatedAt:       evidence.CreatedAt,
		UpdatedAt:       evidence.UpdatedAt,
	}
}

func mapPublicAnnouncementEvidenceResponse(
	evidence *entity.CaseEvidence,
	publicAnnouncement *entity.CaseEvidencePublicAnnouncement,
	credibilityTags []string,
) model.AdminPublicAnnouncementEvidenceResponse {
	return model.AdminPublicAnnouncementEvidenceResponse{
		CaseEvidenceID:  evidence.CaseEvidenceID,
		CaseVersionID:   evidence.CaseVersionID,
		TemplateType:    evidence.TemplateType,
		Label:           evidence.Label,
		CredibilityTags: credibilityTags,
		IsCritical:      evidence.IsCritical,
		SortOrder:       evidence.SortOrder,
		IssuingBody:     publicAnnouncement.IssuingBody,
		Title:           publicAnnouncement.Title,
		Date:            publicAnnouncement.Date,
		BodyText:        publicAnnouncement.BodyText,
		CreatedAt:       evidence.CreatedAt,
		UpdatedAt:       evidence.UpdatedAt,
	}
}

func normalizeCredibilityTagItems(tags []string) ([]string, string, error) {
	normalizedTags := make([]string, 0, len(tags))
	seen := map[string]bool{}

	for _, tag := range tags {
		normalizedTag := strings.ToLower(strings.TrimSpace(tag))
		if normalizedTag == "" {
			continue
		}

		if seen[normalizedTag] {
			continue
		}

		seen[normalizedTag] = true
		normalizedTags = append(normalizedTags, normalizedTag)
	}

	if len(normalizedTags) == 0 {
		return nil, "", appErrors.BadRequest("credibility tags are required")
	}

	payload, err := json.Marshal(normalizedTags)
	if err != nil {
		return nil, "", appErrors.InternalServer("failed to normalize credibility tags")
	}

	return normalizedTags, string(payload), nil
}

func normalizeCredibilityTags(raw string) ([]string, string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, "", appErrors.BadRequest("credibility tags are required")
	}

	var tags []string
	if err := json.Unmarshal([]byte(raw), &tags); err != nil {
		return nil, "", appErrors.BadRequest("credibility tags must be a valid json array")
	}

	normalizedTags := make([]string, 0, len(tags))
	seen := map[string]bool{}

	for _, tag := range tags {
		normalizedTag := strings.ToLower(strings.TrimSpace(tag))
		if normalizedTag == "" {
			continue
		}

		if seen[normalizedTag] {
			continue
		}

		seen[normalizedTag] = true
		normalizedTags = append(normalizedTags, normalizedTag)
	}

	if len(normalizedTags) == 0 {
		return nil, "", appErrors.BadRequest("credibility tags are required")
	}

	payload, err := json.Marshal(normalizedTags)
	if err != nil {
		return nil, "", appErrors.InternalServer("failed to normalize credibility tags")
	}

	return normalizedTags, string(payload), nil
}

func parseEvidenceTimestamp(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, appErrors.BadRequest("timestamp is required")
	}

	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
	}

	for _, layout := range layouts {
		value, err := time.Parse(layout, raw)
		if err == nil {
			return value.UTC(), nil
		}
	}

	return time.Time{}, appErrors.BadRequest("invalid timestamp")
}

func mapSocialPostEvidenceResponse(
	evidence *entity.CaseEvidence,
	socialPost *entity.CaseEvidenceSocialPost,
	credibilityTags []string,
) model.AdminSocialPostEvidenceResponse {
	return model.AdminSocialPostEvidenceResponse{
		CaseEvidenceID:    evidence.CaseEvidenceID,
		CaseVersionID:     evidence.CaseVersionID,
		TemplateType:      evidence.TemplateType,
		Label:             evidence.Label,
		CredibilityTags:   credibilityTags,
		IsCritical:        evidence.IsCritical,
		SortOrder:         evidence.SortOrder,
		AuthorName:        socialPost.AuthorName,
		AuthorHandle:      socialPost.AuthorHandle,
		Platform:          socialPost.Platform,
		PostText:          socialPost.PostText,
		Timestamp:         socialPost.Timestamp,
		LikesCount:        socialPost.LikesCount,
		SharesCount:       socialPost.SharesCount,
		CommentsCount:     socialPost.CommentsCount,
		IsVerifiedAccount: socialPost.IsVerifiedAccount,
		ImagePrompt:       socialPost.ImagePrompt,
		ImageURL:          socialPost.ImageURL,
		CreatedAt:         evidence.CreatedAt,
		UpdatedAt:         evidence.UpdatedAt,
	}
}

func (s *CaseService) uploadCaseThumbnail(file *multipart.FileHeader) (*string, error) {
	return s.uploadOptionalImage(
		file,
		maxCaseThumbnailSize,
		"thumbnail size exceeds 5MB limit",
		"failed to upload thumbnail",
	)
}

func (s *CaseService) uploadEvidenceImage(file *multipart.FileHeader) (*string, error) {
	return s.uploadOptionalImage(
		file,
		maxEvidenceImageSize,
		"evidence image size exceeds 5MB limit",
		"failed to upload evidence image",
	)
}

func (s *CaseService) uploadOptionalImage(file *multipart.FileHeader, maxSize int64, sizeErrMessage string, uploadErrMessage string) (*string, error) {
	if file == nil {
		return nil, nil
	}

	url, err := supabase.UploadOptionalImage(
		s.storage,
		file,
		maxSize,
		sizeErrMessage,
	)
	if err != nil {
		return nil, appErrors.BadRequest(uploadErrMessage)
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

func (s *CaseService) buildUniqueSlugForUpdate(caseID uuid.UUID, title string) (string, error) {
	baseSlug := slugify(title)
	if baseSlug == "" {
		baseSlug = "case"
	}

	existingCase, err := s.caseRepo.GetCase(s.db, model.GetCaseParam{
		Slug: baseSlug,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return baseSlug, nil
		}
		return "", appErrors.InternalServer("failed to check case slug")
	}

	if existingCase.CaseID == caseID {
		return baseSlug, nil
	}

	return fmt.Sprintf("%s-%s", baseSlug, caseID.String()[:8]), nil
}

func mapAdminUpdateCaseResponse(caseEntity *entity.Case) *model.AdminUpdateCaseResponse {
	return &model.AdminUpdateCaseResponse{
		CaseID:                   caseEntity.CaseID,
		Title:                    caseEntity.Title,
		Slug:                     caseEntity.Slug,
		ShortDescription:         caseEntity.ShortDescription,
		Theme:                    caseEntity.Theme,
		ThemeOtherText:           caseEntity.ThemeOtherText,
		CompetencyFocus:          caseEntity.CompetencyFocus,
		DifficultyLevel:          caseEntity.DifficultyLevel,
		RiskLevel:                caseEntity.RiskLevel,
		EstimatedDurationMinutes: caseEntity.EstimatedDurationMinutes,
		AIModel:                  caseEntity.AIModel,
		MinimumLevel:             caseEntity.MinimumLevel,
		MinimumReputation:        caseEntity.MinimumReputation,
		UnlockRequirement:        caseEntity.UnlockRequirement,
		ThumbnailURL:             caseEntity.ThumbnailURL,
		ThumbnailPrompt:          caseEntity.ThumbnailPrompt,
		GenerationSource:         caseEntity.GenerationSource,
		Status:                   caseEntity.Status,
		PublishedAt:              caseEntity.PublishedAt,
		CreatedBy:                caseEntity.CreatedBy,
		CreatedAt:                caseEntity.CreatedAt,
		UpdatedAt:                caseEntity.UpdatedAt,
	}
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

func optionalVersionLabel(versionNumber *int) *string {
	if versionNumber == nil {
		return nil
	}

	label := formatCaseVersionLabel(*versionNumber)
	return &label
}
