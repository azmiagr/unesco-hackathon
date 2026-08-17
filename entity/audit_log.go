package entity

import (
	"time"

	"github.com/google/uuid"
)

type AuditLog struct {
	AuditLogID    uuid.UUID  `json:"audit_log_id" gorm:"type:varchar(36);primaryKey"`
	ActorAdminID  *uuid.UUID `json:"actor_admin_id" gorm:"type:varchar(36);index"`
	ActorName     string     `json:"actor_name" gorm:"type:varchar(100);not null"`
	ActorEmail    string     `json:"actor_email" gorm:"type:varchar(150);not null"`
	ActionType    string     `json:"action_type" gorm:"type:varchar(64);not null;index"`
	Module        string     `json:"module" gorm:"type:varchar(64);not null;index"`
	TargetType    string     `json:"target_type" gorm:"type:varchar(64);not null;index"`
	TargetID      string     `json:"target_id" gorm:"type:varchar(64);not null;index"`
	TargetLabel   string     `json:"target_label" gorm:"type:varchar(255);not null"`
	Detail        string     `json:"detail" gorm:"type:text;not null"`
	PayloadBefore *string    `json:"payload_before" gorm:"type:json"`
	PayloadAfter  *string    `json:"payload_after" gorm:"type:json"`
	IPAddress     *string    `json:"ip_address" gorm:"type:varchar(45)"`
	UserAgent     *string    `json:"user_agent" gorm:"type:varchar(255)"`
	CreatedAt     time.Time  `json:"created_at" gorm:"autoCreateTime;index"`
}
