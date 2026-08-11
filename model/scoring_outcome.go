package model

import (
	"time"

	"github.com/google/uuid"
)

type AdminUpsertCaseScoringOutcomeConfigRequest struct {
	ScoringRules []AdminUpsertCaseScoringRuleRequest `json:"scoring_rules" binding:"required"`
	OutcomeRules []AdminUpsertCaseOutcomeRuleRequest `json:"outcome_rules" binding:"required"`
}

type AdminUpsertCaseScoringRuleRequest struct {
	CategoryKey       string                 `json:"category_key" binding:"required"`
	CategoryLabel     string                 `json:"category_label" binding:"required"`
	WeightBasisPoints int                    `json:"weight_basis_points"`
	Settings          map[string]interface{} `json:"settings" binding:"required"`
	SortOrder         int                    `json:"sort_order"`
}

type AdminUpsertCaseOutcomeRuleRequest struct {
	OutcomeKey         string                                    `json:"outcome_key" binding:"required"`
	OutcomeLabel       string                                    `json:"outcome_label" binding:"required"`
	ScoreMin           int                                       `json:"score_min"`
	ScoreMax           int                                       `json:"score_max"`
	OutcomeID          string                                    `json:"outcome_id" binding:"required"`
	NarrativeText      string                                    `json:"narrative_text" binding:"required"`
	SortOrder          int                                       `json:"sort_order"`
	CityImpactSettings []AdminUpsertCaseOutcomeCityImpactRequest `json:"city_impact_settings" binding:"required"`
}

type AdminUpsertCaseOutcomeCityImpactRequest struct {
	ImpactKey   string `json:"impact_key" binding:"required"`
	ImpactValue int    `json:"impact_value"`
	SortOrder   int    `json:"sort_order"`
}

type AdminCaseScoringOutcomeConfigResponse struct {
	CaseVersionID uuid.UUID                      `json:"case_version_id"`
	ScoringRules  []AdminCaseScoringRuleResponse `json:"scoring_rules"`
	OutcomeRules  []AdminCaseOutcomeRuleResponse `json:"outcome_rules"`
	CreatedAt     time.Time                      `json:"created_at"`
	UpdatedAt     time.Time                      `json:"updated_at"`
}

type AdminCaseScoringRuleResponse struct {
	CaseScoringRuleID uuid.UUID              `json:"case_scoring_rule_id"`
	CaseVersionID     uuid.UUID              `json:"case_version_id"`
	CategoryKey       string                 `json:"category_key"`
	CategoryLabel     string                 `json:"category_label"`
	WeightBasisPoints int                    `json:"weight_basis_points"`
	Settings          map[string]interface{} `json:"settings"`
	SortOrder         int                    `json:"sort_order"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

type AdminCaseOutcomeRuleResponse struct {
	CaseOutcomeRuleID  uuid.UUID                                   `json:"case_outcome_rule_id"`
	CaseVersionID      uuid.UUID                                   `json:"case_version_id"`
	OutcomeKey         string                                      `json:"outcome_key"`
	OutcomeLabel       string                                      `json:"outcome_label"`
	ScoreMin           int                                         `json:"score_min"`
	ScoreMax           int                                         `json:"score_max"`
	OutcomeID          string                                      `json:"outcome_id"`
	NarrativeText      string                                      `json:"narrative_text"`
	SortOrder          int                                         `json:"sort_order"`
	CityImpactSettings []AdminCaseOutcomeCityImpactSettingResponse `json:"city_impact_settings"`
	CreatedAt          time.Time                                   `json:"created_at"`
	UpdatedAt          time.Time                                   `json:"updated_at"`
}

type AdminCaseOutcomeCityImpactSettingResponse struct {
	CaseOutcomeCityImpactSettingID uuid.UUID `json:"case_outcome_city_impact_setting_id"`
	CaseOutcomeRuleID              uuid.UUID `json:"case_outcome_rule_id"`
	ImpactKey                      string    `json:"impact_key"`
	ImpactValue                    int       `json:"impact_value"`
	SortOrder                      int       `json:"sort_order"`
	CreatedAt                      time.Time `json:"created_at"`
	UpdatedAt                      time.Time `json:"updated_at"`
}

type AdminGetCaseScoringOutcomeConfigResponse struct {
	Config AdminCaseScoringOutcomeConfigResponse `json:"config"`
}

type AdminUpsertCaseScoringOutcomeConfigResponse struct {
	Config AdminCaseScoringOutcomeConfigResponse `json:"config"`
}
