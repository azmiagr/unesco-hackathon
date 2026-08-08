package model

import (
	"time"

	"github.com/google/uuid"
)

type GetUserParam struct {
	UserID   uuid.UUID `json:"-"`
	Email    string    `json:"-"`
	Username string    `json:"-"`
}

type GetAdminLoginOtpSessionParam struct {
	AdminLoginOtpSessionID uuid.UUID
	UserID                 uuid.UUID
	SessionTokenHash       string
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type RegisterResponse struct {
	SessionToken string `json:"session_token"`
}

type AdminListUsersParam struct {
	Search string
	Role   string
	Status string
	Limit  int
	Offset int
}

type AdminUserListRow struct {
	UserID       uuid.UUID `json:"user_id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	RoleName     string    `json:"role_name"`
	Status       string    `json:"status"`
	CurrentLevel int       `json:"current_level"`
	AvatarURL    string    `json:"avatar_url"`
	CreatedAt    time.Time `json:"created_at"`
}

type AdminUserDetailRow struct {
	UserID                     uuid.UUID `json:"user_id"`
	Username                   string    `json:"username"`
	Email                      string    `json:"email"`
	RoleID                     uuid.UUID `json:"role_id"`
	RoleName                   string    `json:"role_name"`
	Status                     string    `json:"status"`
	UserProfileID              uuid.UUID `json:"user_profile_id"`
	AvatarID                   uuid.UUID `json:"avatar_id"`
	AvatarURL                  string    `json:"avatar_url"`
	Title                      string    `json:"title"`
	CurrentLevel               int       `json:"current_level"`
	CurrentXP                  int       `json:"current_xp"`
	AuditorReputation          float64   `json:"auditor_reputation"`
	EvidenceEvaluationScore    float64   `json:"evidence_evaluation_score"`
	ClaimAnalysisScore         float64   `json:"claim_analysis_score"`
	ConfidenceCalibrationScore float64   `json:"confidence_calibration_score"`
	ReasoningScore             float64   `json:"reasoning_score"`
	SafetyJudgmentScore        float64   `json:"safety_judgment_score"`
	CreatedAt                  time.Time `json:"created_at"`
	UpdatedAt                  time.Time `json:"updated_at"`
}

type AdminUpdateUserAccessParam struct {
	UserID uuid.UUID
	RoleID uuid.UUID
	Status string
}

type AdminListUsersRequest struct {
	Search string `form:"search"`
	Role   string `form:"role"`
	Status string `form:"status"`
	Page   int    `form:"page"`
	Limit  int    `form:"limit"`
}

type AdminListUsersResponse struct {
	Users      []AdminUserListRow `json:"users"`
	Pagination PaginationResponse `json:"pagination"`
}

type PaginationResponse struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

type AdminUserRecentProgressResponse struct {
	Items []any `json:"items"`
}

type AdminUserDetailResponse struct {
	User           AdminUserDetailRow              `json:"user"`
	RecentProgress AdminUserRecentProgressResponse `json:"recent_progress"`
}

type AdminUpdateUserAccessRequest struct {
	RoleName *string `json:"role_name"`
	Status   *string `json:"status"`
}

type AdminUpdateUserAccessResponse struct {
	UserID   uuid.UUID `json:"user_id"`
	RoleName string    `json:"role_name"`
	Status   string    `json:"status"`
}

type AdminDeleteUserResponse struct {
	UserID uuid.UUID `json:"user_id"`
}
