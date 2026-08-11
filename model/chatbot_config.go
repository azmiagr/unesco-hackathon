package model

import (
	"time"

	"github.com/google/uuid"
)

type GetCaseChatbotConfigParam struct {
	CaseID uuid.UUID
}

type AdminUpsertCaseChatbotConfigRequest struct {
	BotName               string   `json:"bot_name" binding:"required"`
	BotPersonaDescription string   `json:"bot_persona_description" binding:"required"`
	KnowledgeBoundary     string   `json:"knowledge_boundary" binding:"required"`
	ProhibitedBehaviors   []string `json:"prohibited_behaviors" binding:"required"`
	SuggestedQuestions    []string `json:"suggested_questions" binding:"required"`
}

type AdminCaseChatbotConfigResponse struct {
	CaseID                uuid.UUID `json:"case_id"`
	BotName               string    `json:"bot_name"`
	BotPersonaDescription string    `json:"bot_persona_description"`
	KnowledgeBoundary     string    `json:"knowledge_boundary"`
	ProhibitedBehaviors   []string  `json:"prohibited_behaviors"`
	SuggestedQuestions    []string  `json:"suggested_questions"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type AdminGetCaseChatbotConfigResponse struct {
	Config AdminCaseChatbotConfigResponse `json:"config"`
}

type AdminUpsertCaseChatbotConfigResponse struct {
	Config AdminCaseChatbotConfigResponse `json:"config"`
}
