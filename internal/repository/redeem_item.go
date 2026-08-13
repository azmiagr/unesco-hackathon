package repository

import (
	"strings"

	"github.com/azmiagr/unesco-hackathon/entity"
	"github.com/azmiagr/unesco-hackathon/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IRedeemItemRepository interface {
	CreateRedeemItem(tx *gorm.DB, redeemItem *entity.RedeemItem) error
	GetRedeemItem(tx *gorm.DB, param model.GetRedeemItemParam) (*entity.RedeemItem, error)
	GetRedeemItemByNormalizedName(tx *gorm.DB, name string) (*entity.RedeemItem, error)
	ListRedeemItemsByNormalizedNameLookup(tx *gorm.DB) ([]entity.RedeemItem, error)
	GetRedeemItemForUpdate(tx *gorm.DB, redeemItemID uuid.UUID) (*entity.RedeemItem, error)
	ListRedeemItems(tx *gorm.DB, param model.ListRedeemItemsParam) ([]entity.RedeemItem, int64, error)
	UpdateRedeemItem(tx *gorm.DB, redeemItem *entity.RedeemItem) error
	DeleteRedeemItem(tx *gorm.DB, redeemItemID uuid.UUID) error
}

type RedeemItemRepository struct {
	db *gorm.DB
}

func NewRedeemItemRepository(db *gorm.DB) IRedeemItemRepository {
	return &RedeemItemRepository{db: db}
}

func (r *RedeemItemRepository) CreateRedeemItem(tx *gorm.DB, redeemItem *entity.RedeemItem) error {
	err := tx.Debug().Create(redeemItem).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *RedeemItemRepository) GetRedeemItem(tx *gorm.DB, param model.GetRedeemItemParam) (*entity.RedeemItem, error) {
	var redeemItem entity.RedeemItem
	query := tx.Model(&entity.RedeemItem{})

	if param.RedeemItemID != uuid.Nil {
		query = query.Where("redeem_items.redeem_item_id = ?", param.RedeemItemID)
	}
	if param.Name != "" {
		query = query.Where(`
			LOWER(TRIM(
				REPLACE(
					REPLACE(
						REPLACE(
							REPLACE(redeem_items.name, CHAR(13), ' '),
							CHAR(10), ' '
						),
						CHAR(9), ' '
					),
					'  ', ' '
				)
			)) = ?
		`, normalizeRedeemItemName(param.Name))
	}

	err := query.First(&redeemItem).Error
	if err != nil {
		return nil, err
	}

	return &redeemItem, nil
}

func (r *RedeemItemRepository) GetRedeemItemByNormalizedName(tx *gorm.DB, name string) (*entity.RedeemItem, error) {
	normalizedName := normalizeRedeemItemName(name)

	if normalizedName == "" {
		return nil, gorm.ErrRecordNotFound
	}

	redeemItems, err := r.ListRedeemItemsByNormalizedNameLookup(tx)
	if err != nil {
		return nil, err
	}

	for _, redeemItem := range redeemItems {
		if normalizeRedeemItemName(redeemItem.Name) == normalizedName {
			return &redeemItem, nil
		}
	}

	return nil, gorm.ErrRecordNotFound
}

func (r *RedeemItemRepository) ListRedeemItemsByNormalizedNameLookup(tx *gorm.DB) ([]entity.RedeemItem, error) {
	var redeemItems []entity.RedeemItem

	err := tx.Model(&entity.RedeemItem{}).
		Order("redeem_items.created_at DESC").
		Find(&redeemItems).Error
	if err != nil {
		return nil, err
	}

	return redeemItems, nil
}

func normalizeRedeemItemName(name string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(name)), " "))
}

func (r *RedeemItemRepository) GetRedeemItemForUpdate(tx *gorm.DB, redeemItemID uuid.UUID) (*entity.RedeemItem, error) {
	var redeemItem entity.RedeemItem

	err := tx.Model(&entity.RedeemItem{}).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("redeem_items.redeem_item_id = ?", redeemItemID).
		First(&redeemItem).Error
	if err != nil {
		return nil, err
	}

	return &redeemItem, nil
}

func (r *RedeemItemRepository) ListRedeemItems(tx *gorm.DB, param model.ListRedeemItemsParam) ([]entity.RedeemItem, int64, error) {
	var redeemItems []entity.RedeemItem
	var total int64

	query := applyRedeemItemFilters(tx.Model(&entity.RedeemItem{}), param)

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = applyRedeemItemFilters(tx.Model(&entity.RedeemItem{}), param).
		Order("redeem_items.created_at DESC").
		Limit(param.Limit).
		Offset(param.Offset).
		Find(&redeemItems).Error
	if err != nil {
		return nil, 0, err
	}

	return redeemItems, total, nil
}

func (r *RedeemItemRepository) UpdateRedeemItem(tx *gorm.DB, redeemItem *entity.RedeemItem) error {
	err := tx.Debug().Save(redeemItem).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *RedeemItemRepository) DeleteRedeemItem(tx *gorm.DB, redeemItemID uuid.UUID) error {
	err := tx.Debug().
		Where("redeem_item_id = ?", redeemItemID).
		Delete(&entity.RedeemItem{}).Error
	if err != nil {
		return err
	}

	return nil
}

func applyRedeemItemFilters(query *gorm.DB, param model.ListRedeemItemsParam) *gorm.DB {
	if param.Search != "" {
		search := "%" + param.Search + "%"
		query = query.Where(
			"redeem_items.name LIKE ? OR redeem_items.partner_name LIKE ? OR redeem_items.description LIKE ?",
			search,
			search,
			search,
		)
	}
	if param.RedeemTypeID != uuid.Nil {
		query = query.Where("redeem_items.redeem_type_id = ?", param.RedeemTypeID)
	}
	if param.TypeCode != "" {
		query = query.Joins("JOIN redeem_types ON redeem_types.redeem_type_id = redeem_items.redeem_type_id").
			Where("redeem_types.code = ?", param.TypeCode)
	}
	if param.Status != "" {
		query = query.Where("redeem_items.status = ?", param.Status)
	}
	if param.ClaimPeriod != "" {
		query = query.Where("redeem_items.claim_period = ?", param.ClaimPeriod)
	}

	return query
}
