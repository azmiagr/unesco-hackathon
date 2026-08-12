package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	GameConfigDefaultKey = "default"
)

type GetGameConfigParam struct {
	GameConfigID uuid.UUID
	ConfigKey    string
}

type GetGameLevelParam struct {
	GameLevelID uuid.UUID
	Level       int
}

type ListGameLevelsParam struct {
	Search string
	Limit  int
	Offset int
}

type AdminUpsertGameConfigRequest struct {
	MaxCasesPerDay              int     `json:"max_cases_per_day" binding:"required"`
	CooldownBetweenCasesMinutes int     `json:"cooldown_between_cases_minutes" binding:"required"`
	StreakBonusMultiplier       float64 `json:"streak_bonus_multiplier" binding:"required"`
	MaintenanceMode             bool    `json:"maintenance_mode"`
	CompleteCaseBaseMultiplier  float64 `json:"complete_case_base_multiplier" binding:"required"`
	PerfectScoreBonusMultiplier float64 `json:"perfect_score_bonus_multiplier" binding:"required"`
	DailyLoginRewardXP          int     `json:"daily_login_reward_xp" binding:"required"`
	StreakBonusCapMultiplier    float64 `json:"streak_bonus_cap_multiplier" binding:"required"`
	DefaultAIProvider           string  `json:"default_ai_provider" binding:"required"`
	OpenAIAPISecretKey          *string `json:"openai_api_secret_key"`
	DailyTokenBudget            int     `json:"daily_token_budget" binding:"required"`
	PerCaseTokenCap             int     `json:"per_case_token_cap" binding:"required"`
	EnableFailoverSystem        bool    `json:"enable_failover_system"`
	FallbackProvider            string  `json:"fallback_provider" binding:"required"`
	ErrorThresholdPercent       int     `json:"error_threshold_percent" binding:"required"`
	SystemPromptTemplate        string  `json:"system_prompt_template" binding:"required"`
	AITemperature               float64 `json:"ai_temperature" binding:"required"`
}

type AdminUpsertGameGeneralConfigRequest struct {
	MaxCasesPerDay              int     `json:"max_cases_per_day" binding:"required"`
	CooldownBetweenCasesMinutes int     `json:"cooldown_between_cases_minutes" binding:"required"`
	StreakBonusMultiplier       float64 `json:"streak_bonus_multiplier" binding:"required"`
	MaintenanceMode             bool    `json:"maintenance_mode"`
}

type AdminUpsertGameAIConfigRequest struct {
	DefaultAIProvider     string  `json:"default_ai_provider" binding:"required"`
	OpenAIAPISecretKey    *string `json:"openai_api_secret_key"`
	DailyTokenBudget      int     `json:"daily_token_budget" binding:"required"`
	PerCaseTokenCap       int     `json:"per_case_token_cap" binding:"required"`
	EnableFailoverSystem  bool    `json:"enable_failover_system"`
	FallbackProvider      string  `json:"fallback_provider" binding:"required"`
	ErrorThresholdPercent int     `json:"error_threshold_percent" binding:"required"`
	SystemPromptTemplate  string  `json:"system_prompt_template" binding:"required"`
	AITemperature         float64 `json:"ai_temperature" binding:"required"`
}

type AdminGameConfigResponse struct {
	GameConfigID                uuid.UUID `json:"game_config_id"`
	ConfigKey                   string    `json:"config_key"`
	MaxCasesPerDay              int       `json:"max_cases_per_day"`
	CooldownBetweenCasesMinutes int       `json:"cooldown_between_cases_minutes"`
	StreakBonusMultiplier       float64   `json:"streak_bonus_multiplier"`
	MaintenanceMode             bool      `json:"maintenance_mode"`
	CompleteCaseBaseMultiplier  float64   `json:"complete_case_base_multiplier"`
	PerfectScoreBonusMultiplier float64   `json:"perfect_score_bonus_multiplier"`
	DailyLoginRewardXP          int       `json:"daily_login_reward_xp"`
	StreakBonusCapMultiplier    float64   `json:"streak_bonus_cap_multiplier"`
	DefaultAIProvider           string    `json:"default_ai_provider"`
	OpenAIAPISecretKey          *string   `json:"openai_api_secret_key"`
	DailyTokenBudget            int       `json:"daily_token_budget"`
	PerCaseTokenCap             int       `json:"per_case_token_cap"`
	EnableFailoverSystem        bool      `json:"enable_failover_system"`
	FallbackProvider            string    `json:"fallback_provider"`
	ErrorThresholdPercent       int       `json:"error_threshold_percent"`
	SystemPromptTemplate        string    `json:"system_prompt_template"`
	AITemperature               float64   `json:"ai_temperature"`
	CreatedAt                   time.Time `json:"created_at"`
	UpdatedAt                   time.Time `json:"updated_at"`
}

type AdminGameGeneralConfigResponse struct {
	GameConfigID                uuid.UUID `json:"game_config_id"`
	ConfigKey                   string    `json:"config_key"`
	MaxCasesPerDay              int       `json:"max_cases_per_day"`
	CooldownBetweenCasesMinutes int       `json:"cooldown_between_cases_minutes"`
	StreakBonusMultiplier       float64   `json:"streak_bonus_multiplier"`
	MaintenanceMode             bool      `json:"maintenance_mode"`
	CreatedAt                   time.Time `json:"created_at"`
	UpdatedAt                   time.Time `json:"updated_at"`
}

type AdminGameAIConfigResponse struct {
	GameConfigID          uuid.UUID `json:"game_config_id"`
	ConfigKey             string    `json:"config_key"`
	DefaultAIProvider     string    `json:"default_ai_provider"`
	OpenAIAPISecretKey    *string   `json:"openai_api_secret_key"`
	DailyTokenBudget      int       `json:"daily_token_budget"`
	PerCaseTokenCap       int       `json:"per_case_token_cap"`
	EnableFailoverSystem  bool      `json:"enable_failover_system"`
	FallbackProvider      string    `json:"fallback_provider"`
	ErrorThresholdPercent int       `json:"error_threshold_percent"`
	SystemPromptTemplate  string    `json:"system_prompt_template"`
	AITemperature         float64   `json:"ai_temperature"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type AdminGetGameConfigResponse struct {
	Config AdminGameConfigResponse `json:"config"`
}

type AdminUpsertGameConfigResponse struct {
	Config AdminGameConfigResponse `json:"config"`
}

type AdminGetGameGeneralConfigResponse struct {
	Config AdminGameGeneralConfigResponse `json:"config"`
}

type AdminUpsertGameGeneralConfigResponse struct {
	Config AdminGameGeneralConfigResponse `json:"config"`
}

type AdminGetGameAIConfigResponse struct {
	Config AdminGameAIConfigResponse `json:"config"`
}

type AdminUpsertGameAIConfigResponse struct {
	Config AdminGameAIConfigResponse `json:"config"`
}

type AdminListGameLevelsRequest struct {
	Search string `form:"search"`
	Page   int    `form:"page"`
	Limit  int    `form:"limit"`
}

type AdminCreateGameLevelRequest struct {
	Level      int    `json:"level" binding:"required"`
	XPRequired int    `json:"xp_required"`
	Title      string `json:"title" binding:"required"`
	RewardCoin int    `json:"reward_coin"`
}

type AdminUpdateGameLevelRequest = AdminCreateGameLevelRequest

type AdminGameLevelResponse struct {
	GameLevelID uuid.UUID `json:"game_level_id"`
	Level       int       `json:"level"`
	XPRequired  int       `json:"xp_required"`
	Title       string    `json:"title"`
	RewardCoin  int       `json:"reward_coin"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type AdminListGameLevelsResponse struct {
	Levels     []AdminGameLevelResponse `json:"levels"`
	Pagination PaginationResponse       `json:"pagination"`
}

type AdminGetGameLevelDetailResponse struct {
	Level AdminGameLevelResponse `json:"level"`
}

type AdminCreateGameLevelResponse struct {
	Level AdminGameLevelResponse `json:"level"`
}

type AdminUpdateGameLevelResponse struct {
	Level AdminGameLevelResponse `json:"level"`
}

type AdminDeleteGameLevelResponse struct {
	GameLevelID uuid.UUID `json:"game_level_id"`
}
