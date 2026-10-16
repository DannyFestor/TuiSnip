package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
)

func TestNewTrashMark(t *testing.T) {
	t.Parallel()

	created := time.Date(2026, time.March, 1, 9, 0, 0, 0, time.UTC)
	atRoot := testkit.Folder(t, testkit.FolderSpec{CreatedAt: created})
	inFolder := testkit.Snippet(t, testkit.SnippetSpec{
		FolderID: testkit.NewSequentialIDs().NewFolderID(), CreatedAt: created,
	})

	tests := []struct {
		name       string
		item       domain.TrashItem
		deletedAt  time.Time
		formerPath value.FolderPath
		wantErr    []error
	}{
		{name: "marks an item deleted when it was created", item: inFolder, deletedAt: created},
		{
			name:       "marks a detached item at the Root with the path it was in",
			item:       atRoot,
			deletedAt:  created.Add(time.Hour),
			formerPath: folderPath(t, "scripts", "a/b"),
		},
		{
			name:      "refuses a deletion before the creation",
			item:      inFolder,
			deletedAt: created.Add(-time.Nanosecond),
			wantErr:   []error{domain.ErrDeletedBeforeCreate},
		},
		{
			name:      "refuses a deletion date SQLite can't store",
			item:      atRoot,
			deletedAt: time.Date(2300, time.January, 1, 0, 0, 0, 0, time.UTC),
			wantErr:   []error{domain.ErrTimestampOutOfRange},
		},
		{
			name:       "refuses a former path on an item still in its Folder",
			item:       inFolder,
			deletedAt:  created,
			formerPath: folderPath(t, "scripts"),
			wantErr:    []error{domain.ErrFormerPathNotDetached},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := domain.NewTrashMark(tt.item, tt.deletedAt, tt.formerPath)

			requireErrors(t, err, tt.wantErr)

			if len(tt.wantErr) == 0 {
				assert.Equal(t, tt.deletedAt, got.DeletedAt())
				assert.Equal(t, tt.formerPath, got.FormerPath())
			}
		})
	}
}

func folderPath(t *testing.T, raw ...string) value.FolderPath {
	t.Helper()

	path, err := value.NewFolderPath(raw)
	require.NoError(t, err)

	return path
}
