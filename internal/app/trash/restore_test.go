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

func TestNewRestore(t *testing.T) {
	t.Parallel()

	t.Run("names the missing dependency", func(t *testing.T) {
		t.Parallel()

		_, err := trash.NewRestore(nil)

		require.ErrorIs(t, err, domain.ErrMissingDependency)
		assert.ErrorContains(t, err, "trash.NewRestore: repo")
	})
}

func TestRestore_Run(t *testing.T) {
	t.Parallel()

	ids := testkit.NewSequentialIDs()
	landed := domain.Restoration{FolderID: ids.NewFolderID(), Fallback: domain.RestoreFallbackNone}

	t.Run("restores a Snippet item and says where it landed", func(t *testing.T) {
		t.Parallel()

		snippetID := ids.NewSnippetID()
		repo := NewMockRestoreRepository(t)
		repo.EXPECT().RestoreSnippet(mock.Anything, snippetID).Return(landed, nil)

		got, err := newRestore(t, repo).Run(t.Context(), trash.RestoreInput{Item: trash.SnippetItem(snippetID)})

		require.NoError(t, err)
		assert.Equal(t, landed, got)
	})

	t.Run("restores a Folder item and says where it landed", func(t *testing.T) {
		t.Parallel()

		folderID := ids.NewFolderID()
		repo := NewMockRestoreRepository(t)
		repo.EXPECT().RestoreFolder(mock.Anything, folderID).Return(landed, nil)

		got, err := newRestore(t, repo).Run(t.Context(), trash.RestoreInput{Item: trash.FolderItem(folderID)})

		require.NoError(t, err)
		assert.Equal(t, landed, got)
	})

	t.Run("passes up an item no longer in the Trash", func(t *testing.T) {
		t.Parallel()

		repo := NewMockRestoreRepository(t)
		repo.EXPECT().RestoreSnippet(mock.Anything, mock.Anything).Return(domain.Restoration{}, domain.ErrNotFound)

		_, err := newRestore(t, repo).Run(t.Context(), trash.RestoreInput{Item: trash.SnippetItem(ids.NewSnippetID())})

		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.ErrorContains(t, err, "trash.Restore: ")
	})
}

func newRestore(t *testing.T, repo trash.RestoreRepository) *trash.Restore {
	t.Helper()

	restore, err := trash.NewRestore(repo)
	require.NoError(t, err)

	return restore
}
