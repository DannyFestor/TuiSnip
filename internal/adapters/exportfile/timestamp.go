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
	created, createdErr := time.Parse(time.RFC3339Nano, createdAt)
	updated, updatedErr := time.Parse(time.RFC3339Nano, updatedAt)

	err := errors.Join(createdErr, updatedErr)
	if err != nil {
		return dates{}, fmt.Errorf("parse dates: %w", err)
	}

	return dates{created: created.UTC(), updated: updated.UTC()}, nil
}
