package repository

import (
	"github.com/azmiagr/unesco-hackathon/entity"
	"github.com/azmiagr/unesco-hackathon/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IAuditLogRepository interface {
	CreateAuditLog(tx *gorm.DB, auditLog *entity.AuditLog) error
	GetAdminAuditLog(tx *gorm.DB, param model.GetAdminAuditLogParam) (*entity.AuditLog, error)
	ListAdminAuditLogs(tx *gorm.DB, param model.ListAdminAuditLogsParam) ([]model.AdminAuditLogListRow, int64, error)
}

type AuditLogRepository struct {
	db *gorm.DB
}

func NewAuditLogRepository(db *gorm.DB) IAuditLogRepository {
	return &AuditLogRepository{db: db}
}

func (r *AuditLogRepository) CreateAuditLog(tx *gorm.DB, auditLog *entity.AuditLog) error {
	err := tx.Debug().Create(auditLog).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *AuditLogRepository) GetAdminAuditLog(tx *gorm.DB, param model.GetAdminAuditLogParam) (*entity.AuditLog, error) {
	var auditLog entity.AuditLog

	query := tx.Model(&entity.AuditLog{})

	if param.AdminAuditLogID != uuid.Nil {
		query = query.Where("audit_log_id = ?", param.AdminAuditLogID)
	}

	if err := query.First(&auditLog).Error; err != nil {
		return nil, err
	}

	return &auditLog, nil
}

func (r *AuditLogRepository) ListAdminAuditLogs(tx *gorm.DB, param model.ListAdminAuditLogsParam) ([]model.AdminAuditLogListRow, int64, error) {
	var auditLogs []model.AdminAuditLogListRow
	var total int64

	countQuery := applyAdminAuditLogFilters(tx.Model(&entity.AuditLog{}), param)

	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	dataQuery := applyAdminAuditLogFilters(tx.Model(&entity.AuditLog{}), param)

	err := dataQuery.
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
		Limit(param.Limit).
		Offset(param.Offset).
		Scan(&auditLogs).Error
	if err != nil {
		return nil, 0, err
	}

	return auditLogs, total, nil
}

func applyAdminAuditLogFilters(query *gorm.DB, param model.ListAdminAuditLogsParam) *gorm.DB {
	if param.ActorAdminID != uuid.Nil {
		query = query.Where("audit_logs.actor_admin_id = ?", param.ActorAdminID)
	}
	if param.ActionType != "" {
		query = query.Where("audit_logs.action_type = ?", param.ActionType)
	}
	if param.Module != "" {
		query = query.Where("audit_logs.module = ?", param.Module)
	}
	if param.TargetType != "" {
		query = query.Where("audit_logs.target_type = ?", param.TargetType)
	}
	if param.From != nil {
		query = query.Where("audit_logs.created_at >= ?", *param.From)
	}
	if param.To != nil {
		query = query.Where("audit_logs.created_at <= ?", *param.To)
	}

	return query
}
