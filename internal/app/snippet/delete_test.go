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

	t.Run("names the missing dependency", func(t *testing.T) {
		t.Parallel()

		_, err := snippet.NewDelete(nil)

		require.ErrorIs(t, err, domain.ErrMissingDependency)
		assert.ErrorContains(t, err, "snippet.NewDelete: deleter")
	})

	t.Run("accepts every dependency", func(t *testing.T) {
		t.Parallel()

		_, err := snippet.NewDelete(NewMockDeleter(t))

		assert.NoError(t, err)
	})
}

func TestDelete_Run(t *testing.T) {
	t.Parallel()

	t.Run("deletes the Snippet", func(t *testing.T) {
		t.Parallel()

		id := storedSnippet(t, "").ID()
		deleter := NewMockDeleter(t)
		deleter.EXPECT().Delete(mock.Anything, id).Return(nil)

		err := newDelete(t, deleter).Run(t.Context(), snippet.DeleteInput{SnippetID: id})

		assert.NoError(t, err)
	})

	t.Run("passes up a Snippet that is gone", func(t *testing.T) {
		t.Parallel()

		deleter := NewMockDeleter(t)
		deleter.EXPECT().Delete(mock.Anything, mock.Anything).Return(domain.ErrNotFound)

		err := newDelete(t, deleter).Run(t.Context(), snippet.DeleteInput{SnippetID: storedSnippet(t, "").ID()})

		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.ErrorContains(t, err, "snippet.Delete: ")
	})
}

func newDelete(t *testing.T, deleter snippet.Deleter) *snippet.Delete {
	t.Helper()

	deleteAction, err := snippet.NewDelete(deleter)
	require.NoError(t, err)

	return deleteAction
}
