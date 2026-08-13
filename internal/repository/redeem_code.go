package repository

import (
	"time"

	"github.com/azmiagr/unesco-hackathon/entity"
	"github.com/azmiagr/unesco-hackathon/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IRedeemCodeRepository interface {
	CreateRedeemCode(tx *gorm.DB, redeemCode *entity.RedeemCode) error
	CreateRedeemCodes(tx *gorm.DB, redeemCodes []entity.RedeemCode) error
	GetRedeemCode(tx *gorm.DB, param model.GetRedeemCodeParam) (*entity.RedeemCode, error)
	GetRedeemCodeForUpdate(tx *gorm.DB, param model.GetRedeemCodeParam) (*entity.RedeemCode, error)
	GetRedeemCodeRow(tx *gorm.DB, redeemCodeID uuid.UUID) (*model.AdminRedeemCodeListRow, error)
	ListRedeemCodes(tx *gorm.DB, param model.ListRedeemCodesParam) ([]model.AdminRedeemCodeListRow, int64, error)
	DeleteRedeemCode(tx *gorm.DB, redeemCodeID uuid.UUID) error
	ClaimRedeemCode(tx *gorm.DB, redeemCode *entity.RedeemCode, userID uuid.UUID, claimedAt time.Time) error
}

type RedeemCodeRepository struct {
	db *gorm.DB
}

func NewRedeemCodeRepository(db *gorm.DB) IRedeemCodeRepository {
	return &RedeemCodeRepository{db: db}
}

func (r *RedeemCodeRepository) CreateRedeemCode(tx *gorm.DB, redeemCode *entity.RedeemCode) error {
	err := tx.Debug().Create(redeemCode).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *RedeemCodeRepository) CreateRedeemCodes(tx *gorm.DB, redeemCodes []entity.RedeemCode) error {
	if len(redeemCodes) == 0 {
		return nil
	}

	err := tx.Debug().Create(&redeemCodes).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *RedeemCodeRepository) GetRedeemCode(tx *gorm.DB, param model.GetRedeemCodeParam) (*entity.RedeemCode, error) {
	var redeemCode entity.RedeemCode
	query := applyGetRedeemCodeFilters(tx.Model(&entity.RedeemCode{}), param)

	if err := query.First(&redeemCode).Error; err != nil {
		return nil, err
	}

	return &redeemCode, nil
}

func (r *RedeemCodeRepository) GetRedeemCodeForUpdate(tx *gorm.DB, param model.GetRedeemCodeParam) (*entity.RedeemCode, error) {
	var redeemCode entity.RedeemCode
	query := applyGetRedeemCodeFilters(
		tx.Model(&entity.RedeemCode{}).Clauses(clause.Locking{Strength: "UPDATE"}),
		param,
	)

	err := query.First(&redeemCode).Error
	if err != nil {
		return nil, err
	}

	return &redeemCode, nil
}

func (r *RedeemCodeRepository) ListRedeemCodes(tx *gorm.DB, param model.ListRedeemCodesParam) ([]model.AdminRedeemCodeListRow, int64, error) {
	var rows []model.AdminRedeemCodeListRow
	var total int64

	query := applyRedeemCodeFilters(baseRedeemCodeRowQuery(tx), param)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := applyRedeemCodeFilters(baseRedeemCodeRowQuery(tx), param).
		Select(`
			redeem_codes.redeem_code_id,
			redeem_codes.redeem_item_id,
			redeem_items.name AS redeem_item_name,
			redeem_codes.code,
			CASE
				WHEN redeem_codes.claimed_at IS NOT NULL THEN 'claimed'
				WHEN redeem_codes.expires_at < UTC_TIMESTAMP() THEN 'expired'
				ELSE 'available'
			END AS status,
			redeem_codes.claimed_by_user_id,
			COALESCE(users.username, '') AS claimed_by,
			redeem_codes.claimed_at,
			redeem_codes.expires_at,
			redeem_codes.created_at
		`).
		Order("redeem_codes.created_at DESC").
		Limit(param.Limit).
		Offset(param.Offset).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}

func (r *RedeemCodeRepository) GetRedeemCodeRow(tx *gorm.DB, redeemCodeID uuid.UUID) (*model.AdminRedeemCodeListRow, error) {
	var row model.AdminRedeemCodeListRow

	err := baseRedeemCodeRowQuery(tx).
		Where("redeem_codes.redeem_code_id = ?", redeemCodeID).
		Select(`
			redeem_codes.redeem_code_id,
			redeem_codes.redeem_item_id,
			redeem_items.name AS redeem_item_name,
			redeem_codes.code,
			CASE
				WHEN redeem_codes.claimed_at IS NOT NULL THEN 'claimed'
				WHEN redeem_codes.expires_at < UTC_TIMESTAMP() THEN 'expired'
				ELSE 'available'
			END AS status,
			redeem_codes.claimed_by_user_id,
			COALESCE(users.username, '') AS claimed_by,
			redeem_codes.claimed_at,
			redeem_codes.expires_at,
			redeem_codes.created_at
		`).
		Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.RedeemCodeID == uuid.Nil {
		return nil, gorm.ErrRecordNotFound
	}

	return &row, nil
}

func (r *RedeemCodeRepository) DeleteRedeemCode(tx *gorm.DB, redeemCodeID uuid.UUID) error {
	err := tx.Debug().
		Where("redeem_code_id = ?", redeemCodeID).
		Delete(&entity.RedeemCode{}).Error
	if err != nil {
		return err
	}
	return err
}

func (r *RedeemCodeRepository) ClaimRedeemCode(tx *gorm.DB, redeemCode *entity.RedeemCode, userID uuid.UUID, claimedAt time.Time) error {
	err := tx.Debug().
		Model(&entity.RedeemCode{}).
		Where("redeem_code_id = ? AND claimed_at IS NULL", redeemCode.RedeemCodeID).
		Updates(map[string]any{
			"claimed_by_user_id": userID,
			"claimed_at":         claimedAt,
		}).Error
	if err != nil {
		return err
	}
	return err
}

func baseRedeemCodeRowQuery(tx *gorm.DB) *gorm.DB {
	return tx.Table("redeem_codes").
		Joins("JOIN redeem_items ON redeem_items.redeem_item_id = redeem_codes.redeem_item_id").
		Joins("LEFT JOIN users ON users.user_id = redeem_codes.claimed_by_user_id").
		Where("redeem_codes.deleted_at IS NULL")
}

func applyGetRedeemCodeFilters(query *gorm.DB, param model.GetRedeemCodeParam) *gorm.DB {
	if param.RedeemCodeID != uuid.Nil {
		query = query.Where("redeem_codes.redeem_code_id = ?", param.RedeemCodeID)
	}
	if param.RedeemItemID != uuid.Nil {
		query = query.Where("redeem_codes.redeem_item_id = ?", param.RedeemItemID)
	}
	if param.Code != "" {
		query = query.Where("redeem_codes.code = ?", param.Code)
	}

	return query
}

func applyRedeemCodeFilters(query *gorm.DB, param model.ListRedeemCodesParam) *gorm.DB {
	if param.Search != "" {
		search := "%" + param.Search + "%"
		query = query.Where(
			"redeem_codes.code LIKE ? OR redeem_items.name LIKE ? OR users.username LIKE ?",
			search,
			search,
			search,
		)
	}
	if param.RedeemItemID != uuid.Nil {
		query = query.Where("redeem_codes.redeem_item_id = ?", param.RedeemItemID)
	}
	if param.Status != "" {
		switch param.Status {
		case model.RedeemCodeStatusAvailable:
			query = query.Where("redeem_codes.claimed_at IS NULL AND redeem_codes.expires_at >= UTC_TIMESTAMP()")
		case model.RedeemCodeStatusClaimed:
			query = query.Where("redeem_codes.claimed_at IS NOT NULL")
		case model.RedeemCodeStatusExpired:
			query = query.Where("redeem_codes.claimed_at IS NULL AND redeem_codes.expires_at < UTC_TIMESTAMP()")
		}
	}

	return query
}
