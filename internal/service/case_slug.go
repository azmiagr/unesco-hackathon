package service

import (
	"errors"
	"fmt"
	"github.com/azmiagr/unesco-hackathon/model"
	appErrors "github.com/azmiagr/unesco-hackathon/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"regexp"
	"strings"
)

func (s *CaseService) buildUniqueSlug(caseID uuid.UUID, title string) (string, error) {
	baseSlug := slugify(title)
	if baseSlug == "" {
		baseSlug = "case"
	}

	slug := baseSlug
	exists, err := s.caseRepo.CaseExists(s.db, model.GetCaseParam{
		Slug: slug,
	})
	if err != nil {
		return "", appErrors.InternalServer("failed to check case slug")
	}

	if exists {
		slug = fmt.Sprintf("%s-%s", baseSlug, caseID.String()[:8])
	}

	return slug, nil
}

func (s *CaseService) buildUniqueSlugForUpdate(caseID uuid.UUID, title string) (string, error) {
	baseSlug := slugify(title)
	if baseSlug == "" {
		baseSlug = "case"
	}

	existingCase, err := s.caseRepo.GetCase(s.db, model.GetCaseParam{
		Slug: baseSlug,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return baseSlug, nil
		}
		return "", appErrors.InternalServer("failed to check case slug")
	}

	if existingCase.CaseID == caseID {
		return baseSlug, nil
	}

	return fmt.Sprintf("%s-%s", baseSlug, caseID.String()[:8]), nil
}

func slugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))

	nonAlphaNumeric := regexp.MustCompile(`[^a-z0-9]+`)
	value = nonAlphaNumeric.ReplaceAllString(value, "-")

	multipleDashes := regexp.MustCompile(`-+`)
	value = multipleDashes.ReplaceAllString(value, "-")

	return strings.Trim(value, "-")
}
