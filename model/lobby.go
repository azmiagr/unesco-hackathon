package model

import "github.com/google/uuid"

type UserLobbyProfileResponse struct {
	UserID      uuid.UUID  `json:"user_id"`
	Username    string     `json:"username"`
	AvatarID    *uuid.UUID `json:"avatar_id"`
	AvatarURL   string     `json:"avatar_url"`
	Title       string     `json:"title"`
	CoinBalance int        `json:"coin_balance"`
}

type UserLobbyLevelProgressResponse struct {
	CurrentLevel    int    `json:"current_level"`
	CurrentXP       int    `json:"current_xp"`
	CurrentLevelXP  int    `json:"current_level_xp"`
	NextLevel       int    `json:"next_level"`
	NextLevelXP     int    `json:"next_level_xp"`
	ProgressXP      int    `json:"progress_xp"`
	RemainingXP     int    `json:"remaining_xp"`
	ProgressPercent int    `json:"progress_percent"`
	Title           string `json:"title"`
}

type UserLobbyCityStatResponse struct {
	Key    string `json:"key"`
	Label  string `json:"label"`
	Value  int    `json:"value"`
	Delta  int    `json:"delta"`
	Status string `json:"status"`
}

type UserLobbyContinueCaseResponse struct {
	CaseID                   uuid.UUID `json:"case_id"`
	CaseSessionID            uuid.UUID `json:"case_session_id"`
	CaseVersionID            uuid.UUID `json:"case_version_id"`
	Title                    string    `json:"title"`
	Slug                     string    `json:"slug"`
	ShortDescription         string    `json:"short_description"`
	DifficultyLevel          string    `json:"difficulty_level"`
	EstimatedDurationMinutes int       `json:"estimated_duration_minutes"`
	ThumbnailURL             *string   `json:"thumbnail_url"`
	SessionVersion           int       `json:"session_version"`
	ProgressPercent          int       `json:"progress_percent"`
	OpenedEvidenceCount      int       `json:"opened_evidence_count"`
	TotalEvidenceCount       int       `json:"total_evidence_count"`
	AnsweredQuestionCount    int       `json:"answered_question_count"`
	RequiredQuestionCount    int       `json:"required_question_count"`
	LastActivityAt           string    `json:"last_activity_at"`
	StartedAt                string    `json:"started_at"`
}

type UserLobbyResponse struct {
	Profile      UserLobbyProfileResponse       `json:"profile"`
	Level        UserLobbyLevelProgressResponse `json:"level"`
	VisualState  string                         `json:"visual_state"`
	CityStats    []UserLobbyCityStatResponse    `json:"city_stats"`
	FeaturedCase *UserCaseCardResponse          `json:"featured_case"`
	ContinueCase *UserLobbyContinueCaseResponse `json:"continue_case"`
	OtherCases   []UserCaseCardResponse         `json:"other_cases"`
}
