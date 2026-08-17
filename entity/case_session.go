package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CaseSession struct {
	CaseSessionID       uuid.UUID      `json:"case_session_id" gorm:"type:varchar(36);primaryKey"`
	UserID              uuid.UUID      `json:"user_id" gorm:"type:varchar(36);not null;index;index:idx_case_sessions_user_started;index:idx_case_sessions_user_status_started,priority:1"`
	CaseID              uuid.UUID      `json:"case_id" gorm:"type:varchar(36);not null;index"`
	CaseVersionID       uuid.UUID      `json:"case_version_id" gorm:"type:varchar(36);not null;index"`
	ActiveSessionKey    *string        `json:"active_session_key" gorm:"type:varchar(120);uniqueIndex:uq_case_sessions_active_key"`
	StartIdempotencyKey *string        `json:"start_idempotency_key" gorm:"type:varchar(128);uniqueIndex:uq_case_sessions_start_idempotency_key"`
	SessionSnapshot     string         `json:"session_snapshot" gorm:"type:json;not null"`
	SessionVersion      int            `json:"session_version" gorm:"type:int;not null;default:1"`
	Status              string         `json:"status" gorm:"type:enum('active','completed','abandoned');not null;default:'active';index;index:idx_case_sessions_user_status_started,priority:2"`
	InitialAssessment   *string        `json:"initial_assessment" gorm:"type:varchar(80)"`
	InitialConfidence   *int           `json:"initial_confidence" gorm:"type:int"`
	StartedAt           time.Time      `json:"started_at" gorm:"not null;index;index:idx_case_sessions_user_started;index:idx_case_sessions_user_status_started,priority:3"`
	LastActivityAt      time.Time      `json:"last_activity_at" gorm:"not null;index"`
	SubmittedAt         *time.Time     `json:"submitted_at" gorm:"index"`
	AppVersion          *string        `json:"app_version" gorm:"type:varchar(30)"`
	CreatedAt           time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt           time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt           gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	Answers          []CaseSessionAnswer           `gorm:"foreignKey:CaseSessionID;references:CaseSessionID;constraint:onDelete:CASCADE"`
	EvidenceProgress []CaseSessionEvidenceProgress `gorm:"foreignKey:CaseSessionID;references:CaseSessionID;constraint:onDelete:CASCADE"`
	Events           []CaseSessionEvent            `gorm:"foreignKey:CaseSessionID;references:CaseSessionID;constraint:onDelete:CASCADE"`
	IdempotencyKeys  []CaseSessionIdempotencyKey   `gorm:"foreignKey:CaseSessionID;references:CaseSessionID;constraint:onDelete:CASCADE"`
	Result           *CaseSessionResult            `gorm:"foreignKey:CaseSessionID;references:CaseSessionID;constraint:onDelete:CASCADE"`
}

type CaseSessionAnswer struct {
	CaseSessionAnswerID uuid.UUID `json:"case_session_answer_id" gorm:"type:varchar(36);primaryKey"`
	CaseSessionID       uuid.UUID `json:"case_session_id" gorm:"type:varchar(36);not null;index;uniqueIndex:uq_session_question_answer;index:idx_session_answers_session_final_saved,priority:1"`
	CaseQuestionID      uuid.UUID `json:"case_question_id" gorm:"type:varchar(36);not null;index;uniqueIndex:uq_session_question_answer"`
	QuestionType        string    `json:"question_type" gorm:"type:varchar(50);not null;index"`
	Value               string    `json:"value" gorm:"type:json;not null"`
	ConfidenceInitial   *int      `json:"confidence_initial" gorm:"type:int"`
	ConfidenceFinal     *int      `json:"confidence_final" gorm:"type:int"`
	IsFinal             bool      `json:"is_final" gorm:"not null;default:false;index;index:idx_session_answers_session_final_saved,priority:2"`
	SavedAt             time.Time `json:"saved_at" gorm:"not null;index;index:idx_session_answers_session_final_saved,priority:3"`
	CreatedAt           time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt           time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type CaseSessionEvidenceProgress struct {
	CaseSessionEvidenceProgressID uuid.UUID `json:"case_session_evidence_progress_id" gorm:"type:varchar(36);primaryKey"`
	CaseSessionID                 uuid.UUID `json:"case_session_id" gorm:"type:varchar(36);not null;index;uniqueIndex:uq_session_evidence_progress"`
	CaseEvidenceID                uuid.UUID `json:"case_evidence_id" gorm:"type:varchar(36);not null;index;uniqueIndex:uq_session_evidence_progress"`
	OpenedCount                   int       `json:"opened_count" gorm:"type:int;not null;default:0"`
	FirstOpenedAt                 time.Time `json:"first_opened_at" gorm:"not null;index"`
	LastOpenedAt                  time.Time `json:"last_opened_at" gorm:"not null;index"`
	CreatedAt                     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt                     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type CaseSessionEvent struct {
	CaseSessionEventID uuid.UUID  `json:"case_session_event_id" gorm:"type:varchar(36);primaryKey"`
	CaseSessionID      uuid.UUID  `json:"case_session_id" gorm:"type:varchar(36);not null;index;index:idx_session_events_session_type_created,priority:1"`
	EventType          string     `json:"event_type" gorm:"type:varchar(80);not null;index;index:idx_session_events_session_type_created,priority:2"`
	CaseEvidenceID     *uuid.UUID `json:"case_evidence_id" gorm:"type:varchar(36);index"`
	CaseQuestionID     *uuid.UUID `json:"case_question_id" gorm:"type:varchar(36);index"`
	Payload            *string    `json:"payload" gorm:"type:json"`
	SessionVersion     int        `json:"session_version" gorm:"type:int;not null"`
	CreatedAt          time.Time  `json:"created_at" gorm:"autoCreateTime;index;index:idx_session_events_session_type_created,priority:3"`
}

type CaseSessionIdempotencyKey struct {
	CaseSessionIdempotencyKeyID uuid.UUID `json:"case_session_idempotency_key_id" gorm:"type:varchar(36);primaryKey"`
	CaseSessionID               uuid.UUID `json:"case_session_id" gorm:"type:varchar(36);not null;index"`
	UserID                      uuid.UUID `json:"user_id" gorm:"type:varchar(36);not null;index"`
	IdempotencyKey              string    `json:"idempotency_key" gorm:"type:varchar(128);not null;uniqueIndex:uq_case_session_idempotency_key"`
	RequestMethod               string    `json:"request_method" gorm:"type:varchar(12);not null"`
	RequestPath                 string    `json:"request_path" gorm:"type:varchar(255);not null"`
	ResponseCode                int       `json:"response_code" gorm:"type:int;not null"`
	ResponseBody                string    `json:"response_body" gorm:"type:json;not null"`
	CreatedAt                   time.Time `json:"created_at" gorm:"autoCreateTime;index"`
	ExpiresAt                   time.Time `json:"expires_at" gorm:"not null;index"`
}

type CaseSessionResult struct {
	CaseSessionResultID uuid.UUID `json:"case_session_result_id" gorm:"type:varchar(36);primaryKey"`
	CaseSessionID       uuid.UUID `json:"case_session_id" gorm:"type:varchar(36);not null;uniqueIndex"`
	UserID              uuid.UUID `json:"user_id" gorm:"type:varchar(36);not null;index"`
	CaseID              uuid.UUID `json:"case_id" gorm:"type:varchar(36);not null;index"`
	CaseVersionID       uuid.UUID `json:"case_version_id" gorm:"type:varchar(36);not null;index"`
	TotalScore          int       `json:"total_score" gorm:"type:int;not null"`
	ScoreBreakdown      string    `json:"score_breakdown" gorm:"type:json;not null"`
	OutcomeKey          string    `json:"outcome_key" gorm:"type:varchar(80);not null;index"`
	OutcomeID           string    `json:"outcome_id" gorm:"type:varchar(150);not null"`
	OutcomeLabel        string    `json:"outcome_label" gorm:"type:varchar(150);not null"`
	NarrativeText       string    `json:"narrative_text" gorm:"type:longtext;not null"`
	CityImpact          string    `json:"city_impact" gorm:"type:json;not null"`
	FinalDecision       string    `json:"final_decision" gorm:"type:varchar(150);not null"`
	FinalConfidence     int       `json:"final_confidence" gorm:"type:int;not null;default:0"`
	Reason              string    `json:"reason" gorm:"type:longtext;not null"`
	XPGained            int       `json:"xp_gained" gorm:"type:int;not null;default:0"`
	CoinGained          int       `json:"coin_gained" gorm:"type:int;not null;default:0"`
	LevelBefore         int       `json:"level_before" gorm:"type:int;not null"`
	LevelAfter          int       `json:"level_after" gorm:"type:int;not null"`
	XPBefore            int       `json:"xp_before" gorm:"type:int;not null"`
	XPAfter             int       `json:"xp_after" gorm:"type:int;not null"`
	ReputationBefore    float64   `json:"reputation_before" gorm:"type:decimal(8,2);not null;default:0"`
	ReputationAfter     float64   `json:"reputation_after" gorm:"type:decimal(8,2);not null;default:0"`
	CreatedAt           time.Time `json:"created_at" gorm:"autoCreateTime;index"`
	UpdatedAt           time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
