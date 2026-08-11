package repository

import "gorm.io/gorm"

type Repository struct {
	UserRepository                IUserRepository
	OtpRepository                 IOtpRepository
	AvatarRepository              IAvatarRepository
	UserProfileRepository         IUserProfileRepository
	RegistrationSessionRepository IRegistrationSessionRepository
	RoleRepository                IRoleRepository
	CaseRepository                ICaseRepository
	CaseVersionRepository         ICaseVersionRepository
	CaseEvidenceRepository        ICaseEvidenceRepository
	CaseQuestionRepository        ICaseQuestionRepository
	CaseChatbotConfigRepository   ICaseChatbotConfigRepository
	CaseScoringOutcomeRepository  ICaseScoringOutcomeRepository
	ItemCategoryRepository        IItemCategoryRepository
	ItemRepository                IItemRepository
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		UserRepository:                NewUserRepository(db),
		OtpRepository:                 NewOtpRepository(db),
		AvatarRepository:              NewAvatarRepository(db),
		UserProfileRepository:         NewUserProfileRepository(db),
		RegistrationSessionRepository: NewRegistrationSessionRepository(db),
		RoleRepository:                NewRoleRepository(db),
		CaseRepository:                NewCaseRepository(db),
		CaseVersionRepository:         NewCaseVersionRepository(db),
		CaseEvidenceRepository:        NewCaseEvidenceRepository(db),
		CaseQuestionRepository:        NewCaseQuestionRepository(db),
		CaseChatbotConfigRepository:   NewCaseChatbotConfigRepository(db),
		CaseScoringOutcomeRepository:  NewCaseScoringOutcomeRepository(db),
		ItemCategoryRepository:        NewItemCategoryRepository(db),
		ItemRepository:                NewItemRepository(db),
	}
}
