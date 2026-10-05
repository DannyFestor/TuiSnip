package folder_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/folder"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

func TestNewDelete(t *testing.T) {
	t.Parallel()

	t.Run("names the missing dependency", func(t *testing.T) {
		t.Parallel()

		_, err := folder.NewDelete(nil)

		require.ErrorIs(t, err, domain.ErrMissingDependency)
		assert.ErrorContains(t, err, "folder.NewDelete: deleter")
	})

	t.Run("accepts every dependency", func(t *testing.T) {
		t.Parallel()

		_, err := folder.NewDelete(NewMockDeleter(t))

		assert.NoError(t, err)
	})
}

func TestDelete_Run(t *testing.T) {
	t.Parallel()

	t.Run("deletes the Folder", func(t *testing.T) {
		t.Parallel()

		id := storedFolder(t).ID()
		deleter := NewMockDeleter(t)
		deleter.EXPECT().Delete(mock.Anything, id).Return(nil)

		err := newDelete(t, deleter).Run(t.Context(), folder.DeleteInput{FolderID: id})

		assert.NoError(t, err)
	})

	t.Run("passes up a Folder that is gone", func(t *testing.T) {
		t.Parallel()

		deleter := NewMockDeleter(t)
		deleter.EXPECT().Delete(mock.Anything, mock.Anything).Return(domain.ErrNotFound)

		err := newDelete(t, deleter).Run(t.Context(), folder.DeleteInput{FolderID: storedFolder(t).ID()})

		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.ErrorContains(t, err, "folder.Delete")
	})
}

func newDelete(t *testing.T, deleter folder.Deleter) *folder.Delete {
	t.Helper()

	deleteAction, err := folder.NewDelete(deleter)
	require.NoError(t, err)

	return deleteAction
}
