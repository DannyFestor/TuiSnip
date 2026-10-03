package tui_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui"
	"github.com/DannyFestor/TuiSnip/internal/app/search"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/test/keypress"
)

const searchTitle = "Search · "

func TestModel_searchPopup(t *testing.T) {
	t.Parallel()

	t.Run("lists the hits for the query", func(t *testing.T) {
		t.Parallel()

		snippets := sampleSnippets(t)
		searcher := NewMockSnippetSearcher(t)
		searcher.EXPECT().Run(mock.Anything, search.QueryInput{Text: "p"}).Return(hitsOf(snippets[1]), nil)
		screen := start(t, searchingModel(t, searcher), wideWidth, wideHeight)

		screen.press(keypress.Letter('/'), keypress.Letter('p'))

		assert.Contains(t, screen.screen(), "Search · 1 result")
		assert.Contains(t, screen.screen(), "docker system prune")
	})

	t.Run("reports a search failure", func(t *testing.T) {
		t.Parallel()

		searcher := NewMockSnippetSearcher(t)
		searcher.EXPECT().Run(mock.Anything, mock.Anything).Return(nil, errDatabaseLocked)
		screen := start(t, searchingModel(t, searcher), wideWidth, wideHeight)

		screen.press(keypress.Letter('/'), keypress.Letter('x'))

		assert.Contains(t, screen.screen(), "Something went wrong; see the log")
	})

	t.Run("enter selects the Snippet and focuses the Snippet pane", func(t *testing.T) {
		t.Parallel()

		snippets := sampleSnippets(t)
		screen := start(t, modelWith(t, actions{
			lister:   listerReturning(t, snippets, snippets),
			copier:   NewMockSnippetCopier(t),
			creator:  NewMockSnippetCreator(t),
			searcher: NewMockSnippetSearcher(t),
		}), narrowWidth, narrowHeight)

		screen.press(keypress.Letter('/'), keypress.Special(tea.KeyDown), keypress.Special(tea.KeyEnter))

		assert.NotContains(t, screen.screen(), searchTitle)
		assert.Contains(t, screen.screen(), snippetPaneTitle)
		assert.Contains(t, screen.screen(), "Reclaim disk space")
	})

	t.Run("ctrl+y copies the highlighted result", func(t *testing.T) {
		t.Parallel()

		snippets := sampleSnippets(t)
		copier := NewMockSnippetCopier(t)
		copier.EXPECT().
			Run(mock.Anything, snippet.CopyInput{SnippetID: snippets[1].ID()}).
			Return(copied(t, domain.CopyDeliveryPlaced), nil)
		screen := start(t, modelWith(t, actions{
			lister:   listerOf(t, snippets...),
			copier:   copier,
			creator:  NewMockSnippetCreator(t),
			searcher: NewMockSnippetSearcher(t),
		}), wideWidth, wideHeight)

		screen.press(keypress.Letter('/'), keypress.Special(tea.KeyDown), keypress.Ctrl('y'))

		assert.NotContains(t, screen.screen(), searchTitle)
		assert.Contains(t, screen.screen(), "Copied")
	})
}

func searchingModel(t *testing.T, searcher *MockSnippetSearcher) tui.Model {
	t.Helper()

	return modelWith(t, actions{
		lister:   listerOf(t, sampleSnippets(t)...),
		copier:   NewMockSnippetCopier(t),
		creator:  NewMockSnippetCreator(t),
		searcher: searcher,
	})
}

func hitsOf(snippets ...domain.Snippet) []domain.SearchHit {
	hits := make([]domain.SearchHit, 0, len(snippets))
	for index := range snippets {
		hits = append(hits, domain.NewSearchHit(snippets[index], domain.FieldScores{}))
	}

	return hits
}
