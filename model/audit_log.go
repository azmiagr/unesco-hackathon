package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	AuditActionLogin        = "login"
	AuditActionCreate       = "create"
	AuditActionUpdate       = "update"
	AuditActionDelete       = "delete"
	AuditActionConfigChange = "config_change"

	AuditModuleAuth   = "auth"
	AuditModuleCMS    = "cms"
	AuditModuleUsers  = "users"
	AuditModuleShop   = "shop"
	AuditModuleConfig = "config"
	AuditModuleSystem = "system"
)

type GetAdminAuditLogParam struct {
	AdminAuditLogID uuid.UUID
}

type ListAdminAuditLogsParam struct {
	ActorAdminID uuid.UUID
	ActionType   string
	Module       string
	TargetType   string
	From         *time.Time
	To           *time.Time
	Limit        int
	Offset       int
}

type AdminListAuditLogsRequest struct {
	ActorAdminID string `form:"actor_admin_id"`
	ActionType   string `form:"action_type"`
	Action       string `form:"action"`
	Module       string `form:"module"`
	TargetType   string `form:"target_type"`
	From         string `form:"from"`
	To           string `form:"to"`
	Range        string `form:"range"`
	Page         int    `form:"page"`
	Limit        int    `form:"limit"`
}

type AdminAuditLogListRow struct {
	AdminAuditLogID uuid.UUID  `json:"admin_audit_log_id"`
	ActorAdminID    *uuid.UUID `json:"actor_admin_id"`
	ActorName       string     `json:"actor_name"`
	ActorEmail      string     `json:"actor_email"`
	ActionType      string     `json:"action_type"`
	Module          string     `json:"module"`
	TargetType      string     `json:"target_type"`
	TargetID        string     `json:"target_id"`
	TargetLabel     string     `json:"target_label"`
	Detail          string     `json:"detail"`
	CreatedAt       time.Time  `json:"created_at"`
}

type AdminListAuditLogsResponse struct {
	AuditLogs  []AdminAuditLogListRow `json:"audit_logs"`
	Pagination PaginationResponse     `json:"pagination"`
}
