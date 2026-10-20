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

func TestNewDeletePermanently(t *testing.T) {
	t.Parallel()

	t.Run("names the missing dependency", func(t *testing.T) {
		t.Parallel()

		_, err := trash.NewDeletePermanently(nil)

		require.ErrorIs(t, err, domain.ErrMissingDependency)
		assert.ErrorContains(t, err, "trash.NewDeletePermanently: repo")
	})
}

func TestDeletePermanently_Run(t *testing.T) {
	t.Parallel()

	ids := testkit.NewSequentialIDs()

	t.Run("deletes a trashed Snippet", func(t *testing.T) {
		t.Parallel()

		snippetID := ids.NewSnippetID()
		repo := NewMockDeletePermanentlyRepository(t)
		repo.EXPECT().DeleteSnippet(mock.Anything, snippetID).Return(nil)

		err := newDeletePermanently(t, repo).Run(
			t.Context(), trash.DeletePermanentlyInput{Item: trash.SnippetItem(snippetID)},
		)

		require.NoError(t, err)
	})

	t.Run("deletes a trashed Folder", func(t *testing.T) {
		t.Parallel()

		folderID := ids.NewFolderID()
		repo := NewMockDeletePermanentlyRepository(t)
		repo.EXPECT().DeleteFolder(mock.Anything, folderID).Return(nil)

		err := newDeletePermanently(t, repo).Run(
			t.Context(), trash.DeletePermanentlyInput{Item: trash.FolderItem(folderID)},
		)

		require.NoError(t, err)
	})

	t.Run("passes up a Trash item no longer in the Trash", func(t *testing.T) {
		t.Parallel()

		repo := NewMockDeletePermanentlyRepository(t)
		repo.EXPECT().DeleteFolder(mock.Anything, mock.Anything).Return(domain.ErrNotFound)

		err := newDeletePermanently(t, repo).Run(
			t.Context(), trash.DeletePermanentlyInput{Item: trash.FolderItem(ids.NewFolderID())},
		)

		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.ErrorContains(t, err, "trash.DeletePermanently: ")
	})
}

func newDeletePermanently(t *testing.T, repo trash.DeletePermanentlyRepository) *trash.DeletePermanently {
	t.Helper()

	deletePermanently, err := trash.NewDeletePermanently(repo)
	require.NoError(t, err)

	return deletePermanently
}
