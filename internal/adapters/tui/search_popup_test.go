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
)

const searchTitle = "Search · "

func TestModel_searchPopup(t *testing.T) {
	t.Parallel()

	t.Run("/ lists the Browse selection and previews the first", func(t *testing.T) {
		t.Parallel()

		screen := start(t, searchingModel(t, NewMockSnippetSearcher(t)), wideWidth, wideHeight)

		screen.press(letter('/'))

		assert.Contains(t, screen.screen(), "Search · 2 results")
		assert.Contains(t, screen.screen(), "Preview")
		assert.Contains(t, screen.screen(), "down move · enter reveal · ctrl+y Copy · esc close")
	})

	t.Run("lists the hits for the query and previews the first", func(t *testing.T) {
		t.Parallel()

		snippets := sampleSnippets(t)
		searcher := NewMockSnippetSearcher(t)
		searcher.EXPECT().Run(mock.Anything, search.QueryInput{Text: "p"}).Return(hitsOf(snippets[1]), nil)
		screen := start(t, searchingModel(t, searcher), wideWidth, wideHeight)

		screen.press(letter('/'), letter('p'))

		assert.Contains(t, screen.screen(), "Search · 1 result")
		assert.Contains(t, screen.screen(), "docker system prune")
	})

	t.Run("searches a pasted query", func(t *testing.T) {
		t.Parallel()

		searcher := NewMockSnippetSearcher(t)
		searcher.EXPECT().Run(mock.Anything, search.QueryInput{Text: "prune"}).Return(nil, nil)
		screen := start(t, searchingModel(t, searcher), wideWidth, wideHeight)

		screen.press(letter('/'))
		screen.send(tea.PasteMsg{Content: "prune"})

		assert.Contains(t, screen.screen(), "No Snippets match.")
	})

	t.Run("types Pane keys into the query instead of acting on them", func(t *testing.T) {
		t.Parallel()

		searcher := NewMockSnippetSearcher(t)
		searcher.EXPECT().Run(mock.Anything, search.QueryInput{Text: "q"}).Return(nil, nil)
		screen := start(t, searchingModel(t, searcher), wideWidth, wideHeight)

		screen.press(letter('/'), letter('q'))

		assert.NotContains(t, screen.emitted, tea.QuitMsg{})
	})

	t.Run("lists the Browse selection again for a blank query without searching", func(t *testing.T) {
		t.Parallel()

		snippets := sampleSnippets(t)
		searcher := NewMockSnippetSearcher(t)
		searcher.EXPECT().Run(mock.Anything, search.QueryInput{Text: "p"}).Return(nil, nil)
		screen := start(t, searchingModel(t, searcher), wideWidth, wideHeight)

		screen.press(letter('/'), letter('p'), special(tea.KeyBackspace), letter(' '))

		assert.Contains(t, screen.screen(), "Search · 2 results")
		assert.Contains(t, screen.screen(), snippets[0].Description().String())
	})

	t.Run("moving down previews the next result", func(t *testing.T) {
		t.Parallel()

		screen := start(t, searchingModel(t, NewMockSnippetSearcher(t)), wideWidth, wideHeight)

		screen.press(letter('/'), ctrl('n'))

		assert.Contains(t, screen.screen(), "docker system prune")
	})

	t.Run("moving up stops at the first result", func(t *testing.T) {
		t.Parallel()

		screen := start(t, searchingModel(t, NewMockSnippetSearcher(t)), wideWidth, wideHeight)

		screen.press(letter('/'), special(tea.KeyDown), special(tea.KeyUp), special(tea.KeyUp))

		assert.Contains(t, screen.screen(), "srv.Shutdown")
	})

	t.Run("reports a search failure", func(t *testing.T) {
		t.Parallel()

		searcher := NewMockSnippetSearcher(t)
		searcher.EXPECT().Run(mock.Anything, mock.Anything).Return(nil, errDatabaseLocked)
		screen := start(t, searchingModel(t, searcher), wideWidth, wideHeight)

		screen.press(letter('/'), letter('x'))

		assert.Contains(t, screen.screen(), "Something went wrong; see the log")
	})

	t.Run("esc closes and changes nothing", func(t *testing.T) {
		t.Parallel()

		screen := start(t, searchingModel(t, NewMockSnippetSearcher(t)), wideWidth, wideHeight)

		screen.press(letter('/'), special(tea.KeyDown), special(tea.KeyEscape))

		assert.NotContains(t, screen.screen(), searchTitle)
		assert.Contains(t, screen.screen(), "Stop accepting, drain, exit")
	})
}

func TestModel_searchPopupReveal(t *testing.T) {
	t.Parallel()

	t.Run("enter selects the Snippet and focuses the Snippet pane", func(t *testing.T) {
		t.Parallel()

		snippets := sampleSnippets(t)
		screen := start(t, modelWith(t, actions{
			lister:   listerReturning(t, snippets, snippets),
			copier:   NewMockSnippetCopier(t),
			creator:  NewMockSnippetCreator(t),
			searcher: NewMockSnippetSearcher(t),
		}), narrowWidth, narrowHeight)

		screen.press(letter('/'), special(tea.KeyDown), special(tea.KeyEnter))

		assert.NotContains(t, screen.screen(), searchTitle)
		assert.Contains(t, screen.screen(), snippetPaneTitle)
		assert.Contains(t, screen.screen(), "Reclaim disk space")
	})

	t.Run("enter moves the list cursor to the first Snippet", func(t *testing.T) {
		t.Parallel()

		snippets := sampleSnippets(t)
		screen := start(t, modelWith(t, actions{
			lister:   listerReturning(t, snippets, snippets),
			copier:   NewMockSnippetCopier(t),
			creator:  NewMockSnippetCreator(t),
			searcher: NewMockSnippetSearcher(t),
		}), narrowWidth, narrowHeight)

		screen.press(letter('3'), letter('j'), letter('/'), special(tea.KeyEnter))

		assert.Contains(t, screen.screen(), "Stop accepting, drain, exit")
	})

	t.Run("enter does nothing without results", func(t *testing.T) {
		t.Parallel()

		searcher := NewMockSnippetSearcher(t)
		searcher.EXPECT().Run(mock.Anything, mock.Anything).Return(nil, nil)
		screen := start(t, searchingModel(t, searcher), wideWidth, wideHeight)

		screen.press(letter('/'), letter('x'), special(tea.KeyEnter))

		assert.Contains(t, screen.screen(), searchTitle)
	})
}

func TestModel_searchPopupCopy(t *testing.T) {
	t.Parallel()

	t.Run("ctrl+y copies the highlighted result and closes", func(t *testing.T) {
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

		screen.press(letter('/'), special(tea.KeyDown), ctrl('y'))

		assert.NotContains(t, screen.screen(), searchTitle)
		assert.Contains(t, screen.screen(), "Copied")
	})

	t.Run("ctrl+y does nothing without results", func(t *testing.T) {
		t.Parallel()

		searcher := NewMockSnippetSearcher(t)
		searcher.EXPECT().Run(mock.Anything, mock.Anything).Return(nil, nil)
		screen := start(t, searchingModel(t, searcher), wideWidth, wideHeight)

		screen.press(letter('/'), letter('x'), ctrl('y'))

		assert.Contains(t, screen.screen(), searchTitle)
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
