package repository

import "gorm.io/gorm"

type Repository struct {
	UserRepository                IUserRepository
	OtpRepository                 IOtpRepository
	AvatarRepository              IAvatarRepository
	UserProfileRepository         IUserProfileRepository
	RegistrationSessionRepository IRegistrationSessionRepository
	RoleRepository                IRoleRepository
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		UserRepository:                NewUserRepository(db),
		OtpRepository:                 NewOtpRepository(db),
		AvatarRepository:              NewAvatarRepository(db),
		UserProfileRepository:         NewUserProfileRepository(db),
		RegistrationSessionRepository: NewRegistrationSessionRepository(db),
		RoleRepository:                NewRoleRepository(db),
	}
}
