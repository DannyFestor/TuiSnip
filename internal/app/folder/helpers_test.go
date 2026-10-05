package folder_test

import (
	"testing"
	"time"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
)

func fixedClock() testkit.FixedClock {
	return testkit.NewFixedClock(now())
}

func now() time.Time {
	return time.Date(2026, time.March, 1, 12, 0, 0, 0, time.UTC)
}

func storedFolder(t *testing.T) domain.Folder {
	t.Helper()

	return testkit.Folder(t, testkit.FolderSpec{
		ID:        testkit.NewSequentialIDs().NewFolderID(),
		Name:      "go",
		CreatedAt: now().Add(-time.Hour),
	})
}
