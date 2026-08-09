package repository

import (
	"github.com/azmiagr/unesco-hackathon/entity"
	"github.com/azmiagr/unesco-hackathon/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ICaseVersionRepository interface {
	CreateCaseVersion(tx *gorm.DB, caseVersion *entity.CaseVersion) error
	GetCaseVersion(tx *gorm.DB, param model.GetCaseVersionParam) (*entity.CaseVersion, error)
	GetCaseVersionForUpdate(tx *gorm.DB, caseVersionID uuid.UUID) (*entity.CaseVersion, error)
	ListCaseVersionsByCaseID(tx *gorm.DB, caseID uuid.UUID) ([]entity.CaseVersion, error)
	UpdateCaseVersion(tx *gorm.DB, caseVersion *entity.CaseVersion) error
}

type CaseVersionRepository struct {
	db *gorm.DB
}

func NewCaseVersionRepository(db *gorm.DB) ICaseVersionRepository {
	return &CaseVersionRepository{db: db}
}

func (r *CaseVersionRepository) CreateCaseVersion(tx *gorm.DB, caseVersion *entity.CaseVersion) error {
	err := tx.Debug().Create(caseVersion).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *CaseVersionRepository) GetCaseVersion(tx *gorm.DB, param model.GetCaseVersionParam) (*entity.CaseVersion, error) {
	var caseVersion entity.CaseVersion
	query := tx

	if param.CaseVersionID != uuid.Nil {
		query = query.Where("case_version_id = ?", param.CaseVersionID)
	}
	if param.CaseID != uuid.Nil {
		query = query.Where("case_id = ?", param.CaseID)
	}
	if param.VersionNumber > 0 {
		query = query.Where("version_number = ?", param.VersionNumber)
	}
	if param.Status != "" {
		query = query.Where("status = ?", param.Status)
	}

	err := query.First(&caseVersion).Error
	if err != nil {
		return nil, err
	}

	return &caseVersion, nil
}

func (r *CaseVersionRepository) GetCaseVersionForUpdate(tx *gorm.DB, caseVersionID uuid.UUID) (*entity.CaseVersion, error) {
	var caseVersion entity.CaseVersion

	err := tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("case_version_id = ?", caseVersionID).
		First(&caseVersion).Error
	if err != nil {
		return nil, err
	}

	return &caseVersion, nil
}

func (r *CaseVersionRepository) ListCaseVersionsByCaseID(tx *gorm.DB, caseID uuid.UUID) ([]entity.CaseVersion, error) {
	var caseVersions []entity.CaseVersion

	err := tx.
		Where("case_id = ?", caseID).
		Order("version_number DESC").
		Find(&caseVersions).Error
	if err != nil {
		return nil, err
	}

	return caseVersions, nil
}

func (r *CaseVersionRepository) UpdateCaseVersion(tx *gorm.DB, caseVersion *entity.CaseVersion) error {
	err := tx.Debug().Save(caseVersion).Error
	if err != nil {
		return err
	}
	return nil
}
