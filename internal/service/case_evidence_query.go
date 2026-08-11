package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/azmiagr/unesco-hackathon/model"
	appErrors "github.com/azmiagr/unesco-hackathon/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

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

func (s *CaseService) ListEvidenceOptionsByAdmin(caseID uuid.UUID) (*model.AdminListEvidenceOptionsResponse, error) {
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

	options := make([]model.AdminEvidenceOptionResponse, 0, len(evidences))
	for i, evidence := range evidences {
		options = append(options, model.AdminEvidenceOptionResponse{
			CaseEvidenceID: evidence.CaseEvidenceID,
			Code:           formatEvidenceCode(i + 1),
			Label:          evidence.Label,
			TemplateType:   evidence.TemplateType,
			SortOrder:      evidence.SortOrder,
		})
	}

	return &model.AdminListEvidenceOptionsResponse{
		CaseID:        caseDetail.CaseID,
		CaseVersionID: caseDetail.CurrentCaseVersionID,
		Total:         len(options),
		Evidences:     options,
	}, nil
}

func (s *CaseService) GetCaseEvidenceDetailByAdmin(
	caseID uuid.UUID,
	caseVersionID uuid.UUID,
	caseEvidenceID uuid.UUID,
) (*model.AdminEvidenceDetailResponse, error) {
	if caseID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid case id")
	}
	if caseVersionID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid case version id")
	}
	if caseEvidenceID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid case evidence id")
	}

	caseVersion, err := s.caseVersionRepo.GetCaseVersion(s.db, model.GetCaseVersionParam{
		CaseVersionID: caseVersionID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("case version not found")
		}
		return nil, appErrors.InternalServer("failed to get case version")
	}

	if caseVersion.CaseID != caseID {
		return nil, appErrors.NotFound("case version not found")
	}

	evidence, err := s.caseEvidenceRepo.GetCaseEvidence(s.db, model.GetCaseEvidenceParam{
		CaseEvidenceID: caseEvidenceID,
		CaseVersionID:  caseVersionID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("case evidence not found")
		}
		return nil, appErrors.InternalServer("failed to get case evidence")
	}

	credibilityTags, err := parseStoredCredibilityTags(evidence.CredibilityTags)
	if err != nil {
		return nil, err
	}

	result := &model.AdminEvidenceDetailResponse{
		TemplateType: evidence.TemplateType,
	}

	switch evidence.TemplateType {
	case model.CaseEvidenceTemplateSocialPost:
		if evidence.SocialPost == nil {
			return nil, appErrors.InternalServer("social post evidence detail not found")
		}
		socialPost := mapSocialPostEvidenceResponse(evidence, evidence.SocialPost, credibilityTags)
		result.SocialPost = &socialPost
	case model.CaseEvidenceTemplateArticle:
		if evidence.Article == nil {
			return nil, appErrors.InternalServer("article evidence detail not found")
		}
		article := mapArticleEvidenceResponse(evidence, evidence.Article, credibilityTags)
		result.Article = &article
	case model.CaseEvidenceTemplateBlog:
		if evidence.Blog == nil {
			return nil, appErrors.InternalServer("blog evidence detail not found")
		}
		blog := mapBlogEvidenceResponse(evidence, evidence.Blog, credibilityTags)
		result.Blog = &blog
	case model.CaseEvidenceTemplateForumThread:
		if evidence.ForumThread == nil {
			return nil, appErrors.InternalServer("forum thread evidence detail not found")
		}
		forumThread := mapForumThreadEvidenceResponse(evidence, evidence.ForumThread, evidence.ForumThread.Posts, credibilityTags)
		result.ForumThread = &forumThread
	case model.CaseEvidenceTemplateChatTranscript:
		if evidence.ChatTranscript == nil {
			return nil, appErrors.InternalServer("chat transcript evidence detail not found")
		}
		chatTranscript := mapChatTranscriptEvidenceResponse(evidence, evidence.ChatTranscript.Participants, evidence.ChatTranscript.Messages, credibilityTags)
		result.ChatTranscript = &chatTranscript
	case model.CaseEvidenceTemplatePublicAnnouncement:
		if evidence.PublicAnnouncement == nil {
			return nil, appErrors.InternalServer("public announcement evidence detail not found")
		}
		publicAnnouncement := mapPublicAnnouncementEvidenceResponse(evidence, evidence.PublicAnnouncement, credibilityTags)
		result.PublicAnnouncement = &publicAnnouncement
	default:
		return nil, appErrors.BadRequest("unsupported evidence template type")
	}

	return result, nil
}

func buildEvidenceOptionMap(evidences []model.AdminCaseEvidenceListRow) map[uuid.UUID]model.AdminEvidenceOptionResponse {
	options := map[uuid.UUID]model.AdminEvidenceOptionResponse{}
	for i, evidence := range evidences {
		options[evidence.CaseEvidenceID] = model.AdminEvidenceOptionResponse{
			CaseEvidenceID: evidence.CaseEvidenceID,
			Code:           formatEvidenceCode(i + 1),
			Label:          evidence.Label,
			TemplateType:   evidence.TemplateType,
			SortOrder:      evidence.SortOrder,
		}
	}

	return options
}

func formatEvidenceCode(index int) string {
	return fmt.Sprintf("EV-%02d", index)
}

func parseStoredCredibilityTags(raw string) ([]string, error) {
	var tags []string
	if err := json.Unmarshal([]byte(raw), &tags); err != nil {
		return nil, appErrors.InternalServer("failed to parse credibility tags")
	}

	return tags, nil
}
