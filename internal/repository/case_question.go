package repository

import (
	"github.com/azmiagr/unesco-hackathon/entity"
	"github.com/azmiagr/unesco-hackathon/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ICaseQuestionRepository interface {
	CreateMCQQuestion(tx *gorm.DB, question *entity.CaseQuestion, options []entity.CaseQuestionMCQOption, evidenceReferences []entity.CaseQuestionEvidenceReference) error
	CreateOpenEndedQuestion(tx *gorm.DB, question *entity.CaseQuestion, openEndedDetail *entity.CaseQuestionOpenEndedDetail, evidenceReferences []entity.CaseQuestionEvidenceReference) error
	CreateConfidenceSliderQuestion(tx *gorm.DB, question *entity.CaseQuestion, confidenceSliderDetail *entity.CaseQuestionConfidenceSliderDetail, evidenceReferences []entity.CaseQuestionEvidenceReference) error
	CreateClaimClassificationQuestion(tx *gorm.DB, question *entity.CaseQuestion, claimClassificationDetail *entity.CaseQuestionClaimClassificationDetail, evidenceReferences []entity.CaseQuestionEvidenceReference) error
	GetCaseQuestion(tx *gorm.DB, param model.GetCaseQuestionParam) (*entity.CaseQuestion, error)
	GetCaseQuestionForUpdate(tx *gorm.DB, caseQuestionID uuid.UUID) (*entity.CaseQuestion, error)
	ListCaseQuestions(tx *gorm.DB, param model.ListCaseQuestionsParam) ([]entity.CaseQuestion, error)
	CountCaseQuestions(tx *gorm.DB, caseVersionID uuid.UUID) (int64, error)
	UpdateCaseQuestion(tx *gorm.DB, question *entity.CaseQuestion) error
	UpdateOpenEndedQuestion(tx *gorm.DB, openEndedDetail *entity.CaseQuestionOpenEndedDetail) error
	UpdateConfidenceSliderQuestion(tx *gorm.DB, confidenceSliderDetail *entity.CaseQuestionConfidenceSliderDetail) error
	UpdateClaimClassificationQuestion(tx *gorm.DB, claimClassificationDetail *entity.CaseQuestionClaimClassificationDetail) error
	ReplaceMCQOptions(tx *gorm.DB, caseQuestionID uuid.UUID, options []entity.CaseQuestionMCQOption) error
	ReplaceEvidenceReferences(tx *gorm.DB, caseQuestionID uuid.UUID, evidenceReferences []entity.CaseQuestionEvidenceReference) error
	DeleteCaseQuestion(tx *gorm.DB, caseQuestionID uuid.UUID) error
	UpdateCaseQuestionSortOrder(tx *gorm.DB, param model.ReorderCaseQuestionParam) error
}

type CaseQuestionRepository struct {
	db *gorm.DB
}

func NewCaseQuestionRepository(db *gorm.DB) ICaseQuestionRepository {
	return &CaseQuestionRepository{db: db}
}

func (r *CaseQuestionRepository) CreateMCQQuestion(
	tx *gorm.DB,
	question *entity.CaseQuestion,
	options []entity.CaseQuestionMCQOption,
	evidenceReferences []entity.CaseQuestionEvidenceReference,
) error {
	err := tx.Debug().Create(question).Error
	if err != nil {
		return err
	}

	if len(options) > 0 {
		err := tx.Debug().Create(&options).Error
		if err != nil {
			return err
		}
	}

	if len(evidenceReferences) > 0 {
		err := tx.Debug().Create(&evidenceReferences).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *CaseQuestionRepository) CreateOpenEndedQuestion(
	tx *gorm.DB,
	question *entity.CaseQuestion,
	openEndedDetail *entity.CaseQuestionOpenEndedDetail,
	evidenceReferences []entity.CaseQuestionEvidenceReference,
) error {
	err := tx.Debug().Create(question).Error
	if err != nil {
		return err
	}

	err = tx.Debug().Create(openEndedDetail).Error
	if err != nil {
		return err
	}

	if len(evidenceReferences) > 0 {
		if err := tx.Debug().Create(&evidenceReferences).Error; err != nil {
			return err
		}
	}

	return nil
}

func (r *CaseQuestionRepository) CreateConfidenceSliderQuestion(
	tx *gorm.DB,
	question *entity.CaseQuestion,
	confidenceSliderDetail *entity.CaseQuestionConfidenceSliderDetail,
	evidenceReferences []entity.CaseQuestionEvidenceReference,
) error {
	err := tx.Debug().Create(question).Error
	if err != nil {
		return err
	}

	err = tx.Debug().Create(confidenceSliderDetail).Error
	if err != nil {
		return err
	}

	if len(evidenceReferences) > 0 {
		err := tx.Debug().Create(&evidenceReferences).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *CaseQuestionRepository) CreateClaimClassificationQuestion(
	tx *gorm.DB,
	question *entity.CaseQuestion,
	claimClassificationDetail *entity.CaseQuestionClaimClassificationDetail,
	evidenceReferences []entity.CaseQuestionEvidenceReference,
) error {
	err := tx.Debug().Create(question).Error
	if err != nil {
		return err
	}

	err = tx.Debug().Create(claimClassificationDetail).Error
	if err != nil {
		return err
	}

	if len(evidenceReferences) > 0 {
		err := tx.Debug().Create(&evidenceReferences).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *CaseQuestionRepository) GetCaseQuestion(tx *gorm.DB, param model.GetCaseQuestionParam) (*entity.CaseQuestion, error) {
	var question entity.CaseQuestion
	query := preloadQuestionDetails(tx)

	if param.CaseQuestionID != uuid.Nil {
		query = query.Where("case_question_id = ?", param.CaseQuestionID)
	}
	if param.CaseVersionID != uuid.Nil {
		query = query.Where("case_version_id = ?", param.CaseVersionID)
	}
	if param.QuestionType != "" {
		query = query.Where("question_type = ?", param.QuestionType)
	}

	if err := query.First(&question).Error; err != nil {
		return nil, err
	}

	return &question, nil
}

func (r *CaseQuestionRepository) GetCaseQuestionForUpdate(tx *gorm.DB, caseQuestionID uuid.UUID) (*entity.CaseQuestion, error) {
	var question entity.CaseQuestion

	err := tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("case_question_id = ?", caseQuestionID).
		First(&question).Error
	if err != nil {
		return nil, err
	}

	return &question, nil
}

func (r *CaseQuestionRepository) ListCaseQuestions(tx *gorm.DB, param model.ListCaseQuestionsParam) ([]entity.CaseQuestion, error) {
	var questions []entity.CaseQuestion
	query := preloadQuestionDetails(tx)

	if param.CaseVersionID != uuid.Nil {
		query = query.Where("case_version_id = ?", param.CaseVersionID)
	}
	if param.QuestionType != "" {
		query = query.Where("question_type = ?", param.QuestionType)
	}

	err := query.
		Order("sort_order ASC").
		Order("created_at ASC").
		Find(&questions).Error
	if err != nil {
		return nil, err
	}

	return questions, nil
}

func (r *CaseQuestionRepository) CountCaseQuestions(tx *gorm.DB, caseVersionID uuid.UUID) (int64, error) {
	var count int64

	err := tx.Model(&entity.CaseQuestion{}).
		Where("case_version_id = ?", caseVersionID).
		Count(&count).Error
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *CaseQuestionRepository) UpdateCaseQuestion(tx *gorm.DB, question *entity.CaseQuestion) error {
	err := tx.Debug().Save(question).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *CaseQuestionRepository) UpdateOpenEndedQuestion(tx *gorm.DB, openEndedDetail *entity.CaseQuestionOpenEndedDetail) error {
	err := tx.Debug().Save(openEndedDetail).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *CaseQuestionRepository) UpdateConfidenceSliderQuestion(tx *gorm.DB, confidenceSliderDetail *entity.CaseQuestionConfidenceSliderDetail) error {
	err := tx.Debug().Save(confidenceSliderDetail).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *CaseQuestionRepository) UpdateClaimClassificationQuestion(tx *gorm.DB, claimClassificationDetail *entity.CaseQuestionClaimClassificationDetail) error {
	err := tx.Debug().Save(claimClassificationDetail).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *CaseQuestionRepository) ReplaceMCQOptions(
	tx *gorm.DB,
	caseQuestionID uuid.UUID,
	options []entity.CaseQuestionMCQOption,
) error {
	err := tx.Debug().
		Where("case_question_id = ?", caseQuestionID).
		Delete(&entity.CaseQuestionMCQOption{}).Error
	if err != nil {
		return err
	}

	if len(options) > 0 {
		err := tx.Debug().Create(&options).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *CaseQuestionRepository) ReplaceEvidenceReferences(
	tx *gorm.DB,
	caseQuestionID uuid.UUID,
	evidenceReferences []entity.CaseQuestionEvidenceReference,
) error {
	err := tx.Debug().
		Where("case_question_id = ?", caseQuestionID).
		Delete(&entity.CaseQuestionEvidenceReference{}).Error
	if err != nil {
		return err
	}

	if len(evidenceReferences) > 0 {
		err := tx.Debug().Create(&evidenceReferences).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *CaseQuestionRepository) DeleteCaseQuestion(tx *gorm.DB, caseQuestionID uuid.UUID) error {
	err := tx.Debug().
		Where("case_question_id = ?", caseQuestionID).
		Delete(&entity.CaseQuestionMCQOption{}).Error
	if err != nil {
		return err
	}

	err = tx.Debug().
		Where("case_question_id = ?", caseQuestionID).
		Delete(&entity.CaseQuestionEvidenceReference{}).Error
	if err != nil {
		return err
	}

	err = tx.Debug().
		Where("case_question_id = ?", caseQuestionID).
		Delete(&entity.CaseQuestionOpenEndedDetail{}).Error
	if err != nil {
		return err
	}

	err = tx.Debug().
		Where("case_question_id = ?", caseQuestionID).
		Delete(&entity.CaseQuestionConfidenceSliderDetail{}).Error
	if err != nil {
		return err
	}

	err = tx.Debug().
		Where("case_question_id = ?", caseQuestionID).
		Delete(&entity.CaseQuestionClaimClassificationDetail{}).Error
	if err != nil {
		return err
	}

	err = tx.Debug().
		Where("case_question_id = ?", caseQuestionID).
		Delete(&entity.CaseQuestion{}).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *CaseQuestionRepository) UpdateCaseQuestionSortOrder(tx *gorm.DB, param model.ReorderCaseQuestionParam) error {
	err := tx.Debug().
		Model(&entity.CaseQuestion{}).
		Where("case_question_id = ?", param.CaseQuestionID).
		Update("sort_order", param.SortOrder).Error
	if err != nil {
		return err
	}

	return nil
}

func preloadQuestionDetails(query *gorm.DB) *gorm.DB {
	return query.
		Preload("MCQOptions", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC").Order("created_at ASC")
		}).
		Preload("EvidenceReferences", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC").Order("created_at ASC")
		}).
		Preload("OpenEndedDetail").
		Preload("ConfidenceSliderDetail").
		Preload("ClaimClassificationDetail")
}
