package repository

import (
	"github.com/azmiagr/unesco-hackathon/entity"
	"github.com/azmiagr/unesco-hackathon/model"
	"gorm.io/gorm"
)

type IAvatarRepository interface {
	GetAvatar(tx *gorm.DB, param model.GetAvatarParam) (*entity.Avatar, error)
	CreateAvatar(tx *gorm.DB, avatar *entity.Avatar) error
	UpdateAvatar(tx *gorm.DB, avatar *entity.Avatar) error
}

type AvatarRepository struct {
	db *gorm.DB
}

func NewAvatarRepository(db *gorm.DB) IAvatarRepository {
	return &AvatarRepository{db: db}
}

func (r *AvatarRepository) GetAvatar(tx *gorm.DB, param model.GetAvatarParam) (*entity.Avatar, error) {
	var avatar entity.Avatar
	err := tx.Where(&param).First(&avatar).Error
	if err != nil {
		return nil, err
	}

	return &avatar, nil
}

func (r *AvatarRepository) CreateAvatar(tx *gorm.DB, avatar *entity.Avatar) error {
	err := tx.Debug().Create(avatar).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *AvatarRepository) UpdateAvatar(tx *gorm.DB, avatar *entity.Avatar) error {
	err := tx.Debug().Save(avatar).Error
	if err != nil {
		return err
	}
	return nil
}
