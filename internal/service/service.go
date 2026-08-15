package service

import (
	"github.com/azmiagr/unesco-hackathon/internal/repository"
	"github.com/azmiagr/unesco-hackathon/pkg/bcrypt"
	"github.com/azmiagr/unesco-hackathon/pkg/jwt"
	"github.com/azmiagr/unesco-hackathon/pkg/supabase"
)

type Service struct {
	UserService           IUserService
	AuthService           IAuthService
	RoleService           IRoleService
	CaseService           ICaseService
	AvatarService         IAvatarService
	TitleService          ITitleService
	ItemService           IItemService
	RedeemService         IRedeemService
	GameConfigService     IGameConfigService
	AuditLogService       IAuditLogService
	AdminDashboardService IAdminDashboardService
	LeaderboardService    ILeaderboardService
	ProfileService        IProfileService
	LobbyService          ILobbyService
	GameplayService       IGameplayService
}

func NewService(repository *repository.Repository, bcrypt bcrypt.Interface, jwtAuth jwt.Interface, supabase supabase.Interface) *Service {
	userService := NewUserService(repository.UserRepository, repository.RoleRepository, repository.UserProfileRepository, repository.AuditLogRepository, bcrypt)
	authService := NewAuthService(repository.UserRepository, repository.AvatarRepository, repository.UserProfileRepository, repository.RegistrationSessionRepository, repository.RoleRepository, jwtAuth, bcrypt)
	roleService := NewRoleService(repository.RoleRepository)
	caseService := NewCaseService(repository.CaseRepository, repository.CaseVersionRepository, repository.CaseEvidenceRepository, repository.CaseQuestionRepository, repository.CaseChatbotConfigRepository, repository.CaseScoringOutcomeRepository, repository.UserRepository, repository.AuditLogRepository, supabase, repository.UserProfileRepository)
	avatarService := NewAvatarService(repository.AvatarRepository)
	titleService := NewTitleService(repository.TitleRepository, repository.UserItemRepository, repository.UserProfileRepository, supabase)
	itemService := NewItemService(repository.ItemRepository, repository.ItemCategoryRepository, repository.UserItemRepository, repository.UserProfileRepository, repository.AvatarRepository, repository.UserRepository, repository.AuditLogRepository, supabase)
	redeemService := NewRedeemService(repository.RedeemItemRepository, repository.RedeemTypeRepository, repository.RedeemCodeRepository, repository.UserProfileRepository, repository.UserItemRepository, repository.UserRepository, repository.AuditLogRepository, supabase)
	gameConfigService := NewGameConfigService(repository.GameConfigRepository, repository.GameLevelRepository, repository.UserRepository, repository.AuditLogRepository)
	auditLogService := NewAuditLogService(repository.AuditLogRepository)
	adminDashboardService := NewAdminDashboardService(repository.AdminDashboardRepository)
	leaderboardService := NewLeaderboardService(repository.UserProfileRepository)
	profileService := NewProfileService(repository.UserProfileRepository, repository.GameLevelRepository, repository.CaseSessionRepository)
	lobbyService := NewLobbyService(repository.UserProfileRepository, repository.GameLevelRepository, repository.CaseRepository, repository.CityStatisticsRepository, repository.CaseSessionRepository)
	gameplayService := NewGameplayService(repository.CaseRepository, repository.CaseVersionRepository, repository.CaseEvidenceRepository, repository.CaseQuestionRepository, repository.CaseSessionRepository, repository.UserProfileRepository, repository.GameConfigRepository, repository.CaseScoringOutcomeRepository, repository.GameLevelRepository, repository.TitleRepository, repository.UserItemRepository, repository.CityStatisticsRepository)

	return &Service{
		UserService:           userService,
		AuthService:           authService,
		RoleService:           roleService,
		CaseService:           caseService,
		AvatarService:         avatarService,
		TitleService:          titleService,
		ItemService:           itemService,
		RedeemService:         redeemService,
		GameConfigService:     gameConfigService,
		AuditLogService:       auditLogService,
		AdminDashboardService: adminDashboardService,
		LeaderboardService:    leaderboardService,
		ProfileService:        profileService,
		LobbyService:          lobbyService,
		GameplayService:       gameplayService,
	}
}
