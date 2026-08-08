package service

import (
	"github.com/azmiagr/unesco-hackathon/entity"
	"github.com/azmiagr/unesco-hackathon/internal/repository"
	"github.com/azmiagr/unesco-hackathon/model"
	"github.com/azmiagr/unesco-hackathon/pkg/database/mariadb"
	"gorm.io/gorm"
)

type IUserService interface {
	GetUser(param model.GetUserParam) (*entity.User, error)
	GetUserRoleName(user *entity.User) (string, error)
}

type UserService struct {
	db       *gorm.DB
	userRepo repository.IUserRepository
	roleRepo repository.IRoleRepository
}

func NewUserService(userRepo repository.IUserRepository, roleRepo repository.IRoleRepository) IUserService {
	return &UserService{
		db:       mariadb.Connection,
		userRepo: userRepo,
		roleRepo: roleRepo,
	}
}

func (s *UserService) GetUser(param model.GetUserParam) (*entity.User, error) {
	return s.userRepo.GetUser(s.db, param)
}

func (s *UserService) GetUserRoleName(user *entity.User) (string, error) {
	role, err := s.roleRepo.GetRole(s.db, user.RoleID)
	if err != nil {
		return "", err
	}

	return role.RoleName, nil
}
