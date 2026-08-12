package repository

import (
	"github.com/azmiagr/unesco-hackathon/entity"
	"github.com/azmiagr/unesco-hackathon/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IGameConfigRepository interface {
	CreateGameConfig(tx *gorm.DB, config *entity.GameConfig) error
	GetGameConfig(tx *gorm.DB, param model.GetGameConfigParam) (*entity.GameConfig, error)
	GetGameConfigForUpdate(tx *gorm.DB, param model.GetGameConfigParam) (*entity.GameConfig, error)
	UpsertGameConfig(tx *gorm.DB, config *entity.GameConfig) error
	UpdateGameConfig(tx *gorm.DB, config *entity.GameConfig) error
}

type GameConfigRepository struct {
	db *gorm.DB
}

func NewGameConfigRepository(db *gorm.DB) IGameConfigRepository {
	return &GameConfigRepository{db: db}
}

func (r *GameConfigRepository) CreateGameConfig(tx *gorm.DB, config *entity.GameConfig) error {
	err := tx.Debug().Create(config).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *GameConfigRepository) GetGameConfig(tx *gorm.DB, param model.GetGameConfigParam) (*entity.GameConfig, error) {
	var config entity.GameConfig
	query := tx.Model(&entity.GameConfig{})

	if param.GameConfigID != uuid.Nil {
		query = query.Where("game_config_id = ?", param.GameConfigID)
	}
	if param.ConfigKey != "" {
		query = query.Where("config_key = ?", param.ConfigKey)
	}

	err := query.First(&config).Error
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (r *GameConfigRepository) GetGameConfigForUpdate(tx *gorm.DB, param model.GetGameConfigParam) (*entity.GameConfig, error) {
	var config entity.GameConfig
	query := tx.Model(&entity.GameConfig{}).Clauses(clause.Locking{Strength: "UPDATE"})

	if param.GameConfigID != uuid.Nil {
		query = query.Where("game_config_id = ?", param.GameConfigID)
	}
	if param.ConfigKey != "" {
		query = query.Where("config_key = ?", param.ConfigKey)
	}

	err := query.First(&config).Error
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (r *GameConfigRepository) UpsertGameConfig(tx *gorm.DB, config *entity.GameConfig) error {
	err := tx.Debug().
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "config_key"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"max_cases_per_day",
				"cooldown_between_cases_minutes",
				"streak_bonus_multiplier",
				"maintenance_mode",
				"complete_case_base_multiplier",
				"perfect_score_bonus_multiplier",
				"daily_login_reward_xp",
				"streak_bonus_cap_multiplier",
				"default_ai_provider",
				"openai_api_secret_key",
				"daily_token_budget",
				"per_case_token_cap",
				"enable_failover_system",
				"fallback_provider",
				"error_threshold_percent",
				"system_prompt_template",
				"ai_temperature",
				"updated_at",
			}),
		}).
		Create(config).Error

	if err != nil {
		return err
	}

	return nil
}

func (r *GameConfigRepository) UpdateGameConfig(tx *gorm.DB, config *entity.GameConfig) error {
	err := tx.Debug().Save(config).Error
	if err != nil {
		return err
	}

	return nil
}
