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
