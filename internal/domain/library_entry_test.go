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

func TestNewLibraryEntry(t *testing.T) {
	t.Parallel()

	snippet := testkit.Snippet(t, testkit.SnippetSpec{FolderID: testkit.NewSequentialIDs().NewFolderID()})

	tests := []struct {
		name       string
		deletedAt  time.Time
		formerPath value.FolderPath
		wantErr    []error
	}{
		{name: "takes a record outside the Trash", deletedAt: time.Time{}, formerPath: value.FolderPath{}},
		{name: "takes a Trash item", deletedAt: snippet.CreatedAt(), formerPath: value.FolderPath{}},
		{
			name:       "refuses a former path on a record outside the Trash",
			deletedAt:  time.Time{},
			formerPath: folderPath(t, "scripts"),
			wantErr:    []error{domain.ErrFormerPathOutsideTrash},
		},
		{
			name:      "refuses a deletion before the creation",
			deletedAt: snippet.CreatedAt().Add(-time.Nanosecond),
			wantErr:   []error{domain.ErrDeletedBeforeCreate},
		},
		{
			name:       "refuses a former path on a Trash item still in its Folder",
			deletedAt:  snippet.CreatedAt(),
			formerPath: folderPath(t, "scripts"),
			wantErr:    []error{domain.ErrFormerPathNotDetached},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := domain.NewLibraryEntry(snippet, tt.deletedAt, tt.formerPath)

			requireErrors(t, err, tt.wantErr)
		})
	}
}

func TestNewLibrary(t *testing.T) {
	t.Parallel()

	t.Run("keeps every entry and marks only the Trash items", func(t *testing.T) {
		t.Parallel()

		ids := testkit.NewSequentialIDs()
		liveFolder := testkit.Folder(t, testkit.FolderSpec{ID: ids.NewFolderID()})
		trashedFolder := testkit.Folder(t, testkit.FolderSpec{ID: ids.NewFolderID()})
		tag := testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID()})
		liveSnippet := testkit.Snippet(t, testkit.SnippetSpec{ID: ids.NewSnippetID()})
		detached := testkit.Snippet(t, testkit.SnippetSpec{ID: ids.NewSnippetID()})
		deleted := liveFolder.CreatedAt().Add(time.Hour)

		got := domain.NewLibrary(
			[]domain.LibraryEntry[domain.Folder]{
				libraryEntry(t, liveFolder, time.Time{}, value.FolderPath{}),
				libraryEntry(t, trashedFolder, deleted, value.FolderPath{}),
			},
			[]domain.Tag{tag},
			[]domain.LibraryEntry[domain.Snippet]{
				libraryEntry(t, liveSnippet, time.Time{}, value.FolderPath{}),
				libraryEntry(t, detached, deleted, folderPath(t, "gone")),
			},
		)

		assert.Equal(t, domain.Library{
			Folders:  []domain.Folder{liveFolder, trashedFolder},
			Tags:     []domain.Tag{tag},
			Snippets: []domain.Snippet{liveSnippet, detached},
			Trash: domain.Trash{
				Folders: map[domain.FolderID]domain.TrashMark{
					trashedFolder.ID(): testkit.TrashMark(t, testkit.TrashMarkSpec{
						Item: trashedFolder, DeletedAt: deleted,
					}),
				},
				Snippets: map[domain.SnippetID]domain.TrashMark{
					detached.ID(): testkit.TrashMark(t, testkit.TrashMarkSpec{
						Item: detached, DeletedAt: deleted, FormerPath: []string{"gone"},
					}),
				},
			},
		}, got)
	})
}

func libraryEntry[E domain.TrashItem](
	t *testing.T, entity E, deletedAt time.Time, formerPath value.FolderPath,
) domain.LibraryEntry[E] {
	t.Helper()

	entry, err := domain.NewLibraryEntry(entity, deletedAt, formerPath)
	require.NoError(t, err)

	return entry
}
