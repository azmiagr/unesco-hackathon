package entity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	UserID    uuid.UUID `json:"id" gorm:"type:varchar(36);primaryKey"`
	RoleID    uuid.UUID `json:"role_id" gorm:"type:varchar(36)"`
	Username  string    `json:"username" gorm:"type:varchar(50);not null"`
	Email     string    `json:"email" gorm:"type:varchar(150);not null;unique"`
	Password  string    `json:"password" gorm:"type:varchar(255);not null"`
	Status    string    `json:"status" gorm:"type:enum('active','inactive','suspended', 'banned');default:'inactive'"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	UserProfile UserProfile `gorm:"foreignKey:UserID;references:UserID;constraint:onDelete:CASCADE"`
	OtpCodes    []OtpCode   `gorm:"foreignKey:UserID;references:UserID;constraint:onDelete:CASCADE"`
}

type AdminLoginOtpSession struct {
	AdminLoginOtpSessionID uuid.UUID  `json:"admin_login_otp_session_id" gorm:"type:varchar(36);primaryKey"`
	UserID                 uuid.UUID  `json:"user_id" gorm:"type:varchar(36);not null;index"`
	SessionTokenHash       string     `json:"-" gorm:"type:varchar(64);not null;uniqueIndex"`
	OtpCodeHash            string     `json:"-" gorm:"type:varchar(64);not null"`
	ExpiresAt              time.Time  `json:"expires_at" gorm:"not null;index"`
	VerifiedAt             *time.Time `json:"verified_at"`
	RevokedAt              *time.Time `json:"revoked_at"`
	CreatedAt              time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt              time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}
