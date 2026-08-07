package entity

import (
	"time"

	"github.com/google/uuid"
)

type UserProfile struct {
	UserProfileID              uuid.UUID `json:"user_profile_id" gorm:"type:varchar(36);primaryKey"`
	UserID                     uuid.UUID `json:"user_id" gorm:"type:varchar(36);unique;not null"`
	AvatarID                   uuid.UUID `json:"avatar_id" gorm:"type:varchar(36)"`
	CurrentLevel               int       `json:"current_level" gorm:"type:int;default:0"`
	CurrentXP                  int       `json:"current_xp" gorm:"type:int;default:0"`
	AuditorReputation          float64   `json:"auditor_reputation" gorm:"type:decimal(8,2);default:0"`
	EvidenceEvaluationScore    float64   `json:"evidence_evaluation_score" gorm:"type:decimal(5,2);default:0"`
	ClaimAnalysisScore         float64   `json:"claim_analysis_score" gorm:"type:decimal(5,2);default:0"`
	ConfidenceCalibrationScore float64   `json:"confidence_calibration_score" gorm:"type:decimal(5,2);default:0"`
	ReasoningScore             float64   `json:"reasoning_score" gorm:"type:decimal(5,2);default:0"`
	SafetyJudgmentScore        float64   `json:"safety_judgment_score" gorm:"type:decimal(5,2);default:0"`
	CreatedAt                  time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt                  time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
