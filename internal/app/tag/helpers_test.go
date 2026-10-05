package tag_test

import (
	"errors"
	"testing"
	"time"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
)

var errDatabaseLocked = errors.New("database is locked")

func fixedClock() testkit.FixedClock {
	return testkit.NewFixedClock(now())
}

func now() time.Time {
	return time.Date(2026, time.March, 1, 12, 0, 0, 0, time.UTC)
}

func storedTag(t *testing.T) domain.Tag {
	t.Helper()

	return testkit.Tag(t, testkit.TagSpec{
		ID:        testkit.NewSequentialIDs().NewTagID(),
		Name:      "go",
		CreatedAt: now().Add(-time.Hour),
	})
}
