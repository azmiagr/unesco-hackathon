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

	err = s.writeCaseEvidenceAuditLog(tx, adminUserID, model.AuditActionCreate, evidence, nil)
	if err != nil {
		return nil, err
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	return &model.AdminCreateForumThreadEvidenceResponse{
		Evidence: mapForumThreadEvidenceResponse(evidence, forumThread, posts, credibilityTags),
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

	before := newAuditCaseEvidenceSnapshot(evidence)

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

	err = s.writeCaseEvidenceAuditLog(tx, adminUserID, model.AuditActionUpdate, evidence, before)
	if err != nil {
		return nil, err
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	return &model.AdminUpdateForumThreadEvidenceResponse{
		Evidence: mapForumThreadEvidenceResponse(evidence, forumThread, posts, credibilityTags),
	}, nil
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
