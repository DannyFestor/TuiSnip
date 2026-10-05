package browse_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
)

func TestNewSnippetsWithTag(t *testing.T) {
	t.Parallel()

	_, err := browse.NewSnippetsWithTag(nil)

	require.ErrorIs(t, err, domain.ErrMissingDependency)
	assert.ErrorContains(t, err, "browse.NewSnippetsWithTag: lister")
}

func TestSnippetsWithTag_Run(t *testing.T) {
	t.Parallel()

	t.Run("returns the lister's Snippets carrying the Tag in the Input's sort order", func(t *testing.T) {
		t.Parallel()

		ids := testkit.NewSequentialIDs()
		tag := testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "go"})
		want := []domain.Snippet{
			testkit.Snippet(t, testkit.SnippetSpec{ID: ids.NewSnippetID(), Tags: []domain.Tag{tag}}),
		}
		lister := NewMockTagSnippetLister(t)
		lister.EXPECT().ListWithTag(mock.Anything, tag.ID(), domain.SortOrderUpdated).Return(want, nil)

		got, err := newSnippetsWithTag(t, lister).Run(
			t.Context(),
			browse.SnippetsWithTagInput{TagID: tag.ID(), Order: domain.SortOrderUpdated},
		)

		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("returns the lister's error", func(t *testing.T) {
		t.Parallel()

		tagID := testkit.NewSequentialIDs().NewTagID()
		lister := NewMockTagSnippetLister(t)
		lister.EXPECT().ListWithTag(mock.Anything, tagID, domain.SortOrderTitle).Return(nil, errDatabaseLocked)

		_, err := newSnippetsWithTag(t, lister).Run(
			t.Context(),
			browse.SnippetsWithTagInput{TagID: tagID, Order: domain.SortOrderTitle},
		)

		require.ErrorIs(t, err, errDatabaseLocked)
		assert.ErrorContains(t, err, "browse.SnippetsWithTag: ")
	})
}

func newSnippetsWithTag(t *testing.T, lister browse.TagSnippetLister) *browse.SnippetsWithTag {
	t.Helper()

	action, err := browse.NewSnippetsWithTag(lister)
	require.NoError(t, err)

	return action
}
