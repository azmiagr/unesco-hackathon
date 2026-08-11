package service

import (
	"github.com/azmiagr/unesco-hackathon/entity"
	"github.com/azmiagr/unesco-hackathon/internal/repository"
	"github.com/azmiagr/unesco-hackathon/model"
	"github.com/azmiagr/unesco-hackathon/pkg/database/mariadb"
	appErrors "github.com/azmiagr/unesco-hackathon/pkg/errors"
)

type IAvatarService interface {
	ListAvailableAvatars() (*model.ListAvatarsResponse, error)
}

type AvatarService struct {
	avatarRepo repository.IAvatarRepository
}

func NewAvatarService(avatarRepo repository.IAvatarRepository) IAvatarService {
	return &AvatarService{
		avatarRepo: avatarRepo,
	}
}

func (s *AvatarService) ListAvailableAvatars() (*model.ListAvatarsResponse, error) {
	avatars, err := s.avatarRepo.ListAvailableAvatars(mariadb.Connection)
	if err != nil {
		return nil, appErrors.InternalServer("failed to list avatars")
	}

	responses := make([]model.AvatarResponse, 0, len(avatars))
	for _, avatar := range avatars {
		responses = append(responses, mapAvatarResponse(avatar))
	}

	return &model.ListAvatarsResponse{
		Avatars: responses,
	}, nil
}

func mapAvatarResponse(avatar entity.Avatar) model.AvatarResponse {
	return model.AvatarResponse{
		AvatarID:    avatar.AvatarID,
		ImageURL:    avatar.ImageURL,
		UnlockLevel: avatar.UnlockLevel,
		Status:      avatar.Status,
		CreatedAt:   avatar.CreatedAt,
		UpdatedAt:   avatar.UpdatedAt,
	}
}
