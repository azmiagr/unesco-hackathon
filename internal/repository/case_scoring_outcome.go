package repository

import (
	"github.com/azmiagr/unesco-hackathon/entity"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ICaseScoringOutcomeRepository interface {
	CreateCaseScoringOutcomeConfig(tx *gorm.DB, config *entity.CaseScoringOutcomeConfig, scoringRules []entity.CaseScoringRule, outcomeRules []entity.CaseOutcomeRule, cityImpactSettings []entity.CaseOutcomeCityImpactSetting) error
	GetCaseScoringOutcomeConfig(tx *gorm.DB, caseVersionID uuid.UUID) (*entity.CaseScoringOutcomeConfig, error)
	GetCaseScoringOutcomeConfigForUpdate(tx *gorm.DB, caseVersionID uuid.UUID) (*entity.CaseScoringOutcomeConfig, error)
	UpsertCaseScoringOutcomeConfig(tx *gorm.DB, config *entity.CaseScoringOutcomeConfig, scoringRules []entity.CaseScoringRule, outcomeRules []entity.CaseOutcomeRule, cityImpactSettings []entity.CaseOutcomeCityImpactSetting) error
	DeleteCaseScoringOutcomeConfig(tx *gorm.DB, caseVersionID uuid.UUID) error
}

type CaseScoringOutcomeRepository struct {
	db *gorm.DB
}

func NewCaseScoringOutcomeRepository(db *gorm.DB) ICaseScoringOutcomeRepository {
	return &CaseScoringOutcomeRepository{db: db}
}

func (r *CaseScoringOutcomeRepository) CreateCaseScoringOutcomeConfig(
	tx *gorm.DB,
	config *entity.CaseScoringOutcomeConfig,
	scoringRules []entity.CaseScoringRule,
	outcomeRules []entity.CaseOutcomeRule,
	cityImpactSettings []entity.CaseOutcomeCityImpactSetting,
) error {
	err := tx.Debug().Create(config).Error
	if err != nil {
		return err
	}

	return r.createCaseScoringOutcomeChildren(tx, scoringRules, outcomeRules, cityImpactSettings)
}

func (r *CaseScoringOutcomeRepository) GetCaseScoringOutcomeConfig(
	tx *gorm.DB,
	caseVersionID uuid.UUID,
) (*entity.CaseScoringOutcomeConfig, error) {
	var config entity.CaseScoringOutcomeConfig

	err := preloadCaseScoringOutcomeConfig(tx).
		Where("case_version_id = ?", caseVersionID).
		First(&config).Error
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (r *CaseScoringOutcomeRepository) GetCaseScoringOutcomeConfigForUpdate(
	tx *gorm.DB,
	caseVersionID uuid.UUID,
) (*entity.CaseScoringOutcomeConfig, error) {
	var config entity.CaseScoringOutcomeConfig

	err := preloadCaseScoringOutcomeConfig(tx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("case_version_id = ?", caseVersionID).
		First(&config).Error
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (r *CaseScoringOutcomeRepository) UpsertCaseScoringOutcomeConfig(
	tx *gorm.DB,
	config *entity.CaseScoringOutcomeConfig,
	scoringRules []entity.CaseScoringRule,
	outcomeRules []entity.CaseOutcomeRule,
	cityImpactSettings []entity.CaseOutcomeCityImpactSetting,
) error {
	err := tx.Debug().
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "case_version_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"updated_at",
			}),
		}).
		Create(config).Error
	if err != nil {
		return err
	}

	err = r.deleteCaseScoringOutcomeChildren(tx, config.CaseVersionID)
	if err != nil {
		return err
	}

	return r.createCaseScoringOutcomeChildren(tx, scoringRules, outcomeRules, cityImpactSettings)
}

func (r *CaseScoringOutcomeRepository) DeleteCaseScoringOutcomeConfig(tx *gorm.DB, caseVersionID uuid.UUID) error {
	err := r.deleteCaseScoringOutcomeChildren(tx, caseVersionID)
	if err != nil {
		return err
	}

	err = tx.Debug().
		Where("case_version_id = ?", caseVersionID).
		Delete(&entity.CaseScoringOutcomeConfig{}).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *CaseScoringOutcomeRepository) createCaseScoringOutcomeChildren(
	tx *gorm.DB,
	scoringRules []entity.CaseScoringRule,
	outcomeRules []entity.CaseOutcomeRule,
	cityImpactSettings []entity.CaseOutcomeCityImpactSetting,
) error {
	if len(scoringRules) > 0 {
		err := tx.Debug().Create(&scoringRules).Error
		if err != nil {
			return err
		}
	}

	if len(outcomeRules) > 0 {
		err := tx.Debug().Create(&outcomeRules).Error
		if err != nil {
			return err
		}
	}

	if len(cityImpactSettings) > 0 {
		err := tx.Debug().Create(&cityImpactSettings).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *CaseScoringOutcomeRepository) deleteCaseScoringOutcomeChildren(tx *gorm.DB, caseVersionID uuid.UUID) error {
	outcomeRuleIDs := tx.
		Model(&entity.CaseOutcomeRule{}).
		Select("case_outcome_rule_id").
		Where("case_version_id = ?", caseVersionID)

	err := tx.Debug().
		Where("case_outcome_rule_id IN (?)", outcomeRuleIDs).
		Delete(&entity.CaseOutcomeCityImpactSetting{}).Error
	if err != nil {
		return err
	}

	err = tx.Debug().
		Where("case_version_id = ?", caseVersionID).
		Delete(&entity.CaseOutcomeRule{}).Error
	if err != nil {
		return err
	}

	err = tx.Debug().
		Where("case_version_id = ?", caseVersionID).
		Delete(&entity.CaseScoringRule{}).Error
	if err != nil {
		return err
	}

	return nil
}

func preloadCaseScoringOutcomeConfig(query *gorm.DB) *gorm.DB {
	return query.
		Preload("ScoringRules", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC").Order("created_at ASC")
		}).
		Preload("OutcomeRules", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC").Order("created_at ASC")
		}).
		Preload("OutcomeRules.CityImpactSettings", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC").Order("created_at ASC")
		})
}
