package service

import (
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/azmiagr/unesco-hackathon/entity"
	"github.com/azmiagr/unesco-hackathon/model"
	appErrors "github.com/azmiagr/unesco-hackathon/pkg/errors"
	"github.com/azmiagr/unesco-hackathon/pkg/supabase"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *CaseService) DeleteCaseEvidenceByAdmin(
	adminUserID uuid.UUID,
	caseID uuid.UUID,
	caseVersionID uuid.UUID,
	caseEvidenceID uuid.UUID,
) (*model.AdminDeleteCaseEvidenceResponse, error) {
	if err := validateAdminEvidenceIDs(adminUserID, caseID, caseVersionID, caseEvidenceID); err != nil {
		return nil, err
	}

	tx := s.db.Begin()
	defer tx.Rollback()

	evidence, err := s.getEditableEvidenceByAdmin(tx, caseID, caseVersionID, caseEvidenceID, "")
	if err != nil {
		return nil, err
	}

	var imageURL *string
	if evidence.SocialPost != nil {
		imageURL = evidence.SocialPost.ImageURL
	}
	if evidence.Article != nil {
		imageURL = evidence.Article.ImageURL
	}

	before := newAuditCaseEvidenceSnapshot(evidence)

	err = s.caseEvidenceRepo.DeleteCaseEvidence(tx, evidence.CaseEvidenceID)
	if err != nil {
		return nil, appErrors.InternalServer("failed to delete case evidence")
	}

	err = s.writeCaseEvidenceAuditLog(tx, adminUserID, model.AuditActionDelete, evidence, before)
	if err != nil {
		return nil, err
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, appErrors.InternalServer("failed to commit transaction")
	}

	if imageURL != nil {
		_ = supabase.DeleteFileIfPresent(s.storage, *imageURL)
	}

	return &model.AdminDeleteCaseEvidenceResponse{
		CaseEvidenceID: evidence.CaseEvidenceID,
	}, nil
}

func validateAdminEvidenceIDs(adminUserID uuid.UUID, caseID uuid.UUID, caseVersionID uuid.UUID, caseEvidenceID uuid.UUID) error {
	if adminUserID == uuid.Nil {
		return appErrors.Unauthorized("unauthorized")
	}

	if caseID == uuid.Nil {
		return appErrors.BadRequest("invalid case id")
	}

	if caseVersionID == uuid.Nil {
		return appErrors.BadRequest("invalid case version id")
	}

	if caseEvidenceID == uuid.Nil {
		return appErrors.BadRequest("invalid case evidence id")
	}

	return nil
}

func (s *CaseService) getEditableEvidenceByAdmin(
	tx *gorm.DB,
	caseID uuid.UUID,
	caseVersionID uuid.UUID,
	caseEvidenceID uuid.UUID,
	templateType string,
) (*entity.CaseEvidence, error) {
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

	evidence, err := s.caseEvidenceRepo.GetCaseEvidence(tx, model.GetCaseEvidenceParam{
		CaseEvidenceID: caseEvidenceID,
		CaseVersionID:  caseVersionID,
		TemplateType:   templateType,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("case evidence not found")
		}
		return nil, appErrors.InternalServer("failed to get case evidence")
	}

	return evidence, nil
}

func parseEvidencePublishDate(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, appErrors.BadRequest("publish date is required")
	}

	layouts := []string{
		"2006-01-02",
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
	}

	for _, layout := range layouts {
		value, err := time.Parse(layout, raw)
		if err == nil {
			return value.UTC(), nil
		}
	}

	return time.Time{}, appErrors.BadRequest("invalid publish date")
}

func normalizeOptionalURL(raw string) (*string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	parsedURL, err := url.ParseRequestURI(raw)
	if err != nil {
		return nil, appErrors.BadRequest("invalid url")
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, appErrors.BadRequest("url must use http or https")
	}

	return &raw, nil
}

func normalizeCredibilityTagItems(tags []string) ([]string, string, error) {
	normalizedTags := make([]string, 0, len(tags))
	seen := map[string]bool{}

	for _, tag := range tags {
		normalizedTag := strings.ToLower(strings.TrimSpace(tag))
		if normalizedTag == "" {
			continue
		}

		if seen[normalizedTag] {
			continue
		}

		seen[normalizedTag] = true
		normalizedTags = append(normalizedTags, normalizedTag)
	}

	if len(normalizedTags) == 0 {
		return nil, "", appErrors.BadRequest("credibility tags are required")
	}

	payload, err := json.Marshal(normalizedTags)
	if err != nil {
		return nil, "", appErrors.InternalServer("failed to normalize credibility tags")
	}

	return normalizedTags, string(payload), nil
}

func normalizeCredibilityTags(raw string) ([]string, string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, "", appErrors.BadRequest("credibility tags are required")
	}

	var tags []string
	if err := json.Unmarshal([]byte(raw), &tags); err != nil {
		return nil, "", appErrors.BadRequest("credibility tags must be a valid json array")
	}

	normalizedTags := make([]string, 0, len(tags))
	seen := map[string]bool{}

	for _, tag := range tags {
		normalizedTag := strings.ToLower(strings.TrimSpace(tag))
		if normalizedTag == "" {
			continue
		}

		if seen[normalizedTag] {
			continue
		}

		seen[normalizedTag] = true
		normalizedTags = append(normalizedTags, normalizedTag)
	}

	if len(normalizedTags) == 0 {
		return nil, "", appErrors.BadRequest("credibility tags are required")
	}

	payload, err := json.Marshal(normalizedTags)
	if err != nil {
		return nil, "", appErrors.InternalServer("failed to normalize credibility tags")
	}

	return normalizedTags, string(payload), nil
}

func parseEvidenceTimestamp(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, appErrors.BadRequest("timestamp is required")
	}

	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
	}

	for _, layout := range layouts {
		value, err := time.Parse(layout, raw)
		if err == nil {
			return value.UTC(), nil
		}
	}

	return time.Time{}, appErrors.BadRequest("invalid timestamp")
}
