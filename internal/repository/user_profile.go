package repository

import (
	"github.com/azmiagr/unesco-hackathon/entity"
	"github.com/azmiagr/unesco-hackathon/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IUserProfileRepository interface {
	GetUserProfile(tx *gorm.DB, param model.GetUserProfileParam) (*entity.UserProfile, error)
	GetUserProfileForUpdate(tx *gorm.DB, param model.GetUserProfileParam) (*entity.UserProfile, error)
	CreateUserProfile(tx *gorm.DB, userProfile *entity.UserProfile) error
	UpdateUserProfile(tx *gorm.DB, userProfile *entity.UserProfile) error
	ListLeaderboard(tx *gorm.DB, param model.ListLeaderboardParam) ([]model.LeaderboardEntryRow, int64, error)
	GetUserLeaderboardRank(tx *gorm.DB, userID uuid.UUID) (*model.LeaderboardEntryRow, error)
	GetUserProfileDetail(tx *gorm.DB, userID uuid.UUID) (*model.UserProfileDetailRow, error)
}

type UserProfileRepository struct {
	db *gorm.DB
}

func NewUserProfileRepository(db *gorm.DB) IUserProfileRepository {
	return &UserProfileRepository{db: db}
}

func (r *UserProfileRepository) GetUserProfile(tx *gorm.DB, param model.GetUserProfileParam) (*entity.UserProfile, error) {
	var userProfile entity.UserProfile
	err := tx.Where(&param).First(&userProfile).Error
	if err != nil {
		return nil, err
	}

	return &userProfile, nil
}

func (r *UserProfileRepository) GetUserProfileForUpdate(tx *gorm.DB, param model.GetUserProfileParam) (*entity.UserProfile, error) {
	var userProfile entity.UserProfile
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(&param).First(&userProfile).Error
	if err != nil {
		return nil, err
	}

	return &userProfile, nil
}

func (r *UserProfileRepository) CreateUserProfile(tx *gorm.DB, userProfile *entity.UserProfile) error {
	err := tx.Debug().Create(userProfile).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *UserProfileRepository) UpdateUserProfile(tx *gorm.DB, userProfile *entity.UserProfile) error {
	err := tx.Debug().Save(userProfile).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *UserProfileRepository) ListLeaderboard(tx *gorm.DB, param model.ListLeaderboardParam) ([]model.LeaderboardEntryRow, int64, error) {
	var entries []model.LeaderboardEntryRow
	var total int64

	baseQuery := tx.Table("user_profiles").
		Joins("JOIN users ON users.user_id = user_profiles.user_id").
		Joins("LEFT JOIN avatars ON avatars.avatar_id = user_profiles.avatar_id").
		Where("users.status = ?", "active")

	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := tx.Table("user_profiles").
		Joins("JOIN users ON users.user_id = user_profiles.user_id").
		Joins("LEFT JOIN avatars ON avatars.avatar_id = user_profiles.avatar_id").
		Where("users.status = ?", "active").
		Select(`
			user_profiles.user_id,
			RANK() OVER (ORDER BY user_profiles.current_xp DESC, users.created_at ASC, users.user_id ASC) AS rank,
			users.username,
			user_profiles.avatar_id,
			COALESCE(avatars.image_url, '') AS avatar_url,
			user_profiles.current_level,
			user_profiles.current_xp AS score
		`).
		Order("user_profiles.current_xp DESC, users.created_at ASC, users.user_id ASC").
		Limit(param.Limit).
		Offset(param.Offset).
		Scan(&entries).Error
	if err != nil {
		return nil, 0, err
	}

	return entries, total, nil
}

func (r *UserProfileRepository) GetUserLeaderboardRank(tx *gorm.DB, userID uuid.UUID) (*model.LeaderboardEntryRow, error) {
	var entry model.LeaderboardEntryRow

	err := tx.Table("(?) AS ranked_users",
		tx.Table("user_profiles").
			Joins("JOIN users ON users.user_id = user_profiles.user_id").
			Joins("LEFT JOIN avatars ON avatars.avatar_id = user_profiles.avatar_id").
			Where("users.status = ?", "active").
			Select(`
				user_profiles.user_id,
				RANK() OVER (ORDER BY user_profiles.current_xp DESC, users.created_at ASC, users.user_id ASC) AS rank,
				users.username,
				user_profiles.avatar_id,
				COALESCE(avatars.image_url, '') AS avatar_url,
				user_profiles.current_level,
				user_profiles.current_xp AS score
			`),
	).
		Where("ranked_users.user_id = ?", userID).
		Scan(&entry).Error
	if err != nil {
		return nil, err
	}
	if entry.UserID == uuid.Nil {
		return nil, gorm.ErrRecordNotFound
	}

	return &entry, nil
}

func (r *UserProfileRepository) GetUserProfileDetail(tx *gorm.DB, userID uuid.UUID) (*model.UserProfileDetailRow, error) {
	var profile model.UserProfileDetailRow

	err := tx.Table("users").
		Joins("JOIN user_profiles ON user_profiles.user_id = users.user_id").
		Joins("LEFT JOIN avatars ON avatars.avatar_id = user_profiles.avatar_id").
		Where("users.user_id = ?", userID).
		Select(`
			users.user_id,
			users.username,
			users.email,
			user_profiles.user_profile_id,
			user_profiles.avatar_id,
			COALESCE(avatars.image_url, '') AS avatar_url,
			COALESCE(user_profiles.title, '') AS title,
			COALESCE(user_profiles.current_level, 0) AS current_level,
			COALESCE(user_profiles.current_xp, 0) AS current_xp,
			COALESCE(user_profiles.coin_balance, 0) AS coin_balance,
			COALESCE(user_profiles.auditor_reputation, 0) AS auditor_reputation,
			COALESCE(user_profiles.evidence_evaluation_score, 0) AS evidence_evaluation_score,
			COALESCE(user_profiles.claim_analysis_score, 0) AS claim_analysis_score,
			COALESCE(user_profiles.confidence_calibration_score, 0) AS confidence_calibration_score,
			COALESCE(user_profiles.reasoning_score, 0) AS reasoning_score,
			COALESCE(user_profiles.safety_judgment_score, 0) AS safety_judgment_score
		`).
		Scan(&profile).Error
	if err != nil {
		return nil, err
	}
	if profile.UserID == uuid.Nil {
		return nil, gorm.ErrRecordNotFound
	}

	return &profile, nil
}
