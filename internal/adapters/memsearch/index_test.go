package memsearch_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/memsearch"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
)

const concurrentSearches = 8

var errDatabaseLocked = errors.New("database is locked")

func TestIndex_Search(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		snippets  []testkit.SnippetSpec
		query     string
		wantOrder []string
	}{
		{
			name: "ranks a title match above the same match in a Description",
			snippets: []testkit.SnippetSpec{
				{Title: "notes", Description: "docker"},
				{Title: "docker"},
			},
			query:     "docker",
			wantOrder: []string{"docker", "notes"},
		},
		{
			name:      "matches the title fuzzily",
			snippets:  []testkit.SnippetSpec{{Title: "docker container run"}},
			query:     "dkrrun",
			wantOrder: []string{"docker container run"},
		},
		{
			name:      "ignores diacritics in the title",
			snippets:  []testkit.SnippetSpec{{Title: "Café menu"}},
			query:     "cafe",
			wantOrder: []string{"Café menu"},
		},
		{
			name:      "ignores diacritics in the query",
			snippets:  []testkit.SnippetSpec{{Title: "Cafe menu"}},
			query:     "café",
			wantOrder: []string{"Cafe menu"},
		},
		{
			name:      "ignores diacritics in the Description",
			snippets:  []testkit.SnippetSpec{{Title: "menu", Description: "for the café"}},
			query:     "cafe",
			wantOrder: []string{"menu"},
		},
		{
			name:      "matches content as a literal substring",
			snippets:  []testkit.SnippetSpec{{Title: "a", Fragment: testkit.FragmentSpec{Content: "kubectl get pods"}}},
			query:     "get pods",
			wantOrder: []string{"a"},
		},
		{
			name:      "does not match content fuzzily",
			snippets:  []testkit.SnippetSpec{{Title: "a", Fragment: testkit.FragmentSpec{Content: "kubectl get pods"}}},
			query:     "kgp",
			wantOrder: []string{},
		},
		{
			name:      "keeps diacritics in content",
			snippets:  []testkit.SnippetSpec{{Title: "a", Fragment: testkit.FragmentSpec{Content: "echo café"}}},
			query:     "cafe",
			wantOrder: []string{},
		},
		{
			name:      "matches case-insensitively for a lowercase query",
			snippets:  []testkit.SnippetSpec{{Title: "Docker", Fragment: testkit.FragmentSpec{Content: "FROM alpine"}}},
			query:     "from",
			wantOrder: []string{"Docker"},
		},
		{
			name:      "matches case-sensitively once the query has an uppercase letter",
			snippets:  []testkit.SnippetSpec{{Title: "docker"}, {Title: "Docker"}},
			query:     "Docker",
			wantOrder: []string{"Docker"},
		},
		{
			name:      "treats the whole query, spaces included, as one pattern",
			snippets:  []testkit.SnippetSpec{{Title: "docker container run"}, {Title: "run docker"}},
			query:     "docker run",
			wantOrder: []string{"docker container run"},
		},
		{
			name:      "returns no hits when nothing matches",
			snippets:  []testkit.SnippetSpec{{Title: "curl json"}},
			query:     "zzz",
			wantOrder: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			index := memsearch.NewIndex(listerOf(t, buildSnippets(t, tt.snippets)))

			hits, err := index.Search(t.Context(), value.NewSearchQuery(tt.query))

			require.NoError(t, err)
			assert.Equal(t, tt.wantOrder, hitTitles(hits))
		})
	}
}

func TestIndex_SearchBreaksTiesByMostRecentlyUpdated(t *testing.T) {
	t.Parallel()

	older := time.Date(2026, time.March, 1, 12, 0, 0, 0, time.UTC)
	snippets := buildSnippets(t, []testkit.SnippetSpec{
		{Title: "curl older", UpdatedAt: older},
		{Title: "curl newer", UpdatedAt: older.Add(time.Hour)},
	})

	hits, err := memsearch.NewIndex(listerOf(t, snippets)).Search(t.Context(), value.NewSearchQuery("curl"))

	require.NoError(t, err)
	assert.Equal(t, []string{"curl newer", "curl older"}, hitTitles(hits))
}

func TestIndex_SearchReturnsTheListError(t *testing.T) {
	t.Parallel()

	lister := NewMockSnippetLister(t)
	lister.EXPECT().List(mock.Anything).Return(nil, errDatabaseLocked)

	_, err := memsearch.NewIndex(lister).Search(t.Context(), value.NewSearchQuery("curl"))

	require.ErrorIs(t, err, errDatabaseLocked)
	assert.ErrorContains(t, err, "memsearch.Index.Search: ")
}

func TestIndex_SearchConcurrently(t *testing.T) {
	t.Parallel()

	index := memsearch.NewIndex(listerOf(t, buildSnippets(t, []testkit.SnippetSpec{{Title: "curl json"}})))

	var wg sync.WaitGroup
	for range concurrentSearches {
		wg.Go(func() {
			hits, err := index.Search(t.Context(), value.NewSearchQuery("curl"))
			if assert.NoError(t, err) {
				assert.Len(t, hits, 1)
			}
		})
	}

	wg.Wait()
}

func FuzzIndexSearch(f *testing.F) {
	f.Add("curl", "curl json", "POST a body", "curl -d @body.json")
	f.Add("Café", "cafe", "", "é\x00\xff")
	f.Add(" ", "", "", "")

	f.Fuzz(func(t *testing.T, query, title, description, content string) {
		snippet, ok := fuzzSnippet(t, title, description, content)
		if !ok {
			t.Skip()
		}

		_, err := memsearch.NewIndex(listerOf(t, []domain.Snippet{snippet})).
			Search(t.Context(), value.NewSearchQuery(query))

		assert.NoError(t, err)
	})
}

func fuzzSnippet(t *testing.T, title, description, content string) (domain.Snippet, bool) {
	t.Helper()

	_, titleErr := value.NewTitle(title)
	_, descriptionErr := value.NewDescription(description)
	_, contentErr := value.NewContent(content)

	if errors.Join(titleErr, descriptionErr, contentErr) != nil {
		return domain.Snippet{}, false
	}

	spec := testkit.SnippetSpec{
		Title:       title,
		Description: description,
		Fragment:    testkit.FragmentSpec{Content: content},
	}

	return testkit.Snippet(t, spec), true
}

func listerOf(t *testing.T, snippets []domain.Snippet) *MockSnippetLister {
	t.Helper()

	lister := NewMockSnippetLister(t)
	lister.EXPECT().List(mock.Anything).Return(snippets, nil).Maybe()

	return lister
}

func buildSnippets(t *testing.T, specs []testkit.SnippetSpec) []domain.Snippet {
	t.Helper()

	ids := testkit.NewSequentialIDs()
	snippets := make([]domain.Snippet, 0, len(specs))

	for i := range specs {
		spec := specs[i]
		spec.ID = ids.NewSnippetID()
		spec.Fragment.ID = ids.NewFragmentID()
		snippets = append(snippets, testkit.Snippet(t, spec))
	}

	return snippets
}

func hitTitles(hits []domain.SearchHit) []string {
	titles := make([]string, 0, len(hits))
	for i := range hits {
		titles = append(titles, hits[i].Snippet().Title().String())
	}

	return titles
}
