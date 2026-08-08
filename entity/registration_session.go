package entity

import (
	"time"

	"github.com/google/uuid"
)

type RegistrationSession struct {
	RegistrationSessionID uuid.UUID  `json:"registration_session_id" gorm:"type:varchar(36);primaryKey"`
	SessionTokenHash      string     `json:"-" gorm:"type:varchar(64);not null;uniqueIndex"`
	Email                 string     `json:"email" gorm:"type:varchar(150);not null;index"`
	PasswordHash          string     `json:"-" gorm:"type:varchar(255);not null"`
	OtpCodeHash           string     `json:"-" gorm:"type:varchar(64)"`
	OtpSentAt             *time.Time `json:"otp_sent_at"`
	OtpExpiresAt          *time.Time `json:"otp_expires_at"`
	EmailVerifiedAt       *time.Time `json:"email_verified_at"`
	AvatarID              *uuid.UUID `json:"avatar_id" gorm:"type:varchar(36)"`
	Title                 string     `json:"title" gorm:"type:varchar(255)"`
	CurrentStep           string     `json:"current_step" gorm:"type:varchar(30);not null;index"`
	CompletedAt           *time.Time `json:"completed_at"`
	ExpiresAt             time.Time  `json:"expires_at" gorm:"not null;index"`
	CreatedAt             time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt             time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}
