package mainscreen_test

import (
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/mainscreen"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/overlaytest"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

const (
	folderPaneTitle  = "1 Folders"
	tagPaneTitle     = "2 Tags"
	snippetListTitle = "3 Root · by title"
	snippetPaneTitle = "4 Snippet"
	tooSmallHint     = "Terminal too small for all four Panes (80×24)"
	listHint         = "y Copy · n new · / search"
	firstDescription = "Stop accepting, drain, exit"
	secondTitle      = "Prune everything"
	topRightCorner   = "╮"
)

func wide() look.Size {
	return look.Size{Width: 120, Height: 40}
}

func narrow() look.Size {
	return look.Size{Width: 60, Height: 20}
}

func minimum() look.Size {
	return look.Size{Width: 80, Height: 24}
}

func showing(t *testing.T, screen look.Size, snippets ...domain.Snippet) *overlaytest.Driver {
	t.Helper()

	return showingStyled(t, screen, look.NewStyles(), snippets...)
}

func showingStyled(t *testing.T, screen look.Size, styles look.Styles, snippets ...domain.Snippet) *overlaytest.Driver {
	t.Helper()

	driver := overlaytest.Open(t, screen, mainscreen.New(testsettings.Default(t).Keys, styles, time.UTC))
	driver.Send(mainscreen.SnippetsLoaded{Snippets: snippets, Selecting: domain.SnippetID{}})

	return driver
}

func upperFocusedTitle() look.Styles {
	styles := look.NewStyles()
	styles.Focused.Title = lipgloss.NewStyle().Transform(strings.ToUpper)

	return styles
}

func sampleSnippets(t *testing.T) []domain.Snippet {
	t.Helper()

	ids := testkit.NewSequentialIDs()

	return []domain.Snippet{
		testkit.Snippet(t, testkit.SnippetSpec{
			ID:          ids.NewSnippetID(),
			Title:       "Graceful HTTP shutdown",
			Description: firstDescription,
			Fragment:    testkit.FragmentSpec{ID: ids.NewFragmentID(), Language: "Go", Content: "return nil\n"},
		}),
		testkit.Snippet(t, testkit.SnippetSpec{
			ID:          ids.NewSnippetID(),
			Title:       secondTitle,
			Description: "Reclaim disk space",
			Fragment: testkit.FragmentSpec{
				ID:       ids.NewFragmentID(),
				Language: "Bash",
				Content:  "docker system prune\n",
			},
		}),
	}
}

func lines(driver *overlaytest.Driver) []string {
	return strings.Split(driver.Screen(), "\n")
}

func statusLine(driver *overlaytest.Driver) string {
	all := lines(driver)

	return all[len(all)-1]
}

func columnWidths(driver *overlaytest.Driver) []int {
	frames := strings.SplitAfter(lines(driver)[0], topRightCorner)
	widths := make([]int, 0, len(frames))

	for _, frame := range frames {
		if frame != "" {
			widths = append(widths, ansi.StringWidth(frame))
		}
	}

	return widths
}

func lineIndexOf(driver *overlaytest.Driver, text string) int {
	for index, line := range lines(driver) {
		if strings.Contains(line, text) {
			return index
		}
	}

	return -1
}
