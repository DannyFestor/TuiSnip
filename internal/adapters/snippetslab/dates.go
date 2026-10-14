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
	createdAt, createdErr := time.Parse(time.RFC3339, created)
	modifiedAt, modifiedErr := time.Parse(time.RFC3339, modified)

	err := errors.Join(createdErr, modifiedErr)
	if err != nil {
		return dates{}, fmt.Errorf("parse dates: %w", err)
	}

	return dates{created: createdAt.UTC(), modified: modifiedAt.UTC()}, nil
}
