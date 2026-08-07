package mariadb

import (
	"github.com/azmiagr/unesco-hackathon/entity"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&entity.Role{},
		&entity.User{},
		&entity.OtpCode{},
		&entity.Avatar{},
		&entity.UserProfile{},
	)

	if err != nil {
		return err
	}

	return nil
}
