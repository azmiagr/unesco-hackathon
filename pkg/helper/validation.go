package helper

import (
	"strings"

	"github.com/azmiagr/unesco-hackathon/pkg/errors"
)

func RequireTrimmedString(value string, message string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", errors.BadRequest(message)
	}

	return trimmed, nil
}
