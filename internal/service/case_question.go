package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/azmiagr/unesco-hackathon/entity"
	"github.com/azmiagr/unesco-hackathon/model"
	appErrors "github.com/azmiagr/unesco-hackathon/pkg/errors"
	"github.com/azmiagr/unesco-hackathon/pkg/helper"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"strings"
)

func (s *CaseService) ListCaseQuestionsByAdmin(caseID uuid.UUID) (*model.AdminListCaseQuestionsResponse, error) {
	if caseID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid case id")
	}

	caseDetail, err := s.caseRepo.GetAdminCaseDetail(s.db, caseID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("case not found")
		}
		return nil, appErrors.InternalServer("failed to get case detail")
	}

	questionRows := []model.AdminQuestionListRow{}
	if caseDetail.CurrentCaseVersionID != nil {
		evidences, err := s.caseEvidenceRepo.ListAdminCaseEvidenceRows(s.db, *caseDetail.CurrentCaseVersionID)
		if err != nil {
			return nil, appErrors.InternalServer("failed to list case evidences")
		}

		evidenceOptions := buildEvidenceOptionMap(evidences)

		questions, err := s.caseQuestionRepo.ListCaseQuestions(s.db, model.ListCaseQuestionsParam{
			CaseVersionID: *caseDetail.CurrentCaseVersionID,
		})
		if err != nil {
			return nil, appErrors.InternalServer("failed to list case questions")
		}

		for i, question := range questions {
			relatedEvidences := make([]model.AdminQuestionEvidenceReferenceListItem, 0, len(question.EvidenceReferences))
			for _, reference := range question.EvidenceReferences {
				if evidenceOption, ok := evidenceOptions[reference.CaseEvidenceID]; ok {
					relatedEvidences = append(relatedEvidences, model.AdminQuestionEvidenceReferenceListItem{
						CaseEvidenceID: evidenceOption.CaseEvidenceID,
						Code:           evidenceOption.Code,
						Label:          evidenceOption.Label,
						TemplateType:   evidenceOption.TemplateType,
						SortOrder:      reference.SortOrder,
					})
				}
			}

			questionRows = append(questionRows, model.AdminQuestionListRow{
				CaseQuestionID:   question.CaseQuestionID,
				CaseVersionID:    question.CaseVersionID,
				Code:             formatQuestionCode(i + 1),
				QuestionType:     question.QuestionType,
				QuestionText:     question.QuestionText,
				ScoringWeight:    question.ScoringWeight,
				IsRequired:       question.IsRequired,
				SortOrder:        question.SortOrder,
				RelatedEvidences: relatedEvidences,
				CreatedAt:        question.CreatedAt,
				UpdatedAt:        question.UpdatedAt,
			})
		}
	}

	return &model.AdminListCaseQuestionsResponse{
		CaseID:        caseDetail.CaseID,
		CaseVersionID: caseDetail.CurrentCaseVersionID,
		Total:         len(questionRows),
		Questions:     questionRows,
	}, nil
}

func (s *CaseService) GetCaseQuestionDetailByAdmin(
	caseID uuid.UUID,
	caseVersionID uuid.UUID,
	caseQuestionID uuid.UUID,
) (*model.AdminQuestionDetailResponse, error) {
	if caseID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid case id")
	}
	if caseVersionID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid case version id")
	}
	if caseQuestionID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid case question id")
	}

	caseVersion, err := s.caseVersionRepo.GetCaseVersion(s.db, model.GetCaseVersionParam{
		CaseVersionID: caseVersionID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("case version not found")
		}
		return nil, appErrors.InternalServer("failed to get case version")
	}

	if caseVersion.CaseID != caseID {
		return nil, appErrors.NotFound("case version not found")
	}

	question, err := s.caseQuestionRepo.GetCaseQuestion(s.db, model.GetCaseQuestionParam{
		CaseQuestionID: caseQuestionID,
		CaseVersionID:  caseVersionID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("case question not found")
		}
		return nil, appErrors.InternalServer("failed to get case question")
	}

	result := &model.AdminQuestionDetailResponse{
		QuestionType: question.QuestionType,
	}

	switch question.QuestionType {
	case model.CaseQuestionTypeMCQ:
		mcq := mapMCQQuestionResponse(question, question.MCQOptions, question.EvidenceReferences)
		result.MCQ = &mcq
	case model.CaseQuestionTypeOpenEnded:
		if question.OpenEndedDetail == nil {
			return nil, appErrors.InternalServer("open ended question detail not found")
		}

		minimumKeywords, err := parseQuestionKeywords(question.OpenEndedDetail.MinimumKeywords)
		if err != nil {
			return nil, err
		}

		openEnded := mapOpenEndedQuestionResponse(question, question.OpenEndedDetail, minimumKeywords, question.EvidenceReferences)
		result.OpenEnded = &openEnded
	case model.CaseQuestionTypeConfidenceSlider:
		if question.ConfidenceSliderDetail == nil {
			return nil, appErrors.InternalServer("confidence slider question detail not found")
		}

		confidenceSlider := mapConfidenceSliderQuestionResponse(question, question.ConfidenceSliderDetail, question.EvidenceReferences)
		result.ConfidenceSlider = &confidenceSlider
	default:
		return nil, appErrors.BadRequest("unsupported question type")
	}

	return result, nil
}

func (s *CaseService) CreateMCQQuestionByAdmin(
	adminUserID uuid.UUID,
	caseID uuid.UUID,
	caseVersionID uuid.UUID,
	req model.AdminCreateMCQQuestionRequest,
) (*model.AdminCreateMCQQuestionResponse, error) {
	if adminUserID == uuid.Nil {
		return nil, appErrors.Unauthorized("unauthorized")
	}
	if caseID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid case id")
	}
	if caseVersionID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid case version id")
	}

	questionText, err := helper.RequireTrimmedString(req.QuestionText, "question text is required")
	if err != nil {
		return nil, err
	}

	explanation, err := helper.RequireTrimmedString(req.Explanation, "explanation is required")
	if err != nil {
		return nil, err
	}

	if req.ScoringWeight < 0 || req.ScoringWeight > 100 {
		return nil, appErrors.BadRequest("scoring weight must be between 0 and 100")
	}

	if len(req.RelatedEvidenceIDs) == 0 {
		return nil, appErrors.BadRequest("related evidence ids are required")
	}

	if len(req.Options) < 2 {
		return nil, appErrors.BadRequest("mcq options must contain at least 2 items")
	}

	questionID := uuid.New()

	evidenceReferences := make([]entity.CaseQuestionEvidenceReference, 0, len(req.RelatedEvidenceIDs))
	seenEvidenceIDs := map[uuid.UUID]bool{}
	for i, evidenceID := range req.RelatedEvidenceIDs {
		if evidenceID == uuid.Nil {
			return nil, appErrors.BadRequest("invalid related evidence id")
		}
		if seenEvidenceIDs[evidenceID] {
			return nil, appErrors.BadRequest("related evidence ids cannot contain duplicates")
		}

		seenEvidenceIDs[evidenceID] = true
		evidenceReferences = append(evidenceReferences, entity.CaseQuestionEvidenceReference{
			CaseQuestionEvidenceReferenceID: uuid.New(),
			CaseQuestionID:                  questionID,
			CaseEvidenceID:                  evidenceID,
			SortOrder:                       i + 1,
		})
	}

	options := make([]entity.CaseQuestionMCQOption, 0, len(req.Options))
	seenOptionCodes := map[string]bool{}
	correctCount := 0

	for i, optionReq := range req.Options {
		optionCode := strings.ToUpper(strings.TrimSpace(optionReq.OptionCode))
		if optionCode == "" {
			return nil, appErrors.BadRequest("option code is required")
		}
		if seenOptionCodes[optionCode] {
			return nil, appErrors.BadRequest("option codes cannot contain duplicates")
		}

		optionText, err := helper.RequireTrimmedString(optionReq.OptionText, "option text is required")
		if err != nil {
			return nil, err
		}

		if optionReq.IsCorrect {
			correctCount++
		}

		seenOptionCodes[optionCode] = true
		options = append(options, entity.CaseQuestionMCQOption{
			CaseQuestionMCQOptionID: uuid.New(),
			CaseQuestionID:          questionID,
			OptionCode:              optionCode,
			OptionText:              optionText,
			IsCorrect:               optionReq.IsCorrect,
			SortOrder:               i + 1,
		})
	}

	if correctCount != 1 {
		return nil, appErrors.BadRequest("mcq question must have exactly one correct option")
	}

	tx := s.db.Begin()
	defer tx.Rollback()

	caseVersion, err := s.caseVersionRepo.GetCaseVersionForUpdate(tx, caseVersionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("case version not found")
		}
		return nil, appErrors.InternalServer("failed to get case version")
	}

	if caseVersion.CaseID != caseID {
		return nil, appErrors.NotFound("case version not found")
	}

	if caseVersion.Status != model.CaseStatusDraft {
		return nil, appErrors.Conflict("case version is not editable")
	}

	for evidenceID := range seenEvidenceIDs {
		_, err := s.caseEvidenceRepo.GetCaseEvidence(tx, model.GetCaseEvidenceParam{
			CaseEvidenceID: evidenceID,
			CaseVersionID:  caseVersionID,
		})
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, appErrors.BadRequest("related evidence must belong to case version")
			}
			return nil, appErrors.InternalServer("failed to validate related evidence")
		}
	}

	question := &entity.CaseQuestion{
		CaseQuestionID: questionID,
		CaseVersionID:  caseVersion.CaseVersionID,
		QuestionType:   model.CaseQuestionTypeMCQ,
		QuestionText:   questionText,
		Explanation:    explanation,
		ScoringWeight:  req.ScoringWeight,
		IsRequired:     req.IsRequired,
		SortOrder:      req.SortOrder,
	}

	err = s.caseQuestionRepo.CreateMCQQuestion(tx, question, options, evidenceReferences)
	if err != nil {
		return nil, appErrors.InternalServer("failed to create mcq question")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	return &model.AdminCreateMCQQuestionResponse{
		Question: mapMCQQuestionResponse(question, options, evidenceReferences),
	}, nil
}

func (s *CaseService) CreateOpenEndedQuestionByAdmin(
	adminUserID uuid.UUID,
	caseID uuid.UUID,
	caseVersionID uuid.UUID,
	req model.AdminCreateOpenEndedQuestionRequest,
) (*model.AdminCreateOpenEndedQuestionResponse, error) {
	if adminUserID == uuid.Nil {
		return nil, appErrors.Unauthorized("unauthorized")
	}
	if caseID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid case id")
	}
	if caseVersionID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid case version id")
	}

	questionText, err := helper.RequireTrimmedString(req.QuestionText, "question text is required")
	if err != nil {
		return nil, err
	}

	expectedKeyPoints, err := helper.RequireTrimmedString(req.ExpectedKeyPoints, "expected key points are required")
	if err != nil {
		return nil, err
	}

	evaluationRubric, err := helper.RequireTrimmedString(req.EvaluationRubric, "evaluation rubric is required")
	if err != nil {
		return nil, err
	}

	if req.ScoringWeight < 0 || req.ScoringWeight > 100 {
		return nil, appErrors.BadRequest("scoring weight must be between 0 and 100")
	}

	if req.MaxScore < 1 {
		return nil, appErrors.BadRequest("max score must be greater than 0")
	}

	if len(req.RelatedEvidenceIDs) == 0 {
		return nil, appErrors.BadRequest("related evidence ids are required")
	}

	minimumKeywords, minimumKeywordsJSON, err := normalizeQuestionKeywords(req.MinimumKeywords)
	if err != nil {
		return nil, err
	}

	questionID := uuid.New()
	evidenceReferences := make([]entity.CaseQuestionEvidenceReference, 0, len(req.RelatedEvidenceIDs))
	seenEvidenceIDs := map[uuid.UUID]bool{}

	for i, evidenceID := range req.RelatedEvidenceIDs {
		if evidenceID == uuid.Nil {
			return nil, appErrors.BadRequest("invalid related evidence id")
		}
		if seenEvidenceIDs[evidenceID] {
			return nil, appErrors.BadRequest("related evidence ids cannot contain duplicates")
		}

		seenEvidenceIDs[evidenceID] = true
		evidenceReferences = append(evidenceReferences, entity.CaseQuestionEvidenceReference{
			CaseQuestionEvidenceReferenceID: uuid.New(),
			CaseQuestionID:                  questionID,
			CaseEvidenceID:                  evidenceID,
			SortOrder:                       i + 1,
		})
	}

	tx := s.db.Begin()
	defer tx.Rollback()

	caseVersion, err := s.caseVersionRepo.GetCaseVersionForUpdate(tx, caseVersionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("case version not found")
		}
		return nil, appErrors.InternalServer("failed to get case version")
	}

	if caseVersion.CaseID != caseID {
		return nil, appErrors.NotFound("case version not found")
	}

	if caseVersion.Status != model.CaseStatusDraft {
		return nil, appErrors.Conflict("case version is not editable")
	}

	for evidenceID := range seenEvidenceIDs {
		_, err := s.caseEvidenceRepo.GetCaseEvidence(tx, model.GetCaseEvidenceParam{
			CaseEvidenceID: evidenceID,
			CaseVersionID:  caseVersionID,
		})
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, appErrors.BadRequest("related evidence must belong to case version")
			}
			return nil, appErrors.InternalServer("failed to validate related evidence")
		}
	}

	question := &entity.CaseQuestion{
		CaseQuestionID: questionID,
		CaseVersionID:  caseVersion.CaseVersionID,
		QuestionType:   model.CaseQuestionTypeOpenEnded,
		QuestionText:   questionText,
		Explanation:    "",
		ScoringWeight:  req.ScoringWeight,
		IsRequired:     req.IsRequired,
		SortOrder:      req.SortOrder,
	}

	openEndedDetail := &entity.CaseQuestionOpenEndedDetail{
		CaseQuestionID:    question.CaseQuestionID,
		ExpectedKeyPoints: expectedKeyPoints,
		MinimumKeywords:   minimumKeywordsJSON,
		EvaluationRubric:  evaluationRubric,
		MaxScore:          req.MaxScore,
	}

	err = s.caseQuestionRepo.CreateOpenEndedQuestion(tx, question, openEndedDetail, evidenceReferences)
	if err != nil {
		return nil, appErrors.InternalServer("failed to create open ended question")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	return &model.AdminCreateOpenEndedQuestionResponse{
		Question: mapOpenEndedQuestionResponse(question, openEndedDetail, minimumKeywords, evidenceReferences),
	}, nil
}

func (s *CaseService) CreateConfidenceSliderQuestionByAdmin(
	adminUserID uuid.UUID,
	caseID uuid.UUID,
	caseVersionID uuid.UUID,
	req model.AdminCreateConfidenceSliderQuestionRequest,
) (*model.AdminCreateConfidenceSliderQuestionResponse, error) {
	if adminUserID == uuid.Nil {
		return nil, appErrors.Unauthorized("unauthorized")
	}
	if caseID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid case id")
	}
	if caseVersionID == uuid.Nil {
		return nil, appErrors.BadRequest("invalid case version id")
	}

	questionText, err := helper.RequireTrimmedString(req.QuestionText, "question text is required")
	if err != nil {
		return nil, err
	}

	labelLow, err := helper.RequireTrimmedString(req.LabelLow, "label low is required")
	if err != nil {
		return nil, err
	}

	labelHigh, err := helper.RequireTrimmedString(req.LabelHigh, "label high is required")
	if err != nil {
		return nil, err
	}

	if len(labelLow) > 150 {
		return nil, appErrors.BadRequest("label low is too long")
	}

	if len(labelHigh) > 150 {
		return nil, appErrors.BadRequest("label high is too long")
	}

	if req.ScoringWeight < 0 || req.ScoringWeight > 100 {
		return nil, appErrors.BadRequest("scoring weight must be between 0 and 100")
	}

	if len(req.RelatedEvidenceIDs) == 0 {
		return nil, appErrors.BadRequest("related evidence ids are required")
	}

	if req.MinValue >= req.MaxValue {
		return nil, appErrors.BadRequest("min value must be less than max value")
	}

	if req.SnapInterval < 1 {
		return nil, appErrors.BadRequest("snap interval must be greater than 0")
	}

	if req.DefaultValue < req.MinValue || req.DefaultValue > req.MaxValue {
		return nil, appErrors.BadRequest("default value must be between min value and max value")
	}

	if (req.DefaultValue-req.MinValue)%req.SnapInterval != 0 {
		return nil, appErrors.BadRequest("default value must align with snap interval")
	}

	questionID := uuid.New()
	evidenceReferences := make([]entity.CaseQuestionEvidenceReference, 0, len(req.RelatedEvidenceIDs))
	seenEvidenceIDs := map[uuid.UUID]bool{}

	for i, evidenceID := range req.RelatedEvidenceIDs {
		if evidenceID == uuid.Nil {
			return nil, appErrors.BadRequest("invalid related evidence id")
		}
		if seenEvidenceIDs[evidenceID] {
			return nil, appErrors.BadRequest("related evidence ids cannot contain duplicates")
		}

		seenEvidenceIDs[evidenceID] = true
		evidenceReferences = append(evidenceReferences, entity.CaseQuestionEvidenceReference{
			CaseQuestionEvidenceReferenceID: uuid.New(),
			CaseQuestionID:                  questionID,
			CaseEvidenceID:                  evidenceID,
			SortOrder:                       i + 1,
		})
	}

	tx := s.db.Begin()
	defer tx.Rollback()

	caseVersion, err := s.caseVersionRepo.GetCaseVersionForUpdate(tx, caseVersionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("case version not found")
		}
		return nil, appErrors.InternalServer("failed to get case version")
	}

	if caseVersion.CaseID != caseID {
		return nil, appErrors.NotFound("case version not found")
	}

	if caseVersion.Status != model.CaseStatusDraft {
		return nil, appErrors.Conflict("case version is not editable")
	}

	for evidenceID := range seenEvidenceIDs {
		_, err := s.caseEvidenceRepo.GetCaseEvidence(tx, model.GetCaseEvidenceParam{
			CaseEvidenceID: evidenceID,
			CaseVersionID:  caseVersionID,
		})
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, appErrors.BadRequest("related evidence must belong to case version")
			}
			return nil, appErrors.InternalServer("failed to validate related evidence")
		}
	}

	question := &entity.CaseQuestion{
		CaseQuestionID: questionID,
		CaseVersionID:  caseVersion.CaseVersionID,
		QuestionType:   model.CaseQuestionTypeConfidenceSlider,
		QuestionText:   questionText,
		Explanation:    "",
		ScoringWeight:  req.ScoringWeight,
		IsRequired:     req.IsRequired,
		SortOrder:      req.SortOrder,
	}

	confidenceSliderDetail := &entity.CaseQuestionConfidenceSliderDetail{
		CaseQuestionID:           question.CaseQuestionID,
		MinValue:                 req.MinValue,
		MaxValue:                 req.MaxValue,
		SnapInterval:             req.SnapInterval,
		DefaultValue:             req.DefaultValue,
		LabelLow:                 labelLow,
		LabelHigh:                labelHigh,
		ShowWarningOnLargeChange: req.ShowWarningOnLargeChange,
	}

	err = s.caseQuestionRepo.CreateConfidenceSliderQuestion(tx, question, confidenceSliderDetail, evidenceReferences)
	if err != nil {
		return nil, appErrors.InternalServer("failed to create confidence slider question")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	return &model.AdminCreateConfidenceSliderQuestionResponse{
		Question: mapConfidenceSliderQuestionResponse(question, confidenceSliderDetail, evidenceReferences),
	}, nil
}

func mapMCQQuestionResponse(
	question *entity.CaseQuestion,
	options []entity.CaseQuestionMCQOption,
	evidenceReferences []entity.CaseQuestionEvidenceReference,
) model.AdminMCQQuestionResponse {
	optionResponses := make([]model.AdminMCQOptionResponse, 0, len(options))
	for _, option := range options {
		optionResponses = append(optionResponses, model.AdminMCQOptionResponse{
			CaseQuestionMCQOptionID: option.CaseQuestionMCQOptionID,
			CaseQuestionID:          option.CaseQuestionID,
			OptionCode:              option.OptionCode,
			OptionText:              option.OptionText,
			IsCorrect:               option.IsCorrect,
			SortOrder:               option.SortOrder,
			CreatedAt:               option.CreatedAt,
			UpdatedAt:               option.UpdatedAt,
		})
	}

	evidenceReferenceResponses := make([]model.AdminQuestionEvidenceReferenceResponse, 0, len(evidenceReferences))
	for _, reference := range evidenceReferences {
		evidenceReferenceResponses = append(evidenceReferenceResponses, model.AdminQuestionEvidenceReferenceResponse{
			CaseQuestionEvidenceReferenceID: reference.CaseQuestionEvidenceReferenceID,
			CaseQuestionID:                  reference.CaseQuestionID,
			CaseEvidenceID:                  reference.CaseEvidenceID,
			SortOrder:                       reference.SortOrder,
			CreatedAt:                       reference.CreatedAt,
			UpdatedAt:                       reference.UpdatedAt,
		})
	}

	return model.AdminMCQQuestionResponse{
		CaseQuestionID:     question.CaseQuestionID,
		CaseVersionID:      question.CaseVersionID,
		QuestionType:       question.QuestionType,
		QuestionText:       question.QuestionText,
		Explanation:        question.Explanation,
		ScoringWeight:      question.ScoringWeight,
		IsRequired:         question.IsRequired,
		SortOrder:          question.SortOrder,
		Options:            optionResponses,
		EvidenceReferences: evidenceReferenceResponses,
		CreatedAt:          question.CreatedAt,
		UpdatedAt:          question.UpdatedAt,
	}
}

func mapOpenEndedQuestionResponse(
	question *entity.CaseQuestion,
	openEndedDetail *entity.CaseQuestionOpenEndedDetail,
	minimumKeywords []string,
	evidenceReferences []entity.CaseQuestionEvidenceReference,
) model.AdminOpenEndedQuestionResponse {
	evidenceReferenceResponses := make([]model.AdminQuestionEvidenceReferenceResponse, 0, len(evidenceReferences))
	for _, reference := range evidenceReferences {
		evidenceReferenceResponses = append(evidenceReferenceResponses, model.AdminQuestionEvidenceReferenceResponse{
			CaseQuestionEvidenceReferenceID: reference.CaseQuestionEvidenceReferenceID,
			CaseQuestionID:                  reference.CaseQuestionID,
			CaseEvidenceID:                  reference.CaseEvidenceID,
			SortOrder:                       reference.SortOrder,
			CreatedAt:                       reference.CreatedAt,
			UpdatedAt:                       reference.UpdatedAt,
		})
	}

	return model.AdminOpenEndedQuestionResponse{
		CaseQuestionID:     question.CaseQuestionID,
		CaseVersionID:      question.CaseVersionID,
		QuestionType:       question.QuestionType,
		QuestionText:       question.QuestionText,
		ScoringWeight:      question.ScoringWeight,
		IsRequired:         question.IsRequired,
		SortOrder:          question.SortOrder,
		ExpectedKeyPoints:  openEndedDetail.ExpectedKeyPoints,
		MinimumKeywords:    minimumKeywords,
		EvaluationRubric:   openEndedDetail.EvaluationRubric,
		MaxScore:           openEndedDetail.MaxScore,
		EvidenceReferences: evidenceReferenceResponses,
		CreatedAt:          question.CreatedAt,
		UpdatedAt:          question.UpdatedAt,
	}
}

func mapConfidenceSliderQuestionResponse(
	question *entity.CaseQuestion,
	confidenceSliderDetail *entity.CaseQuestionConfidenceSliderDetail,
	evidenceReferences []entity.CaseQuestionEvidenceReference,
) model.AdminConfidenceSliderQuestionResponse {
	evidenceReferenceResponses := make([]model.AdminQuestionEvidenceReferenceResponse, 0, len(evidenceReferences))
	for _, reference := range evidenceReferences {
		evidenceReferenceResponses = append(evidenceReferenceResponses, model.AdminQuestionEvidenceReferenceResponse{
			CaseQuestionEvidenceReferenceID: reference.CaseQuestionEvidenceReferenceID,
			CaseQuestionID:                  reference.CaseQuestionID,
			CaseEvidenceID:                  reference.CaseEvidenceID,
			SortOrder:                       reference.SortOrder,
			CreatedAt:                       reference.CreatedAt,
			UpdatedAt:                       reference.UpdatedAt,
		})
	}

	return model.AdminConfidenceSliderQuestionResponse{
		CaseQuestionID:           question.CaseQuestionID,
		CaseVersionID:            question.CaseVersionID,
		QuestionType:             question.QuestionType,
		QuestionText:             question.QuestionText,
		ScoringWeight:            question.ScoringWeight,
		IsRequired:               question.IsRequired,
		SortOrder:                question.SortOrder,
		MinValue:                 confidenceSliderDetail.MinValue,
		MaxValue:                 confidenceSliderDetail.MaxValue,
		SnapInterval:             confidenceSliderDetail.SnapInterval,
		DefaultValue:             confidenceSliderDetail.DefaultValue,
		LabelLow:                 confidenceSliderDetail.LabelLow,
		LabelHigh:                confidenceSliderDetail.LabelHigh,
		ShowWarningOnLargeChange: confidenceSliderDetail.ShowWarningOnLargeChange,
		EvidenceReferences:       evidenceReferenceResponses,
		CreatedAt:                question.CreatedAt,
		UpdatedAt:                question.UpdatedAt,
	}
}

func normalizeQuestionKeywords(keywords []string) ([]string, string, error) {
	normalizedKeywords := make([]string, 0, len(keywords))
	seen := map[string]bool{}

	for _, keyword := range keywords {
		normalizedKeyword := strings.TrimSpace(keyword)
		if normalizedKeyword == "" {
			continue
		}

		key := strings.ToLower(normalizedKeyword)
		if seen[key] {
			continue
		}

		seen[key] = true
		normalizedKeywords = append(normalizedKeywords, normalizedKeyword)
	}

	if len(normalizedKeywords) == 0 {
		return nil, "", appErrors.BadRequest("minimum keywords are required")
	}

	payload, err := json.Marshal(normalizedKeywords)
	if err != nil {
		return nil, "", appErrors.InternalServer("failed to normalize minimum keywords")
	}

	return normalizedKeywords, string(payload), nil
}

func formatQuestionCode(index int) string {
	return fmt.Sprintf("Q-%02d", index)
}

func parseQuestionKeywords(raw string) ([]string, error) {
	var keywords []string
	if err := json.Unmarshal([]byte(raw), &keywords); err != nil {
		return nil, appErrors.InternalServer("failed to parse minimum keywords")
	}

	return keywords, nil
}
