package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/azmiagr/unesco-hackathon/entity"
	"github.com/azmiagr/unesco-hackathon/model"
	appErrors "github.com/azmiagr/unesco-hackathon/pkg/errors"
	"github.com/azmiagr/unesco-hackathon/pkg/helper"
	"github.com/google/uuid"
	"gorm.io/gorm"
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
	case model.CaseQuestionTypeClaimClassification:
		if question.ClaimClassificationDetail == nil {
			return nil, appErrors.InternalServer("claim classification question detail not found")
		}

		taxonomyTags, err := parseQuestionStringItems(question.ClaimClassificationDetail.TaxonomyTags, "taxonomy tags")
		if err != nil {
			return nil, err
		}

		claimClassification := mapClaimClassificationQuestionResponse(
			question,
			question.ClaimClassificationDetail,
			taxonomyTags,
			question.EvidenceReferences,
		)
		result.ClaimClassification = &claimClassification
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

	err = s.writeCaseQuestionAuditLog(tx, adminUserID, model.AuditActionCreate, question, nil)
	if err != nil {
		return nil, err
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

	err = s.writeCaseQuestionAuditLog(tx, adminUserID, model.AuditActionCreate, question, nil)
	if err != nil {
		return nil, err
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

	err = s.writeCaseQuestionAuditLog(tx, adminUserID, model.AuditActionCreate, question, nil)
	if err != nil {
		return nil, err
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	return &model.AdminCreateConfidenceSliderQuestionResponse{
		Question: mapConfidenceSliderQuestionResponse(question, confidenceSliderDetail, evidenceReferences),
	}, nil
}

func (s *CaseService) CreateClaimClassificationQuestionByAdmin(
	adminUserID uuid.UUID,
	caseID uuid.UUID,
	caseVersionID uuid.UUID,
	req model.AdminCreateClaimClassificationQuestionRequest,
) (*model.AdminCreateClaimClassificationQuestionResponse, error) {
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

	correctAnswerInput, err := helper.RequireTrimmedString(req.CorrectAnswer, "correct answer is required")
	if err != nil {
		return nil, err
	}

	explanation := strings.TrimSpace(req.Explanation)

	if req.ScoringWeight < 0 || req.ScoringWeight > 100 {
		return nil, appErrors.BadRequest("scoring weight must be between 0 and 100")
	}

	if len(req.RelatedEvidenceIDs) == 0 {
		return nil, appErrors.BadRequest("related evidence ids are required")
	}

	taxonomyTags, taxonomyTagsJSON, err := normalizeQuestionStringItems(req.TaxonomyTags, "taxonomy tags are required")
	if err != nil {
		return nil, err
	}

	if len(taxonomyTags) < 2 {
		return nil, appErrors.BadRequest("taxonomy tags must contain at least 2 items")
	}

	correctAnswer, ok := findMatchingQuestionItem(correctAnswerInput, taxonomyTags)
	if !ok {
		return nil, appErrors.BadRequest("correct answer must be one of taxonomy tags")
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
		QuestionType:   model.CaseQuestionTypeClaimClassification,
		QuestionText:   questionText,
		Explanation:    explanation,
		ScoringWeight:  req.ScoringWeight,
		IsRequired:     req.IsRequired,
		SortOrder:      req.SortOrder,
	}

	claimClassificationDetail := &entity.CaseQuestionClaimClassificationDetail{
		CaseQuestionID: question.CaseQuestionID,
		TaxonomyTags:   taxonomyTagsJSON,
		CorrectAnswer:  correctAnswer,
	}

	err = s.caseQuestionRepo.CreateClaimClassificationQuestion(tx, question, claimClassificationDetail, evidenceReferences)
	if err != nil {
		return nil, appErrors.InternalServer("failed to create claim classification question")
	}

	err = s.writeCaseQuestionAuditLog(tx, adminUserID, model.AuditActionCreate, question, nil)
	if err != nil {
		return nil, err
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	return &model.AdminCreateClaimClassificationQuestionResponse{
		Question: mapClaimClassificationQuestionResponse(
			question,
			claimClassificationDetail,
			taxonomyTags,
			evidenceReferences,
		),
	}, nil
}

func (s *CaseService) UpdateMCQQuestionByAdmin(
	adminUserID uuid.UUID,
	caseID uuid.UUID,
	caseVersionID uuid.UUID,
	caseQuestionID uuid.UUID,
	req model.AdminUpdateMCQQuestionRequest,
) (*model.AdminUpdateMCQQuestionResponse, error) {
	if err := validateAdminQuestionIDs(adminUserID, caseID, caseVersionID, caseQuestionID); err != nil {
		return nil, err
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

	if len(req.Options) < 2 {
		return nil, appErrors.BadRequest("mcq options must contain at least 2 items")
	}

	evidenceReferences, seenEvidenceIDs, err := buildQuestionEvidenceReferences(caseQuestionID, req.RelatedEvidenceIDs)
	if err != nil {
		return nil, err
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
			CaseQuestionID:          caseQuestionID,
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

	question, err := s.getEditableQuestionByAdmin(tx, caseID, caseVersionID, caseQuestionID, model.CaseQuestionTypeMCQ)
	if err != nil {
		return nil, err
	}

	before := newAuditCaseQuestionSnapshot(question)

	if err := s.validateQuestionEvidenceReferences(tx, caseVersionID, seenEvidenceIDs); err != nil {
		return nil, err
	}

	question.QuestionText = questionText
	question.Explanation = explanation
	question.ScoringWeight = req.ScoringWeight
	question.IsRequired = req.IsRequired
	question.SortOrder = req.SortOrder

	if err := s.caseQuestionRepo.UpdateCaseQuestion(tx, question); err != nil {
		return nil, appErrors.InternalServer("failed to update mcq question")
	}
	if err := s.caseQuestionRepo.ReplaceMCQOptions(tx, question.CaseQuestionID, options); err != nil {
		return nil, appErrors.InternalServer("failed to replace mcq options")
	}
	if err := s.caseQuestionRepo.ReplaceEvidenceReferences(tx, question.CaseQuestionID, evidenceReferences); err != nil {
		return nil, appErrors.InternalServer("failed to replace question evidence references")
	}
	if err := s.writeCaseQuestionAuditLog(tx, adminUserID, model.AuditActionUpdate, question, before); err != nil {
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	return &model.AdminUpdateMCQQuestionResponse{
		Question: mapMCQQuestionResponse(question, options, evidenceReferences),
	}, nil
}

func (s *CaseService) UpdateOpenEndedQuestionByAdmin(
	adminUserID uuid.UUID,
	caseID uuid.UUID,
	caseVersionID uuid.UUID,
	caseQuestionID uuid.UUID,
	req model.AdminUpdateOpenEndedQuestionRequest,
) (*model.AdminUpdateOpenEndedQuestionResponse, error) {
	if err := validateAdminQuestionIDs(adminUserID, caseID, caseVersionID, caseQuestionID); err != nil {
		return nil, err
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

	minimumKeywords, minimumKeywordsJSON, err := normalizeQuestionKeywords(req.MinimumKeywords)
	if err != nil {
		return nil, err
	}
	evidenceReferences, seenEvidenceIDs, err := buildQuestionEvidenceReferences(caseQuestionID, req.RelatedEvidenceIDs)
	if err != nil {
		return nil, err
	}

	tx := s.db.Begin()
	defer tx.Rollback()

	question, err := s.getEditableQuestionByAdmin(tx, caseID, caseVersionID, caseQuestionID, model.CaseQuestionTypeOpenEnded)
	if err != nil {
		return nil, err
	}
	if question.OpenEndedDetail == nil {
		return nil, appErrors.InternalServer("open ended question detail not found")
	}
	before := newAuditCaseQuestionSnapshot(question)
	if err := s.validateQuestionEvidenceReferences(tx, caseVersionID, seenEvidenceIDs); err != nil {
		return nil, err
	}

	question.QuestionText = questionText
	question.Explanation = ""
	question.ScoringWeight = req.ScoringWeight
	question.IsRequired = req.IsRequired
	question.SortOrder = req.SortOrder
	question.OpenEndedDetail.ExpectedKeyPoints = expectedKeyPoints
	question.OpenEndedDetail.MinimumKeywords = minimumKeywordsJSON
	question.OpenEndedDetail.EvaluationRubric = evaluationRubric
	question.OpenEndedDetail.MaxScore = req.MaxScore

	if err := s.caseQuestionRepo.UpdateCaseQuestion(tx, question); err != nil {
		return nil, appErrors.InternalServer("failed to update open ended question")
	}
	if err := s.caseQuestionRepo.UpdateOpenEndedQuestion(tx, question.OpenEndedDetail); err != nil {
		return nil, appErrors.InternalServer("failed to update open ended question detail")
	}
	if err := s.caseQuestionRepo.ReplaceEvidenceReferences(tx, question.CaseQuestionID, evidenceReferences); err != nil {
		return nil, appErrors.InternalServer("failed to replace question evidence references")
	}
	if err := s.writeCaseQuestionAuditLog(tx, adminUserID, model.AuditActionUpdate, question, before); err != nil {
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	return &model.AdminUpdateOpenEndedQuestionResponse{
		Question: mapOpenEndedQuestionResponse(question, question.OpenEndedDetail, minimumKeywords, evidenceReferences),
	}, nil
}

func (s *CaseService) UpdateConfidenceSliderQuestionByAdmin(
	adminUserID uuid.UUID,
	caseID uuid.UUID,
	caseVersionID uuid.UUID,
	caseQuestionID uuid.UUID,
	req model.AdminUpdateConfidenceSliderQuestionRequest,
) (*model.AdminUpdateConfidenceSliderQuestionResponse, error) {
	if err := validateAdminQuestionIDs(adminUserID, caseID, caseVersionID, caseQuestionID); err != nil {
		return nil, err
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

	evidenceReferences, seenEvidenceIDs, err := buildQuestionEvidenceReferences(caseQuestionID, req.RelatedEvidenceIDs)
	if err != nil {
		return nil, err
	}

	tx := s.db.Begin()
	defer tx.Rollback()

	question, err := s.getEditableQuestionByAdmin(tx, caseID, caseVersionID, caseQuestionID, model.CaseQuestionTypeConfidenceSlider)
	if err != nil {
		return nil, err
	}
	if question.ConfidenceSliderDetail == nil {
		return nil, appErrors.InternalServer("confidence slider question detail not found")
	}
	before := newAuditCaseQuestionSnapshot(question)

	err = s.validateQuestionEvidenceReferences(tx, caseVersionID, seenEvidenceIDs)
	if err != nil {
		return nil, err
	}

	question.QuestionText = questionText
	question.Explanation = ""
	question.ScoringWeight = req.ScoringWeight
	question.IsRequired = req.IsRequired
	question.SortOrder = req.SortOrder
	question.ConfidenceSliderDetail.MinValue = req.MinValue
	question.ConfidenceSliderDetail.MaxValue = req.MaxValue
	question.ConfidenceSliderDetail.SnapInterval = req.SnapInterval
	question.ConfidenceSliderDetail.DefaultValue = req.DefaultValue
	question.ConfidenceSliderDetail.LabelLow = labelLow
	question.ConfidenceSliderDetail.LabelHigh = labelHigh
	question.ConfidenceSliderDetail.ShowWarningOnLargeChange = req.ShowWarningOnLargeChange

	err = s.caseQuestionRepo.UpdateCaseQuestion(tx, question)
	if err != nil {
		return nil, appErrors.InternalServer("failed to update confidence slider question")
	}

	err = s.caseQuestionRepo.UpdateConfidenceSliderQuestion(tx, question.ConfidenceSliderDetail)
	if err != nil {
		return nil, appErrors.InternalServer("failed to update confidence slider question detail")
	}

	err = s.caseQuestionRepo.ReplaceEvidenceReferences(tx, question.CaseQuestionID, evidenceReferences)
	if err != nil {
		return nil, appErrors.InternalServer("failed to replace question evidence references")
	}

	err = s.writeCaseQuestionAuditLog(tx, adminUserID, model.AuditActionUpdate, question, before)
	if err != nil {
		return nil, err
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	return &model.AdminUpdateConfidenceSliderQuestionResponse{
		Question: mapConfidenceSliderQuestionResponse(question, question.ConfidenceSliderDetail, evidenceReferences),
	}, nil
}

func (s *CaseService) UpdateClaimClassificationQuestionByAdmin(
	adminUserID uuid.UUID,
	caseID uuid.UUID,
	caseVersionID uuid.UUID,
	caseQuestionID uuid.UUID,
	req model.AdminUpdateClaimClassificationQuestionRequest,
) (*model.AdminUpdateClaimClassificationQuestionResponse, error) {
	if err := validateAdminQuestionIDs(adminUserID, caseID, caseVersionID, caseQuestionID); err != nil {
		return nil, err
	}

	questionText, err := helper.RequireTrimmedString(req.QuestionText, "question text is required")
	if err != nil {
		return nil, err
	}
	correctAnswerInput, err := helper.RequireTrimmedString(req.CorrectAnswer, "correct answer is required")
	if err != nil {
		return nil, err
	}
	explanation := strings.TrimSpace(req.Explanation)
	if req.ScoringWeight < 0 || req.ScoringWeight > 100 {
		return nil, appErrors.BadRequest("scoring weight must be between 0 and 100")
	}

	taxonomyTags, taxonomyTagsJSON, err := normalizeQuestionStringItems(req.TaxonomyTags, "taxonomy tags are required")
	if err != nil {
		return nil, err
	}
	if len(taxonomyTags) < 2 {
		return nil, appErrors.BadRequest("taxonomy tags must contain at least 2 items")
	}
	correctAnswer, ok := findMatchingQuestionItem(correctAnswerInput, taxonomyTags)
	if !ok {
		return nil, appErrors.BadRequest("correct answer must be one of taxonomy tags")
	}
	evidenceReferences, seenEvidenceIDs, err := buildQuestionEvidenceReferences(caseQuestionID, req.RelatedEvidenceIDs)
	if err != nil {
		return nil, err
	}

	tx := s.db.Begin()
	defer tx.Rollback()

	question, err := s.getEditableQuestionByAdmin(tx, caseID, caseVersionID, caseQuestionID, model.CaseQuestionTypeClaimClassification)
	if err != nil {
		return nil, err
	}
	if question.ClaimClassificationDetail == nil {
		return nil, appErrors.InternalServer("claim classification question detail not found")
	}
	before := newAuditCaseQuestionSnapshot(question)
	if err := s.validateQuestionEvidenceReferences(tx, caseVersionID, seenEvidenceIDs); err != nil {
		return nil, err
	}

	question.QuestionText = questionText
	question.Explanation = explanation
	question.ScoringWeight = req.ScoringWeight
	question.IsRequired = req.IsRequired
	question.SortOrder = req.SortOrder
	question.ClaimClassificationDetail.TaxonomyTags = taxonomyTagsJSON
	question.ClaimClassificationDetail.CorrectAnswer = correctAnswer

	err = s.caseQuestionRepo.UpdateCaseQuestion(tx, question)
	if err != nil {
		return nil, appErrors.InternalServer("failed to update claim classification question")
	}

	err = s.caseQuestionRepo.UpdateClaimClassificationQuestion(tx, question.ClaimClassificationDetail)
	if err != nil {
		return nil, appErrors.InternalServer("failed to update claim classification question detail")
	}

	err = s.caseQuestionRepo.ReplaceEvidenceReferences(tx, question.CaseQuestionID, evidenceReferences)
	if err != nil {
		return nil, appErrors.InternalServer("failed to replace question evidence references")
	}

	err = s.writeCaseQuestionAuditLog(tx, adminUserID, model.AuditActionUpdate, question, before)
	if err != nil {
		return nil, err
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	return &model.AdminUpdateClaimClassificationQuestionResponse{
		Question: mapClaimClassificationQuestionResponse(
			question,
			question.ClaimClassificationDetail,
			taxonomyTags,
			evidenceReferences,
		),
	}, nil
}

func (s *CaseService) DeleteCaseQuestionByAdmin(
	adminUserID uuid.UUID,
	caseID uuid.UUID,
	caseVersionID uuid.UUID,
	caseQuestionID uuid.UUID,
) (*model.AdminDeleteCaseQuestionResponse, error) {
	if err := validateAdminQuestionIDs(adminUserID, caseID, caseVersionID, caseQuestionID); err != nil {
		return nil, err
	}

	tx := s.db.Begin()
	defer tx.Rollback()

	question, err := s.getEditableQuestionByAdmin(tx, caseID, caseVersionID, caseQuestionID, "")
	if err != nil {
		return nil, err
	}

	before := newAuditCaseQuestionSnapshot(question)

	err = s.caseQuestionRepo.DeleteCaseQuestion(tx, question.CaseQuestionID)
	if err != nil {
		return nil, appErrors.InternalServer("failed to delete case question")
	}

	err = s.writeCaseQuestionAuditLog(tx, adminUserID, model.AuditActionDelete, question, before)
	if err != nil {
		return nil, err
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	return &model.AdminDeleteCaseQuestionResponse{
		CaseQuestionID: question.CaseQuestionID,
	}, nil
}

func validateAdminQuestionIDs(adminUserID uuid.UUID, caseID uuid.UUID, caseVersionID uuid.UUID, caseQuestionID uuid.UUID) error {
	if adminUserID == uuid.Nil {
		return appErrors.Unauthorized("unauthorized")
	}
	if caseID == uuid.Nil {
		return appErrors.BadRequest("invalid case id")
	}
	if caseVersionID == uuid.Nil {
		return appErrors.BadRequest("invalid case version id")
	}
	if caseQuestionID == uuid.Nil {
		return appErrors.BadRequest("invalid case question id")
	}

	return nil
}

func (s *CaseService) getEditableQuestionByAdmin(
	tx *gorm.DB,
	caseID uuid.UUID,
	caseVersionID uuid.UUID,
	caseQuestionID uuid.UUID,
	questionType string,
) (*entity.CaseQuestion, error) {
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

	lockedQuestion, err := s.caseQuestionRepo.GetCaseQuestionForUpdate(tx, caseQuestionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("case question not found")
		}
		return nil, appErrors.InternalServer("failed to get case question")
	}
	if lockedQuestion.CaseVersionID != caseVersionID {
		return nil, appErrors.NotFound("case question not found")
	}
	if questionType != "" && lockedQuestion.QuestionType != questionType {
		return nil, appErrors.BadRequest("question type does not match endpoint")
	}

	question, err := s.caseQuestionRepo.GetCaseQuestion(tx, model.GetCaseQuestionParam{
		CaseQuestionID: caseQuestionID,
		CaseVersionID:  caseVersionID,
	})
	if err != nil {
		return nil, appErrors.InternalServer("failed to get case question")
	}

	return question, nil
}

func buildQuestionEvidenceReferences(
	caseQuestionID uuid.UUID,
	relatedEvidenceIDs []uuid.UUID,
) ([]entity.CaseQuestionEvidenceReference, map[uuid.UUID]bool, error) {
	if len(relatedEvidenceIDs) == 0 {
		return nil, nil, appErrors.BadRequest("related evidence ids are required")
	}

	evidenceReferences := make([]entity.CaseQuestionEvidenceReference, 0, len(relatedEvidenceIDs))
	seenEvidenceIDs := map[uuid.UUID]bool{}
	for i, evidenceID := range relatedEvidenceIDs {
		if evidenceID == uuid.Nil {
			return nil, nil, appErrors.BadRequest("invalid related evidence id")
		}
		if seenEvidenceIDs[evidenceID] {
			return nil, nil, appErrors.BadRequest("related evidence ids cannot contain duplicates")
		}

		seenEvidenceIDs[evidenceID] = true
		evidenceReferences = append(evidenceReferences, entity.CaseQuestionEvidenceReference{
			CaseQuestionEvidenceReferenceID: uuid.New(),
			CaseQuestionID:                  caseQuestionID,
			CaseEvidenceID:                  evidenceID,
			SortOrder:                       i + 1,
		})
	}

	return evidenceReferences, seenEvidenceIDs, nil
}

func (s *CaseService) validateQuestionEvidenceReferences(
	tx *gorm.DB,
	caseVersionID uuid.UUID,
	seenEvidenceIDs map[uuid.UUID]bool,
) error {
	for evidenceID := range seenEvidenceIDs {
		_, err := s.caseEvidenceRepo.GetCaseEvidence(tx, model.GetCaseEvidenceParam{
			CaseEvidenceID: evidenceID,
			CaseVersionID:  caseVersionID,
		})
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.BadRequest("related evidence must belong to case version")
			}
			return appErrors.InternalServer("failed to validate related evidence")
		}
	}

	return nil
}

func normalizeQuestionStringItems(items []string, requiredMessage string) ([]string, string, error) {
	normalizedItems := make([]string, 0, len(items))
	seen := map[string]bool{}

	for _, item := range items {
		normalizedItem := strings.TrimSpace(item)
		if normalizedItem == "" {
			continue
		}

		key := strings.ToLower(normalizedItem)
		if seen[key] {
			continue
		}

		seen[key] = true
		normalizedItems = append(normalizedItems, normalizedItem)
	}

	if len(normalizedItems) == 0 {
		return nil, "", appErrors.BadRequest(requiredMessage)
	}

	payload, err := json.Marshal(normalizedItems)
	if err != nil {
		return nil, "", appErrors.InternalServer("failed to normalize question items")
	}

	return normalizedItems, string(payload), nil
}

func parseQuestionStringItems(raw string, fieldName string) ([]string, error) {
	var items []string
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil, appErrors.InternalServer("failed to parse " + fieldName)
	}

	return items, nil
}

func findMatchingQuestionItem(value string, items []string) (string, bool) {
	normalizedValue := strings.ToLower(strings.TrimSpace(value))
	for _, item := range items {
		if strings.ToLower(strings.TrimSpace(item)) == normalizedValue {
			return item, true
		}
	}

	return "", false
}

func mapClaimClassificationQuestionResponse(
	question *entity.CaseQuestion,
	claimClassificationDetail *entity.CaseQuestionClaimClassificationDetail,
	taxonomyTags []string,
	evidenceReferences []entity.CaseQuestionEvidenceReference,
) model.AdminClaimClassificationQuestionResponse {
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

	return model.AdminClaimClassificationQuestionResponse{
		CaseQuestionID:     question.CaseQuestionID,
		CaseVersionID:      question.CaseVersionID,
		QuestionType:       question.QuestionType,
		QuestionText:       question.QuestionText,
		Explanation:        question.Explanation,
		ScoringWeight:      question.ScoringWeight,
		IsRequired:         question.IsRequired,
		SortOrder:          question.SortOrder,
		TaxonomyTags:       taxonomyTags,
		CorrectAnswer:      claimClassificationDetail.CorrectAnswer,
		EvidenceReferences: evidenceReferenceResponses,
		CreatedAt:          question.CreatedAt,
		UpdatedAt:          question.UpdatedAt,
	}
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
