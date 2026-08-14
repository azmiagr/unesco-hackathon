package repository

import (
	"github.com/azmiagr/unesco-hackathon/entity"
	"github.com/azmiagr/unesco-hackathon/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ICityStatisticsRepository interface {
	GetCityStatistics(tx *gorm.DB, param model.GetCityStatisticsParam) (*entity.CityStatistics, error)
	GetCityStatisticsForUpdate(tx *gorm.DB, param model.GetCityStatisticsParam) (*entity.CityStatistics, error)
	CreateCityStatistics(tx *gorm.DB, stats *entity.CityStatistics) error
	UpdateCityStatistics(tx *gorm.DB, stats *entity.CityStatistics) error
}

type CityStatisticsRepository struct {
	db *gorm.DB
}

func NewCityStatisticsRepository(db *gorm.DB) ICityStatisticsRepository {
	return &CityStatisticsRepository{db: db}
}

func (r *CityStatisticsRepository) GetCityStatistics(tx *gorm.DB, param model.GetCityStatisticsParam) (*entity.CityStatistics, error) {
	var stats entity.CityStatistics
	query := tx.Model(&entity.CityStatistics{})

	if param.CityStatisticsID != uuid.Nil {
		query = query.Where("city_statistics_id = ?", param.CityStatisticsID)
	}
	if param.StatKey != "" {
		query = query.Where("stat_key = ?", param.StatKey)
	}

	if err := query.First(&stats).Error; err != nil {
		return nil, err
	}

	return &stats, nil
}

func (r *CityStatisticsRepository) GetCityStatisticsForUpdate(tx *gorm.DB, param model.GetCityStatisticsParam) (*entity.CityStatistics, error) {
	var stats entity.CityStatistics
	query := tx.Model(&entity.CityStatistics{}).Clauses(clause.Locking{Strength: "UPDATE"})

	if param.CityStatisticsID != uuid.Nil {
		query = query.Where("city_statistics_id = ?", param.CityStatisticsID)
	}
	if param.StatKey != "" {
		query = query.Where("stat_key = ?", param.StatKey)
	}

	if err := query.First(&stats).Error; err != nil {
		return nil, err
	}

	return &stats, nil
}

func (r *CityStatisticsRepository) CreateCityStatistics(tx *gorm.DB, stats *entity.CityStatistics) error {
	return tx.Debug().Create(stats).Error
}

func (r *CityStatisticsRepository) UpdateCityStatistics(tx *gorm.DB, stats *entity.CityStatistics) error {
	return tx.Debug().Save(stats).Error
}
