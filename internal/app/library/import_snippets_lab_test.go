package library_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/library"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
)

const snippetsLabPath = "/exports/snippetslab.json"

func TestNewImportSnippetsLab(t *testing.T) {
	t.Parallel()

	_, err := library.NewImportSnippetsLab(nil, nil)

	require.ErrorIs(t, err, domain.ErrMissingDependency)
	require.ErrorContains(t, err, "library.NewImportSnippetsLab: reader")
	assert.ErrorContains(t, err, "inserter")
}

func TestImportSnippetsLab_Run(t *testing.T) {
	t.Parallel()

	t.Run(
		"inserts the Library it reads beside the stored one and reports what was written and lost",
		func(t *testing.T) {
			t.Parallel()

			ids := testkit.NewSequentialIDs()
			folder := testkit.Folder(t, testkit.FolderSpec{ID: ids.NewFolderID()})
			reused := testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "Go"})
			fresh := testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "shell"})
			snippet := testkit.Snippet(t, testkit.SnippetSpec{
				ID:        ids.NewSnippetID(),
				FolderID:  folder.ID(),
				Fragments: []testkit.FragmentSpec{{ID: ids.NewFragmentID()}, {ID: ids.NewFragmentID()}},
				Tags:      []domain.Tag{reused, fresh},
			})
			read := domain.Library{
				Folders:  []domain.Folder{folder},
				Tags:     []domain.Tag{reused, fresh},
				Snippets: []domain.Snippet{snippet},
			}
			losses := domain.ImportLosses{
				LostPins:          []value.Title{snippet.Title()},
				LostGistLinks:     nil,
				LostAttachments:   nil,
				FolderFallbacks:   nil,
				FragmentFallbacks: nil,
			}
			reader := NewMockSnippetsLabReader(t)
			reader.EXPECT().Read(mock.Anything, snippetsLabPath).Return(read, losses, nil)

			written := domain.Library{Folders: read.Folders, Tags: []domain.Tag{fresh}, Snippets: read.Snippets}
			inserter := NewMockAlongsideInserter(t)
			inserter.EXPECT().InsertAlongside(mock.Anything, read).Return(written, nil)

			got, err := newImportSnippetsLab(t, reader, inserter).Run(
				t.Context(), library.ImportSnippetsLabInput{Path: snippetsLabPath},
			)

			require.NoError(t, err)
			assert.Equal(t, library.SnippetsLabReport{
				Counts: library.Counts{Folders: 1, Tags: 1, Snippets: 1, Fragments: 2},
				Losses: losses,
			}, got)
		},
	)

	t.Run("inserts nothing when the file can't be read", func(t *testing.T) {
		t.Parallel()

		reader := NewMockSnippetsLabReader(t)
		reader.EXPECT().Read(mock.Anything, snippetsLabPath).
			Return(domain.Library{}, domain.ImportLosses{}, errPermissionDenied)

		_, err := newImportSnippetsLab(t, reader, NewMockAlongsideInserter(t)).Run(
			t.Context(), library.ImportSnippetsLabInput{Path: snippetsLabPath},
		)

		require.ErrorIs(t, err, errPermissionDenied)
		assert.ErrorContains(t, err, "library.ImportSnippetsLab: ")
	})

	t.Run("inserts nothing when two Tags it reads share a name", func(t *testing.T) {
		t.Parallel()

		ids := testkit.NewSequentialIDs()
		tags := []domain.Tag{
			testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "Go"}),
			testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "go"}),
		}
		reader := NewMockSnippetsLabReader(t)
		reader.EXPECT().
			Read(mock.Anything, snippetsLabPath).
			Return(domain.Library{Tags: tags}, domain.ImportLosses{}, nil)

		_, err := newImportSnippetsLab(t, reader, NewMockAlongsideInserter(t)).Run(
			t.Context(), library.ImportSnippetsLabInput{Path: snippetsLabPath},
		)

		require.ErrorIs(t, err, domain.ErrTagNameTaken)
		assert.ErrorContains(t, err, "library.ImportSnippetsLab: ")
	})
}

func newImportSnippetsLab(
	t *testing.T, reader library.SnippetsLabReader, inserter library.AlongsideInserter,
) *library.ImportSnippetsLab {
	t.Helper()

	imported, err := library.NewImportSnippetsLab(reader, inserter)
	require.NoError(t, err)

	return imported
}
