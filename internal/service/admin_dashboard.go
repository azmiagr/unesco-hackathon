package service

import (
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/azmiagr/unesco-hackathon/internal/repository"
	"github.com/azmiagr/unesco-hackathon/model"
	"github.com/azmiagr/unesco-hackathon/pkg/database/mariadb"
	appErrors "github.com/azmiagr/unesco-hackathon/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	adminDashboardRecentActivityLimit = 5
)

type IAdminDashboardService interface {
	GetDashboardByAdmin(adminID uuid.UUID, adminName string, req model.AdminDashboardRequest) (*model.AdminDashboardResponse, error)
}

type AdminDashboardService struct {
	db            *gorm.DB
	dashboardRepo repository.IAdminDashboardRepository
}

func NewAdminDashboardService(dashboardRepo repository.IAdminDashboardRepository) IAdminDashboardService {
	return &AdminDashboardService{
		db:            mariadb.Connection,
		dashboardRepo: dashboardRepo,
	}
}

func (s *AdminDashboardService) GetDashboardByAdmin(adminID uuid.UUID, adminName string, req model.AdminDashboardRequest) (*model.AdminDashboardResponse, error) {
	if adminID == uuid.Nil {
		return nil, appErrors.Unauthorized("unauthorized")
	}

	now := time.Now().UTC()
	rangeKey, currentStart, currentEnd, previousStart, previousEnd := resolveAdminDashboardRange(now, req.Range)

	metrics, err := s.dashboardRepo.GetAdminDashboardMetrics(s.db, model.AdminDashboardMetricQueryParam{
		CurrentStart:  currentStart,
		CurrentEnd:    currentEnd,
		PreviousStart: previousStart,
		PreviousEnd:   previousEnd,
	})
	if err != nil {
		return nil, appErrors.InternalServer("failed to get dashboard metrics")
	}

	activities, err := s.dashboardRepo.ListRecentAdminDashboardActivities(s.db, adminDashboardRecentActivityLimit)
	if err != nil {
		return nil, appErrors.InternalServer("failed to get recent activities")
	}

	coinRows, err := s.dashboardRepo.ListAdminDashboardCoinEconomy(s.db, currentStart, currentEnd)
	if err != nil {
		return nil, appErrors.InternalServer("failed to get coin economy")
	}

	return &model.AdminDashboardResponse{
		Greeting: "Selamat datang, " + adminName + ".",
		Range: model.AdminDashboardRangeResponse{
			Key:   rangeKey,
			From:  currentStart.Format(time.RFC3339),
			To:    currentEnd.Format(time.RFC3339),
			Label: adminDashboardRangeLabel(rangeKey),
		},
		Cards: []model.AdminDashboardCardResponse{
			{
				Key:      "total_players",
				Label:    "Total Pemain",
				Value:    metrics.TotalPlayers,
				Display:  compactNumber(metrics.TotalPlayers),
				Severity: "normal",
				Trend:    buildDashboardTrend(metrics.CurrentPlayers, metrics.PreviousPlayers),
			},
			{
				Key:      "case_published",
				Label:    "Case Published",
				Value:    metrics.PublishedCases,
				Display:  compactNumber(metrics.PublishedCases),
				Severity: "normal",
				Trend:    buildDashboardTrend(metrics.CurrentPublishedCases, metrics.PreviousPublishedCases),
			},
			{
				Key:      "coin_circulating",
				Label:    "Saldo Koin Beredar",
				Value:    metrics.CoinCirculating,
				Display:  compactNumber(metrics.CoinCirculating),
				Severity: "normal",
				Trend:    buildDashboardTrend(metrics.CurrentCoinEarned, metrics.PreviousCoinEarned),
			},
			{
				Key:      "moderation_pending",
				Label:    "Moderasi Pending",
				Value:    metrics.ModerationPending,
				Display:  compactNumber(metrics.ModerationPending),
				Severity: moderationSeverity(metrics.ModerationPending),
				Trend:    moderationTrend(metrics.ModerationPending),
			},
		},
		RecentActivities: mapAdminDashboardActivities(activities, now),
		CoinEconomy:      mapAdminDashboardCoinEconomy(coinRows),
	}, nil
}

func resolveAdminDashboardRange(now time.Time, rawRange string) (string, time.Time, time.Time, time.Time, time.Time) {
	rangeKey := strings.ToLower(strings.TrimSpace(rawRange))
	if rangeKey == "" {
		rangeKey = "week"
	}

	switch rangeKey {
	case "today":
		currentStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		currentEnd := currentStart.AddDate(0, 0, 1)
		return rangeKey, currentStart, currentEnd, currentStart.AddDate(0, 0, -1), currentStart
	case "month":
		currentStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		currentEnd := currentStart.AddDate(0, 1, 0)
		previousStart := currentStart.AddDate(0, -1, 0)
		return rangeKey, currentStart, currentEnd, previousStart, currentStart
	default:
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		currentStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -(weekday - 1))
		currentEnd := currentStart.AddDate(0, 0, 7)
		return "week", currentStart, currentEnd, currentStart.AddDate(0, 0, -7), currentStart
	}
}

func adminDashboardRangeLabel(rangeKey string) string {
	switch rangeKey {
	case "today":
		return "Hari ini"
	case "month":
		return "Bulan ini"
	default:
		return "Minggu ini"
	}
}

func buildDashboardTrend(current int64, previous int64) model.AdminDashboardTrendResponse {
	change := current - previous
	direction := "stable"
	if change > 0 {
		direction = "up"
	} else if change < 0 {
		direction = "down"
	}

	percent := 0.0
	if previous > 0 {
		percent = math.Abs(float64(change)) / float64(previous) * 100
	} else if current > 0 {
		percent = 100
	}
	percent = roundFloat(percent, 2)

	label := "stabil"
	switch direction {
	case "up":
		label = "+ " + trimFloat(percent) + "%"
	case "down":
		label = "- " + trimFloat(percent) + "%"
	}

	return model.AdminDashboardTrendResponse{
		Direction: direction,
		Percent:   percent,
		Label:     label,
	}
}

func moderationSeverity(value int64) string {
	if value > 0 {
		return "critical"
	}
	return "normal"
}

func moderationTrend(value int64) model.AdminDashboardTrendResponse {
	if value > 0 {
		return model.AdminDashboardTrendResponse{Direction: "up", Percent: 0, Label: "kritis!"}
	}
	return model.AdminDashboardTrendResponse{Direction: "stable", Percent: 0, Label: "aman"}
}

func mapAdminDashboardActivities(rows []model.AdminAuditLogListRow, now time.Time) []model.AdminDashboardActivityResponse {
	items := make([]model.AdminDashboardActivityResponse, 0, len(rows))
	for _, row := range rows {
		title := strings.TrimSpace(row.Detail)
		if title == "" {
			title = strings.TrimSpace(row.TargetLabel)
		}
		if title == "" {
			title = strings.TrimSpace(row.ActionType + " " + row.TargetType)
		}

		timeAgo := relativeTime(row.CreatedAt, now)
		items = append(items, model.AdminDashboardActivityResponse{
			AuditLogID:  row.AdminAuditLogID.String(),
			Title:       title,
			Description: timeAgo + " oleh " + actorDisplayName(row.ActorName),
			Module:      row.Module,
			Tag:         auditModuleTag(row.Module),
			ActorName:   actorDisplayName(row.ActorName),
			CreatedAt:   row.CreatedAt.Format(time.RFC3339),
			TimeAgo:     timeAgo,
		})
	}
	return items
}

func mapAdminDashboardCoinEconomy(rows []model.AdminDashboardCoinEconomyRow) model.AdminDashboardCoinEconomyResponse {
	items := make([]model.AdminDashboardCoinEconomyItemResponse, 0, len(rows))
	var totalEarned int64
	var totalSpent int64

	for _, row := range rows {
		totalEarned += row.CoinEarned
		totalSpent += row.CoinSpent
		items = append(items, model.AdminDashboardCoinEconomyItemResponse{
			Date:       row.Date.Format("2006-01-02"),
			DayLabel:   indonesianShortDay(row.Date),
			CoinEarned: row.CoinEarned,
			CoinSpent:  row.CoinSpent,
		})
	}

	return model.AdminDashboardCoinEconomyResponse{
		Items:            items,
		TotalEarned:      totalEarned,
		TotalSpent:       totalSpent,
		TotalEarnedLabel: "+" + formatNumber(totalEarned) + " KOIN",
		TotalSpentLabel:  "-" + formatNumber(totalSpent) + " KOIN",
	}
}

func actorDisplayName(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "System"
	}
	return value
}

func auditModuleTag(module string) string {
	switch strings.ToLower(strings.TrimSpace(module)) {
	case model.AuditModuleCMS:
		return "CMS"
	case model.AuditModuleShop:
		return "SHOP"
	case model.AuditModuleConfig:
		return "CONFIG"
	case model.AuditModuleUsers:
		return "MOD"
	case model.AuditModuleAuth:
		return "AUTH"
	default:
		if module == "" {
			return "SYS"
		}
		return strings.ToUpper(module)
	}
}

func relativeTime(value time.Time, now time.Time) string {
	if value.IsZero() {
		return ""
	}
	duration := now.Sub(value)
	if duration < time.Minute {
		return "baru saja"
	}
	if duration < time.Hour {
		return strconv.Itoa(int(duration.Minutes())) + " menit yang lalu"
	}
	if duration < 24*time.Hour {
		return strconv.Itoa(int(duration.Hours())) + " jam yang lalu"
	}
	return strconv.Itoa(int(duration.Hours()/24)) + " hari yang lalu"
}

func indonesianShortDay(value time.Time) string {
	switch value.Weekday() {
	case time.Monday:
		return "Sen"
	case time.Tuesday:
		return "Sel"
	case time.Wednesday:
		return "Rab"
	case time.Thursday:
		return "Kam"
	case time.Friday:
		return "Jum"
	case time.Saturday:
		return "Sab"
	default:
		return "Min"
	}
}

func compactNumber(value int64) string {
	absolute := math.Abs(float64(value))
	switch {
	case absolute >= 1000000:
		return trimFloat(float64(value)/1000000) + "M"
	case absolute >= 1000:
		return trimFloat(float64(value)/1000) + "K"
	default:
		return strconv.FormatInt(value, 10)
	}
}

func trimFloat(value float64) string {
	return strconv.FormatFloat(roundFloat(value, 1), 'f', -1, 64)
}

func formatNumber(value int64) string {
	sign := ""
	if value < 0 {
		sign = "-"
		value = -value
	}
	raw := strconv.FormatInt(value, 10)
	if len(raw) <= 3 {
		return sign + raw
	}

	var builder strings.Builder
	prefix := len(raw) % 3
	if prefix == 0 {
		prefix = 3
	}
	builder.WriteString(raw[:prefix])
	for i := prefix; i < len(raw); i += 3 {
		builder.WriteString(",")
		builder.WriteString(raw[i : i+3])
	}
	return sign + builder.String()
}
