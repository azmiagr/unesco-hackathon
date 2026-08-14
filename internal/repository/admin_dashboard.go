package repository

import (
	"time"

	"github.com/azmiagr/unesco-hackathon/entity"
	"github.com/azmiagr/unesco-hackathon/model"
	constants "github.com/azmiagr/unesco-hackathon/pkg/constant"
	"gorm.io/gorm"
)

type IAdminDashboardRepository interface {
	GetAdminDashboardMetrics(tx *gorm.DB, param model.AdminDashboardMetricQueryParam) (*model.AdminDashboardMetricRow, error)
	ListAdminDashboardCoinEconomy(tx *gorm.DB, from time.Time, to time.Time) ([]model.AdminDashboardCoinEconomyRow, error)
	ListRecentAdminDashboardActivities(tx *gorm.DB, limit int) ([]model.AdminAuditLogListRow, error)
}

type AdminDashboardRepository struct {
	db *gorm.DB
}

func NewAdminDashboardRepository(db *gorm.DB) IAdminDashboardRepository {
	return &AdminDashboardRepository{db: db}
}

func (r *AdminDashboardRepository) GetAdminDashboardMetrics(tx *gorm.DB, param model.AdminDashboardMetricQueryParam) (*model.AdminDashboardMetricRow, error) {
	var row model.AdminDashboardMetricRow

	err := tx.Raw(`
		SELECT
			COALESCE((SELECT COUNT(*)
				FROM users
				JOIN roles ON roles.role_id = users.role_id
				WHERE roles.role_name = ?), 0) AS total_players,
			COALESCE((SELECT COUNT(*)
				FROM users
				JOIN roles ON roles.role_id = users.role_id
				WHERE roles.role_name = ? AND users.created_at >= ? AND users.created_at < ?), 0) AS current_players,
			COALESCE((SELECT COUNT(*)
				FROM users
				JOIN roles ON roles.role_id = users.role_id
				WHERE roles.role_name = ? AND users.created_at >= ? AND users.created_at < ?), 0) AS previous_players,
			COALESCE((SELECT COUNT(*)
				FROM cases
				WHERE status = ? AND deleted_at IS NULL), 0) AS published_cases,
			COALESCE((SELECT COUNT(*)
				FROM cases
				WHERE status = ? AND published_at >= ? AND published_at < ? AND deleted_at IS NULL), 0) AS current_published_cases,
			COALESCE((SELECT COUNT(*)
				FROM cases
				WHERE status = ? AND published_at >= ? AND published_at < ? AND deleted_at IS NULL), 0) AS previous_published_cases,
			COALESCE((SELECT SUM(coin_balance)
				FROM user_profiles), 0) AS coin_circulating,
			COALESCE((SELECT SUM(coin_gained)
				FROM case_session_results
				WHERE created_at >= ? AND created_at < ?), 0) AS current_coin_earned,
			COALESCE((SELECT SUM(coin_gained)
				FROM case_session_results
				WHERE created_at >= ? AND created_at < ?), 0) AS previous_coin_earned,
			COALESCE((SELECT COUNT(*)
				FROM cases
				WHERE status = ? AND deleted_at IS NULL), 0) AS moderation_pending
	`,
		constants.RoleUser,
		constants.RoleUser, param.CurrentStart, param.CurrentEnd,
		constants.RoleUser, param.PreviousStart, param.PreviousEnd,
		model.CaseStatusPublished,
		model.CaseStatusPublished, param.CurrentStart, param.CurrentEnd,
		model.CaseStatusPublished, param.PreviousStart, param.PreviousEnd,
		param.CurrentStart, param.CurrentEnd,
		param.PreviousStart, param.PreviousEnd,
		model.CaseStatusDraft,
	).Scan(&row).Error
	if err != nil {
		return nil, err
	}

	return &row, nil
}

func (r *AdminDashboardRepository) ListAdminDashboardCoinEconomy(tx *gorm.DB, from time.Time, to time.Time) ([]model.AdminDashboardCoinEconomyRow, error) {
	var rows []model.AdminDashboardCoinEconomyRow

	err := tx.Raw(`
		SELECT
			days.day_date AS date,
			COALESCE(earned.coin_earned, 0) AS coin_earned,
			COALESCE(spent.coin_spent, 0) AS coin_spent
		FROM (
			SELECT DATE(?) AS day_date
			UNION ALL SELECT DATE(DATE_ADD(?, INTERVAL 1 DAY))
			UNION ALL SELECT DATE(DATE_ADD(?, INTERVAL 2 DAY))
			UNION ALL SELECT DATE(DATE_ADD(?, INTERVAL 3 DAY))
			UNION ALL SELECT DATE(DATE_ADD(?, INTERVAL 4 DAY))
			UNION ALL SELECT DATE(DATE_ADD(?, INTERVAL 5 DAY))
			UNION ALL SELECT DATE(DATE_ADD(?, INTERVAL 6 DAY))
		) AS days
		LEFT JOIN (
			SELECT DATE(created_at) AS day_date, SUM(coin_gained) AS coin_earned
			FROM case_session_results
			WHERE created_at >= ? AND created_at < ?
			GROUP BY DATE(created_at)
		) AS earned ON earned.day_date = days.day_date
		LEFT JOIN (
			SELECT DATE(purchased_at) AS day_date, SUM(coin_spent) AS coin_spent
			FROM user_items
			WHERE purchased_at >= ? AND purchased_at < ?
			GROUP BY DATE(purchased_at)
		) AS spent ON spent.day_date = days.day_date
		ORDER BY days.day_date ASC
	`,
		from, from, from, from, from, from, from,
		from, to,
		from, to,
	).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (r *AdminDashboardRepository) ListRecentAdminDashboardActivities(tx *gorm.DB, limit int) ([]model.AdminAuditLogListRow, error) {
	var rows []model.AdminAuditLogListRow
	if limit <= 0 {
		limit = 5
	}

	err := tx.Model(&entity.AuditLog{}).
		Select(`
			audit_logs.audit_log_id AS admin_audit_log_id,
			audit_logs.actor_admin_id,
			audit_logs.actor_name,
			audit_logs.actor_email,
			audit_logs.action_type,
			audit_logs.module,
			audit_logs.target_type,
			audit_logs.target_id,
			audit_logs.target_label,
			audit_logs.detail,
			audit_logs.created_at
		`).
		Order("audit_logs.created_at DESC").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}
