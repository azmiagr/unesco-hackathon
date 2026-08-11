package service

import (
	appErrors "github.com/azmiagr/unesco-hackathon/pkg/errors"
	"github.com/azmiagr/unesco-hackathon/pkg/supabase"
	"mime/multipart"
)

func (s *CaseService) uploadCaseThumbnail(file *multipart.FileHeader) (*string, error) {
	return s.uploadOptionalImage(
		file,
		maxCaseThumbnailSize,
		"thumbnail size exceeds 5MB limit",
		"failed to upload thumbnail",
	)
}

func (s *CaseService) uploadEvidenceImage(file *multipart.FileHeader) (*string, error) {
	return s.uploadOptionalImage(
		file,
		maxEvidenceImageSize,
		"evidence image size exceeds 5MB limit",
		"failed to upload evidence image",
	)
}

func (s *CaseService) uploadOptionalImage(file *multipart.FileHeader, maxSize int64, sizeErrMessage string, uploadErrMessage string) (*string, error) {
	if file == nil {
		return nil, nil
	}

	url, err := supabase.UploadOptionalImage(
		s.storage,
		file,
		maxSize,
		sizeErrMessage,
	)
	if err != nil {
		return nil, appErrors.BadRequest(uploadErrMessage)
	}

	return &url, nil
}
