package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/azmiagr/unesco-hackathon/entity"
	"github.com/azmiagr/unesco-hackathon/model"
	appErrors "github.com/azmiagr/unesco-hackathon/pkg/errors"
	"github.com/azmiagr/unesco-hackathon/pkg/helper"
	"github.com/azmiagr/unesco-hackathon/pkg/supabase"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	minPublishQuestionCount = 5
	minPublishEvidenceCount = 3

	defaultUserCaseLimit = 10
	maxUserCaseLimit     = 100
	newUserCaseWindow    = 7 * 24 * time.Hour
)

var allowedUserCaseTabs = map[string]bool{
	model.UserCaseTabAll:        true,
	model.UserCaseTabInProgress: true,
	model.UserCaseTabCompleted:  true,
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

	err = writeAdminAuditLog(tx, s.auditLogRepo, s.userRepo, adminAuditLogParam{
		ActorAdminID: adminUserID,
		ActionType:   model.AuditActionCreate,
		Module:       model.AuditModuleCMS,
		TargetType:   "case",
		TargetID:     caseEntity.CaseID.String(),
		TargetLabel:  caseEntity.Title,
		Detail:       fmt.Sprintf("Created case %s", caseEntity.Title),
		PayloadAfter: map[string]any{
			"case":            newAuditCaseSnapshot(caseEntity),
			"case_version_id": caseVersion.CaseVersionID,
			"version_number":  caseVersion.VersionNumber,
		},
	})
	if err != nil {
		return nil, err
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

	before := newAuditCaseSnapshot(caseEntity)

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

	err = writeAdminAuditLog(tx, s.auditLogRepo, s.userRepo, adminAuditLogParam{
		ActorAdminID:  adminUserID,
		ActionType:    model.AuditActionUpdate,
		Module:        model.AuditModuleCMS,
		TargetType:    "case",
		TargetID:      caseEntity.CaseID.String(),
		TargetLabel:   caseEntity.Title,
		Detail:        fmt.Sprintf("Updated case %s", caseEntity.Title),
		PayloadBefore: before,
		PayloadAfter:  newAuditCaseSnapshot(caseEntity),
	})
	if err != nil {
		return nil, err
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

func (s *CaseService) PublishCaseByAdmin(
	adminUserID uuid.UUID,
	caseID uuid.UUID,
) (*model.AdminPublishCaseResponse, error) {
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
	if caseEntity.Status == model.CaseStatusPublished {
		return nil, appErrors.BadRequest("case is already published")
	}
	if caseEntity.Status == model.CaseStatusArchived {
		return nil, appErrors.BadRequest("archived case cannot be published")
	}

	caseVersions, err := s.caseVersionRepo.ListCaseVersionsByCaseID(tx, caseID)
	if err != nil {
		return nil, appErrors.InternalServer("failed to list case versions")
	}
	if len(caseVersions) == 0 {
		return nil, appErrors.BadRequest("case version is required")
	}

	caseVersion, err := s.caseVersionRepo.GetCaseVersionForUpdate(tx, caseVersions[0].CaseVersionID)
	if err != nil {
		return nil, appErrors.InternalServer("failed to get case version")
	}
	if caseVersion.Status == model.CaseStatusPublished {
		return nil, appErrors.BadRequest("case version is already published")
	}
	if caseVersion.Status == model.CaseStatusArchived {
		return nil, appErrors.BadRequest("archived case version cannot be published")
	}

	before := newAuditCasePublishSnapshot(caseEntity, caseVersion)

	requirements, err := s.buildPublishRequirements(tx, caseID, caseVersion.CaseVersionID)
	if err != nil {
		return nil, err
	}
	if !publishRequirementsMet(requirements) {
		return nil, appErrors.BadRequest(buildPublishRequirementsErrorMessage(requirements))
	}

	publishedAt := time.Now()
	caseEntity.Status = model.CaseStatusPublished
	caseEntity.PublishedAt = &publishedAt
	caseVersion.Status = model.CaseStatusPublished
	caseVersion.PublishedAt = &publishedAt

	if err := s.caseRepo.UpdateCase(tx, caseEntity); err != nil {
		return nil, appErrors.InternalServer("failed to publish case")
	}
	if err := s.caseVersionRepo.UpdateCaseVersion(tx, caseVersion); err != nil {
		return nil, appErrors.InternalServer("failed to publish case version")
	}

	err = writeAdminAuditLog(tx, s.auditLogRepo, s.userRepo, adminAuditLogParam{
		ActorAdminID:  adminUserID,
		ActionType:    model.AuditActionUpdate,
		Module:        model.AuditModuleCMS,
		TargetType:    "case",
		TargetID:      caseEntity.CaseID.String(),
		TargetLabel:   caseEntity.Title,
		Detail:        fmt.Sprintf("Published case %s", caseEntity.Title),
		PayloadBefore: before,
		PayloadAfter:  newAuditCasePublishSnapshot(caseEntity, caseVersion),
	})
	if err != nil {
		return nil, err
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	return &model.AdminPublishCaseResponse{
		CaseID:        caseEntity.CaseID,
		CaseVersionID: caseVersion.CaseVersionID,
		Status:        caseEntity.Status,
		PublishedAt:   publishedAt,
		Requirements:  requirements,
	}, nil
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

	before := newAuditCaseSnapshot(caseEntity)
	thumbnailURL := caseEntity.ThumbnailURL
	err = s.caseRepo.HardDeleteCase(tx, caseEntity.CaseID)
	if err != nil {
		return nil, appErrors.InternalServer("failed to delete case")
	}

	err = writeAdminAuditLog(tx, s.auditLogRepo, s.userRepo, adminAuditLogParam{
		ActorAdminID:  adminUserID,
		ActionType:    model.AuditActionDelete,
		Module:        model.AuditModuleCMS,
		TargetType:    "case",
		TargetID:      caseEntity.CaseID.String(),
		TargetLabel:   caseEntity.Title,
		Detail:        fmt.Sprintf("Hard deleted case %s", caseEntity.Title),
		PayloadBefore: before,
	})
	if err != nil {
		return nil, err
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

func (s *CaseService) ListCasesForUser(userID uuid.UUID, req model.ListUserCasesRequest) (*model.ListUserCasesResponse, error) {
	if userID == uuid.Nil {
		return nil, appErrors.Unauthorized("unauthorized")
	}

	page := req.Page
	if page < 1 {
		page = 1
	}

	limit := req.Limit
	if limit < 1 {
		limit = defaultUserCaseLimit
	}
	if limit > maxUserCaseLimit {
		limit = maxUserCaseLimit
	}

	tab := strings.ToLower(strings.TrimSpace(req.Tab))
	if tab == "" {
		tab = model.UserCaseTabAll
	}
	if !allowedUserCaseTabs[tab] {
		return nil, appErrors.BadRequest("invalid case tab")
	}

	if tab == model.UserCaseTabInProgress || tab == model.UserCaseTabCompleted {
		return emptyUserCaseListResponse(page, limit), nil
	}

	profile, err := s.userProfileRepo.GetUserProfile(s.db, model.GetUserProfileParam{UserID: userID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("user profile not found")
		}
		return nil, appErrors.InternalServer("failed to get user profile")
	}

	cases, total, err := s.caseRepo.ListPublishedCasesForUser(s.db, model.ListUserCasesParam{
		Limit:  limit,
		Offset: (page - 1) * limit,
	})
	if err != nil {
		return nil, appErrors.InternalServer("failed to list cases")
	}

	responses := make([]model.UserCaseCardResponse, 0, len(cases))
	for _, caseRow := range cases {
		responses = append(responses, mapUserCaseCardResponse(caseRow, profile))
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}

	return &model.ListUserCasesResponse{
		Cases: responses,
		Pagination: model.PaginationResponse{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

func emptyUserCaseListResponse(page int, limit int) *model.ListUserCasesResponse {
	return &model.ListUserCasesResponse{
		Cases: []model.UserCaseCardResponse{},
		Pagination: model.PaginationResponse{
			Page:       page,
			Limit:      limit,
			Total:      0,
			TotalPages: 0,
		},
	}
}

func mapUserCaseCardResponse(caseRow model.UserCaseListRow, profile *entity.UserProfile) model.UserCaseCardResponse {
	accessStatus := model.UserCaseAccessUnlocked
	progressStatus := model.UserCaseProgressAvailable
	var lockedReason *string

	if profile.CurrentLevel < caseRow.MinimumLevel {
		accessStatus = model.UserCaseAccessLocked
		progressStatus = model.UserCaseProgressLocked
		reason := fmt.Sprintf("Terkunci sampai level %d", caseRow.MinimumLevel)
		lockedReason = &reason
	} else if profile.AuditorReputation < caseRow.MinimumReputation {
		accessStatus = model.UserCaseAccessLocked
		progressStatus = model.UserCaseProgressLocked
		reason := fmt.Sprintf("Butuh reputasi %.0f", caseRow.MinimumReputation)
		lockedReason = &reason
	} else if time.Since(caseRow.CreatedAt) <= newUserCaseWindow {
		progressStatus = model.UserCaseProgressNew
	}

	return model.UserCaseCardResponse{
		CaseID:                   caseRow.CaseID,
		Title:                    caseRow.Title,
		Slug:                     caseRow.Slug,
		ShortDescription:         caseRow.ShortDescription,
		DifficultyLevel:          caseRow.DifficultyLevel,
		EstimatedDurationMinutes: caseRow.EstimatedDurationMinutes,
		MinimumLevel:             caseRow.MinimumLevel,
		MinimumReputation:        caseRow.MinimumReputation,
		ThumbnailURL:             caseRow.ThumbnailURL,
		AccessStatus:             accessStatus,
		ProgressStatus:           progressStatus,
		LockedReason:             lockedReason,
		PublishedAt:              caseRow.PublishedAt,
		CreatedAt:                caseRow.CreatedAt,
	}
}

func (s *CaseService) buildPublishRequirements(
	tx *gorm.DB,
	caseID uuid.UUID,
	caseVersionID uuid.UUID,
) ([]model.AdminPublishCaseRequirementResponse, error) {
	questionCount, err := s.caseQuestionRepo.CountCaseQuestions(tx, caseVersionID)
	if err != nil {
		return nil, appErrors.InternalServer("failed to count case questions")
	}

	evidenceCount, err := s.caseEvidenceRepo.CountCaseEvidences(tx, caseVersionID)
	if err != nil {
		return nil, appErrors.InternalServer("failed to count case evidences")
	}

	hasChatbotConfig, err := s.caseChatbotConfigRepo.CaseChatbotConfigExists(tx, caseID)
	if err != nil {
		return nil, appErrors.InternalServer("failed to check case chatbot config")
	}

	hasScoringOutcomeConfig, err := s.caseScoringOutcomeRepo.CaseScoringOutcomeConfigExists(tx, caseVersionID)
	if err != nil {
		return nil, appErrors.InternalServer("failed to check case scoring outcome config")
	}

	return []model.AdminPublishCaseRequirementResponse{
		{
			Key:      "questions",
			Label:    "Minimal 5 soal sudah dibuat",
			Required: minPublishQuestionCount,
			Actual:   int(questionCount),
			IsMet:    questionCount >= minPublishQuestionCount,
		},
		{
			Key:      "evidences",
			Label:    "Minimal 3 evidence sudah dilengkapi",
			Required: minPublishEvidenceCount,
			Actual:   int(evidenceCount),
			IsMet:    evidenceCount >= minPublishEvidenceCount,
		},
		{
			Key:   "ai_prompt",
			Label: "AI prompt sudah dikonfigurasi",
			IsMet: hasChatbotConfig,
		},
		{
			Key:   "reward",
			Label: "Reward sudah dikonfigurasi",
			IsMet: hasScoringOutcomeConfig,
		},
	}, nil
}

func publishRequirementsMet(requirements []model.AdminPublishCaseRequirementResponse) bool {
	for _, requirement := range requirements {
		if !requirement.IsMet {
			return false
		}
	}

	return true
}

func buildPublishRequirementsErrorMessage(requirements []model.AdminPublishCaseRequirementResponse) string {
	missingRequirements := []string{}
	for _, requirement := range requirements {
		if requirement.IsMet {
			continue
		}

		if requirement.Required > 0 {
			missingRequirements = append(
				missingRequirements,
				fmt.Sprintf("%s (%d/%d)", requirement.Label, requirement.Actual, requirement.Required),
			)
			continue
		}

		missingRequirements = append(missingRequirements, requirement.Label)
	}

	if len(missingRequirements) == 0 {
		return "case publish requirements are not met"
	}

	return "case publish requirements are not met: " + strings.Join(missingRequirements, "; ")
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
