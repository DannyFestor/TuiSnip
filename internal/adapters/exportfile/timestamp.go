package exportfile

import (
	"errors"
	"fmt"
	"time"
)

const timestampLayout = "2006-01-02T15:04:05.000000000Z07:00"

type dates struct {
	created time.Time
	updated time.Time
}

func timestampOf(at time.Time) string {
	return at.UTC().Format(timestampLayout)
}

func parseDates(createdAt, updatedAt string) (dates, error) {
	created, createdErr := parseTimestamp(createdAt)
	updated, updatedErr := parseTimestamp(updatedAt)

	err := errors.Join(createdErr, updatedErr)
	if err != nil {
		return dates{}, err
	}

	return dates{created: created, updated: updated}, nil
}

func parseTimestamp(text string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339Nano, text)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse date: %w", err)
	}

	return parsed.UTC(), nil
}
