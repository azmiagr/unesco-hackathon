package service

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/azmiagr/unesco-hackathon/entity"
	"github.com/azmiagr/unesco-hackathon/model"
	appErrors "github.com/azmiagr/unesco-hackathon/pkg/errors"
	"github.com/azmiagr/unesco-hackathon/pkg/helper"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *CaseService) GetCaseChatbotConfigByAdmin(caseID uuid.UUID) (*model.AdminGetCaseChatbotConfigResponse, error) {
	if caseID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid case id")
	}

	_, err := s.caseRepo.GetCase(s.db, model.GetCaseParam{CaseID: caseID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("case not found")
		}
		return nil, appErrors.InternalServer("failed to get case")
	}

	config, err := s.caseChatbotConfigRepo.GetCaseChatbotConfig(s.db, model.GetCaseChatbotConfigParam{
		CaseID: caseID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("case chatbot config not found")
		}
		return nil, appErrors.InternalServer("failed to get case chatbot config")
	}

	response, err := mapCaseChatbotConfigResponse(config)
	if err != nil {
		return nil, err
	}

	return &model.AdminGetCaseChatbotConfigResponse{
		Config: response,
	}, nil
}

func (s *CaseService) UpsertCaseChatbotConfigByAdmin(
	adminUserID uuid.UUID,
	caseID uuid.UUID,
	req model.AdminUpsertCaseChatbotConfigRequest,
) (*model.AdminUpsertCaseChatbotConfigResponse, error) {
	if adminUserID == uuid.Nil {
		return nil, appErrors.Unauthorized("unauthorized")
	}
	if caseID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid case id")
	}

	botName, err := helper.RequireTrimmedString(req.BotName, "bot name is required")
	if err != nil {
		return nil, err
	}
	if len(botName) > 150 {
		return nil, appErrors.BadRequest("bot name is too long")
	}

	persona, err := helper.RequireTrimmedString(req.BotPersonaDescription, "bot persona description is required")
	if err != nil {
		return nil, err
	}

	boundary, err := helper.RequireTrimmedString(req.KnowledgeBoundary, "knowledge boundary is required")
	if err != nil {
		return nil, err
	}

	_, prohibitedBehaviorsJSON, err := normalizeChatbotConfigItems(req.ProhibitedBehaviors, "prohibited behaviors are required")
	if err != nil {
		return nil, err
	}

	_, suggestedQuestionsJSON, err := normalizeChatbotConfigItems(req.SuggestedQuestions, "suggested questions are required")
	if err != nil {
		return nil, err
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

	var before any
	existingConfig, err := s.caseChatbotConfigRepo.GetCaseChatbotConfig(tx, model.GetCaseChatbotConfigParam{CaseID: caseID})
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appErrors.InternalServer("failed to get case chatbot config")
	}
	if existingConfig != nil {
		before = existingConfig
	}

	config := &entity.CaseChatbotConfig{
		CaseID:                caseID,
		BotName:               botName,
		BotPersonaDescription: persona,
		KnowledgeBoundary:     boundary,
		ProhibitedBehaviors:   prohibitedBehaviorsJSON,
		SuggestedQuestions:    suggestedQuestionsJSON,
	}

	err = s.caseChatbotConfigRepo.UpsertCaseChatbotConfig(tx, config)
	if err != nil {
		return nil, appErrors.InternalServer("failed to save case chatbot config")
	}

	savedConfig, err := s.caseChatbotConfigRepo.GetCaseChatbotConfig(tx, model.GetCaseChatbotConfigParam{
		CaseID: caseID,
	})
	if err != nil {
		return nil, appErrors.InternalServer("failed to get saved case chatbot config")
	}

	err = writeAdminAuditLog(tx, s.auditLogRepo, s.userRepo, adminAuditLogParam{
		ActorAdminID:  adminUserID,
		ActionType:    model.AuditActionConfigChange,
		Module:        model.AuditModuleCMS,
		TargetType:    "case_chatbot_config",
		TargetID:      caseID.String(),
		TargetLabel:   caseEntity.Title,
		Detail:        "Updated case chatbot config",
		PayloadBefore: before,
		PayloadAfter:  savedConfig,
	})
	if err != nil {
		return nil, err
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	response, err := mapCaseChatbotConfigResponse(savedConfig)
	if err != nil {
		return nil, err
	}

	return &model.AdminUpsertCaseChatbotConfigResponse{
		Config: response,
	}, nil
}

func normalizeChatbotConfigItems(items []string, requiredMessage string) ([]string, string, error) {
	normalizedItems := make([]string, 0, len(items))
	seen := map[string]bool{}

	for _, item := range items {
		normalizedItem := strings.TrimSpace(item)
		if normalizedItem == "" {
			continue
		}

		key := strings.ToLower(normalizedItem)
		if seen[key] {
			continue
		}

		seen[key] = true
		normalizedItems = append(normalizedItems, normalizedItem)
	}

	if len(normalizedItems) == 0 {
		return nil, "", appErrors.BadRequest(requiredMessage)
	}

	payload, err := json.Marshal(normalizedItems)
	if err != nil {
		return nil, "", appErrors.InternalServer("failed to normalize chatbot config items")
	}

	return normalizedItems, string(payload), nil
}

func parseChatbotConfigItems(raw string, fieldName string) ([]string, error) {
	var items []string
	err := json.Unmarshal([]byte(raw), &items)
	if err != nil {
		return nil, appErrors.InternalServer("failed to parse " + fieldName)
	}

	return items, nil
}

func mapCaseChatbotConfigResponse(config *entity.CaseChatbotConfig) (model.AdminCaseChatbotConfigResponse, error) {
	prohibitedBehaviors, err := parseChatbotConfigItems(config.ProhibitedBehaviors, "prohibited behaviors")
	if err != nil {
		return model.AdminCaseChatbotConfigResponse{}, err
	}

	suggestedQuestions, err := parseChatbotConfigItems(config.SuggestedQuestions, "suggested questions")
	if err != nil {
		return model.AdminCaseChatbotConfigResponse{}, err
	}

	return model.AdminCaseChatbotConfigResponse{
		CaseID:                config.CaseID,
		BotName:               config.BotName,
		BotPersonaDescription: config.BotPersonaDescription,
		KnowledgeBoundary:     config.KnowledgeBoundary,
		ProhibitedBehaviors:   prohibitedBehaviors,
		SuggestedQuestions:    suggestedQuestions,
		CreatedAt:             config.CreatedAt,
		UpdatedAt:             config.UpdatedAt,
	}, nil
}
