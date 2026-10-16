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
		formerPath []value.FolderName
		wantErr    []error
	}{
		{name: "marks an item deleted when it was created", item: inFolder, deletedAt: created},
		{
			name:       "marks a detached item at the Root with the path it was in",
			item:       atRoot,
			deletedAt:  created.Add(time.Hour),
			formerPath: folderNames(t, "scripts", "a/b"),
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
			formerPath: folderNames(t, "scripts"),
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

func TestTrashMark_FormerPath(t *testing.T) {
	t.Parallel()

	t.Run("hands out a copy", func(t *testing.T) {
		t.Parallel()

		folder := testkit.Folder(t, testkit.FolderSpec{})
		mark, err := domain.NewTrashMark(folder, folder.CreatedAt(), folderNames(t, "scripts"))
		require.NoError(t, err)

		mark.FormerPath()[0] = folderNames(t, "changed")[0]

		assert.Equal(t, folderNames(t, "scripts"), mark.FormerPath())
	})
}

func folderNames(t *testing.T, raw ...string) []value.FolderName {
	t.Helper()

	names := make([]value.FolderName, 0, len(raw))
	for _, name := range raw {
		parsed, err := value.NewFolderName(name)
		require.NoError(t, err)

		names = append(names, parsed)
	}

	return names
}
