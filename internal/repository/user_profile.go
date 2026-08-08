package repository

import (
	"github.com/azmiagr/unesco-hackathon/entity"
	"github.com/azmiagr/unesco-hackathon/model"
	"gorm.io/gorm"
)

type IUserProfileRepository interface {
	GetUserProfile(tx *gorm.DB, param model.GetUserProfileParam) (*entity.UserProfile, error)
	CreateUserProfile(tx *gorm.DB, userProfile *entity.UserProfile) error
	UpdateUserProfile(tx *gorm.DB, userProfile *entity.UserProfile) error
}

type UserProfileRepository struct {
	db *gorm.DB
}

func NewUserProfileRepository(db *gorm.DB) IUserProfileRepository {
	return &UserProfileRepository{db: db}
}

func (r *UserProfileRepository) GetUserProfile(tx *gorm.DB, param model.GetUserProfileParam) (*entity.UserProfile, error) {
	var userProfile entity.UserProfile
	err := tx.Where(&param).First(&userProfile).Error
	if err != nil {
		return nil, err
	}

	return &userProfile, nil
}

func (r *UserProfileRepository) CreateUserProfile(tx *gorm.DB, userProfile *entity.UserProfile) error {
	err := tx.Debug().Create(userProfile).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *UserProfileRepository) UpdateUserProfile(tx *gorm.DB, userProfile *entity.UserProfile) error {
	err := tx.Debug().Save(userProfile).Error
	if err != nil {
		return err
	}

	return nil
}
