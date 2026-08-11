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
