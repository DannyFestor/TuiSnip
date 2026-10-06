package snippet_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
)

func TestNewDuplicate(t *testing.T) {
	t.Parallel()

	t.Run("names every missing dependency", func(t *testing.T) {
		t.Parallel()

		_, err := snippet.NewDuplicate(nil, nil, nil)

		require.ErrorIs(t, err, domain.ErrMissingDependency)
		require.ErrorContains(t, err, "snippet.NewDuplicate: repo")
		require.ErrorContains(t, err, "ids")
		assert.ErrorContains(t, err, "clock")
	})

	t.Run("accepts every dependency", func(t *testing.T) {
		t.Parallel()

		_, err := snippet.NewDuplicate(NewMockDuplicateRepository(t), testkit.NewSequentialIDs(), fixedClock())

		assert.NoError(t, err)
	})
}

func TestDuplicate_Run(t *testing.T) {
	t.Parallel()

	t.Run("inserts a copy of the stored Snippet under new ids and the current time", func(t *testing.T) {
		t.Parallel()

		stored := storedSnippet(t, "curl -d @b.json")
		repo := NewMockDuplicateRepository(t)
		repo.EXPECT().Find(mock.Anything, stored.ID()).Return(stored, nil)

		var inserted domain.Snippet

		repo.EXPECT().Insert(mock.Anything, mock.AnythingOfType("domain.Snippet")).
			Run(func(_ context.Context, s domain.Snippet) { inserted = s }).
			Return(nil)

		duplicate, err := newDuplicate(t, repo).Run(t.Context(), snippet.DuplicateInput{SnippetID: stored.ID()})

		require.NoError(t, err)
		assert.Equal(t, inserted, duplicate)
		assert.NotEqual(t, stored.ID(), duplicate.ID())
		assert.NotEqual(t, stored.FirstFragment().ID(), duplicate.FirstFragment().ID())
		assert.Equal(t, stored.FirstFragment().Content(), duplicate.FirstFragment().Content())
		assert.Equal(t, createdAt(), duplicate.CreatedAt())
	})

	t.Run("passes up a Snippet that is gone", func(t *testing.T) {
		t.Parallel()

		repo := NewMockDuplicateRepository(t)
		repo.EXPECT().Find(mock.Anything, mock.Anything).Return(domain.Snippet{}, domain.ErrNotFound)

		_, err := newDuplicate(t, repo).Run(t.Context(), snippet.DuplicateInput{SnippetID: storedSnippet(t, "").ID()})

		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.ErrorContains(t, err, "snippet.Duplicate: ")
	})

	t.Run("passes up a failed insert", func(t *testing.T) {
		t.Parallel()

		stored := storedSnippet(t, "")
		repo := NewMockDuplicateRepository(t)
		repo.EXPECT().Find(mock.Anything, mock.Anything).Return(stored, nil)
		repo.EXPECT().Insert(mock.Anything, mock.Anything).Return(errDatabaseLocked)

		_, err := newDuplicate(t, repo).Run(t.Context(), snippet.DuplicateInput{SnippetID: stored.ID()})

		require.ErrorIs(t, err, errDatabaseLocked)
		assert.ErrorContains(t, err, "snippet.Duplicate: ")
	})
}

func newDuplicate(t *testing.T, repo snippet.DuplicateRepository) *snippet.Duplicate {
	t.Helper()

	duplicate, err := snippet.NewDuplicate(repo, testkit.NewSequentialIDs(), fixedClock())
	require.NoError(t, err)

	return duplicate
}
