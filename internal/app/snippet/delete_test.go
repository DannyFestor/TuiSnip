package snippet_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

func TestNewDelete(t *testing.T) {
	t.Parallel()

	t.Run("names every missing dependency", func(t *testing.T) {
		t.Parallel()

		_, err := snippet.NewDelete(nil, nil)

		require.ErrorIs(t, err, domain.ErrMissingDependency)
		require.ErrorContains(t, err, "snippet.NewDelete: trasher")
		assert.ErrorContains(t, err, "clock")
	})

	t.Run("accepts every dependency", func(t *testing.T) {
		t.Parallel()

		_, err := snippet.NewDelete(NewMockTrasher(t), fixedClock())

		assert.NoError(t, err)
	})
}

func TestDelete_Run(t *testing.T) {
	t.Parallel()

	t.Run("moves the Snippet to the Trash, deleted now", func(t *testing.T) {
		t.Parallel()

		id := storedSnippet(t, "").ID()
		trasher := NewMockTrasher(t)
		trasher.EXPECT().Trash(mock.Anything, id, createdAt()).Return(nil)

		err := newDelete(t, trasher).Run(t.Context(), snippet.DeleteInput{SnippetID: id})

		assert.NoError(t, err)
	})

	t.Run("passes up a Snippet that is gone", func(t *testing.T) {
		t.Parallel()

		trasher := NewMockTrasher(t)
		trasher.EXPECT().Trash(mock.Anything, mock.Anything, mock.Anything).Return(domain.ErrNotFound)

		err := newDelete(t, trasher).Run(t.Context(), snippet.DeleteInput{SnippetID: storedSnippet(t, "").ID()})

		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.ErrorContains(t, err, "snippet.Delete: ")
	})
}

func newDelete(t *testing.T, trasher snippet.Trasher) *snippet.Delete {
	t.Helper()

	deleteAction, err := snippet.NewDelete(trasher, fixedClock())
	require.NoError(t, err)

	return deleteAction
}
