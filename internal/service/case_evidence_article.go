package service

import (
	"errors"
	"strings"

	"github.com/azmiagr/unesco-hackathon/entity"
	"github.com/azmiagr/unesco-hackathon/model"
	appErrors "github.com/azmiagr/unesco-hackathon/pkg/errors"
	"github.com/azmiagr/unesco-hackathon/pkg/helper"
	"github.com/azmiagr/unesco-hackathon/pkg/supabase"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

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

	err = s.writeCaseEvidenceAuditLog(tx, adminUserID, model.AuditActionCreate, evidence, nil)
	if err != nil {
		return nil, err
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

	before := newAuditCaseEvidenceSnapshot(evidence)

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

	err = s.writeCaseEvidenceAuditLog(tx, adminUserID, model.AuditActionUpdate, evidence, before)
	if err != nil {
		return nil, err
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
