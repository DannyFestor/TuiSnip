package mainscreen_test

import (
	"strings"
	"testing"
	"time"
	"uuid"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/browseselection"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/mainscreen"
	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/foldertree"
	"github.com/DannyFestor/TuiSnip/test/overlaytest"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

const (
	folderPaneTitle  = "1 Folders"
	tagPaneTitle     = "2 Tags"
	snippetListTitle = "3 Root · by title"
	snippetPaneTitle = "4 Snippet"
	tooSmallHint     = "Terminal too small for all four Panes (80×24)"
	listHint         = "y Copy · e edit · n new · s sort · z zoom · / search"
	firstDescription = "Stop accepting, drain, exit"
	secondTitle      = "Prune everything"
	filedTitle       = "Table test skeleton"
	filedSnippetID   = "0194c3a0-0000-7000-8000-0000000f11ed"
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

	driver := opened(t, screen, styles, domain.SortOrderTitle)
	driver.Send(mainscreen.SnippetsLoaded{
		Selection: browseselection.Selection{},
		Snippets:  snippets,
		Selecting: domain.SnippetID{},
		Order:     domain.SortOrderTitle,
	})

	return driver
}

func opened(t *testing.T, screen look.Size, styles look.Styles, order domain.SortOrder) *overlaytest.Driver {
	t.Helper()

	return overlaytest.Open(t, screen, newScreen(t, styles, order))
}

func newScreen(t *testing.T, styles look.Styles, order domain.SortOrder) mainscreen.Screen {
	t.Helper()

	screen, err := mainscreen.New(
		testsettings.Default(t).Keys,
		styles,
		time.UTC,
		mainscreen.Remembered{SortOrder: order, CollapsedFolders: nil},
	)
	require.NoError(t, err)

	return screen
}

func browsing(t *testing.T) (*overlaytest.Driver, foldertree.Sample) {
	t.Helper()

	sample := foldertree.New(t)
	driver := showing(t, wide(), sampleSnippets(t)...)
	driver.Send(mainscreen.TreeLoaded{Tree: sample.Tree})

	return driver, sample
}

func browsingTags(t *testing.T) (*overlaytest.Driver, []browse.TagCount) {
	t.Helper()

	ids := testkit.NewSequentialIDs()
	tags := []browse.TagCount{
		{Tag: testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "docker"}), SnippetCount: 3},
		{Tag: testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "go"}), SnippetCount: 4},
	}

	return browsingWithTags(t, tags...), tags
}

func browsingWithTags(t *testing.T, tags ...browse.TagCount) *overlaytest.Driver {
	t.Helper()

	driver := showingStyled(t, wide(), upperFocusedTitle(), sampleSnippets(t)...)
	driver.Send(mainscreen.TagsLoaded{Tags: tags})

	return driver
}

func filedIn(t *testing.T, folderID domain.FolderID) domain.Snippet {
	t.Helper()

	return testkit.Snippet(t, testkit.SnippetSpec{
		ID:       domain.SnippetID(uuid.MustParse(filedSnippetID)),
		Title:    filedTitle,
		FolderID: folderID,
		Fragment: testkit.FragmentSpec{Language: "Go"},
	})
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
