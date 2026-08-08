package repository

import (
	"github.com/azmiagr/unesco-hackathon/entity"
	"github.com/azmiagr/unesco-hackathon/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IUserRepository interface {
	GetUser(tx *gorm.DB, param model.GetUserParam) (*entity.User, error)
	CreateUser(tx *gorm.DB, user *entity.User) error
	UpdateUser(tx *gorm.DB, user *entity.User) error
}

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) IUserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetUser(tx *gorm.DB, param model.GetUserParam) (*entity.User, error) {
	var user entity.User
	query := tx

	if param.UserID != uuid.Nil {
		query = query.Where("user_id = ?", param.UserID)
	}
	if param.Email != "" {
		query = query.Where("email = ?", param.Email)
	}
	if param.Username != "" {
		query = query.Where("username = ?", param.Username)
	}

	if err := query.First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) CreateUser(tx *gorm.DB, user *entity.User) error {
	err := tx.Debug().Create(user).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *UserRepository) UpdateUser(tx *gorm.DB, user *entity.User) error {
	err := tx.Debug().Save(user).Error
	if err != nil {
		return err
	}
	return nil
}
