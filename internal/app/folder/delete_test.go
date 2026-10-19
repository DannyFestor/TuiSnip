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

	t.Run("names every missing dependency", func(t *testing.T) {
		t.Parallel()

		_, err := folder.NewDelete(nil, nil)

		require.ErrorIs(t, err, domain.ErrMissingDependency)
		require.ErrorContains(t, err, "folder.NewDelete: trasher")
		assert.ErrorContains(t, err, "clock")
	})

	t.Run("accepts every dependency", func(t *testing.T) {
		t.Parallel()

		_, err := folder.NewDelete(NewMockTrasher(t), fixedClock())

		assert.NoError(t, err)
	})
}

func TestDelete_Run(t *testing.T) {
	t.Parallel()

	t.Run("moves the Folder to the Trash, deleted now", func(t *testing.T) {
		t.Parallel()

		id := storedFolder(t).ID()
		trasher := NewMockTrasher(t)
		trasher.EXPECT().Trash(mock.Anything, id, now()).Return(nil)

		err := newDelete(t, trasher).Run(t.Context(), folder.DeleteInput{FolderID: id})

		assert.NoError(t, err)
	})

	t.Run("passes up a Folder that is gone", func(t *testing.T) {
		t.Parallel()

		trasher := NewMockTrasher(t)
		trasher.EXPECT().Trash(mock.Anything, mock.Anything, mock.Anything).Return(domain.ErrNotFound)

		err := newDelete(t, trasher).Run(t.Context(), folder.DeleteInput{FolderID: storedFolder(t).ID()})

		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.ErrorContains(t, err, "folder.Delete")
	})
}

func newDelete(t *testing.T, trasher folder.Trasher) *folder.Delete {
	t.Helper()

	deleteAction, err := folder.NewDelete(trasher, fixedClock())
	require.NoError(t, err)

	return deleteAction
}
