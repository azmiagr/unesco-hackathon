package repository

import (
	"github.com/azmiagr/unesco-hackathon/entity"
	"github.com/azmiagr/unesco-hackathon/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ICaseRepository interface {
	CreateCase(tx *gorm.DB, caseEntity *entity.Case) error
	GetCase(tx *gorm.DB, param model.GetCaseParam) (*entity.Case, error)
	GetCaseForUpdate(tx *gorm.DB, caseID uuid.UUID) (*entity.Case, error)
	ListCases(tx *gorm.DB, param model.ListCasesParam) ([]entity.Case, int64, error)
	UpdateCase(tx *gorm.DB, caseEntity *entity.Case) error
	CaseExists(tx *gorm.DB, param model.GetCaseParam) (bool, error)
}

type CaseRepository struct {
	db *gorm.DB
}

func NewCaseRepository(db *gorm.DB) ICaseRepository {
	return &CaseRepository{db: db}
}

func (r *CaseRepository) CreateCase(tx *gorm.DB, caseEntity *entity.Case) error {
	err := tx.Debug().Create(caseEntity).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *CaseRepository) GetCase(tx *gorm.DB, param model.GetCaseParam) (*entity.Case, error) {
	var caseEntity entity.Case
	query := tx

	if param.CaseID != uuid.Nil {
		query = query.Where("case_id = ?", param.CaseID)
	}
	if param.Slug != "" {
		query = query.Where("slug = ?", param.Slug)
	}

	err := query.First(&caseEntity).Error
	if err != nil {
		return nil, err
	}

	return &caseEntity, nil
}

func (r *CaseRepository) GetCaseForUpdate(tx *gorm.DB, caseID uuid.UUID) (*entity.Case, error) {
	var caseEntity entity.Case

	err := tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("case_id = ?", caseID).
		First(&caseEntity).Error
	if err != nil {
		return nil, err
	}

	return &caseEntity, nil
}

func (r *CaseRepository) ListCases(tx *gorm.DB, param model.ListCasesParam) ([]entity.Case, int64, error) {
	var cases []entity.Case
	var total int64

	query := tx.Model(&entity.Case{})

	if param.Search != "" {
		search := "%" + param.Search + "%"
		query = query.Where("title LIKE ? OR short_description LIKE ?", search, search)
	}
	if param.Status != "" {
		query = query.Where("status = ?", param.Status)
	}
	if param.Theme != "" {
		query = query.Where("theme = ?", param.Theme)
	}
	if param.CompetencyFocus != "" {
		query = query.Where("competency_focus = ?", param.CompetencyFocus)
	}
	if param.DifficultyLevel != "" {
		query = query.Where("difficulty_level = ?", param.DifficultyLevel)
	}
	if param.RiskLevel != "" {
		query = query.Where("risk_level = ?", param.RiskLevel)
	}
	if param.GenerationSource != "" {
		query = query.Where("generation_source = ?", param.GenerationSource)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("created_at DESC").
		Limit(param.Limit).
		Offset(param.Offset).
		Find(&cases).Error
	if err != nil {
		return nil, 0, err
	}

	return cases, total, nil
}

func (r *CaseRepository) UpdateCase(tx *gorm.DB, caseEntity *entity.Case) error {
	err := tx.Debug().Save(caseEntity).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *CaseRepository) CaseExists(tx *gorm.DB, param model.GetCaseParam) (bool, error) {
	var count int64
	query := tx.Model(&entity.Case{})

	if param.CaseID != uuid.Nil {
		query = query.Where("case_id = ?", param.CaseID)
	}
	if param.Slug != "" {
		query = query.Where("slug = ?", param.Slug)
	}

	if err := query.Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}
