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
