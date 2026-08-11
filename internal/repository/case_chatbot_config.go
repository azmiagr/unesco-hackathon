package repository

import (
	"github.com/azmiagr/unesco-hackathon/entity"
	"github.com/azmiagr/unesco-hackathon/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ICaseChatbotConfigRepository interface {
	CreateCaseChatbotConfig(tx *gorm.DB, config *entity.CaseChatbotConfig) error
	GetCaseChatbotConfig(tx *gorm.DB, param model.GetCaseChatbotConfigParam) (*entity.CaseChatbotConfig, error)
	GetCaseChatbotConfigForUpdate(tx *gorm.DB, caseID uuid.UUID) (*entity.CaseChatbotConfig, error)
	CaseChatbotConfigExists(tx *gorm.DB, caseID uuid.UUID) (bool, error)
	UpdateCaseChatbotConfig(tx *gorm.DB, config *entity.CaseChatbotConfig) error
	UpsertCaseChatbotConfig(tx *gorm.DB, config *entity.CaseChatbotConfig) error
	DeleteCaseChatbotConfig(tx *gorm.DB, caseID uuid.UUID) error
}

type CaseChatbotConfigRepository struct {
	db *gorm.DB
}

func NewCaseChatbotConfigRepository(db *gorm.DB) ICaseChatbotConfigRepository {
	return &CaseChatbotConfigRepository{db: db}
}

func (r *CaseChatbotConfigRepository) CreateCaseChatbotConfig(tx *gorm.DB, config *entity.CaseChatbotConfig) error {
	err := tx.Debug().Create(config).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *CaseChatbotConfigRepository) GetCaseChatbotConfig(tx *gorm.DB, param model.GetCaseChatbotConfigParam) (*entity.CaseChatbotConfig, error) {
	var config entity.CaseChatbotConfig
	query := tx

	if param.CaseID != uuid.Nil {
		query = query.Where("case_id = ?", param.CaseID)
	}

	err := query.First(&config).Error
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (r *CaseChatbotConfigRepository) GetCaseChatbotConfigForUpdate(tx *gorm.DB, caseID uuid.UUID) (*entity.CaseChatbotConfig, error) {
	var config entity.CaseChatbotConfig

	err := tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("case_id = ?", caseID).
		First(&config).Error
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (r *CaseChatbotConfigRepository) CaseChatbotConfigExists(tx *gorm.DB, caseID uuid.UUID) (bool, error) {
	var count int64

	err := tx.Model(&entity.CaseChatbotConfig{}).
		Where("case_id = ?", caseID).
		Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *CaseChatbotConfigRepository) UpdateCaseChatbotConfig(tx *gorm.DB, config *entity.CaseChatbotConfig) error {
	err := tx.Debug().Save(config).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *CaseChatbotConfigRepository) UpsertCaseChatbotConfig(tx *gorm.DB, config *entity.CaseChatbotConfig) error {
	err := tx.Debug().
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "case_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"bot_name",
				"bot_persona_description",
				"knowledge_boundary",
				"prohibited_behaviors",
				"suggested_questions",
				"updated_at",
			}),
		}).
		Create(config).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *CaseChatbotConfigRepository) DeleteCaseChatbotConfig(tx *gorm.DB, caseID uuid.UUID) error {
	err := tx.Debug().
		Where("case_id = ?", caseID).
		Delete(&entity.CaseChatbotConfig{}).Error
	if err != nil {
		return err
	}

	return nil
}
