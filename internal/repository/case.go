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
	ListAdminCases(tx *gorm.DB, param model.AdminListCasesParam) ([]model.AdminCaseListRow, int64, error)
	GetAdminCaseDetail(tx *gorm.DB, caseID uuid.UUID) (*model.AdminCaseListRow, error)
	UpdateCase(tx *gorm.DB, caseEntity *entity.Case) error
	HardDeleteCase(tx *gorm.DB, caseID uuid.UUID) error
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

func (r *CaseRepository) ListAdminCases(tx *gorm.DB, param model.AdminListCasesParam) ([]model.AdminCaseListRow, int64, error) {
	var cases []model.AdminCaseListRow
	var total int64

	countQuery := applyAdminCaseFilters(r.buildAdminCaseQuery(tx), param)

	if err := countQuery.Distinct("cases.case_id").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	dataQuery := applyAdminCaseFilters(r.buildAdminCaseQuery(tx), param)

	err := dataQuery.
		Select(adminCaseSelectColumns()).
		Group("cases.case_id, current_versions.case_version_id").
		Order("cases.updated_at DESC").
		Limit(param.Limit).
		Offset(param.Offset).
		Scan(&cases).Error
	if err != nil {
		return nil, 0, err
	}

	return cases, total, nil
}

func (r *CaseRepository) GetAdminCaseDetail(tx *gorm.DB, caseID uuid.UUID) (*model.AdminCaseListRow, error) {
	var caseDetail model.AdminCaseListRow

	err := r.buildAdminCaseQuery(tx).
		Where("cases.case_id = ?", caseID).
		Select(adminCaseSelectColumns()).
		Group("cases.case_id, current_versions.case_version_id").
		First(&caseDetail).Error
	if err != nil {
		return nil, err
	}

	return &caseDetail, nil
}

func (r *CaseRepository) UpdateCase(tx *gorm.DB, caseEntity *entity.Case) error {
	err := tx.Debug().Save(caseEntity).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *CaseRepository) HardDeleteCase(tx *gorm.DB, caseID uuid.UUID) error {
	err := tx.Debug().
		Unscoped().
		Where("case_id = ?", caseID).
		Delete(&entity.Case{}).Error
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

func (r *CaseRepository) buildAdminCaseQuery(tx *gorm.DB) *gorm.DB {
	return tx.Model(&entity.Case{}).
		Joins(`
			LEFT JOIN case_versions current_versions
				ON current_versions.case_id = cases.case_id
				AND current_versions.deleted_at IS NULL
				AND current_versions.version_number = (
					SELECT MAX(version_number)
					FROM case_versions
					WHERE case_versions.case_id = cases.case_id
					AND case_versions.deleted_at IS NULL
				)
		`).
		Joins(`
			LEFT JOIN case_evidences
				ON case_evidences.case_version_id = current_versions.case_version_id
		`)
}

func applyAdminCaseFilters(query *gorm.DB, param model.AdminListCasesParam) *gorm.DB {
	if param.Search != "" {
		search := "%" + param.Search + "%"
		query = query.Where("cases.title LIKE ? OR cases.short_description LIKE ?", search, search)
	}
	if param.Status != "" {
		query = query.Where("cases.status = ?", param.Status)
	}
	if param.Theme != "" {
		query = query.Where("cases.theme = ?", param.Theme)
	}
	if param.CompetencyFocus != "" {
		query = query.Where("cases.competency_focus = ?", param.CompetencyFocus)
	}
	if param.DifficultyLevel != "" {
		query = query.Where("cases.difficulty_level = ?", param.DifficultyLevel)
	}
	if param.RiskLevel != "" {
		query = query.Where("cases.risk_level = ?", param.RiskLevel)
	}
	if param.GenerationSource != "" {
		query = query.Where("cases.generation_source = ?", param.GenerationSource)
	}

	return query
}

func adminCaseSelectColumns() []string {
	return []string{
		"cases.case_id",
		"current_versions.case_version_id AS current_case_version_id",
		"current_versions.version_number",
		"cases.title",
		"cases.slug",
		"cases.short_description",
		"cases.theme",
		"cases.theme_other_text",
		"cases.competency_focus",
		"cases.difficulty_level",
		"cases.risk_level",
		"cases.estimated_duration_minutes",
		"cases.ai_model",
		"cases.minimum_level",
		"cases.minimum_reputation",
		"cases.unlock_requirement",
		"cases.thumbnail_url",
		"cases.thumbnail_prompt",
		"cases.generation_source",
		"cases.status",
		"COALESCE(JSON_LENGTH(current_versions.questions), 0) AS question_count",
		"COUNT(case_evidences.case_evidence_id) AS evidence_count",
		"cases.published_at",
		"cases.created_by",
		"cases.created_at",
		"cases.updated_at",
	}
}
