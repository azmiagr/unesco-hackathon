package model

import "time"

type AdminDashboardMetricQueryParam struct {
	CurrentStart  time.Time
	CurrentEnd    time.Time
	PreviousStart time.Time
	PreviousEnd   time.Time
}

type AdminDashboardRequest struct {
	Range string `form:"range"`
}

type AdminDashboardMetricRow struct {
	TotalPlayers           int64 `json:"total_players"`
	CurrentPlayers         int64 `json:"current_players"`
	PreviousPlayers        int64 `json:"previous_players"`
	PublishedCases         int64 `json:"published_cases"`
	CurrentPublishedCases  int64 `json:"current_published_cases"`
	PreviousPublishedCases int64 `json:"previous_published_cases"`
	CoinCirculating        int64 `json:"coin_circulating"`
	CurrentCoinEarned      int64 `json:"current_coin_earned"`
	PreviousCoinEarned     int64 `json:"previous_coin_earned"`
	ModerationPending      int64 `json:"moderation_pending"`
}

type AdminDashboardCoinEconomyRow struct {
	Date       time.Time `json:"date"`
	CoinEarned int64     `json:"coin_earned"`
	CoinSpent  int64     `json:"coin_spent"`
}

type AdminDashboardResponse struct {
	Greeting         string                            `json:"greeting"`
	Range            AdminDashboardRangeResponse       `json:"range"`
	Cards            []AdminDashboardCardResponse      `json:"cards"`
	RecentActivities []AdminDashboardActivityResponse  `json:"recent_activities"`
	CoinEconomy      AdminDashboardCoinEconomyResponse `json:"coin_economy"`
}

type AdminDashboardRangeResponse struct {
	Key   string `json:"key"`
	From  string `json:"from"`
	To    string `json:"to"`
	Label string `json:"label"`
}

type AdminDashboardCardResponse struct {
	Key      string                      `json:"key"`
	Label    string                      `json:"label"`
	Value    int64                       `json:"value"`
	Display  string                      `json:"display"`
	Severity string                      `json:"severity"`
	Trend    AdminDashboardTrendResponse `json:"trend"`
}

type AdminDashboardTrendResponse struct {
	Direction string  `json:"direction"`
	Percent   float64 `json:"percent"`
	Label     string  `json:"label"`
}

type AdminDashboardActivityResponse struct {
	AuditLogID  string `json:"audit_log_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Module      string `json:"module"`
	Tag         string `json:"tag"`
	ActorName   string `json:"actor_name"`
	CreatedAt   string `json:"created_at"`
	TimeAgo     string `json:"time_ago"`
}

type AdminDashboardCoinEconomyResponse struct {
	Items            []AdminDashboardCoinEconomyItemResponse `json:"items"`
	TotalEarned      int64                                   `json:"total_earned"`
	TotalSpent       int64                                   `json:"total_spent"`
	TotalEarnedLabel string                                  `json:"total_earned_label"`
	TotalSpentLabel  string                                  `json:"total_spent_label"`
}

type AdminDashboardCoinEconomyItemResponse struct {
	Date       string `json:"date"`
	DayLabel   string `json:"day_label"`
	CoinEarned int64  `json:"coin_earned"`
	CoinSpent  int64  `json:"coin_spent"`
}
