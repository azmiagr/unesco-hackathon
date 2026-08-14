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
	UpdateUserItem(tx *gorm.DB, userItem *entity.UserItem) error
	CountRedeemPurchasesInPeriod(tx *gorm.DB, userID uuid.UUID, redeemItemID uuid.UUID, start time.Time, end time.Time) (int64, error)
	ClearEquippedItemsByUserAndCategory(tx *gorm.DB, userID uuid.UUID, categoryCode string) error
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

	if err := query.First(&userItem).Error; err != nil {
		return nil, err
	}

	return &userItem, nil
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
