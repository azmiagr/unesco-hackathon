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

type UserLobbyResponse struct {
	Profile      UserLobbyProfileResponse       `json:"profile"`
	Level        UserLobbyLevelProgressResponse `json:"level"`
	VisualState  string                         `json:"visual_state"`
	CityStats    []UserLobbyCityStatResponse    `json:"city_stats"`
	FeaturedCase *UserCaseCardResponse          `json:"featured_case"`
	ContinueCase *UserCaseCardResponse          `json:"continue_case"`
	OtherCases   []UserCaseCardResponse         `json:"other_cases"`
}
