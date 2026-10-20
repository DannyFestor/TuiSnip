package trash_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/trash"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
)

func TestNewEmpty(t *testing.T) {
	t.Parallel()

	t.Run("names the missing dependency", func(t *testing.T) {
		t.Parallel()

		_, err := trash.NewEmpty(nil)

		require.ErrorIs(t, err, domain.ErrMissingDependency)
		assert.ErrorContains(t, err, "trash.NewEmpty: emptier")
	})
}

func TestEmpty_Run(t *testing.T) {
	t.Parallel()

	ids := testkit.NewSequentialIDs()

	t.Run("empties only the given Trash items and reports what was deleted", func(t *testing.T) {
		t.Parallel()

		folderID, snippetID := ids.NewFolderID(), ids.NewSnippetID()
		emptied := domain.TrashEmptied{Folders: 2, Snippets: 20}
		emptier := NewMockEmptier(t)
		emptier.EXPECT().Empty(mock.Anything, domain.TrashItemIDs{
			Folders:  []domain.FolderID{folderID},
			Snippets: []domain.SnippetID{snippetID},
		}).Return(emptied, nil)

		got, err := newEmpty(t, emptier).Run(t.Context(), trash.EmptyInput{
			Items: []trash.ItemID{trash.SnippetItem(snippetID), trash.FolderItem(folderID)},
		})

		require.NoError(t, err)
		assert.Equal(t, emptied, got)
	})

	t.Run("passes up a failure", func(t *testing.T) {
		t.Parallel()

		emptier := NewMockEmptier(t)
		emptier.EXPECT().Empty(mock.Anything, mock.Anything).Return(domain.TrashEmptied{}, assert.AnError)

		_, err := newEmpty(t, emptier).Run(t.Context(), trash.EmptyInput{Items: nil})

		require.ErrorIs(t, err, assert.AnError)
		assert.ErrorContains(t, err, "trash.Empty: ")
	})
}

func newEmpty(t *testing.T, emptier trash.Emptier) *trash.Empty {
	t.Helper()

	empty, err := trash.NewEmpty(emptier)
	require.NoError(t, err)

	return empty
}
