package library_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/library"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
)

const exportPath = "/backups/tuisnip.json"

var errDatabaseLocked = errors.New("database is locked")

func TestNewExport(t *testing.T) {
	t.Parallel()

	_, err := library.NewExport(nil, nil)

	require.ErrorIs(t, err, domain.ErrMissingDependency)
	require.ErrorContains(t, err, "library.NewExport: loader")
	assert.ErrorContains(t, err, "writer")
}

func TestExport_Run(t *testing.T) {
	t.Parallel()

	t.Run("writes the Library in created order to the path and counts it", func(t *testing.T) {
		t.Parallel()

		ids := testkit.NewSequentialIDs()
		earlier := time.Date(2026, time.March, 1, 9, 0, 0, 0, time.UTC)
		later := earlier.Add(time.Hour)
		newerFolder := testkit.Folder(t, testkit.FolderSpec{ID: ids.NewFolderID(), CreatedAt: later})
		olderFolder := testkit.Folder(t, testkit.FolderSpec{ID: ids.NewFolderID(), CreatedAt: earlier})
		tag := testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID()})
		newerSnippet := testkit.Snippet(t, testkit.SnippetSpec{ID: ids.NewSnippetID(), CreatedAt: later})
		olderSnippet := testkit.Snippet(t, testkit.SnippetSpec{ID: ids.NewSnippetID(), CreatedAt: earlier})
		loader := NewMockLoader(t)
		loader.EXPECT().Load(mock.Anything).Return(domain.Library{
			Folders:  []domain.Folder{newerFolder, olderFolder},
			Tags:     []domain.Tag{tag},
			Snippets: []domain.Snippet{newerSnippet, olderSnippet},
		}, nil)

		writer := NewMockWriter(t)
		writer.EXPECT().Write(mock.Anything, exportPath, domain.Library{
			Folders:  []domain.Folder{olderFolder, newerFolder},
			Tags:     []domain.Tag{tag},
			Snippets: []domain.Snippet{olderSnippet, newerSnippet},
		}).Return(nil)

		got, err := newExport(t, loader, writer).Run(t.Context(), library.ExportInput{Path: exportPath})

		require.NoError(t, err)
		assert.Equal(t, library.Counts{Folders: 2, Tags: 1, Snippets: 2, Fragments: 2}, got)
	})

	t.Run("writes nothing when the Library can't be loaded", func(t *testing.T) {
		t.Parallel()

		loader := NewMockLoader(t)
		loader.EXPECT().Load(mock.Anything).Return(domain.Library{}, errDatabaseLocked)

		_, err := newExport(t, loader, NewMockWriter(t)).Run(t.Context(), library.ExportInput{Path: exportPath})

		require.ErrorIs(t, err, errDatabaseLocked)
		assert.ErrorContains(t, err, "library.Export: ")
	})
}

func newExport(t *testing.T, loader library.Loader, writer library.Writer) *library.Export {
	t.Helper()

	export, err := library.NewExport(loader, writer)
	require.NoError(t, err)

	return export
}
