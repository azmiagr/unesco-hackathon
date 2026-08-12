package service

import (
	"errors"
	"math"
	"strings"

	"github.com/azmiagr/unesco-hackathon/entity"
	"github.com/azmiagr/unesco-hackathon/internal/repository"
	"github.com/azmiagr/unesco-hackathon/model"
	"github.com/azmiagr/unesco-hackathon/pkg/database/mariadb"
	appErrors "github.com/azmiagr/unesco-hackathon/pkg/errors"
	"github.com/azmiagr/unesco-hackathon/pkg/helper"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	defaultAdminGameLevelLimit = 10
	maxAdminGameLevelLimit     = 100
)

type IGameConfigService interface {
	GetGameConfigByAdmin() (*model.AdminGetGameConfigResponse, error)
	UpsertGameConfigByAdmin(adminUserID uuid.UUID, req model.AdminUpsertGameConfigRequest) (*model.AdminUpsertGameConfigResponse, error)
	GetGameGeneralConfigByAdmin() (*model.AdminGetGameGeneralConfigResponse, error)
	UpsertGameGeneralConfigByAdmin(adminUserID uuid.UUID, req model.AdminUpsertGameGeneralConfigRequest) (*model.AdminUpsertGameGeneralConfigResponse, error)
	GetGameAIConfigByAdmin() (*model.AdminGetGameAIConfigResponse, error)
	UpsertGameAIConfigByAdmin(adminUserID uuid.UUID, req model.AdminUpsertGameAIConfigRequest) (*model.AdminUpsertGameAIConfigResponse, error)
	ListGameLevelsByAdmin(req model.AdminListGameLevelsRequest) (*model.AdminListGameLevelsResponse, error)
	GetGameLevelDetailByAdmin(gameLevelID uuid.UUID) (*model.AdminGetGameLevelDetailResponse, error)
	CreateGameLevelByAdmin(adminUserID uuid.UUID, req model.AdminCreateGameLevelRequest) (*model.AdminCreateGameLevelResponse, error)
	UpdateGameLevelByAdmin(adminUserID uuid.UUID, gameLevelID uuid.UUID, req model.AdminUpdateGameLevelRequest) (*model.AdminUpdateGameLevelResponse, error)
	DeleteGameLevelByAdmin(adminUserID uuid.UUID, gameLevelID uuid.UUID) (*model.AdminDeleteGameLevelResponse, error)
}

type GameConfigService struct {
	db             *gorm.DB
	gameConfigRepo repository.IGameConfigRepository
	gameLevelRepo  repository.IGameLevelRepository
}

func NewGameConfigService(
	gameConfigRepo repository.IGameConfigRepository,
	gameLevelRepo repository.IGameLevelRepository,
) IGameConfigService {
	return &GameConfigService{
		db:             mariadb.Connection,
		gameConfigRepo: gameConfigRepo,
		gameLevelRepo:  gameLevelRepo,
	}
}

func (s *GameConfigService) GetGameConfigByAdmin() (*model.AdminGetGameConfigResponse, error) {
	config, err := s.gameConfigRepo.GetGameConfig(s.db, model.GetGameConfigParam{
		ConfigKey: model.GameConfigDefaultKey,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			defaultConfig := defaultGameConfigEntity()
			return &model.AdminGetGameConfigResponse{
				Config: mapAdminGameConfigResponse(defaultConfig),
			}, nil
		}
		return nil, appErrors.InternalServer("failed to get game config")
	}

	return &model.AdminGetGameConfigResponse{
		Config: mapAdminGameConfigResponse(config),
	}, nil
}

func (s *GameConfigService) UpsertGameConfigByAdmin(
	adminUserID uuid.UUID,
	req model.AdminUpsertGameConfigRequest,
) (*model.AdminUpsertGameConfigResponse, error) {
	if adminUserID == uuid.Nil {
		return nil, appErrors.Unauthorized("unauthorized")
	}

	err := validateGameConfigRequest(req)
	if err != nil {
		return nil, err
	}

	tx := s.db.Begin()
	defer tx.Rollback()

	config := &entity.GameConfig{
		GameConfigID:                uuid.New(),
		ConfigKey:                   model.GameConfigDefaultKey,
		MaxCasesPerDay:              req.MaxCasesPerDay,
		CooldownBetweenCasesMinutes: req.CooldownBetweenCasesMinutes,
		StreakBonusMultiplier:       req.StreakBonusMultiplier,
		MaintenanceMode:             req.MaintenanceMode,
		CompleteCaseBaseMultiplier:  req.CompleteCaseBaseMultiplier,
		PerfectScoreBonusMultiplier: req.PerfectScoreBonusMultiplier,
		DailyLoginRewardXP:          req.DailyLoginRewardXP,
		StreakBonusCapMultiplier:    req.StreakBonusCapMultiplier,
		DefaultAIProvider:           strings.TrimSpace(req.DefaultAIProvider),
		OpenAIAPISecretKey:          normalizeOptionalSecret(req.OpenAIAPISecretKey),
		DailyTokenBudget:            req.DailyTokenBudget,
		PerCaseTokenCap:             req.PerCaseTokenCap,
		EnableFailoverSystem:        req.EnableFailoverSystem,
		FallbackProvider:            strings.TrimSpace(req.FallbackProvider),
		ErrorThresholdPercent:       req.ErrorThresholdPercent,
		SystemPromptTemplate:        strings.TrimSpace(req.SystemPromptTemplate),
		AITemperature:               req.AITemperature,
	}

	err = s.gameConfigRepo.UpsertGameConfig(tx, config)
	if err != nil {
		return nil, appErrors.InternalServer("failed to save game config")
	}

	savedConfig, err := s.gameConfigRepo.GetGameConfig(tx, model.GetGameConfigParam{
		ConfigKey: model.GameConfigDefaultKey,
	})
	if err != nil {
		return nil, appErrors.InternalServer("failed to get saved game config")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	return &model.AdminUpsertGameConfigResponse{
		Config: mapAdminGameConfigResponse(savedConfig),
	}, nil
}

func (s *GameConfigService) GetGameGeneralConfigByAdmin() (*model.AdminGetGameGeneralConfigResponse, error) {
	config, err := s.gameConfigRepo.GetGameConfig(s.db, model.GetGameConfigParam{
		ConfigKey: model.GameConfigDefaultKey,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &model.AdminGetGameGeneralConfigResponse{
				Config: mapAdminGameGeneralConfigResponse(defaultGameConfigEntity()),
			}, nil
		}
		return nil, appErrors.InternalServer("failed to get game general config")
	}

	return &model.AdminGetGameGeneralConfigResponse{
		Config: mapAdminGameGeneralConfigResponse(config),
	}, nil
}

func (s *GameConfigService) UpsertGameGeneralConfigByAdmin(
	adminUserID uuid.UUID,
	req model.AdminUpsertGameGeneralConfigRequest,
) (*model.AdminUpsertGameGeneralConfigResponse, error) {
	if adminUserID == uuid.Nil {
		return nil, appErrors.Unauthorized("unauthorized")
	}

	err := validateGameGeneralConfigRequest(req)
	if err != nil {
		return nil, err
	}

	tx := s.db.Begin()
	defer tx.Rollback()

	config, err := s.getOrCreateDefaultGameConfigForUpdate(tx)
	if err != nil {
		return nil, err
	}

	config.MaxCasesPerDay = req.MaxCasesPerDay
	config.CooldownBetweenCasesMinutes = req.CooldownBetweenCasesMinutes
	config.StreakBonusMultiplier = req.StreakBonusMultiplier
	config.MaintenanceMode = req.MaintenanceMode

	err = s.gameConfigRepo.UpdateGameConfig(tx, config)
	if err != nil {
		return nil, appErrors.InternalServer("failed to save game general config")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	return &model.AdminUpsertGameGeneralConfigResponse{
		Config: mapAdminGameGeneralConfigResponse(config),
	}, nil
}

func (s *GameConfigService) GetGameAIConfigByAdmin() (*model.AdminGetGameAIConfigResponse, error) {
	config, err := s.gameConfigRepo.GetGameConfig(s.db, model.GetGameConfigParam{
		ConfigKey: model.GameConfigDefaultKey,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &model.AdminGetGameAIConfigResponse{
				Config: mapAdminGameAIConfigResponse(defaultGameConfigEntity()),
			}, nil
		}
		return nil, appErrors.InternalServer("failed to get game ai config")
	}

	return &model.AdminGetGameAIConfigResponse{
		Config: mapAdminGameAIConfigResponse(config),
	}, nil
}

func (s *GameConfigService) UpsertGameAIConfigByAdmin(
	adminUserID uuid.UUID,
	req model.AdminUpsertGameAIConfigRequest,
) (*model.AdminUpsertGameAIConfigResponse, error) {
	if adminUserID == uuid.Nil {
		return nil, appErrors.Unauthorized("unauthorized")
	}

	err := validateGameAIConfigRequest(req)
	if err != nil {
		return nil, err
	}

	tx := s.db.Begin()
	defer tx.Rollback()

	config, err := s.getOrCreateDefaultGameConfigForUpdate(tx)
	if err != nil {
		return nil, err
	}

	config.DefaultAIProvider = strings.TrimSpace(req.DefaultAIProvider)
	if secret := normalizeOptionalSecret(req.OpenAIAPISecretKey); secret != nil {
		config.OpenAIAPISecretKey = secret
	}
	config.DailyTokenBudget = req.DailyTokenBudget
	config.PerCaseTokenCap = req.PerCaseTokenCap
	config.EnableFailoverSystem = req.EnableFailoverSystem
	config.FallbackProvider = strings.TrimSpace(req.FallbackProvider)
	config.ErrorThresholdPercent = req.ErrorThresholdPercent
	config.SystemPromptTemplate = strings.TrimSpace(req.SystemPromptTemplate)
	config.AITemperature = req.AITemperature

	err = s.gameConfigRepo.UpdateGameConfig(tx, config)
	if err != nil {
		return nil, appErrors.InternalServer("failed to save game ai config")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	return &model.AdminUpsertGameAIConfigResponse{
		Config: mapAdminGameAIConfigResponse(config),
	}, nil
}

func (s *GameConfigService) ListGameLevelsByAdmin(req model.AdminListGameLevelsRequest) (*model.AdminListGameLevelsResponse, error) {
	page := req.Page
	if page < 1 {
		page = 1
	}

	limit := req.Limit
	if limit < 1 {
		limit = defaultAdminGameLevelLimit
	}
	if limit > maxAdminGameLevelLimit {
		limit = maxAdminGameLevelLimit
	}

	levels, total, err := s.gameLevelRepo.ListGameLevels(s.db, model.ListGameLevelsParam{
		Search: strings.TrimSpace(req.Search),
		Limit:  limit,
		Offset: (page - 1) * limit,
	})
	if err != nil {
		return nil, appErrors.InternalServer("failed to list game levels")
	}

	responses := make([]model.AdminGameLevelResponse, 0, len(levels))
	for _, level := range levels {
		responses = append(responses, mapAdminGameLevelResponse(&level))
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}

	return &model.AdminListGameLevelsResponse{
		Levels: responses,
		Pagination: model.PaginationResponse{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

func (s *GameConfigService) GetGameLevelDetailByAdmin(gameLevelID uuid.UUID) (*model.AdminGetGameLevelDetailResponse, error) {
	if gameLevelID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid game level id")
	}

	level, err := s.gameLevelRepo.GetGameLevel(s.db, model.GetGameLevelParam{GameLevelID: gameLevelID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("game level not found")
		}
		return nil, appErrors.InternalServer("failed to get game level")
	}

	return &model.AdminGetGameLevelDetailResponse{
		Level: mapAdminGameLevelResponse(level),
	}, nil
}

func (s *GameConfigService) CreateGameLevelByAdmin(
	adminUserID uuid.UUID,
	req model.AdminCreateGameLevelRequest,
) (*model.AdminCreateGameLevelResponse, error) {
	if adminUserID == uuid.Nil {
		return nil, appErrors.Unauthorized("unauthorized")
	}

	title, err := validateGameLevelRequest(req)
	if err != nil {
		return nil, err
	}

	tx := s.db.Begin()
	defer tx.Rollback()

	level := &entity.GameLevel{
		GameLevelID: uuid.New(),
		Level:       req.Level,
		XPRequired:  req.XPRequired,
		Title:       title,
		RewardCoin:  req.RewardCoin,
	}

	err = s.gameLevelRepo.CreateGameLevel(tx, level)
	if err != nil {
		return nil, appErrors.InternalServer("failed to create game level")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	return &model.AdminCreateGameLevelResponse{
		Level: mapAdminGameLevelResponse(level),
	}, nil
}

func (s *GameConfigService) UpdateGameLevelByAdmin(
	adminUserID uuid.UUID,
	gameLevelID uuid.UUID,
	req model.AdminUpdateGameLevelRequest,
) (*model.AdminUpdateGameLevelResponse, error) {
	if adminUserID == uuid.Nil {
		return nil, appErrors.Unauthorized("unauthorized")
	}
	if gameLevelID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid game level id")
	}

	title, err := validateGameLevelRequest(req)
	if err != nil {
		return nil, err
	}

	tx := s.db.Begin()
	defer tx.Rollback()

	level, err := s.gameLevelRepo.GetGameLevelForUpdate(tx, gameLevelID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("game level not found")
		}
		return nil, appErrors.InternalServer("failed to get game level")
	}

	level.Level = req.Level
	level.XPRequired = req.XPRequired
	level.Title = title
	level.RewardCoin = req.RewardCoin

	err = s.gameLevelRepo.UpdateGameLevel(tx, level)
	if err != nil {
		return nil, appErrors.InternalServer("failed to update game level")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	return &model.AdminUpdateGameLevelResponse{
		Level: mapAdminGameLevelResponse(level),
	}, nil
}

func (s *GameConfigService) DeleteGameLevelByAdmin(
	adminUserID uuid.UUID,
	gameLevelID uuid.UUID,
) (*model.AdminDeleteGameLevelResponse, error) {
	if adminUserID == uuid.Nil {
		return nil, appErrors.Unauthorized("unauthorized")
	}
	if gameLevelID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid game level id")
	}

	tx := s.db.Begin()
	defer tx.Rollback()

	level, err := s.gameLevelRepo.GetGameLevelForUpdate(tx, gameLevelID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("game level not found")
		}
		return nil, appErrors.InternalServer("failed to get game level")
	}

	err = s.gameLevelRepo.DeleteGameLevel(tx, level.GameLevelID)
	if err != nil {
		return nil, appErrors.InternalServer("failed to delete game level")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	return &model.AdminDeleteGameLevelResponse{
		GameLevelID: level.GameLevelID,
	}, nil
}

func validateGameConfigRequest(req model.AdminUpsertGameConfigRequest) error {
	err := validateGameGeneralConfigRequest(model.AdminUpsertGameGeneralConfigRequest{
		MaxCasesPerDay:              req.MaxCasesPerDay,
		CooldownBetweenCasesMinutes: req.CooldownBetweenCasesMinutes,
		StreakBonusMultiplier:       req.StreakBonusMultiplier,
		MaintenanceMode:             req.MaintenanceMode,
	})
	if err != nil {
		return err
	}
	if req.CompleteCaseBaseMultiplier < 0 {
		return appErrors.BadRequest("complete case base multiplier cannot be negative")
	}
	if req.PerfectScoreBonusMultiplier < 0 {
		return appErrors.BadRequest("perfect score bonus multiplier cannot be negative")
	}
	if req.DailyLoginRewardXP < 0 {
		return appErrors.BadRequest("daily login reward xp cannot be negative")
	}
	if req.StreakBonusCapMultiplier < 1 {
		return appErrors.BadRequest("streak bonus cap multiplier must be at least 1")
	}
	return validateGameAIConfigRequest(model.AdminUpsertGameAIConfigRequest{
		DefaultAIProvider:     req.DefaultAIProvider,
		OpenAIAPISecretKey:    req.OpenAIAPISecretKey,
		DailyTokenBudget:      req.DailyTokenBudget,
		PerCaseTokenCap:       req.PerCaseTokenCap,
		EnableFailoverSystem:  req.EnableFailoverSystem,
		FallbackProvider:      req.FallbackProvider,
		ErrorThresholdPercent: req.ErrorThresholdPercent,
		SystemPromptTemplate:  req.SystemPromptTemplate,
		AITemperature:         req.AITemperature,
	})
}

func validateGameGeneralConfigRequest(req model.AdminUpsertGameGeneralConfigRequest) error {
	if req.MaxCasesPerDay < 1 {
		return appErrors.BadRequest("max cases per day must be greater than 0")
	}
	if req.CooldownBetweenCasesMinutes < 0 {
		return appErrors.BadRequest("cooldown between cases minutes cannot be negative")
	}
	if req.StreakBonusMultiplier < 1 {
		return appErrors.BadRequest("streak bonus multiplier must be at least 1")
	}

	return nil
}

func validateGameAIConfigRequest(req model.AdminUpsertGameAIConfigRequest) error {
	if strings.TrimSpace(req.DefaultAIProvider) == "" {
		return appErrors.BadRequest("default ai provider is required")
	}
	if req.DailyTokenBudget < 1 {
		return appErrors.BadRequest("daily token budget must be greater than 0")
	}
	if req.PerCaseTokenCap < 1 {
		return appErrors.BadRequest("per case token cap must be greater than 0")
	}
	if strings.TrimSpace(req.FallbackProvider) == "" {
		return appErrors.BadRequest("fallback provider is required")
	}
	if req.ErrorThresholdPercent < 0 || req.ErrorThresholdPercent > 100 {
		return appErrors.BadRequest("error threshold percent must be between 0 and 100")
	}
	if strings.TrimSpace(req.SystemPromptTemplate) == "" {
		return appErrors.BadRequest("system prompt template is required")
	}
	if req.AITemperature < 0 || req.AITemperature > 2 {
		return appErrors.BadRequest("ai temperature must be between 0 and 2")
	}

	return nil
}

func validateGameLevelRequest(req model.AdminCreateGameLevelRequest) (string, error) {
	if req.Level < 1 {
		return "", appErrors.BadRequest("level must be greater than 0")
	}
	if req.XPRequired < 0 {
		return "", appErrors.BadRequest("xp required cannot be negative")
	}
	title, err := helper.RequireTrimmedString(req.Title, "title is required")
	if err != nil {
		return "", err
	}
	if len(title) > 150 {
		return "", appErrors.BadRequest("title is too long")
	}
	if req.RewardCoin < 0 {
		return "", appErrors.BadRequest("reward coin cannot be negative")
	}

	return title, nil
}

func normalizeOptionalSecret(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func (s *GameConfigService) getOrCreateDefaultGameConfigForUpdate(tx *gorm.DB) (*entity.GameConfig, error) {
	config, err := s.gameConfigRepo.GetGameConfigForUpdate(tx, model.GetGameConfigParam{
		ConfigKey: model.GameConfigDefaultKey,
	})
	if err == nil {
		return config, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appErrors.InternalServer("failed to get game config")
	}

	config = defaultGameConfigEntity()
	config.GameConfigID = uuid.New()

	err = s.gameConfigRepo.CreateGameConfig(tx, config)
	if err != nil {
		return nil, appErrors.InternalServer("failed to create default game config")
	}

	return config, nil
}

func defaultGameConfigEntity() *entity.GameConfig {
	return &entity.GameConfig{
		GameConfigID:                uuid.Nil,
		ConfigKey:                   model.GameConfigDefaultKey,
		MaxCasesPerDay:              5,
		CooldownBetweenCasesMinutes: 15,
		StreakBonusMultiplier:       1.5,
		MaintenanceMode:             false,
		CompleteCaseBaseMultiplier:  1,
		PerfectScoreBonusMultiplier: 1.5,
		DailyLoginRewardXP:          100,
		StreakBonusCapMultiplier:    1.2,
		DefaultAIProvider:           "openai",
		OpenAIAPISecretKey:          nil,
		DailyTokenBudget:            2500000,
		PerCaseTokenCap:             15000,
		EnableFailoverSystem:        true,
		FallbackProvider:            "anthropic",
		ErrorThresholdPercent:       5,
		SystemPromptTemplate:        "{}",
		AITemperature:               0.3,
	}
}

func maskOptionalSecret(secret *string) *string {
	if secret == nil {
		return nil
	}
	value := strings.TrimSpace(*secret)
	if value == "" {
		return nil
	}
	if len(value) <= 4 {
		masked := "********"
		return &masked
	}

	masked := "********" + value[len(value)-4:]
	return &masked
}

func mapAdminGameConfigResponse(config *entity.GameConfig) model.AdminGameConfigResponse {
	return model.AdminGameConfigResponse{
		GameConfigID:                config.GameConfigID,
		ConfigKey:                   config.ConfigKey,
		MaxCasesPerDay:              config.MaxCasesPerDay,
		CooldownBetweenCasesMinutes: config.CooldownBetweenCasesMinutes,
		StreakBonusMultiplier:       config.StreakBonusMultiplier,
		MaintenanceMode:             config.MaintenanceMode,
		CompleteCaseBaseMultiplier:  config.CompleteCaseBaseMultiplier,
		PerfectScoreBonusMultiplier: config.PerfectScoreBonusMultiplier,
		DailyLoginRewardXP:          config.DailyLoginRewardXP,
		StreakBonusCapMultiplier:    config.StreakBonusCapMultiplier,
		DefaultAIProvider:           config.DefaultAIProvider,
		OpenAIAPISecretKey:          config.OpenAIAPISecretKey,
		DailyTokenBudget:            config.DailyTokenBudget,
		PerCaseTokenCap:             config.PerCaseTokenCap,
		EnableFailoverSystem:        config.EnableFailoverSystem,
		FallbackProvider:            config.FallbackProvider,
		ErrorThresholdPercent:       config.ErrorThresholdPercent,
		SystemPromptTemplate:        config.SystemPromptTemplate,
		AITemperature:               config.AITemperature,
		CreatedAt:                   config.CreatedAt,
		UpdatedAt:                   config.UpdatedAt,
	}
}

func mapAdminGameGeneralConfigResponse(config *entity.GameConfig) model.AdminGameGeneralConfigResponse {
	return model.AdminGameGeneralConfigResponse{
		GameConfigID:                config.GameConfigID,
		ConfigKey:                   config.ConfigKey,
		MaxCasesPerDay:              config.MaxCasesPerDay,
		CooldownBetweenCasesMinutes: config.CooldownBetweenCasesMinutes,
		StreakBonusMultiplier:       config.StreakBonusMultiplier,
		MaintenanceMode:             config.MaintenanceMode,
		CreatedAt:                   config.CreatedAt,
		UpdatedAt:                   config.UpdatedAt,
	}
}

func mapAdminGameAIConfigResponse(config *entity.GameConfig) model.AdminGameAIConfigResponse {
	return model.AdminGameAIConfigResponse{
		GameConfigID:          config.GameConfigID,
		ConfigKey:             config.ConfigKey,
		DefaultAIProvider:     config.DefaultAIProvider,
		OpenAIAPISecretKey:    maskOptionalSecret(config.OpenAIAPISecretKey),
		DailyTokenBudget:      config.DailyTokenBudget,
		PerCaseTokenCap:       config.PerCaseTokenCap,
		EnableFailoverSystem:  config.EnableFailoverSystem,
		FallbackProvider:      config.FallbackProvider,
		ErrorThresholdPercent: config.ErrorThresholdPercent,
		SystemPromptTemplate:  config.SystemPromptTemplate,
		AITemperature:         config.AITemperature,
		CreatedAt:             config.CreatedAt,
		UpdatedAt:             config.UpdatedAt,
	}
}

func mapAdminGameLevelResponse(level *entity.GameLevel) model.AdminGameLevelResponse {
	return model.AdminGameLevelResponse{
		GameLevelID: level.GameLevelID,
		Level:       level.Level,
		XPRequired:  level.XPRequired,
		Title:       level.Title,
		RewardCoin:  level.RewardCoin,
		CreatedAt:   level.CreatedAt,
		UpdatedAt:   level.UpdatedAt,
	}
}
