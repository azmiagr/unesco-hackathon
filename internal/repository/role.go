package repository

import (
	"github.com/azmiagr/unesco-hackathon/entity"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IRoleRepository interface {
	GetRole(tx *gorm.DB, roleID uuid.UUID) (*entity.Role, error)
	GetRoleByName(tx *gorm.DB, roleName string) (*entity.Role, error)
	GetAllRole(tx *gorm.DB) ([]*entity.Role, error)
}

type RoleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) IRoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) GetRole(tx *gorm.DB, roleID uuid.UUID) (*entity.Role, error) {
	var role entity.Role
	err := tx.Where("role_id = ?", roleID).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepository) GetRoleByName(tx *gorm.DB, roleName string) (*entity.Role, error) {
	var role entity.Role
	err := tx.Where("role_name = ?", roleName).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepository) GetAllRole(tx *gorm.DB) ([]*entity.Role, error) {
	var roles []*entity.Role
	err := tx.Find(&roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}
