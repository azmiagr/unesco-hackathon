package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"mime/multipart"
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
	ListCasesByAdmin(req model.AdminListCasesRequest) (*model.AdminListCasesResponse, error)
	GetCaseDetailByAdmin(caseID uuid.UUID) (*model.AdminCaseDetailResponse, error)
	GetCaseLookups() (*model.AdminCaseLookupsResponse, error)
	CreateSocialPostEvidenceByAdmin(adminUserID uuid.UUID, caseID uuid.UUID, caseVersionID uuid.UUID, req model.AdminCreateSocialPostEvidenceRequest) (*model.AdminCreateSocialPostEvidenceResponse, error)
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
