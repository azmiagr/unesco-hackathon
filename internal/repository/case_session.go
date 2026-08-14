package repository

import (
	"time"

	"github.com/azmiagr/unesco-hackathon/entity"
	"github.com/azmiagr/unesco-hackathon/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ICaseSessionRepository interface {
	CreateCaseSession(tx *gorm.DB, session *entity.CaseSession) error
	GetCaseSession(tx *gorm.DB, param model.GetCaseSessionParam) (*entity.CaseSession, error)
	GetCaseSessionForUpdate(tx *gorm.DB, caseSessionID uuid.UUID) (*entity.CaseSession, error)
	GetActiveCaseSession(tx *gorm.DB, userID uuid.UUID, caseID uuid.UUID) (*entity.CaseSession, error)
	ListRecentCaseSessions(tx *gorm.DB, param model.ListRecentCaseSessionsParam) ([]entity.CaseSession, error)
	CountCaseSessionsStarted(tx *gorm.DB, param model.CountCaseSessionsStartedParam) (int64, error)
	TouchCaseSessionActivity(tx *gorm.DB, caseSessionID uuid.UUID, lastActivityAt time.Time) error
	IncrementCaseSessionVersion(tx *gorm.DB, caseSessionID uuid.UUID, expectedVersion int, lastActivityAt time.Time) (int64, error)
	UpdateCaseSessionStatus(tx *gorm.DB, caseSessionID uuid.UUID, status string, submittedAt *time.Time, clearActiveSessionKey bool) error
	CreateCaseSessionAnswer(tx *gorm.DB, answer *entity.CaseSessionAnswer) error
	GetCaseSessionAnswer(tx *gorm.DB, param model.GetCaseSessionAnswerParam) (*entity.CaseSessionAnswer, error)
	ListCaseSessionAnswers(tx *gorm.DB, param model.ListCaseSessionAnswersParam) ([]entity.CaseSessionAnswer, error)
	UpsertCaseSessionAnswer(tx *gorm.DB, answer *entity.CaseSessionAnswer) error
	CreateCaseSessionEvidenceProgress(tx *gorm.DB, progress *entity.CaseSessionEvidenceProgress) error
	GetCaseSessionEvidenceProgress(tx *gorm.DB, param model.GetCaseSessionEvidenceProgressParam) (*entity.CaseSessionEvidenceProgress, error)
	ListCaseSessionEvidenceProgress(tx *gorm.DB, param model.ListCaseSessionEvidenceProgressParam) ([]entity.CaseSessionEvidenceProgress, error)
	UpsertCaseSessionEvidenceProgress(tx *gorm.DB, progress *entity.CaseSessionEvidenceProgress) error
	CreateCaseSessionEvent(tx *gorm.DB, event *entity.CaseSessionEvent) error
	ListCaseSessionEvents(tx *gorm.DB, param model.ListCaseSessionEventsParam) ([]entity.CaseSessionEvent, error)
	CreateCaseSessionIdempotencyKey(tx *gorm.DB, idempotencyKey *entity.CaseSessionIdempotencyKey) error
	GetCaseSessionIdempotencyKey(tx *gorm.DB, param model.GetCaseSessionIdempotencyKeyParam) (*entity.CaseSessionIdempotencyKey, error)
	GetCaseSessionIdempotencyKeyForUpdate(tx *gorm.DB, param model.GetCaseSessionIdempotencyKeyParam) (*entity.CaseSessionIdempotencyKey, error)
	DeleteExpiredCaseSessionIdempotencyKeys(tx *gorm.DB, now time.Time) error
	CreateCaseSessionResult(tx *gorm.DB, result *entity.CaseSessionResult) error
	GetCaseSessionResult(tx *gorm.DB, caseSessionID uuid.UUID) (*entity.CaseSessionResult, error)
	GetLatestUserCaseSessionResult(tx *gorm.DB, userID uuid.UUID) (*entity.CaseSessionResult, error)
	ListUserCaseResultHistory(tx *gorm.DB, param model.ListUserCaseResultHistoryParam) ([]model.UserCaseResultHistoryRow, error)
	GetUserCaseResultSummary(tx *gorm.DB, userID uuid.UUID) (*model.UserCaseResultSummaryRow, error)
	ListUserCaseCompletionDates(tx *gorm.DB, userID uuid.UUID) ([]time.Time, error)
}

type CaseSessionRepository struct {
	db *gorm.DB
}

func NewCaseSessionRepository(db *gorm.DB) ICaseSessionRepository {
	return &CaseSessionRepository{db: db}
}

func (r *CaseSessionRepository) CreateCaseSession(tx *gorm.DB, session *entity.CaseSession) error {
	err := tx.Create(session).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *CaseSessionRepository) GetCaseSession(tx *gorm.DB, param model.GetCaseSessionParam) (*entity.CaseSession, error) {
	var session entity.CaseSession
	query := tx

	if param.IncludeDeletedRecords {
		query = query.Unscoped()
	}
	query = applyCaseSessionFilters(query, param)

	err := query.First(&session).Error
	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *CaseSessionRepository) GetCaseSessionForUpdate(tx *gorm.DB, caseSessionID uuid.UUID) (*entity.CaseSession, error) {
	var session entity.CaseSession

	err := tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("case_session_id = ?", caseSessionID).
		First(&session).Error
	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *CaseSessionRepository) GetActiveCaseSession(tx *gorm.DB, userID uuid.UUID, caseID uuid.UUID) (*entity.CaseSession, error) {
	var session entity.CaseSession

	err := tx.
		Where("user_id = ? AND case_id = ? AND status = ?", userID, caseID, model.CaseSessionStatusActive).
		Order("started_at DESC").
		First(&session).Error
	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *CaseSessionRepository) ListRecentCaseSessions(tx *gorm.DB, param model.ListRecentCaseSessionsParam) ([]entity.CaseSession, error) {
	var sessions []entity.CaseSession
	query := tx.Model(&entity.CaseSession{})

	if param.UserID != uuid.Nil {
		query = query.Where("user_id = ?", param.UserID)
	}
	if param.CaseID != uuid.Nil {
		query = query.Where("case_id = ?", param.CaseID)
	}
	if param.Status != "" {
		query = query.Where("status = ?", param.Status)
	}

	limit := param.Limit
	if limit <= 0 {
		limit = 10
	}

	err := query.
		Order("started_at DESC").
		Limit(limit).
		Find(&sessions).Error
	if err != nil {
		return nil, err
	}

	return sessions, nil
}

func (r *CaseSessionRepository) CountCaseSessionsStarted(tx *gorm.DB, param model.CountCaseSessionsStartedParam) (int64, error) {
	var count int64
	query := tx.Model(&entity.CaseSession{}).
		Where("user_id = ?", param.UserID)

	if !param.StartAt.IsZero() {
		query = query.Where("started_at >= ?", param.StartAt)
	}
	if !param.EndAt.IsZero() {
		query = query.Where("started_at < ?", param.EndAt)
	}
	if len(param.Statuses) > 0 {
		query = query.Where("status IN ?", param.Statuses)
	}
	if param.ExcludeID != uuid.Nil {
		query = query.Where("case_session_id <> ?", param.ExcludeID)
	}

	err := query.Count(&count).Error
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *CaseSessionRepository) TouchCaseSessionActivity(tx *gorm.DB, caseSessionID uuid.UUID, lastActivityAt time.Time) error {
	err := tx.Model(&entity.CaseSession{}).
		Where("case_session_id = ?", caseSessionID).
		Update("last_activity_at", lastActivityAt).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *CaseSessionRepository) IncrementCaseSessionVersion(tx *gorm.DB, caseSessionID uuid.UUID, expectedVersion int, lastActivityAt time.Time) (int64, error) {
	result := tx.Model(&entity.CaseSession{}).
		Where("case_session_id = ? AND session_version = ?", caseSessionID, expectedVersion).
		Updates(map[string]interface{}{
			"session_version":  gorm.Expr("session_version + 1"),
			"last_activity_at": lastActivityAt,
		})
	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (r *CaseSessionRepository) UpdateCaseSessionStatus(tx *gorm.DB, caseSessionID uuid.UUID, status string, submittedAt *time.Time, clearActiveSessionKey bool) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if submittedAt != nil {
		updates["submitted_at"] = *submittedAt
		updates["last_activity_at"] = *submittedAt
	}
	if clearActiveSessionKey {
		updates["active_session_key"] = nil
	}

	err := tx.Model(&entity.CaseSession{}).
		Where("case_session_id = ?", caseSessionID).
		Updates(updates).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *CaseSessionRepository) CreateCaseSessionAnswer(tx *gorm.DB, answer *entity.CaseSessionAnswer) error {
	err := tx.Create(answer).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *CaseSessionRepository) GetCaseSessionAnswer(tx *gorm.DB, param model.GetCaseSessionAnswerParam) (*entity.CaseSessionAnswer, error) {
	var answer entity.CaseSessionAnswer
	query := tx.Model(&entity.CaseSessionAnswer{})

	if param.CaseSessionID != uuid.Nil {
		query = query.Where("case_session_id = ?", param.CaseSessionID)
	}
	if param.CaseQuestionID != uuid.Nil {
		query = query.Where("case_question_id = ?", param.CaseQuestionID)
	}

	err := query.First(&answer).Error
	if err != nil {
		return nil, err
	}

	return &answer, nil
}

func (r *CaseSessionRepository) ListCaseSessionAnswers(tx *gorm.DB, param model.ListCaseSessionAnswersParam) ([]entity.CaseSessionAnswer, error) {
	var answers []entity.CaseSessionAnswer
	query := tx.Model(&entity.CaseSessionAnswer{}).
		Where("case_session_id = ?", param.CaseSessionID)

	if param.IsFinal != nil {
		query = query.Where("is_final = ?", *param.IsFinal)
	}

	err := query.
		Order("saved_at ASC").
		Order("created_at ASC").
		Find(&answers).Error
	if err != nil {
		return nil, err
	}

	return answers, nil
}

func (r *CaseSessionRepository) UpsertCaseSessionAnswer(tx *gorm.DB, answer *entity.CaseSessionAnswer) error {
	err := tx.
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "case_session_id"},
				{Name: "case_question_id"},
			},
			DoUpdates: clause.AssignmentColumns([]string{
				"question_type",
				"value",
				"confidence_initial",
				"confidence_final",
				"is_final",
				"saved_at",
				"updated_at",
			}),
		}).
		Create(answer).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *CaseSessionRepository) CreateCaseSessionEvidenceProgress(tx *gorm.DB, progress *entity.CaseSessionEvidenceProgress) error {
	err := tx.Create(progress).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *CaseSessionRepository) GetCaseSessionEvidenceProgress(tx *gorm.DB, param model.GetCaseSessionEvidenceProgressParam) (*entity.CaseSessionEvidenceProgress, error) {
	var progress entity.CaseSessionEvidenceProgress
	query := tx.Model(&entity.CaseSessionEvidenceProgress{})

	if param.CaseSessionID != uuid.Nil {
		query = query.Where("case_session_id = ?", param.CaseSessionID)
	}
	if param.CaseEvidenceID != uuid.Nil {
		query = query.Where("case_evidence_id = ?", param.CaseEvidenceID)
	}

	err := query.First(&progress).Error
	if err != nil {
		return nil, err
	}

	return &progress, nil
}

func (r *CaseSessionRepository) ListCaseSessionEvidenceProgress(tx *gorm.DB, param model.ListCaseSessionEvidenceProgressParam) ([]entity.CaseSessionEvidenceProgress, error) {
	var progress []entity.CaseSessionEvidenceProgress

	err := tx.Model(&entity.CaseSessionEvidenceProgress{}).
		Where("case_session_id = ?", param.CaseSessionID).
		Order("first_opened_at ASC").
		Find(&progress).Error
	if err != nil {
		return nil, err
	}

	return progress, nil
}

func (r *CaseSessionRepository) UpsertCaseSessionEvidenceProgress(tx *gorm.DB, progress *entity.CaseSessionEvidenceProgress) error {
	incrementBy := progress.OpenedCount
	if incrementBy <= 0 {
		incrementBy = 1
	}

	err := tx.
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "case_session_id"},
				{Name: "case_evidence_id"},
			},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"opened_count":   gorm.Expr("opened_count + ?", incrementBy),
				"last_opened_at": progress.LastOpenedAt,
				"updated_at":     gorm.Expr("CURRENT_TIMESTAMP"),
			}),
		}).
		Create(progress).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *CaseSessionRepository) CreateCaseSessionEvent(tx *gorm.DB, event *entity.CaseSessionEvent) error {
	err := tx.Create(event).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *CaseSessionRepository) ListCaseSessionEvents(tx *gorm.DB, param model.ListCaseSessionEventsParam) ([]entity.CaseSessionEvent, error) {
	var events []entity.CaseSessionEvent
	query := tx.Model(&entity.CaseSessionEvent{}).
		Where("case_session_id = ?", param.CaseSessionID)

	if param.EventType != "" {
		query = query.Where("event_type = ?", param.EventType)
	}
	if param.CaseEvidenceID != uuid.Nil {
		query = query.Where("case_evidence_id = ?", param.CaseEvidenceID)
	}
	if param.CaseQuestionID != uuid.Nil {
		query = query.Where("case_question_id = ?", param.CaseQuestionID)
	}

	if param.Offset > 0 {
		query = query.Offset(param.Offset)
	}
	if param.Limit > 0 {
		query = query.Limit(param.Limit)
	}

	err := query.
		Order("created_at ASC").
		Find(&events).Error
	if err != nil {
		return nil, err
	}

	return events, nil
}

func (r *CaseSessionRepository) CreateCaseSessionIdempotencyKey(tx *gorm.DB, idempotencyKey *entity.CaseSessionIdempotencyKey) error {
	err := tx.Create(idempotencyKey).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *CaseSessionRepository) GetCaseSessionIdempotencyKey(tx *gorm.DB, param model.GetCaseSessionIdempotencyKeyParam) (*entity.CaseSessionIdempotencyKey, error) {
	var idempotencyKey entity.CaseSessionIdempotencyKey
	query := applyCaseSessionIdempotencyKeyFilters(tx.Model(&entity.CaseSessionIdempotencyKey{}), param)

	err := query.First(&idempotencyKey).Error
	if err != nil {
		return nil, err
	}

	return &idempotencyKey, nil
}

func (r *CaseSessionRepository) GetCaseSessionIdempotencyKeyForUpdate(tx *gorm.DB, param model.GetCaseSessionIdempotencyKeyParam) (*entity.CaseSessionIdempotencyKey, error) {
	var idempotencyKey entity.CaseSessionIdempotencyKey
	query := applyCaseSessionIdempotencyKeyFilters(
		tx.Model(&entity.CaseSessionIdempotencyKey{}).Clauses(clause.Locking{Strength: "UPDATE"}),
		param,
	)

	err := query.First(&idempotencyKey).Error
	if err != nil {
		return nil, err
	}

	return &idempotencyKey, nil
}

func (r *CaseSessionRepository) DeleteExpiredCaseSessionIdempotencyKeys(tx *gorm.DB, now time.Time) error {
	err := tx.
		Where("expires_at <= ?", now).
		Delete(&entity.CaseSessionIdempotencyKey{}).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *CaseSessionRepository) CreateCaseSessionResult(tx *gorm.DB, result *entity.CaseSessionResult) error {
	err := tx.Create(result).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *CaseSessionRepository) GetCaseSessionResult(tx *gorm.DB, caseSessionID uuid.UUID) (*entity.CaseSessionResult, error) {
	var result entity.CaseSessionResult
	err := tx.
		Where("case_session_id = ?", caseSessionID).
		First(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *CaseSessionRepository) GetLatestUserCaseSessionResult(tx *gorm.DB, userID uuid.UUID) (*entity.CaseSessionResult, error) {
	var result entity.CaseSessionResult
	err := tx.Model(&entity.CaseSessionResult{}).
		Joins("JOIN case_sessions ON case_sessions.case_session_id = case_session_results.case_session_id").
		Where("case_session_results.user_id = ?", userID).
		Where("case_sessions.status = ?", model.CaseSessionStatusCompleted).
		Where("case_sessions.submitted_at IS NOT NULL").
		Order("case_sessions.submitted_at DESC").
		First(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *CaseSessionRepository) ListUserCaseResultHistory(tx *gorm.DB, param model.ListUserCaseResultHistoryParam) ([]model.UserCaseResultHistoryRow, error) {
	var rows []model.UserCaseResultHistoryRow
	query := tx.Table("case_session_results").
		Joins("JOIN case_sessions ON case_sessions.case_session_id = case_session_results.case_session_id").
		Joins("JOIN cases ON cases.case_id = case_session_results.case_id").
		Where("case_session_results.user_id = ?", param.UserID).
		Where("case_sessions.status = ?", model.CaseSessionStatusCompleted).
		Where("case_sessions.submitted_at IS NOT NULL").
		Select(`
			case_session_results.case_id,
			case_session_results.case_session_id,
			cases.title,
			cases.difficulty_level,
			cases.status AS case_status,
			case_session_results.total_score,
			case_session_results.outcome_key,
			case_session_results.outcome_label,
			case_session_results.xp_gained,
			case_sessions.submitted_at AS completed_at
		`)

	if param.Offset > 0 {
		query = query.Offset(param.Offset)
	}
	if param.Limit > 0 {
		query = query.Limit(param.Limit)
	}

	err := query.
		Order("case_sessions.submitted_at DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (r *CaseSessionRepository) GetUserCaseResultSummary(tx *gorm.DB, userID uuid.UUID) (*model.UserCaseResultSummaryRow, error) {
	var row model.UserCaseResultSummaryRow

	err := tx.Table("case_session_results").
		Joins("JOIN case_sessions ON case_sessions.case_session_id = case_session_results.case_session_id").
		Where("case_session_results.user_id = ?", userID).
		Where("case_sessions.status = ?", model.CaseSessionStatusCompleted).
		Where("case_sessions.submitted_at IS NOT NULL").
		Select(`
			COUNT(*) AS cases_completed,
			COALESCE(AVG(case_session_results.total_score), 0) AS accuracy_score
		`).
		Scan(&row).Error
	if err != nil {
		return nil, err
	}

	return &row, nil
}

func (r *CaseSessionRepository) ListUserCaseCompletionDates(tx *gorm.DB, userID uuid.UUID) ([]time.Time, error) {
	var rows []struct {
		CompletedDate string `gorm:"column:completed_date"`
	}

	err := tx.Table("case_session_results").
		Joins("JOIN case_sessions ON case_sessions.case_session_id = case_session_results.case_session_id").
		Where("case_session_results.user_id = ?", userID).
		Where("case_sessions.status = ?", model.CaseSessionStatusCompleted).
		Where("case_sessions.submitted_at IS NOT NULL").
		Select("DATE_FORMAT(case_sessions.submitted_at, '%Y-%m-%d') AS completed_date").
		Group("DATE(case_sessions.submitted_at)").
		Order("completed_date DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	dates := make([]time.Time, 0, len(rows))
	for _, row := range rows {
		completedDate, err := time.Parse("2006-01-02", row.CompletedDate)
		if err != nil {
			return nil, err
		}
		dates = append(dates, completedDate)
	}

	return dates, nil
}

func applyCaseSessionFilters(query *gorm.DB, param model.GetCaseSessionParam) *gorm.DB {
	if param.CaseSessionID != uuid.Nil {
		query = query.Where("case_session_id = ?", param.CaseSessionID)
	}
	if param.UserID != uuid.Nil {
		query = query.Where("user_id = ?", param.UserID)
	}
	if param.CaseID != uuid.Nil {
		query = query.Where("case_id = ?", param.CaseID)
	}
	if param.CaseVersionID != uuid.Nil {
		query = query.Where("case_version_id = ?", param.CaseVersionID)
	}
	if param.Status != "" {
		query = query.Where("status = ?", param.Status)
	}
	if param.ActiveSessionKey != "" {
		query = query.Where("active_session_key = ?", param.ActiveSessionKey)
	}
	if param.StartIdempotencyKey != "" {
		query = query.Where("start_idempotency_key = ?", param.StartIdempotencyKey)
	}

	return query
}

func applyCaseSessionIdempotencyKeyFilters(query *gorm.DB, param model.GetCaseSessionIdempotencyKeyParam) *gorm.DB {
	query = query.Where("idempotency_key = ?", param.IdempotencyKey)

	if param.UserID != uuid.Nil {
		query = query.Where("user_id = ?", param.UserID)
	}
	if param.RequestMethod != "" {
		query = query.Where("request_method = ?", param.RequestMethod)
	}
	if param.RequestPath != "" {
		query = query.Where("request_path = ?", param.RequestPath)
	}
	if !param.Now.IsZero() {
		query = query.Where("expires_at > ?", param.Now)
	}

	return query
}
