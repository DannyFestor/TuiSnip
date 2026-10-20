package snippetslab

import (
	"errors"
	"fmt"
	"time"
)

type dates struct {
	created  time.Time
	modified time.Time
}

func parseDates(created, modified string) (dates, error) {
	createdAt, createdErr := parseDate(created)
	modifiedAt, modifiedErr := parseDate(modified)

	err := errors.Join(createdErr, modifiedErr)
	if err != nil {
		return dates{}, fmt.Errorf("parse dates: %w", err)
	}

	return dates{created: createdAt, modified: modifiedAt}, nil
}

func parseDeletedDate(deleted string) (time.Time, error) {
	if deleted == "" {
		return time.Time{}, nil
	}

	deletedAt, err := parseDate(deleted)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse deleted date: %w", err)
	}

	return deletedAt, nil
}

func parseDate(text string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339, text)

	return parsed.UTC(), err
}
