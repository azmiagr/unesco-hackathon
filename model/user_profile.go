package model

import (
	"time"

	"github.com/google/uuid"
)

type GetUserProfileParam struct {
	UserProfileID uuid.UUID `json:"-"`
	UserID        uuid.UUID `json:"-"`
}

type UserProfileDetailRow struct {
	UserID                     uuid.UUID  `json:"user_id"`
	Username                   string     `json:"username"`
	Email                      string     `json:"email"`
	Status                     string     `json:"status"`
	UserProfileID              uuid.UUID  `json:"user_profile_id"`
	AvatarID                   *uuid.UUID `json:"avatar_id"`
	AvatarURL                  string     `json:"avatar_url"`
	TitleID                    *uuid.UUID `json:"title_id"`
	Title                      string     `json:"title"`
	TitleImageBorder           string     `json:"title_image_border"`
	CurrentLevel               int        `json:"current_level"`
	CurrentXP                  int        `json:"current_xp"`
	CoinBalance                int        `json:"coin_balance"`
	AuditorReputation          float64    `json:"auditor_reputation"`
	EvidenceEvaluationScore    float64    `json:"evidence_evaluation_score"`
	ClaimAnalysisScore         float64    `json:"claim_analysis_score"`
	ConfidenceCalibrationScore float64    `json:"confidence_calibration_score"`
	ReasoningScore             float64    `json:"reasoning_score"`
	SafetyJudgmentScore        float64    `json:"safety_judgment_score"`
}

type UserProfileSummaryResponse struct {
	UserID            uuid.UUID  `json:"user_id"`
	Username          string     `json:"username"`
	Email             string     `json:"email"`
	AvatarID          *uuid.UUID `json:"avatar_id"`
	AvatarURL         string     `json:"avatar_url"`
	TitleID           *uuid.UUID `json:"title_id"`
	Title             string     `json:"title"`
	TitleImageBorder  string     `json:"title_image_border"`
	CurrentLevel      int        `json:"current_level"`
	CurrentXP         int        `json:"current_xp"`
	CoinBalance       int        `json:"coin_balance"`
	AuditorReputation float64    `json:"auditor_reputation"`
	AccuracyPercent   float64    `json:"accuracy_percent"`
	CasesCompleted    int        `json:"cases_completed"`
	StreakCount       int        `json:"streak_count"`
	SeasonLabel       string     `json:"season_label"`
}

type UserProfileLevelProgressResponse struct {
	CurrentLevel    int    `json:"current_level"`
	NextLevel       int    `json:"next_level"`
	CurrentXP       int    `json:"current_xp"`
	NextLevelXP     int    `json:"next_level_xp"`
	RemainingXP     int    `json:"remaining_xp"`
	ProgressXP      int    `json:"progress_xp"`
	ProgressPercent int    `json:"progress_percent"`
	NextUnlockText  string `json:"next_unlock_text"`
}

type UserProfileDetectiveStatResponse struct {
	Key     string  `json:"key"`
	Label   string  `json:"label"`
	Score   float64 `json:"score"`
	Average float64 `json:"average"`
}

type UserProfileAccountResponse struct {
	Email           string `json:"email"`
	IsEmailVerified bool   `json:"is_email_verified"`
	ConnectedTo     string `json:"connected_to"`
}

type UserCaseHistoryItemResponse struct {
	CaseID          uuid.UUID `json:"case_id"`
	Title           string    `json:"title"`
	CompletedAt     string    `json:"completed_at"`
	DifficultyLabel string    `json:"difficulty_label"`
	XPReward        int       `json:"xp_reward"`
	ResultStatus    string    `json:"result_status"`
	ScoreLabel      string    `json:"score_label"`
	IsRetryable     bool      `json:"is_retryable"`
}

type UserCaseHistoryResponse struct {
	Items []UserCaseHistoryItemResponse `json:"items"`
}

type GetUserProfileResponse struct {
	Profile       UserProfileSummaryResponse         `json:"profile"`
	LevelProgress UserProfileLevelProgressResponse   `json:"level_progress"`
	Stats         []UserProfileDetectiveStatResponse `json:"stats"`
	Account       UserProfileAccountResponse         `json:"account"`
	CaseHistory   UserCaseHistoryResponse            `json:"case_history"`
}

type ListUserCaseResultHistoryParam struct {
	UserID uuid.UUID
	Limit  int
	Offset int
}

type UserCaseResultHistoryRow struct {
	CaseID          uuid.UUID `json:"case_id"`
	CaseSessionID   uuid.UUID `json:"case_session_id"`
	Title           string    `json:"title"`
	DifficultyLevel string    `json:"difficulty_level"`
	CaseStatus      string    `json:"case_status"`
	TotalScore      int       `json:"total_score"`
	OutcomeKey      string    `json:"outcome_key"`
	OutcomeLabel    string    `json:"outcome_label"`
	XPGained        int       `json:"xp_gained"`
	CompletedAt     time.Time `json:"completed_at"`
}

type UserCaseResultSummaryRow struct {
	CasesCompleted int     `json:"cases_completed"`
	AccuracyScore  float64 `json:"accuracy_score"`
}

type UserProfileDetectiveStatAverageRow struct {
	EvidenceEvaluationAverage    float64 `json:"evidence_evaluation_average"`
	ClaimAnalysisAverage         float64 `json:"claim_analysis_average"`
	ConfidenceCalibrationAverage float64 `json:"confidence_calibration_average"`
	ReasoningAverage             float64 `json:"reasoning_average"`
	SafetyJudgmentAverage        float64 `json:"safety_judgment_average"`
}
