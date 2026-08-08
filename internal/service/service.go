package service

import (
	"github.com/azmiagr/unesco-hackathon/internal/repository"
	"github.com/azmiagr/unesco-hackathon/pkg/bcrypt"
	"github.com/azmiagr/unesco-hackathon/pkg/jwt"
)

type Service struct {
	UserService IUserService
	AuthService IAuthService
}

func NewService(repository *repository.Repository, bcrypt bcrypt.Interface, jwtAuth jwt.Interface) *Service {
	userService := NewUserService(repository.UserRepository, repository.RoleRepository)
	authService := NewAuthService(repository.UserRepository, repository.AvatarRepository, repository.UserProfileRepository, repository.RegistrationSessionRepository, repository.RoleRepository, jwtAuth, bcrypt)

	return &Service{
		UserService: userService,
		AuthService: authService,
	}
}
