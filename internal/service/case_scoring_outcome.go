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

var allowedScoringCategories = map[string]bool{
	model.ScoringCategoryEvidenceEvaluation:    true,
	model.ScoringCategoryClaimAnalysis:         true,
	model.ScoringCategoryConfidenceCalibration: true,
	model.ScoringCategoryReasoning:             true,
	model.ScoringCategorySafetyJudgment:        true,
}

var allowedOutcomeRules = map[string]bool{
	model.OutcomeRuleExpert:     true,
	model.OutcomeRuleDeveloping: true,
	model.OutcomeRuleBeginner:   true,
}

var allowedCityImpacts = map[string]bool{
	model.CityImpactHealth:    true,
	model.CityImpactTrust:     true,
	model.CityImpactStability: true,
	model.CityImpactWellbeing: true,
}

func (s *CaseService) GetCaseScoringOutcomeConfigByAdmin(
	caseID uuid.UUID,
	caseVersionID uuid.UUID,
) (*model.AdminGetCaseScoringOutcomeConfigResponse, error) {
	if caseID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid case id")
	}
	if caseVersionID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid case version id")
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

	config, err := s.caseScoringOutcomeRepo.GetCaseScoringOutcomeConfig(s.db, caseVersionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("case scoring outcome config not found")
		}
		return nil, appErrors.InternalServer("failed to get case scoring outcome config")
	}

	response, err := mapCaseScoringOutcomeConfigResponse(config)
	if err != nil {
		return nil, err
	}

	return &model.AdminGetCaseScoringOutcomeConfigResponse{
		Config: response,
	}, nil
}

func (s *CaseService) UpsertCaseScoringOutcomeConfigByAdmin(
	adminUserID uuid.UUID,
	caseID uuid.UUID,
	caseVersionID uuid.UUID,
	req model.AdminUpsertCaseScoringOutcomeConfigRequest,
) (*model.AdminUpsertCaseScoringOutcomeConfigResponse, error) {
	if adminUserID == uuid.Nil {
		return nil, appErrors.Unauthorized("unauthorized")
	}
	if caseID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid case id")
	}
	if caseVersionID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid case version id")
	}

	scoringRules, err := buildCaseScoringRules(caseVersionID, req.ScoringRules)
	if err != nil {
		return nil, err
	}

	outcomeRules, cityImpactSettings, err := buildCaseOutcomeRules(caseVersionID, req.OutcomeRules)
	if err != nil {
		return nil, err
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
		return nil, appErrors.BadRequest("case version is not editable")
	}

	config := &entity.CaseScoringOutcomeConfig{
		CaseVersionID: caseVersionID,
	}

	err = s.caseScoringOutcomeRepo.UpsertCaseScoringOutcomeConfig(tx, config, scoringRules, outcomeRules, cityImpactSettings)
	if err != nil {
		return nil, appErrors.InternalServer("failed to save case scoring outcome config")
	}

	savedConfig, err := s.caseScoringOutcomeRepo.GetCaseScoringOutcomeConfig(tx, caseVersionID)
	if err != nil {
		return nil, appErrors.InternalServer("failed to get saved case scoring outcome config")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	response, err := mapCaseScoringOutcomeConfigResponse(savedConfig)
	if err != nil {
		return nil, err
	}

	return &model.AdminUpsertCaseScoringOutcomeConfigResponse{
		Config: response,
	}, nil
}

func buildCaseScoringRules(
	caseVersionID uuid.UUID,
	ruleRequests []model.AdminUpsertCaseScoringRuleRequest,
) ([]entity.CaseScoringRule, error) {
	if len(ruleRequests) == 0 {
		return nil, appErrors.BadRequest("scoring rules are required")
	}

	scoringRules := make([]entity.CaseScoringRule, 0, len(ruleRequests))
	seen := map[string]bool{}
	totalWeight := 0

	for index, ruleReq := range ruleRequests {
		categoryKey, err := helper.RequireTrimmedString(ruleReq.CategoryKey, "category key is required")
		if err != nil {
			return nil, err
		}
		categoryKey = strings.ToLower(categoryKey)
		if !allowedScoringCategories[categoryKey] {
			return nil, appErrors.BadRequest("invalid scoring category")
		}
		if seen[categoryKey] {
			return nil, appErrors.BadRequest("scoring category cannot contain duplicates")
		}
		seen[categoryKey] = true

		categoryLabel, err := helper.RequireTrimmedString(ruleReq.CategoryLabel, "category label is required")
		if err != nil {
			return nil, err
		}
		if len(categoryLabel) > 150 {
			return nil, appErrors.BadRequest("category label is too long")
		}
		if ruleReq.WeightBasisPoints < 0 || ruleReq.WeightBasisPoints > 10000 {
			return nil, appErrors.BadRequest("weight basis points must be between 0 and 10000")
		}

		settingsJSON, err := marshalScoringOutcomeSettings(ruleReq.Settings, "scoring rule settings are required")
		if err != nil {
			return nil, err
		}

		totalWeight += ruleReq.WeightBasisPoints
		scoringRules = append(scoringRules, entity.CaseScoringRule{
			CaseScoringRuleID: uuid.New(),
			CaseVersionID:     caseVersionID,
			CategoryKey:       categoryKey,
			CategoryLabel:     categoryLabel,
			WeightBasisPoints: ruleReq.WeightBasisPoints,
			Settings:          settingsJSON,
			SortOrder:         normalizeSortOrder(ruleReq.SortOrder, index),
		})
	}

	if totalWeight != 10000 {
		return nil, appErrors.BadRequest("scoring rule weights must total 10000")
	}

	return scoringRules, nil
}

func buildCaseOutcomeRules(
	caseVersionID uuid.UUID,
	ruleRequests []model.AdminUpsertCaseOutcomeRuleRequest,
) ([]entity.CaseOutcomeRule, []entity.CaseOutcomeCityImpactSetting, error) {
	if len(ruleRequests) == 0 {
		return nil, nil, appErrors.BadRequest("outcome rules are required")
	}

	outcomeRules := make([]entity.CaseOutcomeRule, 0, len(ruleRequests))
	cityImpactSettings := []entity.CaseOutcomeCityImpactSetting{}
	seen := map[string]bool{}

	for index, ruleReq := range ruleRequests {
		outcomeKey, err := helper.RequireTrimmedString(ruleReq.OutcomeKey, "outcome key is required")
		if err != nil {
			return nil, nil, err
		}
		outcomeKey = strings.ToLower(outcomeKey)
		if !allowedOutcomeRules[outcomeKey] {
			return nil, nil, appErrors.BadRequest("invalid outcome rule")
		}
		if seen[outcomeKey] {
			return nil, nil, appErrors.BadRequest("outcome rules cannot contain duplicates")
		}
		seen[outcomeKey] = true

		outcomeLabel, err := helper.RequireTrimmedString(ruleReq.OutcomeLabel, "outcome label is required")
		if err != nil {
			return nil, nil, err
		}
		if len(outcomeLabel) > 150 {
			return nil, nil, appErrors.BadRequest("outcome label is too long")
		}
		if ruleReq.ScoreMin < 0 || ruleReq.ScoreMin > 100 {
			return nil, nil, appErrors.BadRequest("score min must be between 0 and 100")
		}
		if ruleReq.ScoreMax < 0 || ruleReq.ScoreMax > 100 {
			return nil, nil, appErrors.BadRequest("score max must be between 0 and 100")
		}
		if ruleReq.ScoreMin > ruleReq.ScoreMax {
			return nil, nil, appErrors.BadRequest("score min cannot be greater than score max")
		}

		outcomeID, err := helper.RequireTrimmedString(ruleReq.OutcomeID, "outcome id is required")
		if err != nil {
			return nil, nil, err
		}
		narrativeText, err := helper.RequireTrimmedString(ruleReq.NarrativeText, "narrative text is required")
		if err != nil {
			return nil, nil, err
		}

		outcomeRuleID := uuid.New()
		outcomeRules = append(outcomeRules, entity.CaseOutcomeRule{
			CaseOutcomeRuleID: outcomeRuleID,
			CaseVersionID:     caseVersionID,
			OutcomeKey:        outcomeKey,
			OutcomeLabel:      outcomeLabel,
			ScoreMin:          ruleReq.ScoreMin,
			ScoreMax:          ruleReq.ScoreMax,
			OutcomeID:         outcomeID,
			NarrativeText:     narrativeText,
			SortOrder:         normalizeSortOrder(ruleReq.SortOrder, index),
		})

		settings, err := buildCaseOutcomeCityImpactSettings(outcomeRuleID, ruleReq.CityImpactSettings)
		if err != nil {
			return nil, nil, err
		}
		cityImpactSettings = append(cityImpactSettings, settings...)
	}

	return outcomeRules, cityImpactSettings, nil
}

func buildCaseOutcomeCityImpactSettings(
	outcomeRuleID uuid.UUID,
	settingRequests []model.AdminUpsertCaseOutcomeCityImpactRequest,
) ([]entity.CaseOutcomeCityImpactSetting, error) {
	if len(settingRequests) == 0 {
		return nil, appErrors.BadRequest("city impact settings are required")
	}

	settings := make([]entity.CaseOutcomeCityImpactSetting, 0, len(settingRequests))
	seen := map[string]bool{}

	for index, settingReq := range settingRequests {
		impactKey, err := helper.RequireTrimmedString(settingReq.ImpactKey, "impact key is required")
		if err != nil {
			return nil, err
		}
		impactKey = strings.ToLower(impactKey)
		if !allowedCityImpacts[impactKey] {
			return nil, appErrors.BadRequest("invalid city impact")
		}
		if seen[impactKey] {
			return nil, appErrors.BadRequest("city impact settings cannot contain duplicates")
		}
		seen[impactKey] = true

		settings = append(settings, entity.CaseOutcomeCityImpactSetting{
			CaseOutcomeCityImpactSettingID: uuid.New(),
			CaseOutcomeRuleID:              outcomeRuleID,
			ImpactKey:                      impactKey,
			ImpactValue:                    settingReq.ImpactValue,
			SortOrder:                      normalizeSortOrder(settingReq.SortOrder, index),
		})
	}

	return settings, nil
}

func marshalScoringOutcomeSettings(settings map[string]interface{}, requiredMessage string) (string, error) {
	if len(settings) == 0 {
		return "", appErrors.BadRequest(requiredMessage)
	}

	payload, err := json.Marshal(settings)
	if err != nil {
		return "", appErrors.BadRequest("settings must be valid json")
	}

	return string(payload), nil
}

func parseScoringOutcomeSettings(raw string) (map[string]interface{}, error) {
	var settings map[string]interface{}
	err := json.Unmarshal([]byte(raw), &settings)
	if err != nil {
		return nil, appErrors.InternalServer("failed to parse scoring rule settings")
	}

	return settings, nil
}

func normalizeSortOrder(sortOrder int, index int) int {
	if sortOrder > 0 {
		return sortOrder
	}

	return index + 1
}

func mapCaseScoringOutcomeConfigResponse(
	config *entity.CaseScoringOutcomeConfig,
) (model.AdminCaseScoringOutcomeConfigResponse, error) {
	scoringRules := make([]model.AdminCaseScoringRuleResponse, 0, len(config.ScoringRules))
	for _, scoringRule := range config.ScoringRules {
		settings, err := parseScoringOutcomeSettings(scoringRule.Settings)
		if err != nil {
			return model.AdminCaseScoringOutcomeConfigResponse{}, err
		}

		scoringRules = append(scoringRules, model.AdminCaseScoringRuleResponse{
			CaseScoringRuleID: scoringRule.CaseScoringRuleID,
			CaseVersionID:     scoringRule.CaseVersionID,
			CategoryKey:       scoringRule.CategoryKey,
			CategoryLabel:     scoringRule.CategoryLabel,
			WeightBasisPoints: scoringRule.WeightBasisPoints,
			Settings:          settings,
			SortOrder:         scoringRule.SortOrder,
			CreatedAt:         scoringRule.CreatedAt,
			UpdatedAt:         scoringRule.UpdatedAt,
		})
	}

	outcomeRules := make([]model.AdminCaseOutcomeRuleResponse, 0, len(config.OutcomeRules))
	for _, outcomeRule := range config.OutcomeRules {
		cityImpactSettings := make([]model.AdminCaseOutcomeCityImpactSettingResponse, 0, len(outcomeRule.CityImpactSettings))
		for _, cityImpactSetting := range outcomeRule.CityImpactSettings {
			cityImpactSettings = append(cityImpactSettings, model.AdminCaseOutcomeCityImpactSettingResponse{
				CaseOutcomeCityImpactSettingID: cityImpactSetting.CaseOutcomeCityImpactSettingID,
				CaseOutcomeRuleID:              cityImpactSetting.CaseOutcomeRuleID,
				ImpactKey:                      cityImpactSetting.ImpactKey,
				ImpactValue:                    cityImpactSetting.ImpactValue,
				SortOrder:                      cityImpactSetting.SortOrder,
				CreatedAt:                      cityImpactSetting.CreatedAt,
				UpdatedAt:                      cityImpactSetting.UpdatedAt,
			})
		}

		outcomeRules = append(outcomeRules, model.AdminCaseOutcomeRuleResponse{
			CaseOutcomeRuleID:  outcomeRule.CaseOutcomeRuleID,
			CaseVersionID:      outcomeRule.CaseVersionID,
			OutcomeKey:         outcomeRule.OutcomeKey,
			OutcomeLabel:       outcomeRule.OutcomeLabel,
			ScoreMin:           outcomeRule.ScoreMin,
			ScoreMax:           outcomeRule.ScoreMax,
			OutcomeID:          outcomeRule.OutcomeID,
			NarrativeText:      outcomeRule.NarrativeText,
			SortOrder:          outcomeRule.SortOrder,
			CityImpactSettings: cityImpactSettings,
			CreatedAt:          outcomeRule.CreatedAt,
			UpdatedAt:          outcomeRule.UpdatedAt,
		})
	}

	return model.AdminCaseScoringOutcomeConfigResponse{
		CaseVersionID: config.CaseVersionID,
		ScoringRules:  scoringRules,
		OutcomeRules:  outcomeRules,
		CreatedAt:     config.CreatedAt,
		UpdatedAt:     config.UpdatedAt,
	}, nil
}
