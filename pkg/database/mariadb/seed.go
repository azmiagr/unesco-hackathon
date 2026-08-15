package mariadb

import (
	"errors"
	"strings"
	"time"

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
	if err := seedTitles(db); err != nil {
		return err
	}
	if err := backfillLegacyEquippedTitles(db); err != nil {
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

func seedTitles(db *gorm.DB) error {
	titles := []entity.Title{
		{TitleID: uuid.New(), Title: "Detektif Baru", UnlockLevel: 1, ImageBorder: "", Status: "active"},
		{TitleID: uuid.New(), Title: "Pemburu Fakta", UnlockLevel: 1, ImageBorder: "", Status: "active"},
		{TitleID: uuid.New(), Title: "Si Paling Skeptis", UnlockLevel: 1, ImageBorder: "", Status: "active"},
		{TitleID: uuid.New(), Title: "Penjaga Kota", UnlockLevel: 10, ImageBorder: "", Status: "active"},
	}

	for _, title := range titles {
		var existing entity.Title
		err := db.Where("title = ?", title.Title).First(&existing).Error
		if err == nil {
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err := db.Create(&title).Error; err != nil {
			return err
		}
	}

	return nil
}

func backfillLegacyEquippedTitles(db *gorm.DB) error {
	tx := db.Begin()
	defer tx.Rollback()

	var titles []entity.Title
	if err := tx.Where("status = ?", "active").Find(&titles).Error; err != nil {
		return err
	}
	titlesByName := make(map[string]entity.Title, len(titles))
	for _, title := range titles {
		titlesByName[title.Title] = title
	}

	var profiles []entity.UserProfile
	if err := tx.Where("title_id IS NULL AND title IS NOT NULL AND title <> ''").Find(&profiles).Error; err != nil {
		return err
	}

	now := time.Now().UTC()
	for _, profile := range profiles {
		title, ok := titlesByName[strings.TrimSpace(profile.Title)]
		if !ok {
			continue
		}

		if err := tx.Model(&entity.UserProfile{}).
			Where("user_profile_id = ?", profile.UserProfileID).
			Update("title_id", title.TitleID).Error; err != nil {
			return err
		}

		var userItem entity.UserItem
		err := tx.Where("user_id = ? AND title_id = ?", profile.UserID, title.TitleID).First(&userItem).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := tx.Create(&entity.UserItem{
				UserItemID:   uuid.New(),
				UserID:       profile.UserID,
				TitleID:      &title.TitleID,
				PurchaseType: "grant",
				PurchasedAt:  now,
				EquippedAt:   &now,
			}).Error; err != nil {
				return err
			}
			continue
		}
		if err != nil {
			return err
		}
		if err := tx.Model(&entity.UserItem{}).
			Where("user_item_id = ?", userItem.UserItemID).
			Update("equipped_at", now).Error; err != nil {
			return err
		}
	}

	return tx.Commit().Error
}
