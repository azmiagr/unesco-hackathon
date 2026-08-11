package repository

import (
	"github.com/azmiagr/unesco-hackathon/entity"
	"github.com/azmiagr/unesco-hackathon/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IItemCategoryRepository interface {
	CreateItemCategory(tx *gorm.DB, category *entity.ItemCategory) error
	GetItemCategory(tx *gorm.DB, param model.GetItemCategoryParam) (*entity.ItemCategory, error)
	GetItemCategoryForUpdate(tx *gorm.DB, itemCategoryID uuid.UUID) (*entity.ItemCategory, error)
	ListItemCategories(tx *gorm.DB, param model.ListItemCategoriesParam) ([]entity.ItemCategory, error)
	UpdateItemCategory(tx *gorm.DB, category *entity.ItemCategory) error
}

type ItemCategoryRepository struct {
	db *gorm.DB
}

func NewItemCategoryRepository(db *gorm.DB) IItemCategoryRepository {
	return &ItemCategoryRepository{db: db}
}

func (r *ItemCategoryRepository) CreateItemCategory(tx *gorm.DB, category *entity.ItemCategory) error {
	err := tx.Debug().Create(category).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *ItemCategoryRepository) GetItemCategory(tx *gorm.DB, param model.GetItemCategoryParam) (*entity.ItemCategory, error) {
	var category entity.ItemCategory
	query := tx

	if param.ItemCategoryID != uuid.Nil {
		query = query.Where("item_category_id = ?", param.ItemCategoryID)
	}
	if param.Code != "" {
		query = query.Where("code = ?", param.Code)
	}

	err := query.First(&category).Error
	if err != nil {
		return nil, err
	}

	return &category, nil
}

func (r *ItemCategoryRepository) GetItemCategoryForUpdate(tx *gorm.DB, itemCategoryID uuid.UUID) (*entity.ItemCategory, error) {
	var category entity.ItemCategory

	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("item_category_id = ?", itemCategoryID).
		First(&category).Error
	if err != nil {
		return nil, err
	}

	return &category, nil
}

func (r *ItemCategoryRepository) ListItemCategories(tx *gorm.DB, param model.ListItemCategoriesParam) ([]entity.ItemCategory, error) {
	var categories []entity.ItemCategory
	query := tx.Model(&entity.ItemCategory{})

	if param.Search != "" {
		search := "%" + param.Search + "%"
		query = query.Where("code LIKE ? OR name LIKE ?", search, search)
	}

	err := query.Order("name ASC").Find(&categories).Error
	if err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *ItemCategoryRepository) UpdateItemCategory(tx *gorm.DB, category *entity.ItemCategory) error {
	err := tx.Debug().Save(category).Error
	if err != nil {
		return err
	}

	return nil
}
