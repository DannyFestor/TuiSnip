package searchpopup_test

import (
	"fmt"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/searchpopup"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/snippetpane"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/overlaytest"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

const (
	firstContent    = "srv.Shutdown"
	secondContent   = "docker system prune"
	queryInputWidth = 33
	resultRows      = 28
	popupWidth      = 96
)

func screenSize() look.Size {
	return look.Size{Width: 120, Height: 40}
}

func searching(t *testing.T) *overlaytest.Driver {
	t.Helper()

	return searchingIn(t, sampleSnippets(t))
}

func searchingIn(t *testing.T, browse []domain.Snippet) *overlaytest.Driver {
	t.Helper()

	keys := testsettings.Default(t).Keys
	styles := look.NewStyles()
	opened, _ := searchpopup.New(keys, styles, snippetpane.New(keys, styles, time.UTC), browse)

	return overlaytest.Open(t, screenSize(), opened)
}

func sampleSnippets(t *testing.T) []domain.Snippet {
	t.Helper()

	ids := testkit.NewSequentialIDs()

	return []domain.Snippet{
		testkit.Snippet(t, testkit.SnippetSpec{
			ID:          ids.NewSnippetID(),
			Title:       "Graceful HTTP shutdown",
			Description: "Stop accepting, drain, exit",
			Fragment: testkit.FragmentSpec{
				ID:       ids.NewFragmentID(),
				Language: "Go",
				Content:  "return " + firstContent + "(ctx)\n",
			},
		}),
		testkit.Snippet(t, testkit.SnippetSpec{
			ID:          ids.NewSnippetID(),
			Title:       "Prune everything",
			Description: "Reclaim disk space",
			Fragment: testkit.FragmentSpec{
				ID:       ids.NewFragmentID(),
				Language: "Bash",
				Content:  secondContent + " --all\n",
			},
		}),
	}
}

func numberedSnippets(t *testing.T, count int) []domain.Snippet {
	t.Helper()

	ids := testkit.NewSequentialIDs()
	snippets := make([]domain.Snippet, 0, count)

	for number := 1; number <= count; number++ {
		snippets = append(snippets, testkit.Snippet(t, testkit.SnippetSpec{
			ID:       ids.NewSnippetID(),
			Title:    numberedTitle(number),
			Fragment: testkit.FragmentSpec{ID: ids.NewFragmentID()},
		}))
	}

	return snippets
}

func numberedTitle(number int) string {
	return fmt.Sprintf("Snippet %02d", number)
}

func hitsOf(snippets ...domain.Snippet) []domain.SearchHit {
	hits := make([]domain.SearchHit, 0, len(snippets))
	for index := range snippets {
		hits = append(hits, domain.NewSearchHit(snippets[index], domain.FieldScores{}))
	}

	return hits
}

func down() tea.KeyPressMsg {
	return overlaytest.Special(tea.KeyDown)
}

func up() tea.KeyPressMsg {
	return overlaytest.Special(tea.KeyUp)
}

func enter() tea.KeyPressMsg {
	return overlaytest.Special(tea.KeyEnter)
}
