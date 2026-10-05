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

func TestNewTagList(t *testing.T) {
	t.Parallel()

	_, err := browse.NewTagList(nil, nil)

	require.ErrorIs(t, err, domain.ErrMissingDependency)
	require.ErrorContains(t, err, "browse.NewTagList: tags")
	assert.ErrorContains(t, err, "counter")
}

func TestTagList_Run(t *testing.T) {
	t.Parallel()

	t.Run("lists every Tag by name ignoring case, with its Snippet count", func(t *testing.T) {
		t.Parallel()

		ids := testkit.NewSequentialIDs()
		golang := testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "go"})
		docker := testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "Docker"})
		unused := testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "unused"})
		lister := NewMockTagLister(t)
		lister.EXPECT().List(mock.Anything).Return([]domain.Tag{unused, golang, docker}, nil)

		counter := NewMockTagSnippetCounter(t)
		counter.EXPECT().CountByTag(mock.Anything).Return(map[domain.TagID]int{golang.ID(): 4, docker.ID(): 3}, nil)

		got, err := newTagList(t, lister, counter).Run(t.Context(), browse.TagListInput{})

		require.NoError(t, err)
		assert.Equal(t, []browse.TagCount{
			{Tag: docker, SnippetCount: 3},
			{Tag: golang, SnippetCount: 4},
			{Tag: unused, SnippetCount: 0},
		}, got)
	})

	t.Run("returns the lister's error", func(t *testing.T) {
		t.Parallel()

		lister := NewMockTagLister(t)
		lister.EXPECT().List(mock.Anything).Return(nil, errDatabaseLocked)

		_, err := newTagList(t, lister, NewMockTagSnippetCounter(t)).Run(t.Context(), browse.TagListInput{})

		require.ErrorIs(t, err, errDatabaseLocked)
		assert.ErrorContains(t, err, "browse.TagList: ")
	})

	t.Run("returns the counter's error", func(t *testing.T) {
		t.Parallel()

		lister := NewMockTagLister(t)
		lister.EXPECT().List(mock.Anything).Return(nil, nil)

		counter := NewMockTagSnippetCounter(t)
		counter.EXPECT().CountByTag(mock.Anything).Return(nil, errDatabaseLocked)

		_, err := newTagList(t, lister, counter).Run(t.Context(), browse.TagListInput{})

		require.ErrorIs(t, err, errDatabaseLocked)
		assert.ErrorContains(t, err, "browse.TagList: ")
	})
}

func newTagList(t *testing.T, tags browse.TagLister, counter browse.TagSnippetCounter) *browse.TagList {
	t.Helper()

	action, err := browse.NewTagList(tags, counter)
	require.NoError(t, err)

	return action
}
