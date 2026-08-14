package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/azmiagr/unesco-hackathon/entity"
	"github.com/azmiagr/unesco-hackathon/internal/repository"
	"github.com/azmiagr/unesco-hackathon/model"
	"github.com/azmiagr/unesco-hackathon/pkg/database/mariadb"
	appErrors "github.com/azmiagr/unesco-hackathon/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	gameplayIdempotencyTTL = 24 * time.Hour
)

type IGameplayService interface {
	StartCaseSessionForUser(userID uuid.UUID, caseID uuid.UUID, req model.StartCaseSessionRequest, idempotencyKey string) (*model.StartCaseSessionResponse, error)
	GetGameplayStateForUser(userID uuid.UUID, caseSessionID uuid.UUID) (*model.GameplayStateResponse, error)
	OpenEvidenceForUser(userID uuid.UUID, caseSessionID uuid.UUID, caseEvidenceID uuid.UUID, req model.OpenCaseSessionEvidenceRequest, idempotencyKey string, requestMethod string, requestPath string) (*model.OpenCaseSessionEvidenceResponse, error)
	SaveAnswersForUser(userID uuid.UUID, caseSessionID uuid.UUID, req model.SaveCaseSessionAnswersRequest, idempotencyKey string, requestMethod string, requestPath string) (*model.SaveCaseSessionAnswersResponse, error)
	SubmitCaseSessionForUser(userID uuid.UUID, caseSessionID uuid.UUID, req model.SubmitCaseSessionRequest, idempotencyKey string, requestMethod string, requestPath string) (*model.SubmitCaseSessionResponse, error)
}

type GameplayService struct {
	db               *gorm.DB
	caseRepo         repository.ICaseRepository
	caseVersionRepo  repository.ICaseVersionRepository
	caseEvidenceRepo repository.ICaseEvidenceRepository
	caseQuestionRepo repository.ICaseQuestionRepository
	caseSessionRepo  repository.ICaseSessionRepository
	userProfileRepo  repository.IUserProfileRepository
	gameConfigRepo   repository.IGameConfigRepository
	caseScoringRepo  repository.ICaseScoringOutcomeRepository
	gameLevelRepo    repository.IGameLevelRepository
	cityStatsRepo    repository.ICityStatisticsRepository
}

func NewGameplayService(
	caseRepo repository.ICaseRepository,
	caseVersionRepo repository.ICaseVersionRepository,
	caseEvidenceRepo repository.ICaseEvidenceRepository,
	caseQuestionRepo repository.ICaseQuestionRepository,
	caseSessionRepo repository.ICaseSessionRepository,
	userProfileRepo repository.IUserProfileRepository,
	gameConfigRepo repository.IGameConfigRepository,
	caseScoringRepo repository.ICaseScoringOutcomeRepository,
	gameLevelRepo repository.IGameLevelRepository,
	cityStatsRepo repository.ICityStatisticsRepository,
) IGameplayService {
	return &GameplayService{
		db:               mariadb.Connection,
		caseRepo:         caseRepo,
		caseVersionRepo:  caseVersionRepo,
		caseEvidenceRepo: caseEvidenceRepo,
		caseQuestionRepo: caseQuestionRepo,
		caseSessionRepo:  caseSessionRepo,
		userProfileRepo:  userProfileRepo,
		gameConfigRepo:   gameConfigRepo,
		caseScoringRepo:  caseScoringRepo,
		gameLevelRepo:    gameLevelRepo,
		cityStatsRepo:    cityStatsRepo,
	}
}

func (s *GameplayService) StartCaseSessionForUser(userID uuid.UUID, caseID uuid.UUID, req model.StartCaseSessionRequest, idempotencyKey string) (*model.StartCaseSessionResponse, error) {
	if userID == uuid.Nil {
		return nil, appErrors.Unauthorized("unauthorized")
	}
	if caseID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid case id")
	}

	initialAssessment, err := normalizeInitialAssessment(req.InitialAssessment)
	if err != nil {
		return nil, err
	}
	if req.InitialConfidence != nil && (*req.InitialConfidence < 0 || *req.InitialConfidence > 100) {
		return nil, appErrors.BadRequest("initial confidence must be between 0 and 100")
	}

	now := time.Now().UTC()
	tx := s.db.Begin()
	defer tx.Rollback()

	if idempotencyKey != "" {
		existing, err := s.caseSessionRepo.GetCaseSession(tx, model.GetCaseSessionParam{
			UserID:              userID,
			CaseID:              caseID,
			StartIdempotencyKey: idempotencyKey,
		})
		if err == nil {
			state, err := s.buildGameplayState(tx, existing)
			if err != nil {
				return nil, err
			}
			err = tx.Commit().Error
			if err != nil {
				return nil, appErrors.InternalServer("failed to commit transaction")
			}
			return &model.StartCaseSessionResponse{Session: state.Session, Gameplay: *state}, nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.InternalServer("failed to get existing session")
		}
	}

	config, err := s.getGameplayConfig(tx)
	if err != nil {
		return nil, err
	}
	if config.MaintenanceMode {
		return nil, appErrors.Conflict("game is currently under maintenance")
	}

	profile, err := s.userProfileRepo.GetUserProfileForUpdate(tx, model.GetUserProfileParam{UserID: userID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("user profile not found")
		}
		return nil, appErrors.InternalServer("failed to get user profile")
	}

	caseEntity, err := s.caseRepo.GetCase(tx, model.GetCaseParam{CaseID: caseID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("case not found")
		}
		return nil, appErrors.InternalServer("failed to get case")
	}
	if caseEntity.Status != model.CaseStatusPublished {
		return nil, appErrors.Conflict("case is not available")
	}
	if profile.CurrentLevel < caseEntity.MinimumLevel {
		return nil, appErrors.Conflict("minimum level requirement not met")
	}
	if profile.AuditorReputation < caseEntity.MinimumReputation {
		return nil, appErrors.Conflict("minimum reputation requirement not met")
	}

	activeSession, err := s.caseSessionRepo.GetActiveCaseSession(tx, userID, caseID)
	if err == nil {
		state, err := s.buildGameplayState(tx, activeSession)
		if err != nil {
			return nil, err
		}
		err = tx.Commit().Error
		if err != nil {
			return nil, appErrors.InternalServer("failed to commit transaction")
		}
		return &model.StartCaseSessionResponse{Session: state.Session, Gameplay: *state}, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appErrors.InternalServer("failed to get active session")
	}

	if config.MaxCasesPerDay > 0 {
		dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		startedCount, err := s.caseSessionRepo.CountCaseSessionsStarted(tx, model.CountCaseSessionsStartedParam{
			UserID:   userID,
			StartAt:  dayStart,
			EndAt:    dayStart.Add(24 * time.Hour),
			Statuses: []string{model.CaseSessionStatusActive, model.CaseSessionStatusCompleted, model.CaseSessionStatusAbandoned},
		})
		if err != nil {
			return nil, appErrors.InternalServer("failed to count started sessions")
		}
		if startedCount >= int64(config.MaxCasesPerDay) {
			return nil, appErrors.TooManyRequests("daily case limit reached")
		}
	}

	if config.CooldownBetweenCasesMinutes > 0 {
		recentSessions, err := s.caseSessionRepo.ListRecentCaseSessions(tx, model.ListRecentCaseSessionsParam{
			UserID: userID,
			Limit:  1,
		})
		if err != nil {
			return nil, appErrors.InternalServer("failed to get recent session")
		}
		if len(recentSessions) > 0 && now.Sub(recentSessions[0].StartedAt) < time.Duration(config.CooldownBetweenCasesMinutes)*time.Minute {
			return nil, appErrors.TooManyRequests("case cooldown is still active")
		}
	}

	caseVersion, err := s.getLatestPublishedCaseVersion(tx, caseID)
	if err != nil {
		return nil, err
	}

	snapshot, err := s.buildSessionSnapshot(tx, caseEntity, caseVersion)
	if err != nil {
		return nil, err
	}
	rawSnapshot, err := json.Marshal(snapshot)
	if err != nil {
		return nil, appErrors.InternalServer("failed to build session snapshot")
	}

	caseSessionID := uuid.New()
	activeSessionKey := fmt.Sprintf("%s:%s", userID.String(), caseID.String())
	var startIdempotencyKey *string
	if idempotencyKey != "" {
		key := idempotencyKey
		startIdempotencyKey = &key
	}

	session := &entity.CaseSession{
		CaseSessionID:       caseSessionID,
		UserID:              userID,
		CaseID:              caseID,
		CaseVersionID:       caseVersion.CaseVersionID,
		ActiveSessionKey:    &activeSessionKey,
		StartIdempotencyKey: startIdempotencyKey,
		SessionSnapshot:     string(rawSnapshot),
		SessionVersion:      1,
		Status:              model.CaseSessionStatusActive,
		InitialAssessment:   initialAssessment,
		InitialConfidence:   req.InitialConfidence,
		StartedAt:           now,
		LastActivityAt:      now,
		AppVersion:          req.AppVersion,
	}
	err = s.caseSessionRepo.CreateCaseSession(tx, session)
	if err != nil {
		return nil, appErrors.Conflict("active session already exists")
	}

	err = s.caseSessionRepo.CreateCaseSessionEvent(tx, &entity.CaseSessionEvent{
		CaseSessionEventID: uuid.New(),
		CaseSessionID:      caseSessionID,
		EventType:          model.CaseSessionEventSessionStarted,
		SessionVersion:     session.SessionVersion,
		CreatedAt:          now,
	})
	if err != nil {
		return nil, appErrors.InternalServer("failed to create session event")
	}

	state, err := s.buildGameplayStateFromSnapshot(tx, session, snapshot)
	if err != nil {
		return nil, err
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	return &model.StartCaseSessionResponse{Session: state.Session, Gameplay: *state}, nil
}

func (s *GameplayService) GetGameplayStateForUser(userID uuid.UUID, caseSessionID uuid.UUID) (*model.GameplayStateResponse, error) {
	if userID == uuid.Nil {
		return nil, appErrors.Unauthorized("unauthorized")
	}
	if caseSessionID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid session id")
	}

	session, err := s.caseSessionRepo.GetCaseSession(s.db, model.GetCaseSessionParam{
		CaseSessionID: caseSessionID,
		UserID:        userID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("session not found")
		}
		return nil, appErrors.InternalServer("failed to get session")
	}

	return s.buildGameplayState(s.db, session)
}

func (s *GameplayService) OpenEvidenceForUser(
	userID uuid.UUID,
	caseSessionID uuid.UUID,
	caseEvidenceID uuid.UUID,
	req model.OpenCaseSessionEvidenceRequest,
	idempotencyKey string,
	requestMethod string,
	requestPath string,
) (*model.OpenCaseSessionEvidenceResponse, error) {
	if idempotencyKey == "" {
		return nil, appErrors.BadRequest("idempotency key is required")
	}

	var replay model.OpenCaseSessionEvidenceResponse
	if ok, err := s.replayIdempotencyKey(userID, idempotencyKey, requestMethod, requestPath, &replay); err != nil {
		return nil, err
	} else if ok {
		return &replay, nil
	}

	if req.SessionVersion < 1 {
		return nil, appErrors.BadRequest("session version is required")
	}

	now := time.Now().UTC()
	tx := s.db.Begin()
	defer tx.Rollback()

	session, err := s.lockActiveUserSession(tx, userID, caseSessionID)
	if err != nil {
		return nil, err
	}

	snapshot, err := parseGameplaySnapshot(session.SessionSnapshot)
	if err != nil {
		return nil, err
	}
	evidence, ok := findSnapshotEvidence(snapshot, caseEvidenceID)
	if !ok {
		return nil, appErrors.NotFound("evidence not found in session")
	}

	rowsAffected, err := s.caseSessionRepo.IncrementCaseSessionVersion(tx, caseSessionID, req.SessionVersion, now)
	if err != nil {
		return nil, appErrors.InternalServer("failed to update session version")
	}
	if rowsAffected == 0 {
		return nil, appErrors.Conflict("session version conflict")
	}
	session.SessionVersion++
	session.LastActivityAt = now

	err = s.caseSessionRepo.UpsertCaseSessionEvidenceProgress(tx, &entity.CaseSessionEvidenceProgress{
		CaseSessionEvidenceProgressID: uuid.New(),
		CaseSessionID:                 caseSessionID,
		CaseEvidenceID:                caseEvidenceID,
		OpenedCount:                   1,
		FirstOpenedAt:                 now,
		LastOpenedAt:                  now,
	})
	if err != nil {
		return nil, appErrors.InternalServer("failed to update evidence progress")
	}

	payloadBytes, _ := json.Marshal(map[string]string{"case_evidence_id": caseEvidenceID.String()})
	payload := string(payloadBytes)
	err = s.caseSessionRepo.CreateCaseSessionEvent(tx, &entity.CaseSessionEvent{
		CaseSessionEventID: uuid.New(),
		CaseSessionID:      caseSessionID,
		EventType:          model.CaseSessionEventEvidenceOpened,
		CaseEvidenceID:     &caseEvidenceID,
		Payload:            &payload,
		SessionVersion:     session.SessionVersion,
		CreatedAt:          now,
	})
	if err != nil {
		return nil, appErrors.InternalServer("failed to create session event")
	}

	state, err := s.buildGameplayStateFromSnapshot(tx, session, snapshot)
	if err != nil {
		return nil, err
	}
	evidence.Opened = true
	result := &model.OpenCaseSessionEvidenceResponse{
		Session:          state.Session,
		Evidence:         evidence,
		EvidenceProgress: progressForEvidence(state.EvidenceProgress, caseEvidenceID),
		Progress:         state.Progress,
	}

	err = s.storeIdempotencyKey(tx, session.CaseSessionID, userID, idempotencyKey, requestMethod, requestPath, 200, result, now)
	if err != nil {
		return nil, err
	}
	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	return result, nil
}

func (s *GameplayService) SaveAnswersForUser(
	userID uuid.UUID,
	caseSessionID uuid.UUID,
	req model.SaveCaseSessionAnswersRequest,
	idempotencyKey string,
	requestMethod string,
	requestPath string,
) (*model.SaveCaseSessionAnswersResponse, error) {
	if idempotencyKey == "" {
		return nil, appErrors.BadRequest("idempotency key is required")
	}

	var replay model.SaveCaseSessionAnswersResponse
	if ok, err := s.replayIdempotencyKey(userID, idempotencyKey, requestMethod, requestPath, &replay); err != nil {
		return nil, err
	} else if ok {
		return &replay, nil
	}

	if req.SessionVersion < 1 {
		return nil, appErrors.BadRequest("session version is required")
	}
	if len(req.Answers) == 0 {
		return nil, appErrors.BadRequest("answers are required")
	}

	now := time.Now().UTC()
	tx := s.db.Begin()
	defer tx.Rollback()

	session, err := s.lockActiveUserSession(tx, userID, caseSessionID)
	if err != nil {
		return nil, err
	}

	snapshot, err := parseGameplaySnapshot(session.SessionSnapshot)
	if err != nil {
		return nil, err
	}
	questionMap := buildSnapshotQuestionMap(snapshot)

	rowsAffected, err := s.caseSessionRepo.IncrementCaseSessionVersion(tx, caseSessionID, req.SessionVersion, now)
	if err != nil {
		return nil, appErrors.InternalServer("failed to update session version")
	}
	if rowsAffected == 0 {
		return nil, appErrors.Conflict("session version conflict")
	}
	session.SessionVersion++
	session.LastActivityAt = now

	for _, answerReq := range req.Answers {
		question, ok := questionMap[answerReq.CaseQuestionID]
		if !ok {
			return nil, appErrors.NotFound("question not found in session")
		}
		if question.QuestionType != answerReq.QuestionType {
			return nil, appErrors.BadRequest("question type does not match")
		}
		if len(answerReq.Value) == 0 || !json.Valid(answerReq.Value) {
			return nil, appErrors.BadRequest("answer value must be valid json")
		}
		if answerReq.ConfidenceInitial != nil && (*answerReq.ConfidenceInitial < 0 || *answerReq.ConfidenceInitial > 100) {
			return nil, appErrors.BadRequest("confidence initial must be between 0 and 100")
		}
		if answerReq.ConfidenceFinal != nil && (*answerReq.ConfidenceFinal < 0 || *answerReq.ConfidenceFinal > 100) {
			return nil, appErrors.BadRequest("confidence final must be between 0 and 100")
		}

		err := s.caseSessionRepo.UpsertCaseSessionAnswer(tx, &entity.CaseSessionAnswer{
			CaseSessionAnswerID: uuid.New(),
			CaseSessionID:       caseSessionID,
			CaseQuestionID:      answerReq.CaseQuestionID,
			QuestionType:        answerReq.QuestionType,
			Value:               string(answerReq.Value),
			ConfidenceInitial:   answerReq.ConfidenceInitial,
			ConfidenceFinal:     answerReq.ConfidenceFinal,
			IsFinal:             answerReq.IsFinal,
			SavedAt:             now,
		})
		if err != nil {
			return nil, appErrors.InternalServer("failed to save answer")
		}
	}

	payloadBytes, _ := json.Marshal(map[string]int{"answer_count": len(req.Answers)})
	payload := string(payloadBytes)
	err = s.caseSessionRepo.CreateCaseSessionEvent(tx, &entity.CaseSessionEvent{
		CaseSessionEventID: uuid.New(),
		CaseSessionID:      caseSessionID,
		EventType:          model.CaseSessionEventAnswerSaved,
		Payload:            &payload,
		SessionVersion:     session.SessionVersion,
		CreatedAt:          now,
	})
	if err != nil {
		return nil, appErrors.InternalServer("failed to create session event")
	}

	state, err := s.buildGameplayStateFromSnapshot(tx, session, snapshot)
	if err != nil {
		return nil, err
	}
	result := &model.SaveCaseSessionAnswersResponse{
		Session:  state.Session,
		Answers:  state.Answers,
		Progress: state.Progress,
	}

	err = s.storeIdempotencyKey(tx, session.CaseSessionID, userID, idempotencyKey, requestMethod, requestPath, 200, result, now)
	if err != nil {
		return nil, err
	}
	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	return result, nil
}

func (s *GameplayService) SubmitCaseSessionForUser(
	userID uuid.UUID,
	caseSessionID uuid.UUID,
	req model.SubmitCaseSessionRequest,
	idempotencyKey string,
	requestMethod string,
	requestPath string,
) (*model.SubmitCaseSessionResponse, error) {
	if idempotencyKey == "" {
		return nil, appErrors.BadRequest("idempotency key is required")
	}

	var replay model.SubmitCaseSessionResponse
	if ok, err := s.replayIdempotencyKey(userID, idempotencyKey, requestMethod, requestPath, &replay); err != nil {
		return nil, err
	} else if ok {
		return &replay, nil
	}

	finalDecision := strings.TrimSpace(req.FinalDecision)
	reason := strings.TrimSpace(req.Reason)
	if req.SessionVersion < 1 {
		return nil, appErrors.BadRequest("session version is required")
	}
	if finalDecision == "" {
		return nil, appErrors.BadRequest("final decision is required")
	}
	if req.FinalConfidence < 0 || req.FinalConfidence > 100 {
		return nil, appErrors.BadRequest("final confidence must be between 0 and 100")
	}
	if reason == "" {
		return nil, appErrors.BadRequest("reason is required")
	}

	now := time.Now().UTC()
	tx := s.db.Begin()
	defer tx.Rollback()

	session, err := s.lockActiveUserSession(tx, userID, caseSessionID)
	if err != nil {
		return nil, err
	}

	snapshot, err := parseGameplaySnapshot(session.SessionSnapshot)
	if err != nil {
		return nil, err
	}

	state, err := s.buildGameplayStateFromSnapshot(tx, session, snapshot)
	if err != nil {
		return nil, err
	}
	if !state.Progress.CanTakeDecision {
		return nil, appErrors.ConflictWithData("session is incomplete", buildIncompleteSessionDetail(state))
	}

	rowsAffected, err := s.caseSessionRepo.IncrementCaseSessionVersion(tx, caseSessionID, req.SessionVersion, now)
	if err != nil {
		return nil, appErrors.InternalServer("failed to update session version")
	}
	if rowsAffected == 0 {
		return nil, appErrors.Conflict("session version conflict")
	}
	session.SessionVersion++
	session.LastActivityAt = now

	config, err := s.getGameplayConfig(tx)
	if err != nil {
		return nil, err
	}
	if config.MaintenanceMode {
		return nil, appErrors.Conflict("game is currently under maintenance")
	}

	scoringConfig, err := s.caseScoringRepo.GetCaseScoringOutcomeConfig(tx, session.CaseVersionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.Conflict("case scoring outcome config not found")
		}
		return nil, appErrors.InternalServer("failed to get case scoring outcome config")
	}

	profile, err := s.userProfileRepo.GetUserProfileForUpdate(tx, model.GetUserProfileParam{UserID: userID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("user profile not found")
		}
		return nil, appErrors.InternalServer("failed to get user profile")
	}

	answersByQuestion := answersByQuestionID(state.Answers)
	progressByEvidence := evidenceProgressByEvidenceID(state.EvidenceProgress)
	scoreBreakdown, totalScore, err := scoreGameplaySession(scoringConfig, snapshot, answersByQuestion, progressByEvidence, session, finalDecision, req.FinalConfidence, reason)
	if err != nil {
		return nil, err
	}

	outcomeRule, err := selectOutcomeRule(scoringConfig.OutcomeRules, totalScore)
	if err != nil {
		return nil, err
	}

	cityImpact, err := s.applyCityImpact(tx, outcomeRule.CityImpactSettings)
	if err != nil {
		return nil, err
	}

	levelBefore := profile.CurrentLevel
	if levelBefore < 1 {
		levelBefore = 1
	}
	xpBefore := profile.CurrentXP
	reputationBefore := profile.AuditorReputation
	xpGained := calculateXPGained(snapshot.Case.EstimatedDurationMinutes, totalScore, config)
	coinGained := calculateCoinGained(xpGained, totalScore)
	xpAfter := xpBefore + xpGained
	levelAfter, levelRewardCoin, titleAfter, err := s.resolveLevelAfter(tx, levelBefore, xpAfter)
	if err != nil {
		return nil, err
	}
	coinGained += levelRewardCoin
	reputationAfter := clampFloat64(reputationBefore+float64(totalScore)/20, 0, 1000)

	profile.CurrentXP = xpAfter
	profile.CurrentLevel = levelAfter
	profile.CoinBalance += coinGained
	profile.AuditorReputation = reputationAfter
	profile.EvidenceEvaluationScore = mergeProfileScore(profile.EvidenceEvaluationScore, scoreByCategory(scoreBreakdown, model.ScoringCategoryEvidenceEvaluation))
	profile.ClaimAnalysisScore = mergeProfileScore(profile.ClaimAnalysisScore, scoreByCategory(scoreBreakdown, model.ScoringCategoryClaimAnalysis))
	profile.ConfidenceCalibrationScore = mergeProfileScore(profile.ConfidenceCalibrationScore, scoreByCategory(scoreBreakdown, model.ScoringCategoryConfidenceCalibration))
	profile.ReasoningScore = mergeProfileScore(profile.ReasoningScore, scoreByCategory(scoreBreakdown, model.ScoringCategoryReasoning))
	profile.SafetyJudgmentScore = mergeProfileScore(profile.SafetyJudgmentScore, scoreByCategory(scoreBreakdown, model.ScoringCategorySafetyJudgment))
	if titleAfter != "" {
		profile.Title = titleAfter
	}
	err = s.userProfileRepo.UpdateUserProfile(tx, profile)
	if err != nil {
		return nil, appErrors.InternalServer("failed to update user profile")
	}

	submittedAt := now
	err = s.caseSessionRepo.UpdateCaseSessionStatus(tx, caseSessionID, model.CaseSessionStatusCompleted, &submittedAt, true)
	if err != nil {
		return nil, appErrors.InternalServer("failed to complete session")
	}
	session.Status = model.CaseSessionStatusCompleted
	session.SubmittedAt = &submittedAt
	session.LastActivityAt = submittedAt

	scoreBreakdownJSON, err := json.Marshal(scoreBreakdown)
	if err != nil {
		return nil, appErrors.InternalServer("failed to serialize score breakdown")
	}
	cityImpactJSON, err := json.Marshal(cityImpact)
	if err != nil {
		return nil, appErrors.InternalServer("failed to serialize city impact")
	}

	err = s.caseSessionRepo.CreateCaseSessionResult(tx, &entity.CaseSessionResult{
		CaseSessionResultID: uuid.New(),
		CaseSessionID:       caseSessionID,
		UserID:              userID,
		CaseID:              session.CaseID,
		CaseVersionID:       session.CaseVersionID,
		TotalScore:          totalScore,
		ScoreBreakdown:      string(scoreBreakdownJSON),
		OutcomeKey:          outcomeRule.OutcomeKey,
		OutcomeID:           outcomeRule.OutcomeID,
		OutcomeLabel:        outcomeRule.OutcomeLabel,
		NarrativeText:       outcomeRule.NarrativeText,
		CityImpact:          string(cityImpactJSON),
		FinalDecision:       finalDecision,
		FinalConfidence:     req.FinalConfidence,
		Reason:              reason,
		XPGained:            xpGained,
		CoinGained:          coinGained,
		LevelBefore:         levelBefore,
		LevelAfter:          levelAfter,
		XPBefore:            xpBefore,
		XPAfter:             xpAfter,
		ReputationBefore:    reputationBefore,
		ReputationAfter:     reputationAfter,
	})
	if err != nil {
		return nil, appErrors.InternalServer("failed to save session result")
	}

	payloadBytes, _ := json.Marshal(map[string]interface{}{"total_score": totalScore, "outcome_key": outcomeRule.OutcomeKey})
	payload := string(payloadBytes)
	err = s.caseSessionRepo.CreateCaseSessionEvent(tx, &entity.CaseSessionEvent{
		CaseSessionEventID: uuid.New(),
		CaseSessionID:      caseSessionID,
		EventType:          model.CaseSessionEventSubmitted,
		Payload:            &payload,
		SessionVersion:     session.SessionVersion,
		CreatedAt:          now,
	})
	if err != nil {
		return nil, appErrors.InternalServer("failed to create session event")
	}

	result := &model.SubmitCaseSessionResponse{
		Session: mapGameplaySessionResponse(session),
		Outcome: model.GameplayOutcomeResponse{
			OutcomeKey:   outcomeRule.OutcomeKey,
			OutcomeID:    outcomeRule.OutcomeID,
			OutcomeLabel: outcomeRule.OutcomeLabel,
			Narrative:    outcomeRule.NarrativeText,
			TotalScore:   totalScore,
		},
		ScoreBreakdown: scoreBreakdown,
		CityImpact:     cityImpact,
		Rewards: model.GameplayRewardResponse{
			XPGained:   xpGained,
			CoinGained: coinGained,
		},
		Progression: model.GameplayProgressionResponse{
			LevelBefore:      levelBefore,
			LevelAfter:       levelAfter,
			LevelUp:          levelAfter > levelBefore,
			XPBefore:         xpBefore,
			XPAfter:          xpAfter,
			CoinBalanceAfter: profile.CoinBalance,
			ReputationBefore: reputationBefore,
			ReputationAfter:  reputationAfter,
		},
		Feedback: buildGameplayFeedback(scoreBreakdown),
	}

	err = s.storeIdempotencyKey(tx, session.CaseSessionID, userID, idempotencyKey, requestMethod, requestPath, 200, result, now)
	if err != nil {
		return nil, err
	}
	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	return result, nil
}

func (s *GameplayService) getGameplayConfig(tx *gorm.DB) (*entity.GameConfig, error) {
	config, err := s.gameConfigRepo.GetGameConfig(tx, model.GetGameConfigParam{ConfigKey: model.GameConfigDefaultKey})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return defaultGameConfigEntity(), nil
		}
		return nil, appErrors.InternalServer("failed to get game config")
	}
	return config, nil
}

func (s *GameplayService) getLatestPublishedCaseVersion(tx *gorm.DB, caseID uuid.UUID) (*entity.CaseVersion, error) {
	versions, err := s.caseVersionRepo.ListCaseVersionsByCaseID(tx, caseID)
	if err != nil {
		return nil, appErrors.InternalServer("failed to list case versions")
	}

	for _, version := range versions {
		if version.Status == model.CaseStatusPublished {
			return &version, nil
		}
	}

	return nil, appErrors.Conflict("published case version not found")
}

func (s *GameplayService) buildSessionSnapshot(tx *gorm.DB, caseEntity *entity.Case, caseVersion *entity.CaseVersion) (*model.GameplaySessionSnapshot, error) {
	evidences, err := s.caseEvidenceRepo.ListCaseEvidences(tx, model.ListCaseEvidencesParam{CaseVersionID: caseVersion.CaseVersionID})
	if err != nil {
		return nil, appErrors.InternalServer("failed to list case evidences")
	}
	questions, err := s.caseQuestionRepo.ListCaseQuestions(tx, model.ListCaseQuestionsParam{CaseVersionID: caseVersion.CaseVersionID})
	if err != nil {
		return nil, appErrors.InternalServer("failed to list case questions")
	}
	if len(evidences) == 0 {
		return nil, appErrors.Conflict("case has no evidence")
	}
	if len(questions) == 0 {
		return nil, appErrors.Conflict("case has no questions")
	}

	snapshot := &model.GameplaySessionSnapshot{
		Case: model.GameplayCaseSnapshotResponse{
			CaseID:                   caseEntity.CaseID,
			CaseVersionID:            caseVersion.CaseVersionID,
			VersionNumber:            caseVersion.VersionNumber,
			Title:                    caseEntity.Title,
			Slug:                     caseEntity.Slug,
			ShortDescription:         caseEntity.ShortDescription,
			DifficultyLevel:          caseEntity.DifficultyLevel,
			RiskLevel:                caseEntity.RiskLevel,
			EstimatedDurationMinutes: caseEntity.EstimatedDurationMinutes,
			MinimumLevel:             caseEntity.MinimumLevel,
			MinimumReputation:        caseEntity.MinimumReputation,
			ThumbnailURL:             caseEntity.ThumbnailURL,
			PublishedAt:              caseEntity.PublishedAt,
		},
		Evidences: make([]model.GameplayEvidenceResponse, 0, len(evidences)),
		Questions: make([]model.GameplayQuestionResponse, 0, len(questions)),
	}

	for i, evidence := range evidences {
		snapshot.Evidences = append(snapshot.Evidences, mapGameplayEvidenceResponse(evidence, i+1))
	}
	for i, question := range questions {
		mappedQuestion, err := mapGameplayQuestionResponse(question, i+1)
		if err != nil {
			return nil, err
		}
		snapshot.Questions = append(snapshot.Questions, mappedQuestion)
	}

	return snapshot, nil
}

func (s *GameplayService) lockActiveUserSession(tx *gorm.DB, userID uuid.UUID, caseSessionID uuid.UUID) (*entity.CaseSession, error) {
	if userID == uuid.Nil {
		return nil, appErrors.Unauthorized("unauthorized")
	}
	if caseSessionID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid session id")
	}

	session, err := s.caseSessionRepo.GetCaseSessionForUpdate(tx, caseSessionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("session not found")
		}
		return nil, appErrors.InternalServer("failed to get session")
	}
	if session.UserID != userID {
		return nil, appErrors.NotFound("session not found")
	}
	if session.Status != model.CaseSessionStatusActive {
		return nil, appErrors.Conflict("session is not active")
	}

	return session, nil
}

func (s *GameplayService) buildGameplayState(tx *gorm.DB, session *entity.CaseSession) (*model.GameplayStateResponse, error) {
	snapshot, err := parseGameplaySnapshot(session.SessionSnapshot)
	if err != nil {
		return nil, err
	}

	return s.buildGameplayStateFromSnapshot(tx, session, snapshot)
}

func (s *GameplayService) buildGameplayStateFromSnapshot(tx *gorm.DB, session *entity.CaseSession, snapshot *model.GameplaySessionSnapshot) (*model.GameplayStateResponse, error) {
	progressRows, err := s.caseSessionRepo.ListCaseSessionEvidenceProgress(tx, model.ListCaseSessionEvidenceProgressParam{CaseSessionID: session.CaseSessionID})
	if err != nil {
		return nil, appErrors.InternalServer("failed to list evidence progress")
	}
	answerRows, err := s.caseSessionRepo.ListCaseSessionAnswers(tx, model.ListCaseSessionAnswersParam{CaseSessionID: session.CaseSessionID})
	if err != nil {
		return nil, appErrors.InternalServer("failed to list answers")
	}

	progressByEvidence := map[uuid.UUID]entity.CaseSessionEvidenceProgress{}
	for _, progress := range progressRows {
		progressByEvidence[progress.CaseEvidenceID] = progress
	}

	evidences := make([]model.GameplayEvidenceResponse, 0, len(snapshot.Evidences))
	evidenceProgress := make([]model.GameplayEvidenceProgressResponse, 0, len(snapshot.Evidences))
	for _, evidence := range snapshot.Evidences {
		progress, opened := progressByEvidence[evidence.CaseEvidenceID]
		evidence.Opened = opened
		evidences = append(evidences, evidence)
		if opened {
			evidenceProgress = append(evidenceProgress, model.GameplayEvidenceProgressResponse{
				CaseEvidenceID: evidence.CaseEvidenceID,
				Opened:         true,
				OpenedCount:    progress.OpenedCount,
				FirstOpenedAt:  progress.FirstOpenedAt,
				LastOpenedAt:   progress.LastOpenedAt,
			})
		} else {
			evidenceProgress = append(evidenceProgress, model.GameplayEvidenceProgressResponse{
				CaseEvidenceID: evidence.CaseEvidenceID,
				Opened:         false,
			})
		}
	}

	answers := make([]model.GameplayAnswerResponse, 0, len(answerRows))
	answeredRequired := map[uuid.UUID]bool{}
	for _, answer := range answerRows {
		rawValue := json.RawMessage(answer.Value)
		answers = append(answers, model.GameplayAnswerResponse{
			CaseQuestionID:    answer.CaseQuestionID,
			QuestionType:      answer.QuestionType,
			Value:             rawValue,
			ConfidenceInitial: answer.ConfidenceInitial,
			ConfidenceFinal:   answer.ConfidenceFinal,
			IsFinal:           answer.IsFinal,
			SavedAt:           answer.SavedAt,
		})
		if answer.IsFinal {
			answeredRequired[answer.CaseQuestionID] = true
		}
	}

	requiredQuestionCount := 0
	for _, question := range snapshot.Questions {
		if question.IsRequired {
			requiredQuestionCount++
		}
	}

	progress := model.GameplayProgressResponse{
		OpenedEvidenceCount:   len(progressRows),
		TotalEvidenceCount:    len(snapshot.Evidences),
		AnsweredQuestionCount: len(answeredRequired),
		RequiredQuestionCount: requiredQuestionCount,
	}
	progress.CanTakeDecision = progress.OpenedEvidenceCount >= progress.TotalEvidenceCount &&
		progress.AnsweredQuestionCount >= progress.RequiredQuestionCount

	return &model.GameplayStateResponse{
		Session:          mapGameplaySessionResponse(session),
		Case:             snapshot.Case,
		Evidences:        evidences,
		Questions:        snapshot.Questions,
		Answers:          answers,
		EvidenceProgress: evidenceProgress,
		Progress:         progress,
	}, nil
}

func (s *GameplayService) replayIdempotencyKey(userID uuid.UUID, idempotencyKey string, requestMethod string, requestPath string, target interface{}) (bool, error) {
	existing, err := s.caseSessionRepo.GetCaseSessionIdempotencyKey(s.db, model.GetCaseSessionIdempotencyKeyParam{
		IdempotencyKey: idempotencyKey,
		UserID:         userID,
		RequestMethod:  requestMethod,
		RequestPath:    requestPath,
		Now:            time.Now().UTC(),
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, appErrors.InternalServer("failed to get idempotency key")
	}

	if err := json.Unmarshal([]byte(existing.ResponseBody), target); err != nil {
		return false, appErrors.InternalServer("failed to replay idempotent response")
	}

	return true, nil
}

func (s *GameplayService) storeIdempotencyKey(tx *gorm.DB, caseSessionID uuid.UUID, userID uuid.UUID, idempotencyKey string, requestMethod string, requestPath string, responseCode int, body interface{}, now time.Time) error {
	responseBody, err := json.Marshal(body)
	if err != nil {
		return appErrors.InternalServer("failed to store idempotent response")
	}

	err = s.caseSessionRepo.CreateCaseSessionIdempotencyKey(tx, &entity.CaseSessionIdempotencyKey{
		CaseSessionIdempotencyKeyID: uuid.New(),
		CaseSessionID:               caseSessionID,
		UserID:                      userID,
		IdempotencyKey:              idempotencyKey,
		RequestMethod:               requestMethod,
		RequestPath:                 requestPath,
		ResponseCode:                responseCode,
		ResponseBody:                string(responseBody),
		CreatedAt:                   now,
		ExpiresAt:                   now.Add(gameplayIdempotencyTTL),
	})
	if err != nil {
		return appErrors.Conflict("idempotency key already used")
	}

	return nil
}

func normalizeInitialAssessment(value *string) (*string, error) {
	if value == nil {
		return nil, nil
	}

	normalized := strings.ToLower(strings.TrimSpace(*value))
	switch normalized {
	case model.InitialAssessmentTrusted, model.InitialAssessmentNeedCheck, model.InitialAssessmentMisleading:
		return &normalized, nil
	default:
		return nil, appErrors.BadRequest("invalid initial assessment")
	}
}

func parseGameplaySnapshot(raw string) (*model.GameplaySessionSnapshot, error) {
	var snapshot model.GameplaySessionSnapshot
	if err := json.Unmarshal([]byte(raw), &snapshot); err != nil {
		return nil, appErrors.InternalServer("failed to parse session snapshot")
	}
	return &snapshot, nil
}

func mapGameplaySessionResponse(session *entity.CaseSession) model.GameplaySessionResponse {
	return model.GameplaySessionResponse{
		CaseSessionID:     session.CaseSessionID,
		UserID:            session.UserID,
		CaseID:            session.CaseID,
		CaseVersionID:     session.CaseVersionID,
		SessionVersion:    session.SessionVersion,
		Status:            session.Status,
		InitialAssessment: session.InitialAssessment,
		InitialConfidence: session.InitialConfidence,
		StartedAt:         session.StartedAt,
		LastActivityAt:    session.LastActivityAt,
		SubmittedAt:       session.SubmittedAt,
	}
}

func findSnapshotEvidence(snapshot *model.GameplaySessionSnapshot, caseEvidenceID uuid.UUID) (model.GameplayEvidenceResponse, bool) {
	for _, evidence := range snapshot.Evidences {
		if evidence.CaseEvidenceID == caseEvidenceID {
			return evidence, true
		}
	}
	return model.GameplayEvidenceResponse{}, false
}

func buildSnapshotQuestionMap(snapshot *model.GameplaySessionSnapshot) map[uuid.UUID]model.GameplayQuestionResponse {
	questions := map[uuid.UUID]model.GameplayQuestionResponse{}
	for _, question := range snapshot.Questions {
		questions[question.CaseQuestionID] = question
	}
	return questions
}

func progressForEvidence(progress []model.GameplayEvidenceProgressResponse, caseEvidenceID uuid.UUID) model.GameplayEvidenceProgressResponse {
	for _, item := range progress {
		if item.CaseEvidenceID == caseEvidenceID {
			return item
		}
	}
	return model.GameplayEvidenceProgressResponse{CaseEvidenceID: caseEvidenceID}
}

func mapGameplayEvidenceResponse(evidence entity.CaseEvidence, index int) model.GameplayEvidenceResponse {
	result := model.GameplayEvidenceResponse{
		CaseEvidenceID: evidence.CaseEvidenceID,
		CaseVersionID:  evidence.CaseVersionID,
		Code:           formatEvidenceCode(index),
		TemplateType:   evidence.TemplateType,
		Label:          evidence.Label,
		SortOrder:      evidence.SortOrder,
	}

	if evidence.SocialPost != nil {
		result.SocialPost = &model.GameplaySocialPostEvidenceResponse{
			AuthorName:        evidence.SocialPost.AuthorName,
			AuthorHandle:      evidence.SocialPost.AuthorHandle,
			Platform:          evidence.SocialPost.Platform,
			PostText:          evidence.SocialPost.PostText,
			Timestamp:         evidence.SocialPost.Timestamp,
			LikesCount:        evidence.SocialPost.LikesCount,
			SharesCount:       evidence.SocialPost.SharesCount,
			CommentsCount:     evidence.SocialPost.CommentsCount,
			IsVerifiedAccount: evidence.SocialPost.IsVerifiedAccount,
			ImageURL:          evidence.SocialPost.ImageURL,
		}
	}
	if evidence.Article != nil {
		result.Article = &model.GameplayArticleEvidenceResponse{
			Headline:    evidence.Article.Headline,
			SourceName:  evidence.Article.SourceName,
			AuthorName:  evidence.Article.AuthorName,
			PublishDate: evidence.Article.PublishDate,
			URL:         evidence.Article.URL,
			BodyText:    evidence.Article.BodyText,
			ImageURL:    evidence.Article.ImageURL,
		}
	}
	if evidence.Blog != nil {
		result.Blog = &model.GameplayBlogEvidenceResponse{
			Title:       evidence.Blog.Title,
			AuthorName:  evidence.Blog.AuthorName,
			BlogName:    evidence.Blog.BlogName,
			PublishDate: evidence.Blog.PublishDate,
			BodyText:    evidence.Blog.BodyText,
		}
	}
	if evidence.ForumThread != nil {
		posts := make([]model.GameplayForumThreadPostResponse, 0, len(evidence.ForumThread.Posts))
		for _, post := range evidence.ForumThread.Posts {
			posts = append(posts, model.GameplayForumThreadPostResponse{
				AuthorName:  post.AuthorName,
				Text:        post.Text,
				Timestamp:   post.Timestamp,
				UpvoteCount: post.UpvoteCount,
				SortOrder:   post.SortOrder,
			})
		}
		result.ForumThread = &model.GameplayForumThreadEvidenceResponse{
			ThreadTitle: evidence.ForumThread.ThreadTitle,
			ForumName:   evidence.ForumThread.ForumName,
			Posts:       posts,
		}
	}
	if evidence.ChatTranscript != nil {
		participants := make([]model.GameplayChatTranscriptParticipantResponse, 0, len(evidence.ChatTranscript.Participants))
		for _, participant := range evidence.ChatTranscript.Participants {
			participants = append(participants, model.GameplayChatTranscriptParticipantResponse{
				Name:      participant.Name,
				SortOrder: participant.SortOrder,
			})
		}
		messages := make([]model.GameplayChatTranscriptMessageResponse, 0, len(evidence.ChatTranscript.Messages))
		for _, message := range evidence.ChatTranscript.Messages {
			messages = append(messages, model.GameplayChatTranscriptMessageResponse{
				Sender:    message.Sender,
				Text:      message.Text,
				Timestamp: message.Timestamp,
				SortOrder: message.SortOrder,
			})
		}
		result.ChatTranscript = &model.GameplayChatTranscriptEvidenceResponse{
			Participants: participants,
			Messages:     messages,
		}
	}
	if evidence.PublicAnnouncement != nil {
		result.PublicAnnouncement = &model.GameplayPublicAnnouncementResponse{
			IssuingBody: evidence.PublicAnnouncement.IssuingBody,
			Title:       evidence.PublicAnnouncement.Title,
			BodyText:    evidence.PublicAnnouncement.BodyText,
			Date:        evidence.PublicAnnouncement.Date,
		}
	}

	return result
}

func mapGameplayQuestionResponse(question entity.CaseQuestion, index int) (model.GameplayQuestionResponse, error) {
	relatedEvidenceIDs := make([]uuid.UUID, 0, len(question.EvidenceReferences))
	for _, reference := range question.EvidenceReferences {
		relatedEvidenceIDs = append(relatedEvidenceIDs, reference.CaseEvidenceID)
	}

	result := model.GameplayQuestionResponse{
		CaseQuestionID:     question.CaseQuestionID,
		CaseVersionID:      question.CaseVersionID,
		Code:               formatQuestionCode(index),
		QuestionType:       question.QuestionType,
		QuestionText:       question.QuestionText,
		IsRequired:         question.IsRequired,
		SortOrder:          question.SortOrder,
		RelatedEvidenceIDs: relatedEvidenceIDs,
	}

	switch question.QuestionType {
	case model.CaseQuestionTypeMCQ:
		result.Options = make([]model.GameplayMCQOptionResponse, 0, len(question.MCQOptions))
		for _, option := range question.MCQOptions {
			result.Options = append(result.Options, model.GameplayMCQOptionResponse{
				OptionCode: option.OptionCode,
				OptionText: option.OptionText,
				SortOrder:  option.SortOrder,
			})
		}
	case model.CaseQuestionTypeOpenEnded:
	case model.CaseQuestionTypeConfidenceSlider:
		if question.ConfidenceSliderDetail == nil {
			return result, appErrors.InternalServer("confidence slider question detail not found")
		}
		result.ConfidenceSlider = &model.GameplayConfidenceSliderResponse{
			MinValue:                 question.ConfidenceSliderDetail.MinValue,
			MaxValue:                 question.ConfidenceSliderDetail.MaxValue,
			SnapInterval:             question.ConfidenceSliderDetail.SnapInterval,
			DefaultValue:             question.ConfidenceSliderDetail.DefaultValue,
			LabelLow:                 question.ConfidenceSliderDetail.LabelLow,
			LabelHigh:                question.ConfidenceSliderDetail.LabelHigh,
			ShowWarningOnLargeChange: question.ConfidenceSliderDetail.ShowWarningOnLargeChange,
		}
	case model.CaseQuestionTypeClaimClassification:
		if question.ClaimClassificationDetail == nil {
			return result, appErrors.InternalServer("claim classification question detail not found")
		}
		tags, err := parseQuestionStringItems(question.ClaimClassificationDetail.TaxonomyTags, "taxonomy tags")
		if err != nil {
			return result, err
		}
		result.ClaimClassification = tags
	default:
		return result, appErrors.BadRequest("unsupported question type")
	}

	return result, nil
}

func answersByQuestionID(answers []model.GameplayAnswerResponse) map[uuid.UUID]model.GameplayAnswerResponse {
	result := map[uuid.UUID]model.GameplayAnswerResponse{}
	for _, answer := range answers {
		result[answer.CaseQuestionID] = answer
	}
	return result
}

func evidenceProgressByEvidenceID(progress []model.GameplayEvidenceProgressResponse) map[uuid.UUID]model.GameplayEvidenceProgressResponse {
	result := map[uuid.UUID]model.GameplayEvidenceProgressResponse{}
	for _, item := range progress {
		result[item.CaseEvidenceID] = item
	}
	return result
}

func scoreGameplaySession(
	config *entity.CaseScoringOutcomeConfig,
	snapshot *model.GameplaySessionSnapshot,
	answers map[uuid.UUID]model.GameplayAnswerResponse,
	progress map[uuid.UUID]model.GameplayEvidenceProgressResponse,
	session *entity.CaseSession,
	finalDecision string,
	finalConfidence int,
	reason string,
) ([]model.GameplayScoreBreakdownResponse, int, error) {
	breakdown := make([]model.GameplayScoreBreakdownResponse, 0, len(config.ScoringRules))
	totalWeighted := 0

	for _, rule := range config.ScoringRules {
		settings, err := parseScoringOutcomeSettings(rule.Settings)
		if err != nil {
			return nil, 0, err
		}

		score := scoreCategory(rule.CategoryKey, settings, snapshot, answers, progress, session, finalDecision, finalConfidence, reason)
		weighted := int(math.Round(float64(score*rule.WeightBasisPoints) / 10000))
		totalWeighted += weighted
		breakdown = append(breakdown, model.GameplayScoreBreakdownResponse{
			CategoryKey:       rule.CategoryKey,
			CategoryLabel:     rule.CategoryLabel,
			Score:             score,
			WeightBasisPoints: rule.WeightBasisPoints,
			WeightedScore:     weighted,
		})
	}

	return breakdown, clampInt(totalWeighted, 0, 100), nil
}

func scoreCategory(
	categoryKey string,
	settings map[string]interface{},
	snapshot *model.GameplaySessionSnapshot,
	answers map[uuid.UUID]model.GameplayAnswerResponse,
	progress map[uuid.UUID]model.GameplayEvidenceProgressResponse,
	session *entity.CaseSession,
	finalDecision string,
	finalConfidence int,
	reason string,
) int {
	switch categoryKey {
	case model.ScoringCategoryEvidenceEvaluation:
		return scoreEvidenceEvaluation(settings, snapshot, progress)
	case model.ScoringCategoryClaimAnalysis:
		return scoreClaimAnalysis(settings, answers)
	case model.ScoringCategoryConfidenceCalibration:
		return scoreConfidenceCalibration(settings, answers, session, finalConfidence)
	case model.ScoringCategoryReasoning:
		return scoreReasoning(settings, answers, reason)
	case model.ScoringCategorySafetyJudgment:
		return scoreSafetyJudgment(settings, answers, finalDecision)
	default:
		return 0
	}
}

func scoreEvidenceEvaluation(settings map[string]interface{}, snapshot *model.GameplaySessionSnapshot, progress map[uuid.UUID]model.GameplayEvidenceProgressResponse) int {
	ids := stringSliceSetting(settings, "critical_evidence_ids")
	if len(ids) == 0 {
		for _, evidence := range snapshot.Evidences {
			ids = append(ids, evidence.CaseEvidenceID.String())
		}
	}
	if len(ids) == 0 {
		return 100
	}

	opened := 0
	for _, rawID := range ids {
		evidenceID, err := uuid.Parse(rawID)
		if err != nil {
			continue
		}
		if progress[evidenceID].Opened {
			opened++
		}
	}

	return int(math.Round(float64(opened) / float64(len(ids)) * 100))
}

func scoreClaimAnalysis(settings map[string]interface{}, answers map[uuid.UUID]model.GameplayAnswerResponse) int {
	mapping, _ := settings["mapping"].([]interface{})
	if len(mapping) == 0 {
		return scoreAnsweredRatio(answers)
	}

	correct := 0
	total := 0
	for _, item := range mapping {
		row, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		questionID, err := uuid.Parse(strings.TrimSpace(fmt.Sprint(row["question_id"])))
		if err != nil {
			continue
		}
		expected := strings.ToLower(settingString(row, "expected_classification"))
		if expected == "" {
			expected = strings.ToLower(settingString(row, "expected_answer"))
		}
		answer, ok := answers[questionID]
		if !ok {
			total++
			continue
		}
		actual := strings.ToLower(extractAnswerString(answer.Value, "classification", "answer", "value", "option_code"))
		if actual == expected {
			correct++
		}
		total++
	}
	if total == 0 {
		return scoreAnsweredRatio(answers)
	}
	return int(math.Round(float64(correct) / float64(total) * 100))
}

func scoreConfidenceCalibration(settings map[string]interface{}, answers map[uuid.UUID]model.GameplayAnswerResponse, session *entity.CaseSession, finalConfidence int) int {
	initial := 50
	if session.InitialConfidence != nil {
		initial = *session.InitialConfidence
	}
	if rawID := settingString(settings, "initial_question_id"); rawID != "" {
		if questionID, err := uuid.Parse(rawID); err == nil {
			if answer, ok := answers[questionID]; ok && answer.ConfidenceFinal != nil {
				initial = *answer.ConfidenceFinal
			}
		}
	}

	target := intSetting(settings, "target_final_confidence", 80)
	if finalID := settingString(settings, "final_question_id"); finalID != "" {
		if questionID, err := uuid.Parse(finalID); err == nil {
			if answer, ok := answers[questionID]; ok && answer.ConfidenceFinal != nil {
				finalConfidence = *answer.ConfidenceFinal
			}
		}
	}

	improvement := finalConfidence - initial
	score := 100 - int(math.Abs(float64(target-finalConfidence)))
	if improvement > 0 {
		score += min(improvement, 15)
	}
	return clampInt(score, 0, 100)
}

func scoreReasoning(settings map[string]interface{}, answers map[uuid.UUID]model.GameplayAnswerResponse, reason string) int {
	if rawID := settingString(settings, "open_question_id"); rawID != "" {
		if questionID, err := uuid.Parse(rawID); err == nil {
			if answer, ok := answers[questionID]; ok {
				reason = extractAnswerString(answer.Value, "text", "reason", "answer", "value")
			}
		}
	}

	length := len(strings.Fields(reason))
	switch {
	case length >= 35:
		return 100
	case length >= 20:
		return 80
	case length >= 10:
		return 60
	case length >= 5:
		return 40
	default:
		return 20
	}
}

func scoreSafetyJudgment(settings map[string]interface{}, answers map[uuid.UUID]model.GameplayAnswerResponse, finalDecision string) int {
	actual := strings.ToLower(strings.TrimSpace(finalDecision))
	if rawID := settingString(settings, "final_decision_question_id"); rawID != "" {
		if questionID, err := uuid.Parse(rawID); err == nil {
			if answer, ok := answers[questionID]; ok {
				actual = strings.ToLower(extractAnswerString(answer.Value, "decision", "verdict", "answer", "value", "option_code"))
			}
		}
	}

	expected := strings.ToLower(settingString(settings, "correct_decision"))
	if expected == "" {
		expected = strings.ToLower(settingString(settings, "expected_decision"))
	}
	if expected == "" {
		return 70
	}
	if actual == expected {
		return 100
	}
	if strings.Contains(actual, expected) || strings.Contains(expected, actual) {
		return 75
	}
	return 25
}

func scoreAnsweredRatio(answers map[uuid.UUID]model.GameplayAnswerResponse) int {
	if len(answers) == 0 {
		return 0
	}
	finalCount := 0
	for _, answer := range answers {
		if answer.IsFinal {
			finalCount++
		}
	}
	return int(math.Round(float64(finalCount) / float64(len(answers)) * 100))
}

func selectOutcomeRule(rules []entity.CaseOutcomeRule, totalScore int) (*entity.CaseOutcomeRule, error) {
	sort.SliceStable(rules, func(i, j int) bool {
		return rules[i].ScoreMin > rules[j].ScoreMin
	})
	for _, rule := range rules {
		if totalScore >= rule.ScoreMin && totalScore <= rule.ScoreMax {
			return &rule, nil
		}
	}
	return nil, appErrors.Conflict("outcome rule not found for score")
}

func (s *GameplayService) applyCityImpact(tx *gorm.DB, settings []entity.CaseOutcomeCityImpactSetting) ([]model.GameplayCityImpactResponse, error) {
	stats, err := s.cityStatsRepo.GetCityStatisticsForUpdate(tx, model.GetCityStatisticsParam{StatKey: model.CityStatisticsDefaultKey})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			stats = defaultCityStatistics()
			stats.CityStatisticsID = uuid.New()
			if err := s.cityStatsRepo.CreateCityStatistics(tx, stats); err != nil {
				return nil, appErrors.InternalServer("failed to create city statistics")
			}
		} else {
			return nil, appErrors.InternalServer("failed to get city statistics")
		}
	}

	impact := make([]model.GameplayCityImpactResponse, 0, len(settings))
	for _, setting := range settings {
		before := cityStatValue(stats, setting.ImpactKey)
		after := clampCityStat(before + setting.ImpactValue)
		setCityStatValue(stats, setting.ImpactKey, after)
		impact = append(impact, model.GameplayCityImpactResponse{
			Key:    setting.ImpactKey,
			Label:  cityImpactLabel(setting.ImpactKey),
			Delta:  setting.ImpactValue,
			Before: before,
			After:  after,
		})
	}
	stats.VisualState = cityVisualState(stats)
	if err := s.cityStatsRepo.UpdateCityStatistics(tx, stats); err != nil {
		return nil, appErrors.InternalServer("failed to update city statistics")
	}
	return impact, nil
}

func (s *GameplayService) resolveLevelAfter(tx *gorm.DB, currentLevel int, xpAfter int) (int, int, string, error) {
	levels, _, err := s.gameLevelRepo.ListGameLevels(tx, model.ListGameLevelsParam{Limit: 1000})
	if err != nil {
		return currentLevel, 0, "", appErrors.InternalServer("failed to list game levels")
	}

	levelAfter := currentLevel
	rewardCoin := 0
	title := ""
	for _, level := range levels {
		if xpAfter >= level.XPRequired && level.Level > levelAfter {
			levelAfter = level.Level
			rewardCoin += level.RewardCoin
			title = level.Title
		}
	}
	return levelAfter, rewardCoin, title, nil
}

func calculateXPGained(durationMinutes int, totalScore int, config *entity.GameConfig) int {
	baseXP := max(durationMinutes, 10) * 10
	multiplier := config.CompleteCaseBaseMultiplier
	if multiplier <= 0 {
		multiplier = 1
	}
	if totalScore >= 100 && config.PerfectScoreBonusMultiplier > 0 {
		multiplier *= config.PerfectScoreBonusMultiplier
	}
	return int(math.Round(float64(baseXP) * multiplier * float64(totalScore) / 100))
}

func calculateCoinGained(xpGained int, totalScore int) int {
	coin := int(math.Round(float64(xpGained) / 3))
	if totalScore >= 90 {
		coin += 10
	}
	return max(coin, 0)
}

func mergeProfileScore(current float64, latest int) float64 {
	if current <= 0 {
		return float64(latest)
	}
	return math.Round(((current*0.7)+float64(latest)*0.3)*100) / 100
}

func scoreByCategory(scores []model.GameplayScoreBreakdownResponse, category string) int {
	for _, score := range scores {
		if score.CategoryKey == category {
			return score.Score
		}
	}
	return 0
}

func buildGameplayFeedback(scores []model.GameplayScoreBreakdownResponse) model.GameplayFeedbackResponse {
	if len(scores) == 0 {
		return model.GameplayFeedbackResponse{Message: "Keputusanmu sudah tercatat."}
	}
	strength := scores[0]
	improvement := scores[0]
	for _, score := range scores {
		if score.Score > strength.Score {
			strength = score
		}
		if score.Score < improvement.Score {
			improvement = score
		}
	}
	return model.GameplayFeedbackResponse{
		StrengthCategory:    strength.CategoryKey,
		ImprovementCategory: improvement.CategoryKey,
		Message:             fmt.Sprintf("Kekuatanmu ada di %s. Latih lagi %s untuk investigasi berikutnya.", strength.CategoryLabel, improvement.CategoryLabel),
	}
}

func buildIncompleteSessionDetail(state *model.GameplayStateResponse) model.GameplayIncompleteDetailResponse {
	openedEvidence := map[uuid.UUID]bool{}
	for _, progress := range state.EvidenceProgress {
		if progress.Opened {
			openedEvidence[progress.CaseEvidenceID] = true
		}
	}

	finalAnswers := map[uuid.UUID]bool{}
	for _, answer := range state.Answers {
		if answer.IsFinal {
			finalAnswers[answer.CaseQuestionID] = true
		}
	}

	missingEvidence := []model.GameplayMissingEvidenceResponse{}
	for _, evidence := range state.Evidences {
		if !openedEvidence[evidence.CaseEvidenceID] {
			missingEvidence = append(missingEvidence, model.GameplayMissingEvidenceResponse{
				CaseEvidenceID: evidence.CaseEvidenceID,
				Code:           evidence.Code,
				Label:          evidence.Label,
				TemplateType:   evidence.TemplateType,
			})
		}
	}

	missingQuestions := []model.GameplayMissingQuestionResponse{}
	for _, question := range state.Questions {
		if question.IsRequired && !finalAnswers[question.CaseQuestionID] {
			missingQuestions = append(missingQuestions, model.GameplayMissingQuestionResponse{
				CaseQuestionID: question.CaseQuestionID,
				Code:           question.Code,
				QuestionType:   question.QuestionType,
				QuestionText:   question.QuestionText,
			})
		}
	}

	return model.GameplayIncompleteDetailResponse{
		OpenedEvidenceCount:   state.Progress.OpenedEvidenceCount,
		TotalEvidenceCount:    state.Progress.TotalEvidenceCount,
		AnsweredQuestionCount: state.Progress.AnsweredQuestionCount,
		RequiredQuestionCount: state.Progress.RequiredQuestionCount,
		CanTakeDecision:       state.Progress.CanTakeDecision,
		MissingEvidences:      missingEvidence,
		MissingQuestions:      missingQuestions,
	}
}

func stringSliceSetting(settings map[string]interface{}, key string) []string {
	raw, ok := settings[key]
	if !ok {
		return nil
	}
	items, ok := raw.([]interface{})
	if !ok {
		return nil
	}
	result := make([]string, 0, len(items))
	for _, item := range items {
		value := strings.TrimSpace(fmt.Sprint(item))
		if value != "" {
			result = append(result, value)
		}
	}
	return result
}

func settingString(settings map[string]interface{}, key string) string {
	raw, ok := settings[key]
	if !ok || raw == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(raw))
}

func intSetting(settings map[string]interface{}, key string, fallback int) int {
	raw, ok := settings[key]
	if !ok {
		return fallback
	}
	switch value := raw.(type) {
	case float64:
		return int(value)
	case int:
		return value
	default:
		return fallback
	}
}

func extractAnswerString(raw json.RawMessage, keys ...string) string {
	var object map[string]interface{}
	if err := json.Unmarshal(raw, &object); err == nil {
		for _, key := range keys {
			if value, ok := object[key]; ok {
				return strings.ToLower(strings.TrimSpace(fmt.Sprint(value)))
			}
		}
	}
	var value string
	if err := json.Unmarshal(raw, &value); err == nil {
		return strings.ToLower(strings.TrimSpace(value))
	}
	return ""
}

func cityStatValue(stats *entity.CityStatistics, key string) int {
	switch key {
	case model.CityImpactHealth:
		return stats.InformationHealth
	case model.CityImpactTrust:
		return stats.PublicTrust
	case model.CityImpactStability:
		return stats.SocialStability
	case model.CityImpactWellbeing:
		return stats.PublicWellbeing
	default:
		return 0
	}
}

func setCityStatValue(stats *entity.CityStatistics, key string, value int) {
	switch key {
	case model.CityImpactHealth:
		stats.InformationHealth = value
	case model.CityImpactTrust:
		stats.PublicTrust = value
	case model.CityImpactStability:
		stats.SocialStability = value
	case model.CityImpactWellbeing:
		stats.PublicWellbeing = value
	}
}

func cityImpactLabel(key string) string {
	switch key {
	case model.CityImpactHealth:
		return "Information Health"
	case model.CityImpactTrust:
		return "Public Trust"
	case model.CityImpactStability:
		return "Social Stability"
	case model.CityImpactWellbeing:
		return "Public Wellbeing"
	default:
		return key
	}
}

func cityVisualState(stats *entity.CityStatistics) string {
	lowest := min(min(stats.InformationHealth, stats.PublicTrust), min(stats.SocialStability, stats.PublicWellbeing))
	if lowest < 40 {
		return "critical"
	}
	if lowest < 70 {
		return "warning"
	}
	return "normal"
}

func clampInt(value int, minValue int, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func clampFloat64(value float64, minValue float64, maxValue float64) float64 {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}
