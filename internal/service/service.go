package service

import (
	"github.com/azmiagr/unesco-hackathon/internal/repository"
	"github.com/azmiagr/unesco-hackathon/pkg/bcrypt"
	"github.com/azmiagr/unesco-hackathon/pkg/jwt"
)

type Service struct {
	UserService IUserService
}

func NewService(repository *repository.Repository, bcrypt bcrypt.Interface, jwtAuth jwt.Interface) *Service {
	return &Service{
		UserService: NewUserService(repository.UserRepository),
	}
}
