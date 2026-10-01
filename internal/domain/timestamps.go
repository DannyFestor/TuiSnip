package domain

import (
	"errors"
	"math"
	"time"
)

// SQLite stores timestamps as positive Unix nanoseconds, so a time outside that range
// would fail its CHECK or overflow the column.
func requireTimestamps(createdAt, updatedAt time.Time) error {
	return errors.Join(
		requireStorable(createdAt),
		requireStorable(updatedAt),
		requireOrder(createdAt, updatedAt),
	)
}

func requireStorable(at time.Time) error {
	latest := time.Unix(0, math.MaxInt64)
	if !at.After(time.Unix(0, 0)) || at.After(latest) {
		return ErrTimestampOutOfRange
	}

	return nil
}

func requireOrder(createdAt, updatedAt time.Time) error {
	if updatedAt.Before(createdAt) {
		return ErrUpdatedBeforeCreate
	}

	return nil
}
