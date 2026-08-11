package model

import (
	"time"

	"github.com/google/uuid"
)

type GetCaseQuestionParam struct {
	CaseQuestionID uuid.UUID
	CaseVersionID  uuid.UUID
	QuestionType   string
}

type ListCaseQuestionsParam struct {
	CaseVersionID uuid.UUID
	QuestionType  string
}

type ReorderCaseQuestionParam struct {
	CaseQuestionID uuid.UUID
	SortOrder      int
}

type AdminCreateMCQQuestionRequest struct {
	QuestionText       string                        `json:"question_text" binding:"required"`
	ScoringWeight      int                           `json:"scoring_weight" binding:"required"`
	RelatedEvidenceIDs []uuid.UUID                   `json:"related_evidence_ids" binding:"required"`
	Options            []AdminCreateMCQOptionRequest `json:"options" binding:"required"`
	Explanation        string                        `json:"explanation" binding:"required"`
	IsRequired         bool                          `json:"is_required"`
	SortOrder          int                           `json:"sort_order"`
}

type AdminCreateMCQOptionRequest struct {
	OptionCode string `json:"option_code" binding:"required"`
	OptionText string `json:"option_text" binding:"required"`
	IsCorrect  bool   `json:"is_correct"`
}

type AdminMCQOptionResponse struct {
	CaseQuestionMCQOptionID uuid.UUID `json:"case_question_mcq_option_id"`
	CaseQuestionID          uuid.UUID `json:"case_question_id"`
	OptionCode              string    `json:"option_code"`
	OptionText              string    `json:"option_text"`
	IsCorrect               bool      `json:"is_correct"`
	SortOrder               int       `json:"sort_order"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}

type AdminQuestionEvidenceReferenceResponse struct {
	CaseQuestionEvidenceReferenceID uuid.UUID `json:"case_question_evidence_reference_id"`
	CaseQuestionID                  uuid.UUID `json:"case_question_id"`
	CaseEvidenceID                  uuid.UUID `json:"case_evidence_id"`
	SortOrder                       int       `json:"sort_order"`
	CreatedAt                       time.Time `json:"created_at"`
	UpdatedAt                       time.Time `json:"updated_at"`
}

type AdminMCQQuestionResponse struct {
	CaseQuestionID     uuid.UUID                                `json:"case_question_id"`
	CaseVersionID      uuid.UUID                                `json:"case_version_id"`
	QuestionType       string                                   `json:"question_type"`
	QuestionText       string                                   `json:"question_text"`
	Explanation        string                                   `json:"explanation"`
	ScoringWeight      int                                      `json:"scoring_weight"`
	IsRequired         bool                                     `json:"is_required"`
	SortOrder          int                                      `json:"sort_order"`
	Options            []AdminMCQOptionResponse                 `json:"options"`
	EvidenceReferences []AdminQuestionEvidenceReferenceResponse `json:"evidence_references"`
	CreatedAt          time.Time                                `json:"created_at"`
	UpdatedAt          time.Time                                `json:"updated_at"`
}

type AdminCreateMCQQuestionResponse struct {
	Question AdminMCQQuestionResponse `json:"question"`
}

type AdminCreateOpenEndedQuestionRequest struct {
	QuestionText       string      `json:"question_text" binding:"required"`
	ScoringWeight      int         `json:"scoring_weight" binding:"required"`
	RelatedEvidenceIDs []uuid.UUID `json:"related_evidence_ids" binding:"required"`
	ExpectedKeyPoints  string      `json:"expected_key_points" binding:"required"`
	MinimumKeywords    []string    `json:"minimum_keywords" binding:"required"`
	EvaluationRubric   string      `json:"evaluation_rubric" binding:"required"`
	MaxScore           int         `json:"max_score" binding:"required"`
	IsRequired         bool        `json:"is_required"`
	SortOrder          int         `json:"sort_order"`
}

type AdminOpenEndedQuestionResponse struct {
	CaseQuestionID     uuid.UUID                                `json:"case_question_id"`
	CaseVersionID      uuid.UUID                                `json:"case_version_id"`
	QuestionType       string                                   `json:"question_type"`
	QuestionText       string                                   `json:"question_text"`
	ScoringWeight      int                                      `json:"scoring_weight"`
	IsRequired         bool                                     `json:"is_required"`
	SortOrder          int                                      `json:"sort_order"`
	ExpectedKeyPoints  string                                   `json:"expected_key_points"`
	MinimumKeywords    []string                                 `json:"minimum_keywords"`
	EvaluationRubric   string                                   `json:"evaluation_rubric"`
	MaxScore           int                                      `json:"max_score"`
	EvidenceReferences []AdminQuestionEvidenceReferenceResponse `json:"evidence_references"`
	CreatedAt          time.Time                                `json:"created_at"`
	UpdatedAt          time.Time                                `json:"updated_at"`
}

type AdminCreateOpenEndedQuestionResponse struct {
	Question AdminOpenEndedQuestionResponse `json:"question"`
}

type AdminCreateConfidenceSliderQuestionRequest struct {
	QuestionText             string      `json:"question_text" binding:"required"`
	ScoringWeight            int         `json:"scoring_weight" binding:"required"`
	RelatedEvidenceIDs       []uuid.UUID `json:"related_evidence_ids" binding:"required"`
	MinValue                 int         `json:"min_value"`
	MaxValue                 int         `json:"max_value" binding:"required"`
	SnapInterval             int         `json:"snap_interval" binding:"required"`
	DefaultValue             int         `json:"default_value"`
	LabelLow                 string      `json:"label_low" binding:"required"`
	LabelHigh                string      `json:"label_high" binding:"required"`
	ShowWarningOnLargeChange bool        `json:"show_warning_on_large_change"`
	IsRequired               bool        `json:"is_required"`
	SortOrder                int         `json:"sort_order"`
}

type AdminConfidenceSliderQuestionResponse struct {
	CaseQuestionID           uuid.UUID                                `json:"case_question_id"`
	CaseVersionID            uuid.UUID                                `json:"case_version_id"`
	QuestionType             string                                   `json:"question_type"`
	QuestionText             string                                   `json:"question_text"`
	ScoringWeight            int                                      `json:"scoring_weight"`
	IsRequired               bool                                     `json:"is_required"`
	SortOrder                int                                      `json:"sort_order"`
	MinValue                 int                                      `json:"min_value"`
	MaxValue                 int                                      `json:"max_value"`
	SnapInterval             int                                      `json:"snap_interval"`
	DefaultValue             int                                      `json:"default_value"`
	LabelLow                 string                                   `json:"label_low"`
	LabelHigh                string                                   `json:"label_high"`
	ShowWarningOnLargeChange bool                                     `json:"show_warning_on_large_change"`
	EvidenceReferences       []AdminQuestionEvidenceReferenceResponse `json:"evidence_references"`
	CreatedAt                time.Time                                `json:"created_at"`
	UpdatedAt                time.Time                                `json:"updated_at"`
}

type AdminCreateConfidenceSliderQuestionResponse struct {
	Question AdminConfidenceSliderQuestionResponse `json:"question"`
}

type AdminEvidenceOptionResponse struct {
	CaseEvidenceID uuid.UUID `json:"case_evidence_id"`
	Code           string    `json:"code"`
	Label          string    `json:"label"`
	TemplateType   string    `json:"template_type"`
	SortOrder      int       `json:"sort_order"`
}

type AdminListEvidenceOptionsResponse struct {
	CaseID        uuid.UUID                     `json:"case_id"`
	CaseVersionID *uuid.UUID                    `json:"case_version_id"`
	Total         int                           `json:"total"`
	Evidences     []AdminEvidenceOptionResponse `json:"evidences"`
}

type AdminQuestionEvidenceReferenceListItem struct {
	CaseEvidenceID uuid.UUID `json:"case_evidence_id"`
	Code           string    `json:"code"`
	Label          string    `json:"label"`
	TemplateType   string    `json:"template_type"`
	SortOrder      int       `json:"sort_order"`
}

type AdminQuestionListRow struct {
	CaseQuestionID   uuid.UUID                                `json:"case_question_id"`
	CaseVersionID    uuid.UUID                                `json:"case_version_id"`
	Code             string                                   `json:"code"`
	QuestionType     string                                   `json:"question_type"`
	QuestionText     string                                   `json:"question_text"`
	ScoringWeight    int                                      `json:"scoring_weight"`
	IsRequired       bool                                     `json:"is_required"`
	SortOrder        int                                      `json:"sort_order"`
	RelatedEvidences []AdminQuestionEvidenceReferenceListItem `json:"related_evidences"`
	CreatedAt        time.Time                                `json:"created_at"`
	UpdatedAt        time.Time                                `json:"updated_at"`
}

type AdminListCaseQuestionsResponse struct {
	CaseID        uuid.UUID              `json:"case_id"`
	CaseVersionID *uuid.UUID             `json:"case_version_id"`
	Total         int                    `json:"total"`
	Questions     []AdminQuestionListRow `json:"questions"`
}

type AdminQuestionDetailResponse struct {
	QuestionType     string                                 `json:"question_type"`
	MCQ              *AdminMCQQuestionResponse              `json:"mcq,omitempty"`
	OpenEnded        *AdminOpenEndedQuestionResponse        `json:"open_ended,omitempty"`
	ConfidenceSlider *AdminConfidenceSliderQuestionResponse `json:"confidence_slider,omitempty"`
}
