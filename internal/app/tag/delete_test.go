package tag_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/tag"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

func TestNewDelete(t *testing.T) {
	t.Parallel()

	t.Run("names the missing dependency", func(t *testing.T) {
		t.Parallel()

		_, err := tag.NewDelete(nil)

		require.ErrorIs(t, err, domain.ErrMissingDependency)
		assert.ErrorContains(t, err, "tag.NewDelete: deleter")
	})

	t.Run("accepts every dependency", func(t *testing.T) {
		t.Parallel()

		_, err := tag.NewDelete(NewMockDeleter(t))

		assert.NoError(t, err)
	})
}

func TestDelete_Run(t *testing.T) {
	t.Parallel()

	t.Run("deletes the Tag", func(t *testing.T) {
		t.Parallel()

		id := storedTag(t).ID()
		deleter := NewMockDeleter(t)
		deleter.EXPECT().Delete(mock.Anything, id).Return(nil)

		err := newDelete(t, deleter).Run(t.Context(), tag.DeleteInput{TagID: id})

		assert.NoError(t, err)
	})

	t.Run("passes up a Tag that is gone", func(t *testing.T) {
		t.Parallel()

		deleter := NewMockDeleter(t)
		deleter.EXPECT().Delete(mock.Anything, mock.Anything).Return(domain.ErrNotFound)

		err := newDelete(t, deleter).Run(t.Context(), tag.DeleteInput{TagID: storedTag(t).ID()})

		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.ErrorContains(t, err, "tag.Delete")
	})
}

func newDelete(t *testing.T, deleter tag.Deleter) *tag.Delete {
	t.Helper()

	deleteAction, err := tag.NewDelete(deleter)
	require.NoError(t, err)

	return deleteAction
}
