package repository

import (
	"github.com/azmiagr/unesco-hackathon/entity"
	"github.com/azmiagr/unesco-hackathon/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IGameLevelRepository interface {
	CreateGameLevel(tx *gorm.DB, level *entity.GameLevel) error
	GetGameLevel(tx *gorm.DB, param model.GetGameLevelParam) (*entity.GameLevel, error)
	GetGameLevelForUpdate(tx *gorm.DB, gameLevelID uuid.UUID) (*entity.GameLevel, error)
	ListGameLevels(tx *gorm.DB, param model.ListGameLevelsParam) ([]entity.GameLevel, int64, error)
	UpdateGameLevel(tx *gorm.DB, level *entity.GameLevel) error
	DeleteGameLevel(tx *gorm.DB, gameLevelID uuid.UUID) error
}

type GameLevelRepository struct {
	db *gorm.DB
}

func NewGameLevelRepository(db *gorm.DB) IGameLevelRepository {
	return &GameLevelRepository{db: db}
}

func (r *GameLevelRepository) CreateGameLevel(tx *gorm.DB, level *entity.GameLevel) error {
	err := tx.Debug().Create(level).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *GameLevelRepository) GetGameLevel(tx *gorm.DB, param model.GetGameLevelParam) (*entity.GameLevel, error) {
	var level entity.GameLevel
	query := tx.Model(&entity.GameLevel{})

	if param.GameLevelID != uuid.Nil {
		query = query.Where("game_level_id = ?", param.GameLevelID)
	}
	if param.Level > 0 {
		query = query.Where("level = ?", param.Level)
	}

	err := query.First(&level).Error
	if err != nil {
		return nil, err
	}

	return &level, nil
}

func (r *GameLevelRepository) GetGameLevelForUpdate(tx *gorm.DB, gameLevelID uuid.UUID) (*entity.GameLevel, error) {
	var level entity.GameLevel
	err := tx.Model(&entity.GameLevel{}).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("game_level_id = ?", gameLevelID).
		First(&level).Error
	if err != nil {
		return nil, err
	}
	return &level, nil
}

func (r *GameLevelRepository) ListGameLevels(tx *gorm.DB, param model.ListGameLevelsParam) ([]entity.GameLevel, int64, error) {
	var levels []entity.GameLevel
	var total int64

	query := tx.Model(&entity.GameLevel{})
	if param.Search != "" {
		search := "%" + param.Search + "%"
		query = query.Where("title LIKE ?", search)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.
		Order("level ASC").
		Limit(param.Limit).
		Offset(param.Offset).
		Find(&levels).Error
	if err != nil {
		return nil, 0, err
	}

	return levels, total, nil
}

func (r *GameLevelRepository) UpdateGameLevel(tx *gorm.DB, level *entity.GameLevel) error {
	err := tx.Debug().Save(level).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *GameLevelRepository) DeleteGameLevel(tx *gorm.DB, gameLevelID uuid.UUID) error {
	err := tx.Debug().
		Where("game_level_id = ?", gameLevelID).
		Delete(&entity.GameLevel{}).Error
	if err != nil {
		return err
	}

	return nil
}
