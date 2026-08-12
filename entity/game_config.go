package entity

import (
	"time"

	"github.com/google/uuid"
)

type GameConfig struct {
	GameConfigID                uuid.UUID `json:"game_config_id" gorm:"type:varchar(36);primaryKey"`
	ConfigKey                   string    `json:"config_key" gorm:"type:varchar(80);not null;uniqueIndex"`
	MaxCasesPerDay              int       `json:"max_cases_per_day" gorm:"type:int;not null;default:5"`
	CooldownBetweenCasesMinutes int       `json:"cooldown_between_cases_minutes" gorm:"type:int;not null;default:15"`
	StreakBonusMultiplier       float64   `json:"streak_bonus_multiplier" gorm:"type:decimal(8,2);not null;default:1.50"`
	MaintenanceMode             bool      `json:"maintenance_mode" gorm:"type:boolean;not null;default:false"`
	CompleteCaseBaseMultiplier  float64   `json:"complete_case_base_multiplier" gorm:"type:decimal(8,2);not null;default:1.00"`
	PerfectScoreBonusMultiplier float64   `json:"perfect_score_bonus_multiplier" gorm:"type:decimal(8,2);not null;default:1.50"`
	DailyLoginRewardXP          int       `json:"daily_login_reward_xp" gorm:"type:int;not null;default:100"`
	StreakBonusCapMultiplier    float64   `json:"streak_bonus_cap_multiplier" gorm:"type:decimal(8,2);not null;default:1.20"`
	DefaultAIProvider           string    `json:"default_ai_provider" gorm:"type:varchar(120);not null;default:'openai'"`
	OpenAIAPISecretKey          *string   `json:"openai_api_secret_key" gorm:"type:text"`
	DailyTokenBudget            int       `json:"daily_token_budget" gorm:"type:int;not null;default:2500000"`
	PerCaseTokenCap             int       `json:"per_case_token_cap" gorm:"type:int;not null;default:15000"`
	EnableFailoverSystem        bool      `json:"enable_failover_system" gorm:"type:boolean;not null;default:true"`
	FallbackProvider            string    `json:"fallback_provider" gorm:"type:varchar(120);not null;default:'anthropic'"`
	ErrorThresholdPercent       int       `json:"error_threshold_percent" gorm:"type:int;not null;default:5"`
	SystemPromptTemplate        string    `json:"system_prompt_template" gorm:"type:longtext;not null"`
	AITemperature               float64   `json:"ai_temperature" gorm:"type:decimal(4,2);not null;default:0.30"`
	CreatedAt                   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt                   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
