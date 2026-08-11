package service

import (
	"github.com/azmiagr/unesco-hackathon/internal/repository"
	"github.com/azmiagr/unesco-hackathon/model"
	"github.com/azmiagr/unesco-hackathon/pkg/database/mariadb"
	"github.com/azmiagr/unesco-hackathon/pkg/supabase"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	maxCaseThumbnailSize     = 5 * 1024 * 1024
	maxEvidenceImageSize     = 5 * 1024 * 1024
	maxSocialPostLabelLength = 150
	defaultAdminCaseLimit    = 10
	maxAdminCaseLimit        = 100
	maxEvidenceLabelLength   = 150
)

var allowedCaseThemes = map[string]bool{
	model.CaseThemeMisleadingHealthAdvice: true,
	model.CaseThemeChatbotHallucination:   true,
	model.CaseThemeClickbaitHeadline:      true,
	model.CaseThemeStatisticOutOfContext:  true,
	model.CaseThemeForumMisinformation:    true,
	model.CaseThemeViralConflictContent:   true,
	model.CaseThemeAlgorithmicEchoChamber: true,
	model.CaseThemeOther:                  true,
}

var allowedCaseCompetencyFocuses = map[string]bool{
	model.CaseCompetencyEvidenceEvaluation:    true,
	model.CaseCompetencyClaimAnalysis:         true,
	model.CaseCompetencyConfidenceCalibration: true,
	model.CaseCompetencyReasoning:             true,
	model.CaseCompetencySafetyJudgment:        true,
}

var allowedCaseDifficultyLevels = map[string]bool{
	model.CaseDifficultyLow:    true,
	model.CaseDifficultyMedium: true,
	model.CaseDifficultyHigh:   true,
}

var allowedCaseRiskLevels = map[string]bool{
	model.CaseRiskLow:    true,
	model.CaseRiskMedium: true,
	model.CaseRiskHigh:   true,
}

var allowedCaseGenerationSources = map[string]bool{
	model.CaseGenerationManual:     true,
	model.CaseGenerationAIAssisted: true,
}

var allowedCaseStatuses = map[string]bool{
	model.CaseStatusDraft:     true,
	model.CaseStatusPublished: true,
	model.CaseStatusArchived:  true,
}

type ICaseService interface {
	CreateCaseByAdmin(adminUserID uuid.UUID, req model.AdminCreateCaseRequest) (*model.AdminCreateCaseResponse, error)
	UpdateCaseByAdmin(adminUserID uuid.UUID, caseID uuid.UUID, req model.AdminUpdateCaseRequest) (*model.AdminUpdateCaseResponse, error)
	HardDeleteCaseByAdmin(adminUserID uuid.UUID, caseID uuid.UUID) (*model.AdminDeleteCaseResponse, error)
	ListCasesByAdmin(req model.AdminListCasesRequest) (*model.AdminListCasesResponse, error)
	GetCaseDetailByAdmin(caseID uuid.UUID) (*model.AdminCaseDetailResponse, error)
	ListCaseEvidencesByAdmin(caseID uuid.UUID) (*model.AdminListCaseEvidencesResponse, error)
	ListEvidenceOptionsByAdmin(caseID uuid.UUID) (*model.AdminListEvidenceOptionsResponse, error)
	GetCaseEvidenceDetailByAdmin(caseID uuid.UUID, caseVersionID uuid.UUID, caseEvidenceID uuid.UUID) (*model.AdminEvidenceDetailResponse, error)
	ListCaseQuestionsByAdmin(caseID uuid.UUID) (*model.AdminListCaseQuestionsResponse, error)
	GetCaseQuestionDetailByAdmin(caseID uuid.UUID, caseVersionID uuid.UUID, caseQuestionID uuid.UUID) (*model.AdminQuestionDetailResponse, error)
	GetCaseChatbotConfigByAdmin(caseID uuid.UUID) (*model.AdminGetCaseChatbotConfigResponse, error)
	UpsertCaseChatbotConfigByAdmin(adminUserID uuid.UUID, caseID uuid.UUID, req model.AdminUpsertCaseChatbotConfigRequest) (*model.AdminUpsertCaseChatbotConfigResponse, error)
	GetCaseLookups() (*model.AdminCaseLookupsResponse, error)
	CreateSocialPostEvidenceByAdmin(adminUserID uuid.UUID, caseID uuid.UUID, caseVersionID uuid.UUID, req model.AdminCreateSocialPostEvidenceRequest) (*model.AdminCreateSocialPostEvidenceResponse, error)
	CreateArticleEvidenceByAdmin(adminUserID uuid.UUID, caseID uuid.UUID, caseVersionID uuid.UUID, req model.AdminCreateArticleEvidenceRequest) (*model.AdminCreateArticleEvidenceResponse, error)
	CreateBlogEvidenceByAdmin(adminUserID uuid.UUID, caseID uuid.UUID, caseVersionID uuid.UUID, req model.AdminCreateBlogEvidenceRequest) (*model.AdminCreateBlogEvidenceResponse, error)
	CreateForumThreadEvidenceByAdmin(adminUserID uuid.UUID, caseID uuid.UUID, caseVersionID uuid.UUID, req model.AdminCreateForumThreadEvidenceRequest) (*model.AdminCreateForumThreadEvidenceResponse, error)
	CreateChatTranscriptEvidenceByAdmin(adminUserID uuid.UUID, caseID uuid.UUID, caseVersionID uuid.UUID, req model.AdminCreateChatTranscriptEvidenceRequest) (*model.AdminCreateChatTranscriptEvidenceResponse, error)
	CreatePublicAnnouncementEvidenceByAdmin(adminUserID uuid.UUID, caseID uuid.UUID, caseVersionID uuid.UUID, req model.AdminCreatePublicAnnouncementEvidenceRequest) (*model.AdminCreatePublicAnnouncementEvidenceResponse, error)
	CreateMCQQuestionByAdmin(adminUserID uuid.UUID, caseID uuid.UUID, caseVersionID uuid.UUID, req model.AdminCreateMCQQuestionRequest) (*model.AdminCreateMCQQuestionResponse, error)
	CreateOpenEndedQuestionByAdmin(adminUserID uuid.UUID, caseID uuid.UUID, caseVersionID uuid.UUID, req model.AdminCreateOpenEndedQuestionRequest) (*model.AdminCreateOpenEndedQuestionResponse, error)
	UpdateSocialPostEvidenceByAdmin(adminUserID uuid.UUID, caseID uuid.UUID, caseVersionID uuid.UUID, caseEvidenceID uuid.UUID, req model.AdminUpdateSocialPostEvidenceRequest) (*model.AdminUpdateSocialPostEvidenceResponse, error)
	UpdateArticleEvidenceByAdmin(adminUserID uuid.UUID, caseID uuid.UUID, caseVersionID uuid.UUID, caseEvidenceID uuid.UUID, req model.AdminUpdateArticleEvidenceRequest) (*model.AdminUpdateArticleEvidenceResponse, error)
	UpdateBlogEvidenceByAdmin(adminUserID uuid.UUID, caseID uuid.UUID, caseVersionID uuid.UUID, caseEvidenceID uuid.UUID, req model.AdminUpdateBlogEvidenceRequest) (*model.AdminUpdateBlogEvidenceResponse, error)
	UpdateForumThreadEvidenceByAdmin(adminUserID uuid.UUID, caseID uuid.UUID, caseVersionID uuid.UUID, caseEvidenceID uuid.UUID, req model.AdminUpdateForumThreadEvidenceRequest) (*model.AdminUpdateForumThreadEvidenceResponse, error)
	UpdateChatTranscriptEvidenceByAdmin(adminUserID uuid.UUID, caseID uuid.UUID, caseVersionID uuid.UUID, caseEvidenceID uuid.UUID, req model.AdminUpdateChatTranscriptEvidenceRequest) (*model.AdminUpdateChatTranscriptEvidenceResponse, error)
	UpdatePublicAnnouncementEvidenceByAdmin(adminUserID uuid.UUID, caseID uuid.UUID, caseVersionID uuid.UUID, caseEvidenceID uuid.UUID, req model.AdminUpdatePublicAnnouncementEvidenceRequest) (*model.AdminUpdatePublicAnnouncementEvidenceResponse, error)
	DeleteCaseEvidenceByAdmin(adminUserID uuid.UUID, caseID uuid.UUID, caseVersionID uuid.UUID, caseEvidenceID uuid.UUID) (*model.AdminDeleteCaseEvidenceResponse, error)
}

type CaseService struct {
	db                    *gorm.DB
	caseRepo              repository.ICaseRepository
	caseVersionRepo       repository.ICaseVersionRepository
	caseEvidenceRepo      repository.ICaseEvidenceRepository
	caseQuestionRepo      repository.ICaseQuestionRepository
	caseChatbotConfigRepo repository.ICaseChatbotConfigRepository
	storage               supabase.Interface
}

func NewCaseService(
	caseRepo repository.ICaseRepository,
	caseVersionRepo repository.ICaseVersionRepository,
	caseEvidenceRepo repository.ICaseEvidenceRepository,
	caseQuestionRepo repository.ICaseQuestionRepository,
	caseChatbotConfigRepo repository.ICaseChatbotConfigRepository,
	storage supabase.Interface,
) ICaseService {
	return &CaseService{
		db:                    mariadb.Connection,
		caseRepo:              caseRepo,
		caseVersionRepo:       caseVersionRepo,
		caseEvidenceRepo:      caseEvidenceRepo,
		caseQuestionRepo:      caseQuestionRepo,
		caseChatbotConfigRepo: caseChatbotConfigRepo,
		storage:               storage,
	}
}
