package service

import (
	"github.com/azmiagr/unesco-hackathon/internal/repository"
	"github.com/azmiagr/unesco-hackathon/pkg/bcrypt"
	"github.com/azmiagr/unesco-hackathon/pkg/jwt"
	"github.com/azmiagr/unesco-hackathon/pkg/supabase"
)

type Service struct {
	UserService        IUserService
	AuthService        IAuthService
	RoleService        IRoleService
	CaseService        ICaseService
	AvatarService      IAvatarService
	ItemService        IItemService
	RedeemService      IRedeemService
	GameConfigService  IGameConfigService
	AuditLogService    IAuditLogService
	LeaderboardService ILeaderboardService
	ProfileService     IProfileService
}

func NewService(repository *repository.Repository, bcrypt bcrypt.Interface, jwtAuth jwt.Interface, supabase supabase.Interface) *Service {
	userService := NewUserService(repository.UserRepository, repository.RoleRepository, repository.UserProfileRepository, repository.AuditLogRepository, bcrypt)
	authService := NewAuthService(repository.UserRepository, repository.AvatarRepository, repository.UserProfileRepository, repository.RegistrationSessionRepository, repository.RoleRepository, jwtAuth, bcrypt)
	roleService := NewRoleService(repository.RoleRepository)
	caseService := NewCaseService(repository.CaseRepository, repository.CaseVersionRepository, repository.CaseEvidenceRepository, repository.CaseQuestionRepository, repository.CaseChatbotConfigRepository, repository.CaseScoringOutcomeRepository, repository.UserRepository, repository.AuditLogRepository, supabase, repository.UserProfileRepository)
	avatarService := NewAvatarService(repository.AvatarRepository)
	itemService := NewItemService(repository.ItemRepository, repository.ItemCategoryRepository, repository.UserRepository, repository.AuditLogRepository, supabase)
	redeemService := NewRedeemService(repository.RedeemItemRepository, repository.RedeemTypeRepository, repository.RedeemCodeRepository, repository.UserRepository, repository.AuditLogRepository, supabase)
	gameConfigService := NewGameConfigService(repository.GameConfigRepository, repository.GameLevelRepository, repository.UserRepository, repository.AuditLogRepository)
	auditLogService := NewAuditLogService(repository.AuditLogRepository)
	leaderboardService := NewLeaderboardService(repository.UserProfileRepository)
	profileService := NewProfileService(repository.UserProfileRepository, repository.GameLevelRepository)

	return &Service{
		UserService:        userService,
		AuthService:        authService,
		RoleService:        roleService,
		CaseService:        caseService,
		AvatarService:      avatarService,
		ItemService:        itemService,
		RedeemService:      redeemService,
		GameConfigService:  gameConfigService,
		AuditLogService:    auditLogService,
		LeaderboardService: leaderboardService,
		ProfileService:     profileService,
	}
}
