package service

import (
	"github.com/azmiagr/unesco-hackathon/internal/repository"
	"github.com/azmiagr/unesco-hackathon/pkg/bcrypt"
	"github.com/azmiagr/unesco-hackathon/pkg/jwt"
	"github.com/azmiagr/unesco-hackathon/pkg/supabase"
)

type Service struct {
	UserService IUserService
	AuthService IAuthService
	RoleService IRoleService
	CaseService ICaseService
}

func NewService(repository *repository.Repository, bcrypt bcrypt.Interface, jwtAuth jwt.Interface, supabase supabase.Interface) *Service {
	userService := NewUserService(repository.UserRepository, repository.RoleRepository, repository.UserProfileRepository, bcrypt)
	authService := NewAuthService(repository.UserRepository, repository.AvatarRepository, repository.UserProfileRepository, repository.RegistrationSessionRepository, repository.RoleRepository, jwtAuth, bcrypt)
	roleService := NewRoleService(repository.RoleRepository)
	caseService := NewCaseService(repository.CaseRepository, repository.CaseVersionRepository, supabase)

	return &Service{
		UserService: userService,
		AuthService: authService,
		RoleService: roleService,
		CaseService: caseService,
	}
}
