package tag_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/tag"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

func TestNewPreviewDelete(t *testing.T) {
	t.Parallel()

	t.Run("names the missing dependency", func(t *testing.T) {
		t.Parallel()

		_, err := tag.NewPreviewDelete(nil)

		require.ErrorIs(t, err, domain.ErrMissingDependency)
		assert.ErrorContains(t, err, "tag.NewPreviewDelete: repo")
	})

	t.Run("accepts every dependency", func(t *testing.T) {
		t.Parallel()

		_, err := tag.NewPreviewDelete(NewMockPreviewDeleteRepository(t))

		assert.NoError(t, err)
	})
}

func TestPreviewDelete_Run(t *testing.T) {
	t.Parallel()

	t.Run("returns the Tag with the count of the Snippets carrying it", func(t *testing.T) {
		t.Parallel()

		stored := storedTag(t)
		repo := NewMockPreviewDeleteRepository(t)
		repo.EXPECT().Find(mock.Anything, stored.ID()).Return(stored, nil)
		repo.EXPECT().CountSnippets(mock.Anything, stored.ID()).Return(4, nil)

		got, err := newPreviewDelete(t, repo).Run(t.Context(), tag.PreviewDeleteInput{TagID: stored.ID()})

		require.NoError(t, err)
		assert.Equal(t, tag.DeletePreview{Tag: stored, SnippetCount: 4}, got)
	})

	t.Run("passes up a Tag that is gone", func(t *testing.T) {
		t.Parallel()

		repo := NewMockPreviewDeleteRepository(t)
		repo.EXPECT().Find(mock.Anything, mock.Anything).Return(domain.Tag{}, domain.ErrNotFound)

		_, err := newPreviewDelete(t, repo).Run(t.Context(), tag.PreviewDeleteInput{TagID: storedTag(t).ID()})

		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.ErrorContains(t, err, "tag.PreviewDelete")
	})

	t.Run("passes up a failed count", func(t *testing.T) {
		t.Parallel()

		stored := storedTag(t)
		repo := NewMockPreviewDeleteRepository(t)
		repo.EXPECT().Find(mock.Anything, mock.Anything).Return(stored, nil)
		repo.EXPECT().CountSnippets(mock.Anything, mock.Anything).Return(0, errDatabaseLocked)

		_, err := newPreviewDelete(t, repo).Run(t.Context(), tag.PreviewDeleteInput{TagID: stored.ID()})

		require.ErrorIs(t, err, errDatabaseLocked)
		assert.ErrorContains(t, err, "tag.PreviewDelete")
	})
}

func newPreviewDelete(t *testing.T, repo tag.PreviewDeleteRepository) *tag.PreviewDelete {
	t.Helper()

	preview, err := tag.NewPreviewDelete(repo)
	require.NoError(t, err)

	return preview
}
