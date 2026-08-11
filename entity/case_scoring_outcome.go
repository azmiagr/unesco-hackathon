package entity

import (
	"time"

	"github.com/google/uuid"
)

type CaseScoringOutcomeConfig struct {
	CaseVersionID uuid.UUID `json:"case_version_id" gorm:"type:varchar(36);primaryKey"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	ScoringRules []CaseScoringRule `json:"scoring_rules" gorm:"foreignKey:CaseVersionID;references:CaseVersionID;constraint:onDelete:CASCADE"`
	OutcomeRules []CaseOutcomeRule `json:"outcome_rules" gorm:"foreignKey:CaseVersionID;references:CaseVersionID;constraint:onDelete:CASCADE"`
}

type CaseScoringRule struct {
	CaseScoringRuleID uuid.UUID `json:"case_scoring_rule_id" gorm:"type:varchar(36);primaryKey"`
	CaseVersionID     uuid.UUID `json:"case_version_id" gorm:"type:varchar(36);not null;index;uniqueIndex:idx_case_scoring_rule_category"`
	CategoryKey       string    `json:"category_key" gorm:"type:varchar(80);not null;uniqueIndex:idx_case_scoring_rule_category"`
	CategoryLabel     string    `json:"category_label" gorm:"type:varchar(150);not null"`
	WeightBasisPoints int       `json:"weight_basis_points" gorm:"type:int;not null;default:0"`
	Settings          string    `json:"settings" gorm:"type:json;not null"`
	SortOrder         int       `json:"sort_order" gorm:"type:int;not null;default:0"`
	CreatedAt         time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type CaseOutcomeRule struct {
	CaseOutcomeRuleID uuid.UUID `json:"case_outcome_rule_id" gorm:"type:varchar(36);primaryKey"`
	CaseVersionID     uuid.UUID `json:"case_version_id" gorm:"type:varchar(36);not null;index;uniqueIndex:idx_case_outcome_rule_key"`
	OutcomeKey        string    `json:"outcome_key" gorm:"type:varchar(80);not null;uniqueIndex:idx_case_outcome_rule_key"`
	OutcomeLabel      string    `json:"outcome_label" gorm:"type:varchar(150);not null"`
	ScoreMin          int       `json:"score_min" gorm:"type:int;not null"`
	ScoreMax          int       `json:"score_max" gorm:"type:int;not null"`
	OutcomeID         string    `json:"outcome_id" gorm:"type:varchar(150);not null"`
	NarrativeText     string    `json:"narrative_text" gorm:"type:longtext;not null"`
	SortOrder         int       `json:"sort_order" gorm:"type:int;not null;default:0"`
	CreatedAt         time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	CityImpactSettings []CaseOutcomeCityImpactSetting `json:"city_impact_settings" gorm:"foreignKey:CaseOutcomeRuleID;references:CaseOutcomeRuleID;constraint:onDelete:CASCADE"`
}

type CaseOutcomeCityImpactSetting struct {
	CaseOutcomeCityImpactSettingID uuid.UUID `json:"case_outcome_city_impact_setting_id" gorm:"type:varchar(36);primaryKey"`
	CaseOutcomeRuleID              uuid.UUID `json:"case_outcome_rule_id" gorm:"type:varchar(36);not null;index;uniqueIndex:idx_outcome_city_impact"`
	ImpactKey                      string    `json:"impact_key" gorm:"type:varchar(80);not null;uniqueIndex:idx_outcome_city_impact"`
	ImpactValue                    int       `json:"impact_value" gorm:"type:int;not null;default:0"`
	SortOrder                      int       `json:"sort_order" gorm:"type:int;not null;default:0"`
	CreatedAt                      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt                      time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
