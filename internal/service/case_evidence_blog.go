package service

import (
	"errors"
	"github.com/azmiagr/unesco-hackathon/entity"
	"github.com/azmiagr/unesco-hackathon/model"
	appErrors "github.com/azmiagr/unesco-hackathon/pkg/errors"
	"github.com/azmiagr/unesco-hackathon/pkg/helper"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

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

	err = s.writeCaseEvidenceAuditLog(tx, adminUserID, model.AuditActionCreate, evidence, nil)
	if err != nil {
		return nil, err
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	return &model.AdminCreateBlogEvidenceResponse{
		Evidence: mapBlogEvidenceResponse(evidence, blog, credibilityTags),
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

	before := newAuditCaseEvidenceSnapshot(evidence)

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

	err = s.writeCaseEvidenceAuditLog(tx, adminUserID, model.AuditActionUpdate, evidence, before)
	if err != nil {
		return nil, err
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	return &model.AdminUpdateBlogEvidenceResponse{
		Evidence: mapBlogEvidenceResponse(evidence, blog, credibilityTags),
	}, nil
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
