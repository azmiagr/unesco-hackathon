package service

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/azmiagr/unesco-hackathon/entity"
	"github.com/azmiagr/unesco-hackathon/internal/repository"
	"github.com/azmiagr/unesco-hackathon/model"
	appErrors "github.com/azmiagr/unesco-hackathon/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type adminAuditLogParam struct {
	ActorAdminID  uuid.UUID
	ActionType    string
	Module        string
	TargetType    string
	TargetID      string
	TargetLabel   string
	Detail        string
	PayloadBefore any
	PayloadAfter  any
}

func writeAdminAuditLog(
	tx *gorm.DB,
	auditLogRepo repository.IAuditLogRepository,
	userRepo repository.IUserRepository,
	param adminAuditLogParam,
) error {
	actor, err := userRepo.GetUser(tx, model.GetUserParam{UserID: param.ActorAdminID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return appErrors.Unauthorized("unauthorized")
		}
		return appErrors.InternalServer("failed to get audit actor")
	}

	payloadBefore, err := marshalAuditPayload(param.PayloadBefore)
	if err != nil {
		return appErrors.InternalServer("failed to marshal audit payload")
	}
	payloadAfter, err := marshalAuditPayload(param.PayloadAfter)
	if err != nil {
		return appErrors.InternalServer("failed to marshal audit payload")
	}

	actorID := actor.UserID
	auditLog := &entity.AuditLog{
		AuditLogID:    uuid.New(),
		ActorAdminID:  &actorID,
		ActorName:     actor.Username,
		ActorEmail:    actor.Email,
		ActionType:    param.ActionType,
		Module:        param.Module,
		TargetType:    param.TargetType,
		TargetID:      param.TargetID,
		TargetLabel:   param.TargetLabel,
		Detail:        param.Detail,
		PayloadBefore: payloadBefore,
		PayloadAfter:  payloadAfter,
	}

	if err := auditLogRepo.CreateAuditLog(tx, auditLog); err != nil {
		return appErrors.InternalServer("failed to write audit log")
	}

	return nil
}

func marshalAuditPayload(payload any) (*string, error) {
	if payload == nil {
		return nil, nil
	}

	bytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	value := string(bytes)
	return &value, nil
}

type auditUserSnapshot struct {
	UserID   uuid.UUID `json:"user_id"`
	RoleID   uuid.UUID `json:"role_id"`
	RoleName string    `json:"role_name"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Status   string    `json:"status"`
}

func newAuditUserSnapshot(user *entity.User, roleName string) auditUserSnapshot {
	return auditUserSnapshot{
		UserID:   user.UserID,
		RoleID:   user.RoleID,
		RoleName: roleName,
		Username: user.Username,
		Email:    user.Email,
		Status:   user.Status,
	}
}

type auditCaseSnapshot struct {
	CaseID                   uuid.UUID  `json:"case_id"`
	Title                    string     `json:"title"`
	Slug                     string     `json:"slug"`
	Status                   string     `json:"status"`
	Theme                    string     `json:"theme"`
	CompetencyFocus          string     `json:"competency_focus"`
	DifficultyLevel          string     `json:"difficulty_level"`
	RiskLevel                string     `json:"risk_level"`
	EstimatedDurationMinutes int        `json:"estimated_duration_minutes"`
	MinimumLevel             int        `json:"minimum_level"`
	MinimumReputation        float64    `json:"minimum_reputation"`
	GenerationSource         string     `json:"generation_source"`
	ThumbnailURL             *string    `json:"thumbnail_url"`
	PublishedAt              *time.Time `json:"published_at,omitempty"`
}

func newAuditCaseSnapshot(caseEntity *entity.Case) auditCaseSnapshot {
	return auditCaseSnapshot{
		CaseID:                   caseEntity.CaseID,
		Title:                    caseEntity.Title,
		Slug:                     caseEntity.Slug,
		Status:                   caseEntity.Status,
		Theme:                    caseEntity.Theme,
		CompetencyFocus:          caseEntity.CompetencyFocus,
		DifficultyLevel:          caseEntity.DifficultyLevel,
		RiskLevel:                caseEntity.RiskLevel,
		EstimatedDurationMinutes: caseEntity.EstimatedDurationMinutes,
		MinimumLevel:             caseEntity.MinimumLevel,
		MinimumReputation:        caseEntity.MinimumReputation,
		GenerationSource:         caseEntity.GenerationSource,
		ThumbnailURL:             caseEntity.ThumbnailURL,
		PublishedAt:              caseEntity.PublishedAt,
	}
}

type auditCasePublishSnapshot struct {
	CaseID        uuid.UUID  `json:"case_id"`
	CaseVersionID uuid.UUID  `json:"case_version_id"`
	CaseStatus    string     `json:"case_status"`
	VersionStatus string     `json:"version_status"`
	PublishedAt   *time.Time `json:"published_at,omitempty"`
}

func newAuditCasePublishSnapshot(caseEntity *entity.Case, caseVersion *entity.CaseVersion) auditCasePublishSnapshot {
	return auditCasePublishSnapshot{
		CaseID:        caseEntity.CaseID,
		CaseVersionID: caseVersion.CaseVersionID,
		CaseStatus:    caseEntity.Status,
		VersionStatus: caseVersion.Status,
		PublishedAt:   caseEntity.PublishedAt,
	}
}

type auditItemSnapshot struct {
	ItemID         uuid.UUID  `json:"item_id"`
	ItemCategoryID uuid.UUID  `json:"item_category_id"`
	AvatarID       *uuid.UUID `json:"avatar_id"`
	Name           string     `json:"name"`
	PriceCoin      int        `json:"price_coin"`
	ImageURL       string     `json:"image_url"`
	IsVisible      bool       `json:"is_visible"`
	IsFeatured     bool       `json:"is_featured"`
	Status         string     `json:"status"`
}

func newAuditItemSnapshot(item *entity.Item) auditItemSnapshot {
	return auditItemSnapshot{
		ItemID:         item.ItemID,
		ItemCategoryID: item.ItemCategoryID,
		AvatarID:       item.AvatarID,
		Name:           item.Name,
		PriceCoin:      item.PriceCoin,
		ImageURL:       item.ImageURL,
		IsVisible:      item.IsVisible,
		IsFeatured:     item.IsFeatured,
		Status:         item.Status,
	}
}

type auditRedeemItemSnapshot struct {
	RedeemItemID      uuid.UUID `json:"redeem_item_id"`
	RedeemTypeID      uuid.UUID `json:"redeem_type_id"`
	Name              string    `json:"name"`
	PartnerName       string    `json:"partner_name"`
	PriceCoin         int       `json:"price_coin"`
	MaxClaimPerPeriod int       `json:"max_claim_per_period"`
	ClaimPeriod       string    `json:"claim_period"`
	MinimumLevel      int       `json:"minimum_level"`
	ImageURL          string    `json:"image_url"`
	IsStockVisible    bool      `json:"is_stock_visible"`
	Status            string    `json:"status"`
}

func newAuditRedeemItemSnapshot(redeemItem *entity.RedeemItem) auditRedeemItemSnapshot {
	return auditRedeemItemSnapshot{
		RedeemItemID:      redeemItem.RedeemItemID,
		RedeemTypeID:      redeemItem.RedeemTypeID,
		Name:              redeemItem.Name,
		PartnerName:       redeemItem.PartnerName,
		PriceCoin:         redeemItem.PriceCoin,
		MaxClaimPerPeriod: redeemItem.MaxClaimPerPeriod,
		ClaimPeriod:       redeemItem.ClaimPeriod,
		MinimumLevel:      redeemItem.MinimumLevel,
		ImageURL:          redeemItem.ImageURL,
		IsStockVisible:    redeemItem.IsStockVisible,
		Status:            redeemItem.Status,
	}
}

type auditGameConfigSnapshot struct {
	GameConfigID                uuid.UUID `json:"game_config_id"`
	ConfigKey                   string    `json:"config_key"`
	MaxCasesPerDay              int       `json:"max_cases_per_day"`
	CooldownBetweenCasesMinutes int       `json:"cooldown_between_cases_minutes"`
	StreakBonusMultiplier       float64   `json:"streak_bonus_multiplier"`
	MaintenanceMode             bool      `json:"maintenance_mode"`
	CompleteCaseBaseMultiplier  float64   `json:"complete_case_base_multiplier"`
	PerfectScoreBonusMultiplier float64   `json:"perfect_score_bonus_multiplier"`
	DailyLoginRewardXP          int       `json:"daily_login_reward_xp"`
	StreakBonusCapMultiplier    float64   `json:"streak_bonus_cap_multiplier"`
	DefaultAIProvider           string    `json:"default_ai_provider"`
	OpenAIAPISecretConfigured   bool      `json:"openai_api_secret_configured"`
	DailyTokenBudget            int       `json:"daily_token_budget"`
	PerCaseTokenCap             int       `json:"per_case_token_cap"`
	EnableFailoverSystem        bool      `json:"enable_failover_system"`
	FallbackProvider            string    `json:"fallback_provider"`
	ErrorThresholdPercent       int       `json:"error_threshold_percent"`
	SystemPromptTemplate        string    `json:"system_prompt_template"`
	AITemperature               float64   `json:"ai_temperature"`
}

func newAuditGameConfigSnapshot(config *entity.GameConfig) auditGameConfigSnapshot {
	return auditGameConfigSnapshot{
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
		OpenAIAPISecretConfigured:   config.OpenAIAPISecretKey != nil && *config.OpenAIAPISecretKey != "",
		DailyTokenBudget:            config.DailyTokenBudget,
		PerCaseTokenCap:             config.PerCaseTokenCap,
		EnableFailoverSystem:        config.EnableFailoverSystem,
		FallbackProvider:            config.FallbackProvider,
		ErrorThresholdPercent:       config.ErrorThresholdPercent,
		SystemPromptTemplate:        config.SystemPromptTemplate,
		AITemperature:               config.AITemperature,
	}
}

type auditGameLevelSnapshot struct {
	GameLevelID uuid.UUID `json:"game_level_id"`
	Level       int       `json:"level"`
	XPRequired  int       `json:"xp_required"`
	Title       string    `json:"title"`
	RewardCoin  int       `json:"reward_coin"`
}

func newAuditGameLevelSnapshot(level *entity.GameLevel) auditGameLevelSnapshot {
	return auditGameLevelSnapshot{
		GameLevelID: level.GameLevelID,
		Level:       level.Level,
		XPRequired:  level.XPRequired,
		Title:       level.Title,
		RewardCoin:  level.RewardCoin,
	}
}

type auditCaseEvidenceSnapshot struct {
	CaseEvidenceID  uuid.UUID `json:"case_evidence_id"`
	CaseVersionID   uuid.UUID `json:"case_version_id"`
	TemplateType    string    `json:"template_type"`
	Label           string    `json:"label"`
	CredibilityTags string    `json:"credibility_tags"`
	IsCritical      bool      `json:"is_critical"`
	SortOrder       int       `json:"sort_order"`
}

func newAuditCaseEvidenceSnapshot(evidence *entity.CaseEvidence) auditCaseEvidenceSnapshot {
	return auditCaseEvidenceSnapshot{
		CaseEvidenceID:  evidence.CaseEvidenceID,
		CaseVersionID:   evidence.CaseVersionID,
		TemplateType:    evidence.TemplateType,
		Label:           evidence.Label,
		CredibilityTags: evidence.CredibilityTags,
		IsCritical:      evidence.IsCritical,
		SortOrder:       evidence.SortOrder,
	}
}

type auditCaseQuestionSnapshot struct {
	CaseQuestionID uuid.UUID `json:"case_question_id"`
	CaseVersionID  uuid.UUID `json:"case_version_id"`
	QuestionType   string    `json:"question_type"`
	QuestionText   string    `json:"question_text"`
	ScoringWeight  int       `json:"scoring_weight"`
	IsRequired     bool      `json:"is_required"`
	SortOrder      int       `json:"sort_order"`
}

func newAuditCaseQuestionSnapshot(question *entity.CaseQuestion) auditCaseQuestionSnapshot {
	return auditCaseQuestionSnapshot{
		CaseQuestionID: question.CaseQuestionID,
		CaseVersionID:  question.CaseVersionID,
		QuestionType:   question.QuestionType,
		QuestionText:   question.QuestionText,
		ScoringWeight:  question.ScoringWeight,
		IsRequired:     question.IsRequired,
		SortOrder:      question.SortOrder,
	}
}

func (s *CaseService) writeCaseEvidenceAuditLog(tx *gorm.DB, adminUserID uuid.UUID, actionType string, evidence *entity.CaseEvidence, before any) error {
	return writeAdminAuditLog(tx, s.auditLogRepo, s.userRepo, adminAuditLogParam{
		ActorAdminID:  adminUserID,
		ActionType:    actionType,
		Module:        model.AuditModuleCMS,
		TargetType:    "case_evidence",
		TargetID:      evidence.CaseEvidenceID.String(),
		TargetLabel:   evidence.Label,
		Detail:        actionDetail(actionType, "case evidence", evidence.Label),
		PayloadBefore: before,
		PayloadAfter:  payloadAfterForAction(actionType, newAuditCaseEvidenceSnapshot(evidence)),
	})
}

func (s *CaseService) writeCaseQuestionAuditLog(tx *gorm.DB, adminUserID uuid.UUID, actionType string, question *entity.CaseQuestion, before any) error {
	return writeAdminAuditLog(tx, s.auditLogRepo, s.userRepo, adminAuditLogParam{
		ActorAdminID:  adminUserID,
		ActionType:    actionType,
		Module:        model.AuditModuleCMS,
		TargetType:    "case_question",
		TargetID:      question.CaseQuestionID.String(),
		TargetLabel:   question.QuestionText,
		Detail:        actionDetail(actionType, "case question", question.QuestionText),
		PayloadBefore: before,
		PayloadAfter:  payloadAfterForAction(actionType, newAuditCaseQuestionSnapshot(question)),
	})
}

func actionDetail(actionType string, targetType string, label string) string {
	switch actionType {
	case model.AuditActionCreate:
		return "Created " + targetType + " " + label
	case model.AuditActionUpdate, model.AuditActionConfigChange:
		return "Updated " + targetType + " " + label
	case model.AuditActionDelete:
		return "Deleted " + targetType + " " + label
	default:
		return "Changed " + targetType + " " + label
	}
}

func payloadAfterForAction(actionType string, payload any) any {
	if actionType == model.AuditActionDelete {
		return nil
	}

	return payload
}
