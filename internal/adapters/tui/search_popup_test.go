package tui_test

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/mainscreen"
	"github.com/DannyFestor/TuiSnip/internal/app/search"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/test/foldertree"
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
			lister:     listerReturning(t, snippets, snippets),
			treeLister: treeOf(t, emptyTree()),
			copier:     NewMockSnippetCopier(t),
			creator:    NewMockSnippetCreator(t),
			searcher:   NewMockSnippetSearcher(t),
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
			lister:     listerOf(t, snippets...),
			treeLister: treeOf(t, emptyTree()),
			copier:     copier,
			creator:    NewMockSnippetCreator(t),
			searcher:   NewMockSnippetSearcher(t),
		}), wideWidth, wideHeight)

		screen.press(keypress.Letter('/'), keypress.Special(tea.KeyDown), keypress.Ctrl('y'))

		assert.NotContains(t, screen.screen(), searchTitle)
		assert.Contains(t, screen.screen(), "Copied")
	})
}

func TestModel_searchPopupFolderPaths(t *testing.T) {
	t.Parallel()

	sample := foldertree.New(t)

	tests := []struct {
		name    string
		arrived tea.Msg
	}{
		{name: "shows the Folder paths once the Folder tree loads", arrived: mainscreen.TreeLoaded{Tree: sample.Tree}},
		{
			name:    "shows the Folder paths once the Folder tree changes",
			arrived: mainscreen.TreeChanged{Tree: sample.Tree},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			screen := start(t, modelWith(t, actions{
				lister:     listerOf(t, filedIn(t, sample.Tests.ID())),
				treeLister: treeOf(t, emptyTree()),
				copier:     NewMockSnippetCopier(t),
				creator:    NewMockSnippetCreator(t),
				searcher:   NewMockSnippetSearcher(t),
			}), wideWidth, wideHeight)
			screen.press(keypress.Letter('/'))

			screen.send(tt.arrived)

			assert.Regexp(t, `Table test skeleton +go / testing│`, screen.screen())
			assert.Equal(
				t,
				2,
				strings.Count(screen.screen(), "Root / go / testing · Go"),
				"main screen and popup headers",
			)
		})
	}
}

func searchingModel(t *testing.T, searcher *MockSnippetSearcher) tui.Model {
	t.Helper()

	return modelWith(t, actions{
		lister:     listerOf(t, sampleSnippets(t)...),
		treeLister: treeOf(t, emptyTree()),
		copier:     NewMockSnippetCopier(t),
		creator:    NewMockSnippetCreator(t),
		searcher:   searcher,
	})
}

func hitsOf(snippets ...domain.Snippet) []domain.SearchHit {
	hits := make([]domain.SearchHit, 0, len(snippets))
	for index := range snippets {
		hits = append(hits, domain.NewSearchHit(snippets[index], domain.FieldScores{}))
	}

	return hits
}
