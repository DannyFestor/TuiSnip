package folder_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/folder"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

func TestNewPreviewDelete(t *testing.T) {
	t.Parallel()

	t.Run("names the missing dependency", func(t *testing.T) {
		t.Parallel()

		_, err := folder.NewPreviewDelete(nil)

		require.ErrorIs(t, err, domain.ErrMissingDependency)
		assert.ErrorContains(t, err, "folder.NewPreviewDelete: repo")
	})

	t.Run("accepts every dependency", func(t *testing.T) {
		t.Parallel()

		_, err := folder.NewPreviewDelete(NewMockPreviewDeleteRepository(t))

		assert.NoError(t, err)
	})
}

func TestPreviewDelete_Run(t *testing.T) {
	t.Parallel()

	t.Run("returns the Folder with the counts of its subfolders and Snippets", func(t *testing.T) {
		t.Parallel()

		stored := storedFolder(t)
		repo := NewMockPreviewDeleteRepository(t)
		repo.EXPECT().Find(mock.Anything, stored.ID()).Return(stored, nil)
		repo.EXPECT().CountSubfolders(mock.Anything, stored.ID()).Return(1, nil)
		repo.EXPECT().CountSnippetsInSubtree(mock.Anything, stored.ID()).Return(15, nil)

		got, err := newPreviewDelete(t, repo).Run(t.Context(), folder.PreviewDeleteInput{FolderID: stored.ID()})

		require.NoError(t, err)
		assert.Equal(t, folder.DeletePreview{Folder: stored, SubfolderCount: 1, SnippetCount: 15}, got)
	})

	t.Run("passes up a Folder that is gone", func(t *testing.T) {
		t.Parallel()

		repo := NewMockPreviewDeleteRepository(t)
		repo.EXPECT().Find(mock.Anything, mock.Anything).Return(domain.Folder{}, domain.ErrNotFound)

		_, err := newPreviewDelete(t, repo).Run(t.Context(), folder.PreviewDeleteInput{FolderID: storedFolder(t).ID()})

		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.ErrorContains(t, err, "folder.PreviewDelete")
	})
}

func newPreviewDelete(t *testing.T, repo folder.PreviewDeleteRepository) *folder.PreviewDelete {
	t.Helper()

	preview, err := folder.NewPreviewDelete(repo)
	require.NoError(t, err)

	return preview
}
