package repository

import (
	"time"

	"github.com/azmiagr/unesco-hackathon/entity"
	"github.com/azmiagr/unesco-hackathon/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IUserItemRepository interface {
	CreateUserItem(tx *gorm.DB, userItem *entity.UserItem) error
	GetUserItem(tx *gorm.DB, param model.GetUserItemParam) (*entity.UserItem, error)
	ListUserInventoryItems(tx *gorm.DB, userID uuid.UUID) ([]model.UserInventoryItemRow, error)
	UpdateUserItem(tx *gorm.DB, userItem *entity.UserItem) error
	CountRedeemPurchasesInPeriod(tx *gorm.DB, userID uuid.UUID, redeemItemID uuid.UUID, start time.Time, end time.Time) (int64, error)
	ClearEquippedItemsByUserAndCategory(tx *gorm.DB, userID uuid.UUID, categoryCode string) error
	ClearEquippedTitlesByUser(tx *gorm.DB, userID uuid.UUID) error
}

type UserItemRepository struct {
	db *gorm.DB
}

func NewUserItemRepository(db *gorm.DB) IUserItemRepository {
	return &UserItemRepository{db: db}
}

func (r *UserItemRepository) CreateUserItem(tx *gorm.DB, userItem *entity.UserItem) error {
	return tx.Debug().Create(userItem).Error
}

func (r *UserItemRepository) GetUserItem(tx *gorm.DB, param model.GetUserItemParam) (*entity.UserItem, error) {
	var userItem entity.UserItem
	query := tx.Model(&entity.UserItem{})

	if param.UserID != uuid.Nil {
		query = query.Where("user_items.user_id = ?", param.UserID)
	}
	if param.ItemID != uuid.Nil {
		query = query.Where("user_items.item_id = ?", param.ItemID)
	}
	if param.TitleID != uuid.Nil {
		query = query.Where("user_items.title_id = ?", param.TitleID)
	}

	if err := query.First(&userItem).Error; err != nil {
		return nil, err
	}

	return &userItem, nil
}

func (r *UserItemRepository) ListUserInventoryItems(tx *gorm.DB, userID uuid.UUID) ([]model.UserInventoryItemRow, error) {
	var rows []model.UserInventoryItemRow

	err := tx.Table("user_items").
		Joins("LEFT JOIN items ON items.item_id = user_items.item_id").
		Joins("LEFT JOIN item_categories ON item_categories.item_category_id = items.item_category_id").
		Joins("LEFT JOIN redeem_items ON redeem_items.redeem_item_id = user_items.redeem_item_id").
		Joins("LEFT JOIN redeem_types ON redeem_types.redeem_type_id = redeem_items.redeem_type_id").
		Joins("LEFT JOIN redeem_codes ON redeem_codes.redeem_code_id = user_items.redeem_code_id").
		Joins("LEFT JOIN titles ON titles.title_id = user_items.title_id").
		Where("user_items.user_id = ?", userID).
		Select(`
			user_items.user_item_id,
			user_items.purchase_type,
			user_items.coin_spent,
			user_items.purchased_at,
			user_items.equipped_at,
			user_items.item_id,
			user_items.title_id,
			items.item_category_id,
			COALESCE(item_categories.code, '') AS category_code,
			COALESCE(item_categories.name, '') AS category_name,
			items.avatar_id,
			COALESCE(items.name, '') AS item_name,
			COALESCE(items.description, '') AS item_description,
			COALESCE(items.image_url, '') AS item_image_url,
			COALESCE(items.status, '') AS item_status,
			COALESCE(titles.title, '') AS title_name,
			COALESCE(titles.unlock_level, 0) AS title_unlock_level,
			COALESCE(titles.image_border, '') AS title_image_border,
			user_items.redeem_item_id,
			user_items.redeem_code_id,
			redeem_items.redeem_type_id,
			COALESCE(redeem_types.code, '') AS redeem_type_code,
			COALESCE(redeem_types.name, '') AS redeem_type_name,
			COALESCE(redeem_items.name, '') AS redeem_name,
			COALESCE(redeem_items.partner_name, '') AS partner_name,
			COALESCE(redeem_items.description, '') AS redeem_description,
			COALESCE(redeem_items.image_url, '') AS redeem_image_url,
			COALESCE(redeem_codes.code, '') AS redeem_code,
			redeem_codes.expires_at AS redeem_code_expires_at,
			redeem_codes.claimed_at AS redeem_code_claimed_at
		`).
		Order("user_items.purchased_at DESC").
		Order("user_items.created_at DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (r *UserItemRepository) UpdateUserItem(tx *gorm.DB, userItem *entity.UserItem) error {
	return tx.Debug().Save(userItem).Error
}

func (r *UserItemRepository) CountRedeemPurchasesInPeriod(tx *gorm.DB, userID uuid.UUID, redeemItemID uuid.UUID, start time.Time, end time.Time) (int64, error) {
	var count int64

	err := tx.Model(&entity.UserItem{}).
		Where(
			"user_id = ? AND redeem_item_id = ? AND purchase_type = ? AND purchased_at >= ? AND purchased_at < ?",
			userID,
			redeemItemID,
			model.UserItemPurchaseTypeRedeem,
			start,
			end,
		).
		Count(&count).Error
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *UserItemRepository) ClearEquippedItemsByUserAndCategory(tx *gorm.DB, userID uuid.UUID, categoryCode string) error {
	return tx.Debug().Exec(`
		UPDATE user_items
		JOIN items ON items.item_id = user_items.item_id
		JOIN item_categories ON item_categories.item_category_id = items.item_category_id
		SET user_items.equipped_at = NULL
		WHERE user_items.user_id = ?
		AND item_categories.code = ?
	`, userID, categoryCode).Error
}

func (r *UserItemRepository) ClearEquippedTitlesByUser(tx *gorm.DB, userID uuid.UUID) error {
	return tx.Debug().Model(&entity.UserItem{}).
		Where("user_id = ? AND title_id IS NOT NULL", userID).
		Update("equipped_at", nil).Error
}
