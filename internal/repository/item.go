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
