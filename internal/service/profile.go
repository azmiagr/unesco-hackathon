package service

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"time"

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
	defaultCaseHistoryLimit   = 5
)

type IProfileService interface {
	GetUserProfile(userID uuid.UUID) (*model.GetUserProfileResponse, error)
}

type ProfileService struct {
	db              *gorm.DB
	userProfileRepo repository.IUserProfileRepository
	gameLevelRepo   repository.IGameLevelRepository
	caseSessionRepo repository.ICaseSessionRepository
}

func NewProfileService(
	userProfileRepo repository.IUserProfileRepository,
	gameLevelRepo repository.IGameLevelRepository,
	caseSessionRepo repository.ICaseSessionRepository,
) IProfileService {
	return &ProfileService{
		db:              mariadb.Connection,
		userProfileRepo: userProfileRepo,
		gameLevelRepo:   gameLevelRepo,
		caseSessionRepo: caseSessionRepo,
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

	currentLevel := max(profile.CurrentLevel, 1)
	levelForXP, err := s.gameLevelRepo.GetGameLevelForXP(s.db, profile.CurrentXP)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appErrors.InternalServer("failed to resolve level from xp")
	}
	if levelForXP != nil {
		currentLevel = levelForXP.Level
	}
	if currentLevel != profile.CurrentLevel {
		if err := s.userProfileRepo.SyncUserProfileLevel(s.db, userID, currentLevel); err != nil {
			return nil, appErrors.InternalServer("failed to synchronize level from xp")
		}
	}
	currentGameLevel, err := s.gameLevelRepo.GetGameLevel(s.db, model.GetGameLevelParam{Level: currentLevel})
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appErrors.InternalServer("failed to get current level")
	}

	nextLevel, err := s.gameLevelRepo.GetNextGameLevel(s.db, currentLevel)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appErrors.InternalServer("failed to get next level")
	}

	currentLevelXP := 0
	if currentGameLevel != nil {
		currentLevelXP = currentGameLevel.XPRequired
	}
	nextLevelXP := 0
	nextLevelNumber := currentLevel
	if nextLevel != nil {
		nextLevelXP = nextLevel.XPRequired
		nextLevelNumber = nextLevel.Level
	}

	levelProgress := mapUserProfileLevelProgress(profile, currentLevel, currentLevelXP, nextLevelXP, nextLevelNumber)
	resultSummary, err := s.caseSessionRepo.GetUserCaseResultSummary(s.db, userID)
	if err != nil {
		return nil, appErrors.InternalServer("failed to get case result summary")
	}
	caseHistoryRows, err := s.caseSessionRepo.ListUserCaseResultHistory(s.db, model.ListUserCaseResultHistoryParam{
		UserID: userID,
		Limit:  defaultCaseHistoryLimit,
	})
	if err != nil {
		return nil, appErrors.InternalServer("failed to get case history")
	}
	completionDates, err := s.caseSessionRepo.ListUserCaseCompletionDates(s.db, userID)
	if err != nil {
		return nil, appErrors.InternalServer("failed to get completion streak")
	}
	statAverages, err := s.userProfileRepo.GetDetectiveStatAverages(s.db)
	if err != nil {
		return nil, appErrors.InternalServer("failed to get detective stat averages")
	}

	return &model.GetUserProfileResponse{
		Profile: model.UserProfileSummaryResponse{
			UserID:            profile.UserID,
			Username:          profile.Username,
			Email:             profile.Email,
			AvatarID:          profile.AvatarID,
			AvatarURL:         profile.AvatarURL,
			TitleID:           profile.TitleID,
			Title:             profile.Title,
			TitleImageBorder:  profile.TitleImageBorder,
			CurrentLevel:      currentLevel,
			CurrentXP:         profile.CurrentXP,
			CoinBalance:       profile.CoinBalance,
			AuditorReputation: profile.AuditorReputation,
			AccuracyPercent:   roundFloat(resultSummary.AccuracyScore, 2),
			CasesCompleted:    resultSummary.CasesCompleted,
			StreakCount:       calculateCompletionStreak(completionDates),
			SeasonLabel:       defaultProfileSeasonLabel,
		},
		LevelProgress: levelProgress,
		Stats: []model.UserProfileDetectiveStatResponse{
			{Key: "evidence_evaluation", Label: "Evidence Evaluation", Score: profile.EvidenceEvaluationScore, Average: roundFloat(statAverages.EvidenceEvaluationAverage, 2)},
			{Key: "claim_analysis", Label: "Claim Analysis", Score: profile.ClaimAnalysisScore, Average: roundFloat(statAverages.ClaimAnalysisAverage, 2)},
			{Key: "confidence_calibration", Label: "Confidence Calibration", Score: profile.ConfidenceCalibrationScore, Average: roundFloat(statAverages.ConfidenceCalibrationAverage, 2)},
			{Key: "reasoning", Label: "Reasoning", Score: profile.ReasoningScore, Average: roundFloat(statAverages.ReasoningAverage, 2)},
			{Key: "safety_judgment", Label: "Safety Judgment", Score: profile.SafetyJudgmentScore, Average: roundFloat(statAverages.SafetyJudgmentAverage, 2)},
		},
		Account: model.UserProfileAccountResponse{
			Email:           profile.Email,
			IsEmailVerified: profile.Status == "active",
			ConnectedTo:     defaultConnectedProvider,
		},
		CaseHistory: model.UserCaseHistoryResponse{
			Items: mapUserCaseHistory(caseHistoryRows),
		},
	}, nil
}

func mapUserProfileLevelProgress(profile *model.UserProfileDetailRow, currentLevel int, currentLevelXP int, nextLevelXP int, nextLevel int) model.UserProfileLevelProgressResponse {
	currentXP := profile.CurrentXP
	progressXP := max(currentXP-currentLevelXP, 0)
	remainingXP := 0
	progressPercent := 100

	if nextLevelXP > currentLevelXP {
		remainingXP = nextLevelXP - currentXP
		if remainingXP < 0 {
			remainingXP = 0
		}

		levelRangeXP := nextLevelXP - currentLevelXP
		if progressXP > levelRangeXP {
			progressXP = levelRangeXP
		}
		progressPercent = int(math.Round((float64(progressXP) / float64(levelRangeXP)) * 100))
		progressPercent = min(max(progressPercent, 0), 100)
	}

	return model.UserProfileLevelProgressResponse{
		CurrentLevel:    currentLevel,
		NextLevel:       nextLevel,
		CurrentXP:       currentXP,
		NextLevelXP:     nextLevelXP,
		RemainingXP:     remainingXP,
		ProgressXP:      progressXP,
		ProgressPercent: progressPercent,
		NextUnlockText:  buildNextUnlockText(nextLevel, nextLevelXP),
	}
}

func mapUserCaseHistory(rows []model.UserCaseResultHistoryRow) []model.UserCaseHistoryItemResponse {
	items := make([]model.UserCaseHistoryItemResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, model.UserCaseHistoryItemResponse{
			CaseID:          row.CaseID,
			Title:           row.Title,
			CompletedAt:     row.CompletedAt.Format(time.RFC3339),
			DifficultyLabel: mapDifficultyLabel(row.DifficultyLevel),
			XPReward:        row.XPGained,
			ResultStatus:    row.OutcomeLabel,
			ScoreLabel:      mapScoreLabel(row.TotalScore),
			IsRetryable:     row.CaseStatus == model.CaseStatusPublished,
		})
	}
	return items
}

func buildNextUnlockText(nextLevel int, nextLevelXP int) string {
	if nextLevel <= 0 || nextLevelXP <= 0 {
		return defaultNextUnlockText
	}
	return "Level " + strconv.Itoa(nextLevel) + " terbuka di " + strconv.Itoa(nextLevelXP) + " XP"
}

func mapDifficultyLabel(difficulty string) string {
	switch strings.ToLower(strings.TrimSpace(difficulty)) {
	case "low", "easy":
		return "Mudah"
	case "medium":
		return "Sedang"
	case "high", "hard":
		return "Sulit"
	default:
		if strings.TrimSpace(difficulty) == "" {
			return "-"
		}
		return difficulty
	}
}

func mapScoreLabel(score int) string {
	switch {
	case score >= 90:
		return "Sempurna"
	case score >= 80:
		return "Bagus"
	case score >= 70:
		return "Cukup"
	default:
		return "Perlu Latihan"
	}
}

func calculateCompletionStreak(completionDates []time.Time) int {
	if len(completionDates) == 0 {
		return 0
	}

	streak := 0
	expectedDate := dateOnly(completionDates[0])
	for _, completedAt := range completionDates {
		completedDate := dateOnly(completedAt)
		if completedDate.Equal(expectedDate) {
			streak++
			expectedDate = expectedDate.AddDate(0, 0, -1)
			continue
		}
		if completedDate.Before(expectedDate) {
			break
		}
	}

	return streak
}

func dateOnly(value time.Time) time.Time {
	utc := value.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}

func roundFloat(value float64, precision int) float64 {
	if precision < 0 {
		return value
	}
	multiplier := math.Pow(10, float64(precision))
	return math.Round(value*multiplier) / multiplier
}
