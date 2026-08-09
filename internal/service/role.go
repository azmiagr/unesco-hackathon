package service

import (
	"github.com/azmiagr/unesco-hackathon/internal/repository"
	"github.com/azmiagr/unesco-hackathon/model"
	"github.com/azmiagr/unesco-hackathon/pkg/database/mariadb"
	appErrors "github.com/azmiagr/unesco-hackathon/pkg/errors"
	"gorm.io/gorm"
)

type IRoleService interface {
	GetAllRoles() ([]*model.RoleResponse, error)
}

type RoleService struct {
	db       *gorm.DB
	roleRepo repository.IRoleRepository
}

func NewRoleService(roleRepo repository.IRoleRepository) IRoleService {
	return &RoleService{
		db:       mariadb.Connection,
		roleRepo: roleRepo,
	}
}

func (s *RoleService) GetAllRoles() ([]*model.RoleResponse, error) {
	var response []*model.RoleResponse

	roles, err := s.roleRepo.GetAllRole(s.db)
	if err != nil {
		return nil, appErrors.InternalServer("failed to get all roles")
	}

	for _, role := range roles {
		response = append(response, &model.RoleResponse{
			RoleID:   role.RoleID,
			RoleName: role.RoleName,
		})
	}

	return response, nil
}
