package search_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/search"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
)

var errDatabaseLocked = errors.New("database is locked")

func TestNewQuery(t *testing.T) {
	t.Parallel()

	_, err := search.NewQuery(nil)

	require.ErrorIs(t, err, domain.ErrMissingDependency)
	assert.ErrorContains(t, err, "search.NewQuery: searcher")
}

func TestQuery_Run(t *testing.T) {
	t.Parallel()

	t.Run("returns no hits for an empty query without searching", func(t *testing.T) {
		t.Parallel()

		hits, err := newQuery(t, NewMockSearcher(t)).Run(t.Context(), search.QueryInput{Text: ""})

		require.NoError(t, err)
		assert.Empty(t, hits)
		assert.NotNil(t, hits)
	})

	t.Run("passes a query of spaces to the searcher", func(t *testing.T) {
		t.Parallel()

		searcher := NewMockSearcher(t)
		searcher.EXPECT().Search(mock.Anything, " ").Return([]domain.SearchHit{}, nil)

		_, err := newQuery(t, searcher).Run(t.Context(), search.QueryInput{Text: " "})

		assert.NoError(t, err)
	})

	t.Run("returns the searcher's hits in order", func(t *testing.T) {
		t.Parallel()

		want := []domain.SearchHit{
			domain.NewSearchHit(
				testkit.Snippet(t, testkit.SnippetSpec{Title: "docker run"}),
				domain.FieldScores{Title: 9},
			),
			domain.NewSearchHit(
				testkit.Snippet(t, testkit.SnippetSpec{Title: "run docker"}),
				domain.FieldScores{Title: 2},
			),
		}
		searcher := NewMockSearcher(t)
		searcher.EXPECT().Search(mock.Anything, "docker run").Return(want, nil)

		hits, err := newQuery(t, searcher).Run(t.Context(), search.QueryInput{Text: "docker run"})

		require.NoError(t, err)
		assert.Equal(t, want, hits)
	})

	t.Run("returns the searcher's error", func(t *testing.T) {
		t.Parallel()

		searcher := NewMockSearcher(t)
		searcher.EXPECT().Search(mock.Anything, "curl").Return(nil, errDatabaseLocked)

		_, err := newQuery(t, searcher).Run(t.Context(), search.QueryInput{Text: "curl"})

		require.ErrorIs(t, err, errDatabaseLocked)
		assert.ErrorContains(t, err, "search.Query: ")
	})
}

func newQuery(t *testing.T, searcher search.Searcher) *search.Query {
	t.Helper()

	query, err := search.NewQuery(searcher)
	require.NoError(t, err)

	return query
}
