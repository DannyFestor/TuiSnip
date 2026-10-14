package library_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/library"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
)

const importPath = "/backups/tuisnip.json"

var errPermissionDenied = errors.New("permission denied")

func TestNewImport(t *testing.T) {
	t.Parallel()

	_, err := library.NewImport(nil, nil)

	require.ErrorIs(t, err, domain.ErrMissingDependency)
	require.ErrorContains(t, err, "library.NewImport: reader")
	assert.ErrorContains(t, err, "inserter")
}

func TestImport_Run(t *testing.T) {
	t.Parallel()

	t.Run("inserts the Library it reads into the empty one and counts it", func(t *testing.T) {
		t.Parallel()

		ids := testkit.NewSequentialIDs()
		folder := testkit.Folder(t, testkit.FolderSpec{ID: ids.NewFolderID()})
		tags := []domain.Tag{
			testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "shell"}),
			testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "unused"}),
		}
		snippet := testkit.Snippet(t, testkit.SnippetSpec{
			ID: ids.NewSnippetID(), FolderID: folder.ID(), Tags: tags[:1],
		})
		read := domain.Library{Folders: []domain.Folder{folder}, Tags: tags, Snippets: []domain.Snippet{snippet}}
		reader := NewMockReader(t)
		reader.EXPECT().Read(mock.Anything, importPath).Return(read, nil)

		inserter := NewMockEmptyInserter(t)
		inserter.EXPECT().InsertIntoEmpty(mock.Anything, read).Return(nil)

		got, err := newImport(t, reader, inserter).Run(t.Context(), library.ImportInput{Path: importPath})

		require.NoError(t, err)
		assert.Equal(t, library.Counts{Folders: 1, Tags: 2, Snippets: 1, Fragments: 1}, got)
	})

	t.Run("inserts nothing when the file can't be read", func(t *testing.T) {
		t.Parallel()

		reader := NewMockReader(t)
		reader.EXPECT().Read(mock.Anything, importPath).Return(domain.Library{}, errPermissionDenied)

		_, err := newImport(t, reader, NewMockEmptyInserter(t)).Run(t.Context(), library.ImportInput{Path: importPath})

		require.ErrorIs(t, err, errPermissionDenied)
		assert.ErrorContains(t, err, "library.Import: ")
	})

	t.Run("inserts nothing when a Snippet is filed in a Folder the file lacks", func(t *testing.T) {
		t.Parallel()

		ids := testkit.NewSequentialIDs()
		snippet := testkit.Snippet(t, testkit.SnippetSpec{ID: ids.NewSnippetID(), FolderID: ids.NewFolderID()})
		reader := NewMockReader(t)
		reader.EXPECT().Read(mock.Anything, importPath).Return(domain.Library{Snippets: []domain.Snippet{snippet}}, nil)

		_, err := newImport(t, reader, NewMockEmptyInserter(t)).Run(t.Context(), library.ImportInput{Path: importPath})

		require.ErrorIs(t, err, domain.ErrMissingFolder)
		assert.ErrorContains(t, err, "library.Import: ")
	})
}

func newImport(t *testing.T, reader library.Reader, inserter library.EmptyInserter) *library.Import {
	t.Helper()

	imported, err := library.NewImport(reader, inserter)
	require.NoError(t, err)

	return imported
}
