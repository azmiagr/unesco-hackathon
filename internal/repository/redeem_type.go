package repository

import (
	"github.com/azmiagr/unesco-hackathon/entity"
	"github.com/azmiagr/unesco-hackathon/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IRedeemTypeRepository interface {
	CreateRedeemType(tx *gorm.DB, redeemType *entity.RedeemType) error
	GetRedeemType(tx *gorm.DB, param model.GetRedeemTypeParam) (*entity.RedeemType, error)
	GetRedeemTypeForUpdate(tx *gorm.DB, redeemTypeID uuid.UUID) (*entity.RedeemType, error)
	ListRedeemTypes(tx *gorm.DB, param model.ListRedeemTypesParam) ([]entity.RedeemType, error)
	UpdateRedeemType(tx *gorm.DB, redeemType *entity.RedeemType) error
}

type RedeemTypeRepository struct {
	db *gorm.DB
}

func NewRedeemTypeRepository(db *gorm.DB) IRedeemTypeRepository {
	return &RedeemTypeRepository{db: db}
}

func (r *RedeemTypeRepository) CreateRedeemType(tx *gorm.DB, redeemType *entity.RedeemType) error {
	err := tx.Debug().Create(redeemType).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *RedeemTypeRepository) GetRedeemType(tx *gorm.DB, param model.GetRedeemTypeParam) (*entity.RedeemType, error) {
	var redeemType entity.RedeemType
	query := tx.Model(&entity.RedeemType{})

	if param.RedeemTypeID != uuid.Nil {
		query = query.Where("redeem_type_id = ?", param.RedeemTypeID)
	}
	if param.Code != "" {
		query = query.Where("code = ?", param.Code)
	}

	err := query.First(&redeemType).Error
	if err != nil {
		return nil, err
	}

	return &redeemType, nil
}

func (r *RedeemTypeRepository) GetRedeemTypeForUpdate(tx *gorm.DB, redeemTypeID uuid.UUID) (*entity.RedeemType, error) {
	var redeemType entity.RedeemType

	err := tx.Model(&entity.RedeemType{}).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("redeem_type_id = ?", redeemTypeID).
		First(&redeemType).Error
	if err != nil {
		return nil, err
	}

	return &redeemType, nil
}

func (r *RedeemTypeRepository) ListRedeemTypes(tx *gorm.DB, param model.ListRedeemTypesParam) ([]entity.RedeemType, error) {
	var redeemTypes []entity.RedeemType
	query := tx.Model(&entity.RedeemType{})

	if param.Search != "" {
		search := "%" + param.Search + "%"
		query = query.Where("code LIKE ? OR name LIKE ?", search, search)
	}

	err := query.Order("name ASC").Find(&redeemTypes).Error
	if err != nil {
		return nil, err
	}

	return redeemTypes, nil
}

func (r *RedeemTypeRepository) UpdateRedeemType(tx *gorm.DB, redeemType *entity.RedeemType) error {
	err := tx.Debug().Save(redeemType).Error
	if err != nil {
		return err
	}

	return nil
}
