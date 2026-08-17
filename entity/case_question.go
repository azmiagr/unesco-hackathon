package entity

import (
	"time"

	"github.com/google/uuid"
)

type CaseQuestion struct {
	CaseQuestionID uuid.UUID `json:"case_question_id" gorm:"type:varchar(36);primaryKey"`
	CaseVersionID  uuid.UUID `json:"case_version_id" gorm:"type:varchar(36);not null;index;index:idx_case_questions_version_sort_created,priority:1"`
	QuestionType   string    `json:"question_type" gorm:"type:varchar(50);not null;index"`
	QuestionText   string    `json:"question_text" gorm:"type:text;not null"`
	Explanation    string    `json:"explanation" gorm:"type:text;not null"`
	ScoringWeight  int       `json:"scoring_weight" gorm:"type:int;not null;default:0"`
	IsRequired     bool      `json:"is_required" gorm:"not null;default:true"`
	SortOrder      int       `json:"sort_order" gorm:"type:int;not null;default:0;index:idx_case_questions_version_sort_created,priority:2"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime;index:idx_case_questions_version_sort_created,priority:3"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	MCQOptions                []CaseQuestionMCQOption                `json:"mcq_options" gorm:"foreignKey:CaseQuestionID;references:CaseQuestionID;constraint:onDelete:CASCADE"`
	EvidenceReferences        []CaseQuestionEvidenceReference        `json:"evidence_references" gorm:"foreignKey:CaseQuestionID;references:CaseQuestionID;constraint:onDelete:CASCADE"`
	OpenEndedDetail           *CaseQuestionOpenEndedDetail           `json:"open_ended_detail" gorm:"foreignKey:CaseQuestionID;references:CaseQuestionID;constraint:onDelete:CASCADE"`
	ConfidenceSliderDetail    *CaseQuestionConfidenceSliderDetail    `json:"confidence_slider_detail" gorm:"foreignKey:CaseQuestionID;references:CaseQuestionID;constraint:onDelete:CASCADE"`
	ClaimClassificationDetail *CaseQuestionClaimClassificationDetail `json:"claim_classification_detail" gorm:"foreignKey:CaseQuestionID;references:CaseQuestionID;constraint:onDelete:CASCADE"`
}

type CaseQuestionMCQOption struct {
	CaseQuestionMCQOptionID uuid.UUID `json:"case_question_mcq_option_id" gorm:"type:varchar(36);primaryKey"`
	CaseQuestionID          uuid.UUID `json:"case_question_id" gorm:"type:varchar(36);not null;index"`
	OptionCode              string    `json:"option_code" gorm:"type:varchar(20);not null"`
	OptionText              string    `json:"option_text" gorm:"type:text;not null"`
	IsCorrect               bool      `json:"is_correct" gorm:"not null;default:false"`
	SortOrder               int       `json:"sort_order" gorm:"type:int;not null;default:0"`
	CreatedAt               time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt               time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type CaseQuestionEvidenceReference struct {
	CaseQuestionEvidenceReferenceID uuid.UUID `json:"case_question_evidence_reference_id" gorm:"type:varchar(36);primaryKey"`
	CaseQuestionID                  uuid.UUID `json:"case_question_id" gorm:"type:varchar(36);not null;index"`
	CaseEvidenceID                  uuid.UUID `json:"case_evidence_id" gorm:"type:varchar(36);not null;index"`
	SortOrder                       int       `json:"sort_order" gorm:"type:int;not null;default:0"`
	CreatedAt                       time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt                       time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type CaseQuestionOpenEndedDetail struct {
	CaseQuestionID    uuid.UUID `json:"case_question_id" gorm:"type:varchar(36);primaryKey"`
	ExpectedKeyPoints string    `json:"expected_key_points" gorm:"type:longtext;not null"`
	MinimumKeywords   string    `json:"minimum_keywords" gorm:"type:json;not null"`
	EvaluationRubric  string    `json:"evaluation_rubric" gorm:"type:longtext;not null"`
	MaxScore          int       `json:"max_score" gorm:"type:int;not null;default:1"`
	CreatedAt         time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type CaseQuestionConfidenceSliderDetail struct {
	CaseQuestionID           uuid.UUID `json:"case_question_id" gorm:"type:varchar(36);primaryKey"`
	MinValue                 int       `json:"min_value" gorm:"type:int;not null;default:0"`
	MaxValue                 int       `json:"max_value" gorm:"type:int;not null;default:100"`
	SnapInterval             int       `json:"snap_interval" gorm:"type:int;not null;default:1"`
	DefaultValue             int       `json:"default_value" gorm:"type:int;not null;default:0"`
	LabelLow                 string    `json:"label_low" gorm:"type:varchar(150);not null"`
	LabelHigh                string    `json:"label_high" gorm:"type:varchar(150);not null"`
	ShowWarningOnLargeChange bool      `json:"show_warning_on_large_change" gorm:"not null;default:false"`
	CreatedAt                time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt                time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type CaseQuestionClaimClassificationDetail struct {
	CaseQuestionID uuid.UUID `json:"case_question_id" gorm:"type:varchar(36);primaryKey"`
	TaxonomyTags   string    `json:"taxonomy_tags" gorm:"type:json;not null"`
	CorrectAnswer  string    `json:"correct_answer" gorm:"type:varchar(150);not null"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
