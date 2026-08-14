package mariadb

import (
	"errors"

	"github.com/azmiagr/unesco-hackathon/entity"
	constants "github.com/azmiagr/unesco-hackathon/pkg/constant"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func Seed(db *gorm.DB) error {
	if err := seedRoles(db); err != nil {
		return err
	}
	if err := seedCityStatistics(db); err != nil {
		return err
	}

	return nil
}

func seedRoles(db *gorm.DB) error {
	roles := []entity.Role{
		{
			RoleID:   uuid.New(),
			RoleName: constants.RoleUser,
		},
		{
			RoleID:   uuid.New(),
			RoleName: constants.RoleAdmin,
		},
	}

	for _, role := range roles {
		var existing entity.Role
		err := db.Where("role_name = ?", role.RoleName).First(&existing).Error
		if err == nil {
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if err := db.Create(&role).Error; err != nil {
			return err
		}
	}

	return nil
}

func seedCityStatistics(db *gorm.DB) error {
	stats := entity.CityStatistics{
		CityStatisticsID:  uuid.New(),
		StatKey:           "default",
		InformationHealth: 70,
		PublicTrust:       70,
		SocialStability:   70,
		PublicWellbeing:   70,
		VisualState:       "normal",
	}

	var existing entity.CityStatistics
	err := db.Where("stat_key = ?", stats.StatKey).First(&existing).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return db.Create(&stats).Error
}
