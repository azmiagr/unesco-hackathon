package service

import (
	"encoding/json"
	"errors"
	"math"
	"time"

	"github.com/azmiagr/unesco-hackathon/entity"
	"github.com/azmiagr/unesco-hackathon/internal/repository"
	"github.com/azmiagr/unesco-hackathon/model"
	"github.com/azmiagr/unesco-hackathon/pkg/database/mariadb"
	appErrors "github.com/azmiagr/unesco-hackathon/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const userLobbyCaseLimit = 5

type ILobbyService interface {
	GetLobbyForUser(userID uuid.UUID) (*model.UserLobbyResponse, error)
}

type LobbyService struct {
	db              *gorm.DB
	userProfileRepo repository.IUserProfileRepository
	gameLevelRepo   repository.IGameLevelRepository
	caseRepo        repository.ICaseRepository
	cityStatsRepo   repository.ICityStatisticsRepository
	caseSessionRepo repository.ICaseSessionRepository
}

func NewLobbyService(
	userProfileRepo repository.IUserProfileRepository,
	gameLevelRepo repository.IGameLevelRepository,
	caseRepo repository.ICaseRepository,
	cityStatsRepo repository.ICityStatisticsRepository,
	caseSessionRepo repository.ICaseSessionRepository,
) ILobbyService {
	return &LobbyService{
		db:              mariadb.Connection,
		userProfileRepo: userProfileRepo,
		gameLevelRepo:   gameLevelRepo,
		caseRepo:        caseRepo,
		cityStatsRepo:   cityStatsRepo,
		caseSessionRepo: caseSessionRepo,
	}
}

func (s *LobbyService) GetLobbyForUser(userID uuid.UUID) (*model.UserLobbyResponse, error) {
	if userID == uuid.Nil {
		return nil, appErrors.Unauthorized("unauthorized")
	}

	profile, err := s.userProfileRepo.GetUserProfileDetail(s.db, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("user profile not found")
		}
		return nil, appErrors.InternalServer("failed to get user profile")
	}
	if profile == nil {
		return nil, appErrors.InternalServer("user profile not found")
	}

	currentLevelNumber := max(profile.CurrentLevel, 1)

	currentLevel, err := s.getCurrentGameLevel(currentLevelNumber)
	if err != nil {
		return nil, err
	}

	nextLevel, err := s.gameLevelRepo.GetNextGameLevel(s.db, currentLevelNumber)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appErrors.InternalServer("failed to get next level")
	}

	cityStats, err := s.cityStatsRepo.GetCityStatistics(s.db, model.GetCityStatisticsParam{
		StatKey: model.CityStatisticsDefaultKey,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			cityStats = defaultCityStatistics()
		} else {
			return nil, appErrors.InternalServer("failed to get city statistics")
		}
	}
	if cityStats == nil {
		cityStats = defaultCityStatistics()
	}
	cityStatDeltas, err := s.getLatestCityStatDeltas(userID)
	if err != nil {
		return nil, err
	}

	caseRows, _, err := s.caseRepo.ListPublishedCasesForUser(s.db, model.ListUserCasesParam{
		Limit:  userLobbyCaseLimit,
		Offset: 0,
	})
	if err != nil {
		return nil, appErrors.InternalServer("failed to list lobby cases")
	}

	cases := make([]model.UserCaseCardResponse, 0, len(caseRows))
	for _, caseRow := range caseRows {
		cases = append(cases, mapUserCaseCardResponse(caseRow, &entity.UserProfile{
			CurrentLevel:      currentLevelNumber,
			AuditorReputation: profile.AuditorReputation,
		}))
	}

	var featuredCase *model.UserCaseCardResponse
	if len(cases) > 0 {
		featuredCase = &cases[0]
	}

	otherCases := []model.UserCaseCardResponse{}
	if len(cases) > 1 {
		otherCases = cases[1:]
	}

	continueCase, err := s.getContinueCaseForUser(userID)
	if err != nil {
		return nil, err
	}

	return &model.UserLobbyResponse{
		Profile: model.UserLobbyProfileResponse{
			UserID:      profile.UserID,
			Username:    profile.Username,
			AvatarID:    profile.AvatarID,
			AvatarURL:   profile.AvatarURL,
			Title:       profile.Title,
			CoinBalance: profile.CoinBalance,
		},
		Level:        mapUserLobbyLevelProgress(profile, currentLevelNumber, currentLevel, nextLevel),
		VisualState:  cityStats.VisualState,
		CityStats:    mapUserLobbyCityStats(cityStats, cityStatDeltas),
		FeaturedCase: featuredCase,
		ContinueCase: continueCase,
		OtherCases:   otherCases,
	}, nil
}

func (s *LobbyService) getContinueCaseForUser(userID uuid.UUID) (*model.UserLobbyContinueCaseResponse, error) {
	sessions, err := s.caseSessionRepo.ListRecentCaseSessions(s.db, model.ListRecentCaseSessionsParam{
		UserID: userID,
		Status: model.CaseSessionStatusActive,
		Limit:  1,
	})
	if err != nil {
		return nil, appErrors.InternalServer("failed to get active session")
	}
	if len(sessions) == 0 {
		return nil, nil
	}

	session := sessions[0]
	snapshot, err := parseGameplaySnapshot(session.SessionSnapshot)
	if err != nil {
		return nil, err
	}

	progressRows, err := s.caseSessionRepo.ListCaseSessionEvidenceProgress(s.db, model.ListCaseSessionEvidenceProgressParam{
		CaseSessionID: session.CaseSessionID,
	})
	if err != nil {
		return nil, appErrors.InternalServer("failed to list evidence progress")
	}
	answerRows, err := s.caseSessionRepo.ListCaseSessionAnswers(s.db, model.ListCaseSessionAnswersParam{
		CaseSessionID: session.CaseSessionID,
	})
	if err != nil {
		return nil, appErrors.InternalServer("failed to list answers")
	}

	requiredQuestionIDs := map[uuid.UUID]bool{}
	for _, question := range snapshot.Questions {
		if question.IsRequired {
			requiredQuestionIDs[question.CaseQuestionID] = true
		}
	}
	answeredRequired := map[uuid.UUID]bool{}
	for _, answer := range answerRows {
		if answer.IsFinal && requiredQuestionIDs[answer.CaseQuestionID] {
			answeredRequired[answer.CaseQuestionID] = true
		}
	}

	totalEvidenceCount := len(snapshot.Evidences)
	requiredQuestionCount := len(requiredQuestionIDs)
	totalProgressUnits := totalEvidenceCount + requiredQuestionCount
	completedProgressUnits := len(progressRows) + len(answeredRequired)
	progressPercent := 0
	if totalProgressUnits > 0 {
		progressPercent = int(math.Round(float64(completedProgressUnits) / float64(totalProgressUnits) * 100))
		progressPercent = min(max(progressPercent, 0), 100)
	}

	return &model.UserLobbyContinueCaseResponse{
		CaseID:                   session.CaseID,
		CaseSessionID:            session.CaseSessionID,
		CaseVersionID:            session.CaseVersionID,
		Title:                    snapshot.Case.Title,
		Slug:                     snapshot.Case.Slug,
		ShortDescription:         snapshot.Case.ShortDescription,
		DifficultyLevel:          snapshot.Case.DifficultyLevel,
		EstimatedDurationMinutes: snapshot.Case.EstimatedDurationMinutes,
		ThumbnailURL:             snapshot.Case.ThumbnailURL,
		SessionVersion:           session.SessionVersion,
		ProgressPercent:          progressPercent,
		OpenedEvidenceCount:      len(progressRows),
		TotalEvidenceCount:       totalEvidenceCount,
		AnsweredQuestionCount:    len(answeredRequired),
		RequiredQuestionCount:    requiredQuestionCount,
		LastActivityAt:           session.LastActivityAt.Format(time.RFC3339),
		StartedAt:                session.StartedAt.Format(time.RFC3339),
	}, nil
}

func defaultCityStatistics() *entity.CityStatistics {
	return &entity.CityStatistics{
		StatKey:           model.CityStatisticsDefaultKey,
		InformationHealth: 70,
		PublicTrust:       70,
		SocialStability:   70,
		PublicWellbeing:   70,
		VisualState:       "normal",
	}
}

func (s *LobbyService) getLatestCityStatDeltas(userID uuid.UUID) (map[string]int, error) {
	result, err := s.caseSessionRepo.GetLatestUserCaseSessionResult(s.db, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return map[string]int{}, nil
		}
		return nil, appErrors.InternalServer("failed to get latest city impact")
	}

	var impacts []model.GameplayCityImpactResponse
	if err := json.Unmarshal([]byte(result.CityImpact), &impacts); err != nil {
		return nil, appErrors.InternalServer("failed to parse latest city impact")
	}

	deltas := make(map[string]int, len(impacts))
	for _, impact := range impacts {
		deltas[impact.Key] = impact.Delta
	}

	return deltas, nil
}

func mapUserLobbyCityStats(stats *entity.CityStatistics, deltas map[string]int) []model.UserLobbyCityStatResponse {
	return []model.UserLobbyCityStatResponse{
		mapUserLobbyCityStat(model.CityImpactHealth, "Information Health", stats.InformationHealth, deltas[model.CityImpactHealth]),
		mapUserLobbyCityStat(model.CityImpactTrust, "Public Trust", stats.PublicTrust, deltas[model.CityImpactTrust]),
		mapUserLobbyCityStat(model.CityImpactStability, "Social Stability", stats.SocialStability, deltas[model.CityImpactStability]),
		mapUserLobbyCityStat(model.CityImpactWellbeing, "Public Wellbeing", stats.PublicWellbeing, deltas[model.CityImpactWellbeing]),
	}
}

func mapUserLobbyCityStat(key string, label string, value int, delta int) model.UserLobbyCityStatResponse {
	return model.UserLobbyCityStatResponse{
		Key:    key,
		Label:  label,
		Value:  clampCityStat(value),
		Delta:  delta,
		Status: cityStatStatus(value),
	}
}

func clampCityStat(value int) int {
	return min(max(value, 0), 100)
}

func cityStatStatus(value int) string {
	value = clampCityStat(value)
	if value >= 70 {
		return "aman"
	}
	if value >= 40 {
		return "terancam"
	}
	return "kritis"
}

func (s *LobbyService) getCurrentGameLevel(currentLevelNumber int) (*entity.GameLevel, error) {
	if currentLevelNumber < 1 {
		currentLevelNumber = 1
	}

	level, err := s.gameLevelRepo.GetGameLevel(s.db, model.GetGameLevelParam{Level: currentLevelNumber})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, appErrors.InternalServer("failed to get current level")
	}

	return level, nil
}

func mapUserLobbyLevelProgress(
	profile *model.UserProfileDetailRow,
	currentLevelNumber int,
	currentLevel *entity.GameLevel,
	nextLevel *entity.GameLevel,
) model.UserLobbyLevelProgressResponse {
	currentLevelXP := 0
	title := profile.Title
	if currentLevel != nil {
		currentLevelXP = currentLevel.XPRequired
		if title == "" {
			title = currentLevel.Title
		}
	}

	nextLevelNumber := currentLevelNumber
	nextLevelXP := profile.CurrentXP
	if nextLevel != nil {
		nextLevelNumber = nextLevel.Level
		nextLevelXP = nextLevel.XPRequired
	}

	progressXP := max(profile.CurrentXP-currentLevelXP, 0)
	remainingXP := max(nextLevelXP-profile.CurrentXP, 0)
	progressPercent := 100
	if nextLevelXP > currentLevelXP {
		progressPercent = int(math.Round(float64(progressXP) / float64(nextLevelXP-currentLevelXP) * 100))
		progressPercent = min(max(progressPercent, 0), 100)
	}

	return model.UserLobbyLevelProgressResponse{
		CurrentLevel:    currentLevelNumber,
		CurrentXP:       profile.CurrentXP,
		CurrentLevelXP:  currentLevelXP,
		NextLevel:       nextLevelNumber,
		NextLevelXP:     nextLevelXP,
		ProgressXP:      progressXP,
		RemainingXP:     remainingXP,
		ProgressPercent: progressPercent,
		Title:           title,
	}
}
