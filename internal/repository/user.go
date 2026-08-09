package repository

import (
	"github.com/azmiagr/unesco-hackathon/entity"
	"github.com/azmiagr/unesco-hackathon/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IUserRepository interface {
	GetUser(tx *gorm.DB, param model.GetUserParam) (*entity.User, error)
	CreateUser(tx *gorm.DB, user *entity.User) error
	UpdateUser(tx *gorm.DB, user *entity.User) error
	CreateAdminLoginOtpSession(tx *gorm.DB, session *entity.AdminLoginOtpSession) error
	GetAdminLoginOtpSessionForUpdate(tx *gorm.DB, param model.GetAdminLoginOtpSessionParam) (*entity.AdminLoginOtpSession, error)
	UpdateAdminLoginOtpSession(tx *gorm.DB, session *entity.AdminLoginOtpSession) error
	RevokeActiveAdminLoginOtpSessions(tx *gorm.DB, userID uuid.UUID) error
	ListUsers(tx *gorm.DB, param model.AdminListUsersParam) ([]model.AdminUserListRow, int64, error)
	GetUserDetail(tx *gorm.DB, userID uuid.UUID) (*model.AdminUserDetailRow, error)
	GetUserForUpdate(tx *gorm.DB, userID uuid.UUID) (*entity.User, error)
	UpdateUserAccess(tx *gorm.DB, param model.AdminUpdateUserAccessParam) error
	HardDeleteUser(tx *gorm.DB, userID uuid.UUID) error
	UserExists(tx *gorm.DB, param model.GetUserParam) (bool, error)
}

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) IUserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetUser(tx *gorm.DB, param model.GetUserParam) (*entity.User, error) {
	var user entity.User
	query := tx

	if param.UserID != uuid.Nil {
		query = query.Where("user_id = ?", param.UserID)
	}
	if param.Email != "" {
		query = query.Where("email = ?", param.Email)
	}
	if param.Username != "" {
		query = query.Where("username = ?", param.Username)
	}

	if err := query.First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) CreateUser(tx *gorm.DB, user *entity.User) error {
	err := tx.Debug().Create(user).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *UserRepository) UpdateUser(tx *gorm.DB, user *entity.User) error {
	err := tx.Debug().Save(user).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *UserRepository) CreateAdminLoginOtpSession(tx *gorm.DB, session *entity.AdminLoginOtpSession) error {
	err := tx.Debug().Create(session).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *UserRepository) GetAdminLoginOtpSessionForUpdate(tx *gorm.DB, param model.GetAdminLoginOtpSessionParam) (*entity.AdminLoginOtpSession, error) {
	var session entity.AdminLoginOtpSession
	query := tx.Clauses(clause.Locking{Strength: "UPDATE"})

	if param.AdminLoginOtpSessionID != uuid.Nil {
		query = query.Where("admin_login_otp_session_id = ?", param.AdminLoginOtpSessionID)
	}
	if param.UserID != uuid.Nil {
		query = query.Where("user_id = ?", param.UserID)
	}
	if param.SessionTokenHash != "" {
		query = query.Where("session_token_hash = ?", param.SessionTokenHash)
	}

	if err := query.First(&session).Error; err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *UserRepository) UpdateAdminLoginOtpSession(tx *gorm.DB, session *entity.AdminLoginOtpSession) error {
	err := tx.Debug().Save(session).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *UserRepository) RevokeActiveAdminLoginOtpSessions(tx *gorm.DB, userID uuid.UUID) error {
	err := tx.Debug().Model(&entity.AdminLoginOtpSession{}).
		Where("user_id = ? AND verified_at IS NULL AND revoked_at IS NULL", userID).
		Update("revoked_at", gorm.Expr("UTC_TIMESTAMP()")).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *UserRepository) ListUsers(tx *gorm.DB, param model.AdminListUsersParam) ([]model.AdminUserListRow, int64, error) {
	var users []model.AdminUserListRow
	var total int64

	query := tx.Table("users").
		Joins("JOIN roles ON roles.role_id = users.role_id").
		Joins("LEFT JOIN user_profiles ON user_profiles.user_id = users.user_id").
		Joins("LEFT JOIN avatars ON avatars.avatar_id = user_profiles.avatar_id")

	if param.Search != "" {
		search := "%" + param.Search + "%"
		query = query.Where("users.username LIKE ? OR users.email LIKE ?", search, search)
	}

	if param.Role != "" {
		query = query.Where("roles.role_name = ?", param.Role)
	}

	if param.Status != "" {
		query = query.Where("users.status = ?", param.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Select(`
			users.user_id,
			users.username,
			users.email,
			roles.role_name,
			users.status,
			COALESCE(user_profiles.current_level, 0) AS current_level,
			COALESCE(avatars.image_url, '') AS avatar_url,
			users.created_at
		`).
		Order("users.created_at DESC").
		Limit(param.Limit).
		Offset(param.Offset).
		Scan(&users).Error
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil

}

func (r *UserRepository) GetUserDetail(tx *gorm.DB, userID uuid.UUID) (*model.AdminUserDetailRow, error) {
	var user model.AdminUserDetailRow

	err := tx.Table("users").
		Joins("JOIN roles ON roles.role_id = users.role_id").
		Joins("LEFT JOIN user_profiles ON user_profiles.user_id = users.user_id").
		Joins("LEFT JOIN avatars ON avatars.avatar_id = user_profiles.avatar_id").
		Where("users.user_id = ?", userID).
		Select(`
			users.user_id,
			users.username,
			users.email,
			users.role_id,
			roles.role_name,
			users.status,
			user_profiles.user_profile_id,
			user_profiles.avatar_id,
			COALESCE(avatars.image_url, '') AS avatar_url,
			COALESCE(user_profiles.title, '') AS title,
			COALESCE(user_profiles.current_level, 0) AS current_level,
			COALESCE(user_profiles.current_xp, 0) AS current_xp,
			COALESCE(user_profiles.auditor_reputation, 0) AS auditor_reputation,
			COALESCE(user_profiles.evidence_evaluation_score, 0) AS evidence_evaluation_score,
			COALESCE(user_profiles.claim_analysis_score, 0) AS claim_analysis_score,
			COALESCE(user_profiles.confidence_calibration_score, 0) AS confidence_calibration_score,
			COALESCE(user_profiles.reasoning_score, 0) AS reasoning_score,
			COALESCE(user_profiles.safety_judgment_score, 0) AS safety_judgment_score,
			users.created_at,
			users.updated_at
		`).
		Scan(&user).Error
	if err != nil {
		return nil, err
	}

	if user.UserID == uuid.Nil {
		return nil, gorm.ErrRecordNotFound
	}

	return &user, nil
}

func (r *UserRepository) GetUserForUpdate(tx *gorm.DB, userID uuid.UUID) (*entity.User, error) {
	var user entity.User

	err := tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ?", userID).
		First(&user).Error
	if err != nil {
		return nil, err
	}

	return &user, nil

}

func (r *UserRepository) UpdateUserAccess(tx *gorm.DB, param model.AdminUpdateUserAccessParam) error {
	updates := map[string]interface{}{}

	if param.RoleID != uuid.Nil {
		updates["role_id"] = param.RoleID
	}

	if param.Status != "" {
		updates["status"] = param.Status
	}

	if len(updates) == 0 {
		return nil
	}

	err := tx.Model(&entity.User{}).
		Where("user_id = ?", param.UserID).
		Updates(updates).Error

	if err != nil {
		return err
	}

	return nil
}

func (r *UserRepository) HardDeleteUser(tx *gorm.DB, userID uuid.UUID) error {
	err := tx.Unscoped().
		Where("user_id = ?", userID).
		Delete(&entity.User{}).Error

	if err != nil {
		return err
	}

	return nil
}

func (r *UserRepository) UserExists(tx *gorm.DB, param model.GetUserParam) (bool, error) {
	var count int64

	query := tx.Model(&entity.User{})

	if param.UserID != uuid.Nil {
		query = query.Where("user_id = ?", param.UserID)
	}
	if param.Email != "" {
		query = query.Where("email = ?", param.Email)
	}
	if param.Username != "" {
		query = query.Where("username = ?", param.Username)
	}

	err := query.Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
