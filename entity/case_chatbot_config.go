package entity

import (
	"time"

	"github.com/google/uuid"
)

type CaseChatbotConfig struct {
	CaseID                uuid.UUID `json:"case_id" gorm:"type:varchar(36);primaryKey"`
	BotName               string    `json:"bot_name" gorm:"type:varchar(150);not null"`
	BotPersonaDescription string    `json:"bot_persona_description" gorm:"type:text;not null"`
	KnowledgeBoundary     string    `json:"knowledge_boundary" gorm:"type:text;not null"`
	ProhibitedBehaviors   string    `json:"prohibited_behaviors" gorm:"type:json;not null"`
	SuggestedQuestions    string    `json:"suggested_questions" gorm:"type:json;not null"`
	CreatedAt             time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt             time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
