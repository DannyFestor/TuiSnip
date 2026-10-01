package search_test

import (
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/search"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
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

	for _, blank := range []string{"", "   ", "\t\n"} {
		t.Run("returns no hits for the blank query "+strconv.Quote(blank)+" without searching", func(t *testing.T) {
			t.Parallel()

			hits, err := newQuery(t, NewMockSearcher(t)).Run(t.Context(), search.QueryInput{Text: blank})

			require.NoError(t, err)
			assert.Empty(t, hits)
			assert.NotNil(t, hits)
		})
	}

	t.Run("passes the spaces around text to the searcher", func(t *testing.T) {
		t.Parallel()

		searcher := NewMockSearcher(t)
		searcher.EXPECT().Search(mock.Anything, value.NewSearchQuery(" docker ")).Return([]domain.SearchHit{}, nil)

		_, err := newQuery(t, searcher).Run(t.Context(), search.QueryInput{Text: " docker "})

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
		searcher.EXPECT().Search(mock.Anything, value.NewSearchQuery("docker run")).Return(want, nil)

		hits, err := newQuery(t, searcher).Run(t.Context(), search.QueryInput{Text: "docker run"})

		require.NoError(t, err)
		assert.Equal(t, want, hits)
	})

	t.Run("returns the searcher's error", func(t *testing.T) {
		t.Parallel()

		searcher := NewMockSearcher(t)
		searcher.EXPECT().Search(mock.Anything, value.NewSearchQuery("curl")).Return(nil, errDatabaseLocked)

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
