package repository

import (
	"github.com/azmiagr/unesco-hackathon/entity"
	"github.com/azmiagr/unesco-hackathon/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ICaseEvidenceRepository interface {
	CreateSocialPostEvidence(tx *gorm.DB, evidence *entity.CaseEvidence, socialPost *entity.CaseEvidenceSocialPost) error
	GetCaseEvidence(tx *gorm.DB, param model.GetCaseEvidenceParam) (*entity.CaseEvidence, error)
	GetCaseEvidenceForUpdate(tx *gorm.DB, evidenceID uuid.UUID) (*entity.CaseEvidence, error)
	ListCaseEvidences(tx *gorm.DB, param model.ListCaseEvidencesParam) ([]entity.CaseEvidence, error)
	ListAdminCaseEvidenceRows(tx *gorm.DB, caseVersionID uuid.UUID) ([]model.AdminCaseEvidenceListRow, error)
	UpdateCaseEvidence(tx *gorm.DB, evidence *entity.CaseEvidence) error
	UpdateSocialPostEvidence(tx *gorm.DB, socialPost *entity.CaseEvidenceSocialPost) error
	DeleteCaseEvidence(tx *gorm.DB, evidenceID uuid.UUID) error
	UpdateCaseEvidenceSortOrder(tx *gorm.DB, param model.ReorderCaseEvidenceParam) error
}

type CaseEvidenceRepository struct {
	db *gorm.DB
}

func NewCaseEvidenceRepository(db *gorm.DB) ICaseEvidenceRepository {
	return &CaseEvidenceRepository{db: db}
}

func (r *CaseEvidenceRepository) CreateSocialPostEvidence(tx *gorm.DB, evidence *entity.CaseEvidence, socialPost *entity.CaseEvidenceSocialPost) error {
	err := tx.Debug().Create(evidence).Error
	if err != nil {
		return err
	}

	err = tx.Debug().Create(socialPost).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *CaseEvidenceRepository) GetCaseEvidence(tx *gorm.DB, param model.GetCaseEvidenceParam) (*entity.CaseEvidence, error) {
	var evidence entity.CaseEvidence
	query := tx.Preload("SocialPost")

	if param.CaseEvidenceID != uuid.Nil {
		query = query.Where("case_evidence_id = ?", param.CaseEvidenceID)
	}
	if param.CaseVersionID != uuid.Nil {
		query = query.Where("case_version_id = ?", param.CaseVersionID)
	}
	if param.TemplateType != "" {
		query = query.Where("template_type = ?", param.TemplateType)
	}

	if err := query.First(&evidence).Error; err != nil {
		return nil, err
	}

	return &evidence, nil
}

func (r *CaseEvidenceRepository) GetCaseEvidenceForUpdate(tx *gorm.DB, evidenceID uuid.UUID) (*entity.CaseEvidence, error) {
	var evidence entity.CaseEvidence

	err := tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("case_evidence_id = ?", evidenceID).
		First(&evidence).Error
	if err != nil {
		return nil, err
	}

	return &evidence, nil
}

func (r *CaseEvidenceRepository) ListCaseEvidences(tx *gorm.DB, param model.ListCaseEvidencesParam) ([]entity.CaseEvidence, error) {
	var evidences []entity.CaseEvidence
	query := tx.Preload("SocialPost")

	if param.CaseVersionID != uuid.Nil {
		query = query.Where("case_version_id = ?", param.CaseVersionID)
	}
	if param.TemplateType != "" {
		query = query.Where("template_type = ?", param.TemplateType)
	}

	err := query.
		Order("sort_order ASC").
		Order("created_at ASC").
		Find(&evidences).Error
	if err != nil {
		return nil, err
	}

	return evidences, nil
}

func (r *CaseEvidenceRepository) ListAdminCaseEvidenceRows(tx *gorm.DB, caseVersionID uuid.UUID) ([]model.AdminCaseEvidenceListRow, error) {
	var evidences []model.AdminCaseEvidenceListRow

	err := tx.Model(&entity.CaseEvidence{}).
		Select([]string{
			"case_evidences.case_evidence_id",
			"case_evidences.case_version_id",
			"case_evidences.template_type",
			"case_evidences.label",
			"case_evidences.is_critical",
			"CASE WHEN social_posts.image_url IS NULL OR social_posts.image_url = '' THEN false ELSE true END AS has_image",
			"case_evidences.sort_order",
			"case_evidences.created_at",
			"case_evidences.updated_at",
		}).
		Joins(`
			LEFT JOIN case_evidence_social_posts social_posts
				ON social_posts.case_evidence_id = case_evidences.case_evidence_id
		`).
		Where("case_evidences.case_version_id = ?", caseVersionID).
		Order("case_evidences.sort_order ASC").
		Order("case_evidences.created_at ASC").
		Scan(&evidences).Error
	if err != nil {
		return nil, err
	}

	return evidences, nil
}

func (r *CaseEvidenceRepository) UpdateCaseEvidence(tx *gorm.DB, evidence *entity.CaseEvidence) error {
	err := tx.Debug().Save(evidence).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *CaseEvidenceRepository) UpdateSocialPostEvidence(tx *gorm.DB, socialPost *entity.CaseEvidenceSocialPost) error {
	err := tx.Debug().Save(socialPost).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *CaseEvidenceRepository) DeleteCaseEvidence(tx *gorm.DB, evidenceID uuid.UUID) error {
	err := tx.Debug().
		Where("case_evidence_id = ?", evidenceID).
		Delete(&entity.CaseEvidence{}).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *CaseEvidenceRepository) UpdateCaseEvidenceSortOrder(tx *gorm.DB, param model.ReorderCaseEvidenceParam) error {
	err := tx.Debug().
		Model(&entity.CaseEvidence{}).
		Where("case_evidence_id = ?", param.CaseEvidenceID).
		Update("sort_order", param.SortOrder).Error
	if err != nil {
		return err
	}
	return nil
}
