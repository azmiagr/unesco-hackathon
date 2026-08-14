package service

import (
	"errors"
	"math"

	"github.com/azmiagr/unesco-hackathon/internal/repository"
	"github.com/azmiagr/unesco-hackathon/model"
	"github.com/azmiagr/unesco-hackathon/pkg/database/mariadb"
	appErrors "github.com/azmiagr/unesco-hackathon/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	defaultProfileSeasonLabel = "SEASON 1"
	defaultConnectedProvider  = "local"
	defaultNextUnlockText     = ""
)

type IProfileService interface {
	GetUserProfile(userID uuid.UUID) (*model.GetUserProfileResponse, error)
}

type ProfileService struct {
	db              *gorm.DB
	userProfileRepo repository.IUserProfileRepository
	gameLevelRepo   repository.IGameLevelRepository
}

func NewProfileService(
	userProfileRepo repository.IUserProfileRepository,
	gameLevelRepo repository.IGameLevelRepository,
) IProfileService {
	return &ProfileService{
		db:              mariadb.Connection,
		userProfileRepo: userProfileRepo,
		gameLevelRepo:   gameLevelRepo,
	}
}

func (s *ProfileService) GetUserProfile(userID uuid.UUID) (*model.GetUserProfileResponse, error) {
	if userID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid user id")
	}

	profile, err := s.userProfileRepo.GetUserProfileDetail(s.db, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("user profile not found")
		}
		return nil, appErrors.InternalServer("failed to get user profile")
	}

	nextLevel, err := s.gameLevelRepo.GetNextGameLevel(s.db, profile.CurrentLevel)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appErrors.InternalServer("failed to get next level")
	}

	nextLevelXP := 0
	nextLevelNumber := profile.CurrentLevel
	if nextLevel != nil {
		nextLevelXP = nextLevel.XPRequired
		nextLevelNumber = nextLevel.Level
	}

	levelProgress := mapUserProfileLevelProgress(profile, nextLevelXP, nextLevelNumber)

	return &model.GetUserProfileResponse{
		Profile: model.UserProfileSummaryResponse{
			UserID:            profile.UserID,
			Username:          profile.Username,
			Email:             profile.Email,
			AvatarID:          profile.AvatarID,
			AvatarURL:         profile.AvatarURL,
			Title:             profile.Title,
			CurrentLevel:      profile.CurrentLevel,
			CurrentXP:         profile.CurrentXP,
			CoinBalance:       profile.CoinBalance,
			AuditorReputation: profile.AuditorReputation,
			AccuracyPercent:   0,
			CasesCompleted:    0,
			StreakCount:       0,
			SeasonLabel:       defaultProfileSeasonLabel,
		},
		LevelProgress: levelProgress,
		Stats: []model.UserProfileDetectiveStatResponse{
			{Key: "evidence_evaluation", Label: "Evidence Evaluation", Score: profile.EvidenceEvaluationScore, Average: 74},
			{Key: "claim_analysis", Label: "Claim Analysis", Score: profile.ClaimAnalysisScore, Average: 74},
			{Key: "confidence_calibration", Label: "Confidence Calibration", Score: profile.ConfidenceCalibrationScore, Average: 74},
			{Key: "reasoning", Label: "Reasoning", Score: profile.ReasoningScore, Average: 74},
			{Key: "safety_judgment", Label: "Safety Judgment", Score: profile.SafetyJudgmentScore, Average: 74},
		},
		Account: model.UserProfileAccountResponse{
			Email:           profile.Email,
			IsEmailVerified: true,
			ConnectedTo:     defaultConnectedProvider,
		},
		CaseHistory: model.UserCaseHistoryResponse{
			Items: []model.UserCaseHistoryItemResponse{},
		},
	}, nil
}

func mapUserProfileLevelProgress(profile *model.UserProfileDetailRow, nextLevelXP int, nextLevel int) model.UserProfileLevelProgressResponse {
	currentXP := profile.CurrentXP
	progressXP := currentXP
	remainingXP := 0
	progressPercent := 100

	if nextLevelXP > 0 {
		progressXP = currentXP
		if progressXP > nextLevelXP {
			progressXP = nextLevelXP
		}

		remainingXP = nextLevelXP - currentXP
		if remainingXP < 0 {
			remainingXP = 0
		}

		progressPercent = int(math.Round((float64(progressXP) / float64(nextLevelXP)) * 100))
	}

	return model.UserProfileLevelProgressResponse{
		CurrentLevel:    profile.CurrentLevel,
		NextLevel:       nextLevel,
		CurrentXP:       currentXP,
		NextLevelXP:     nextLevelXP,
		RemainingXP:     remainingXP,
		ProgressXP:      progressXP,
		ProgressPercent: progressPercent,
		NextUnlockText:  defaultNextUnlockText,
	}
}
