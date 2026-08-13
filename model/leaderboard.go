package model

import "github.com/google/uuid"

type ListLeaderboardRequest struct {
	Page  int `form:"page"`
	Limit int `form:"limit"`
}

type ListLeaderboardParam struct {
	Limit  int
	Offset int
}

type LeaderboardEntryRow struct {
	UserID        uuid.UUID  `json:"-"`
	Rank          int        `json:"rank"`
	Username      string     `json:"username"`
	AvatarID      *uuid.UUID `json:"avatar_id"`
	AvatarURL     string     `json:"avatar_url"`
	CurrentLevel  int        `json:"level"`
	Score         int        `json:"score"`
	IsCurrentUser bool       `json:"is_current_user"`
}

type LeaderboardCurrentUserRow = LeaderboardEntryRow

type ListLeaderboardResponse struct {
	Entries    []LeaderboardEntryRow `json:"entries"`
	Me         *LeaderboardEntryRow  `json:"me"`
	Pagination PaginationResponse    `json:"pagination"`
}
