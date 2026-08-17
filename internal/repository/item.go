package repository

import (
	"github.com/azmiagr/unesco-hackathon/entity"
	"github.com/azmiagr/unesco-hackathon/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IItemRepository interface {
	CreateItem(tx *gorm.DB, item *entity.Item) error
	GetItem(tx *gorm.DB, param model.GetItemParam) (*entity.Item, error)
	GetItemForUpdate(tx *gorm.DB, itemID uuid.UUID) (*entity.Item, error)
	ListItems(tx *gorm.DB, param model.ListItemsParam) ([]entity.Item, int64, error)
	ListVisibleShopItems(tx *gorm.DB, param model.ListVisibleShopItemsParam) ([]model.UserShopItemRow, int64, error)
	UpdateItem(tx *gorm.DB, item *entity.Item) error
	DeleteItem(tx *gorm.DB, itemID uuid.UUID) error
}

type ItemRepository struct {
	db *gorm.DB
}

func NewItemRepository(db *gorm.DB) IItemRepository {
	return &ItemRepository{db: db}
}

func (r *ItemRepository) CreateItem(tx *gorm.DB, item *entity.Item) error {
	err := tx.Debug().Create(item).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *ItemRepository) GetItem(tx *gorm.DB, param model.GetItemParam) (*entity.Item, error) {
	var item entity.Item
	query := tx.Model(&entity.Item{})

	if param.ItemID != uuid.Nil {
		query = query.Where("items.item_id = ?", param.ItemID)
	}

	err := query.First(&item).Error
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (r *ItemRepository) GetItemForUpdate(tx *gorm.DB, itemID uuid.UUID) (*entity.Item, error) {
	var item entity.Item

	err := tx.Model(&entity.Item{}).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("items.item_id = ?", itemID).
		First(&item).Error
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (r *ItemRepository) ListItems(tx *gorm.DB, param model.ListItemsParam) ([]entity.Item, int64, error) {
	var items []entity.Item
	var total int64

	query := applyItemFilters(tx.Model(&entity.Item{}), param)

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = applyItemFilters(tx.Model(&entity.Item{}), param).
		Order("items.created_at DESC").
		Limit(param.Limit).
		Offset(param.Offset).
		Find(&items).Error
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *ItemRepository) ListVisibleShopItems(tx *gorm.DB, param model.ListVisibleShopItemsParam) ([]model.UserShopItemRow, int64, error) {
	var rows []model.UserShopItemRow
	var total int64

	query := applyVisibleShopItemFilters(tx.Table("items"), param)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query = applyVisibleShopItemFilters(tx.Table("items"), param).
		Joins("LEFT JOIN user_items ON user_items.item_id = items.item_id AND user_items.user_id = ? AND user_items.purchase_type = ?", param.UserID, model.UserItemPurchaseTypeShop).
		Joins("LEFT JOIN user_profiles ON user_profiles.user_id = ?", param.UserID).
		Select(`
			items.item_id,
			items.item_category_id,
			item_categories.code AS category_code,
			item_categories.name AS category_name,
			items.avatar_id,
			items.name,
			items.description,
			items.price_coin,
			items.image_url,
			user_items.user_item_id,
			user_items.equipped_at,
			user_profiles.avatar_id AS current_avatar_id,
			items.created_at
		`)

	if param.PrioritizeOwnedAvatars {
		hasPurchasedAvatar := clause.Expr{
			SQL: `EXISTS (
				SELECT 1
				FROM user_items owned_user_items
				INNER JOIN items owned_items ON owned_items.item_id = owned_user_items.item_id
				INNER JOIN item_categories owned_categories ON owned_categories.item_category_id = owned_items.item_category_id
				WHERE owned_user_items.user_id = ?
					AND owned_user_items.purchase_type = ?
					AND owned_categories.code = ?
			)`,
			Vars: []interface{}{param.UserID, model.UserItemPurchaseTypeShop, model.ItemCategoryAvatar},
		}

		query = query.
			Order(clause.Expr{
				SQL: `CASE WHEN ? THEN
					CASE WHEN item_categories.code = ? AND user_items.user_item_id IS NOT NULL THEN 0 ELSE 1 END
					ELSE 0
				END ASC`,
				Vars: []interface{}{hasPurchasedAvatar, model.ItemCategoryAvatar},
			}).
			Order("RAND()")
	} else if param.Random {
		query = query.Order("RAND()")
	} else {
		query = query.Order("items.created_at ASC")
	}

	err := query.Limit(param.Limit).
		Offset(param.Offset).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}

func (r *ItemRepository) UpdateItem(tx *gorm.DB, item *entity.Item) error {
	err := tx.Debug().Save(item).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *ItemRepository) DeleteItem(tx *gorm.DB, itemID uuid.UUID) error {
	err := tx.Debug().
		Where("item_id = ?", itemID).
		Delete(&entity.Item{}).Error
	if err != nil {
		return err
	}

	return nil
}

func applyItemFilters(query *gorm.DB, param model.ListItemsParam) *gorm.DB {
	if param.Search != "" {
		search := "%" + param.Search + "%"
		query = query.Where("items.name LIKE ? OR items.description LIKE ?", search, search)
	}
	if param.ItemCategoryID != uuid.Nil {
		query = query.Where("items.item_category_id = ?", param.ItemCategoryID)
	}
	if param.CategoryCode != "" {
		query = query.Joins("JOIN item_categories ON item_categories.item_category_id = items.item_category_id").
			Where("item_categories.code = ?", param.CategoryCode)
	}
	if param.Status != "" {
		query = query.Where("items.status = ?", param.Status)
	}
	if param.IsVisible != nil {
		query = query.Where("items.is_visible = ?", *param.IsVisible)
	}
	if param.IsFeatured != nil {
		query = query.Where("items.is_featured = ?", *param.IsFeatured)
	}

	return query
}

func applyVisibleShopItemFilters(query *gorm.DB, param model.ListVisibleShopItemsParam) *gorm.DB {
	query = query.
		Joins("JOIN item_categories ON item_categories.item_category_id = items.item_category_id").
		Where("items.status = ? AND items.is_visible = ? AND items.deleted_at IS NULL", model.ItemStatusActive, true)

	if param.Search != "" {
		search := "%" + param.Search + "%"
		query = query.Where("items.name LIKE ? OR items.description LIKE ?", search, search)
	}
	if param.ItemID != uuid.Nil {
		query = query.Where("items.item_id = ?", param.ItemID)
	}
	if param.ExcludeItemID != uuid.Nil {
		query = query.Where("items.item_id <> ?", param.ExcludeItemID)
	}
	if param.CategoryCode != "" {
		query = query.Where("item_categories.code = ?", param.CategoryCode)
	}

	return query
}
