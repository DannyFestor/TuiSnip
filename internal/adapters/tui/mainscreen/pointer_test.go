package mainscreen_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/browseselection"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/mainscreen"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/pointer"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/foldertree"
	"github.com/DannyFestor/TuiSnip/test/keypress"
)

func TestScreen_click(t *testing.T) {
	t.Parallel()

	t.Run("focuses the clicked Pane", func(t *testing.T) {
		t.Parallel()

		screen := showingStyled(t, wide(), upperFocusedTitle(), sampleSnippets(t)...)

		screen.Click(firstDescription)

		assert.Contains(t, screen.Screen(), "4 SNIPPET")
		assert.Empty(t, screen.Outcomes())
	})

	t.Run("makes a clicked Folder the Browse selection", func(t *testing.T) {
		t.Parallel()

		screen, sample := browsing(t)

		screen.Click("docker")

		assert.Equal(t, []outcome.Outcome{outcome.FolderSelected{ID: sample.Docker.ID()}}, screen.Outcomes())
		assert.Contains(t, screen.Screen(), "3 Root / docker · by title")
	})

	t.Run("selects nothing again on the Folder that holds the selection", func(t *testing.T) {
		t.Parallel()

		screen, _ := browsing(t)

		screen.Click("Root")

		assert.Empty(t, screen.Outcomes())
	})

	t.Run("makes a clicked Tag the Browse selection", func(t *testing.T) {
		t.Parallel()

		screen, tags := browsingTags(t)

		screen.Click("# go")

		assert.Equal(t, []outcome.Outcome{outcome.TagSelected{ID: tags[1].Tag.ID()}}, screen.Outcomes())
		assert.Contains(t, screen.Screen(), "3 # go · by title")
	})

	t.Run("takes the selection back from a Tag on the Folder under the cursor", func(t *testing.T) {
		t.Parallel()

		screen, tags := browsingTags(t)
		screen.Click("# go")

		screen.Click("Root")

		assert.Equal(t, []outcome.Outcome{
			outcome.TagSelected{ID: tags[1].Tag.ID()},
			outcome.FolderSelected{ID: domain.FolderID{}},
		}, screen.Outcomes())
	})

	t.Run("collapses a Folder whose marker is clicked and selects it", func(t *testing.T) {
		t.Parallel()

		screen, sample := browsing(t)

		screen.Click("▾")

		assert.Equal(t, []outcome.Outcome{
			outcome.CollapsedFoldersChanged{IDs: []domain.FolderID{sample.Go.ID()}},
			outcome.FolderSelected{ID: sample.Go.ID()},
		}, screen.Outcomes())
		assert.Contains(t, screen.Screen(), "▸ go")
	})

	t.Run("shows a clicked Snippet in the Snippet pane", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, wide(), sampleSnippets(t)...)

		screen.Click(secondTitle)

		assert.Contains(t, screen.Screen(), "Reclaim disk space")
	})

	t.Run("does nothing on the status line", func(t *testing.T) {
		t.Parallel()

		screen := showingStyled(t, wide(), upperFocusedTitle(), sampleSnippets(t)...)

		screen.Click("? help")

		assert.Contains(t, screen.Screen(), "1 FOLDERS")
	})

	t.Run("does nothing while a Folder name is typed", func(t *testing.T) {
		t.Parallel()

		screen := showingStyled(t, wide(), upperFocusedTitle(), sampleSnippets(t)...)
		screen.Press(keypress.Letter('N'))

		screen.Click(firstDescription)

		assert.Contains(t, screen.Screen(), "1 FOLDERS")
	})

	t.Run("clicks the only Pane shown below the minimum size", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		screen := showing(t, narrow(), sampleSnippets(t)...)
		screen.Send(mainscreen.TreeLoaded{Tree: sample.Tree})

		screen.Click("docker")

		assert.Equal(t, []outcome.Outcome{outcome.FolderSelected{ID: sample.Docker.ID()}}, screen.Outcomes())
	})
}

func TestScreen_doubleClick(t *testing.T) {
	t.Parallel()

	t.Run("opens a Folder and focuses the Snippet list", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		screen := showingStyled(t, wide(), upperFocusedTitle(), sampleSnippets(t)...)
		screen.Send(mainscreen.TreeLoaded{Tree: sample.Tree})

		screen.DoubleClick("docker")

		assert.Equal(t, []outcome.Outcome{outcome.FolderSelected{ID: sample.Docker.ID()}}, screen.Outcomes())
		assert.Contains(t, screen.Screen(), strings.ToUpper("3 Root / docker · by title"))
	})

	t.Run("focuses the Snippet pane from the Snippet list", func(t *testing.T) {
		t.Parallel()

		screen := showingStyled(t, wide(), upperFocusedTitle(), sampleSnippets(t)...)

		screen.DoubleClick(secondTitle)

		assert.Contains(t, screen.Screen(), "4 SNIPPET")
		assert.Contains(t, screen.Screen(), "Reclaim disk space")
	})

	t.Run("only focuses the Pane below its last row", func(t *testing.T) {
		t.Parallel()

		screen := showingStyled(t, wide(), upperFocusedTitle(), sampleSnippets(t)...)
		below := screen.CellOf(secondTitle)
		below.Y++

		screen.Send(pointer.Clicked{At: below, Double: true})

		assert.Contains(t, screen.Screen(), strings.ToUpper(snippetListTitle))
	})
}

func TestScreen_wheel(t *testing.T) {
	t.Parallel()

	t.Run("scrolls the Pane under the pointer without moving focus", func(t *testing.T) {
		t.Parallel()

		screen := showingStyled(t, wide(), upperFocusedTitle(), manySnippets(t, wide().Height)...)

		screen.Send(pointer.Wheeled{At: screen.CellOf("Snippet 02"), Lines: 3})

		assert.Equal(t, 1, lineIndexOf(screen, "Snippet 04"))
		assert.Contains(t, screen.Screen(), "1 FOLDERS")
	})

	t.Run("keeps the scroll when another Pane moves its cursor", func(t *testing.T) {
		t.Parallel()

		screen := opened(t, wide(), look.NewStyles(look.SchemeDark), domain.SortOrderTitle)
		screen.Send(mainscreen.SnippetsLoaded{
			Selection: browseselection.Selection{},
			Snippets:  manySnippets(t, wide().Height),
			Selecting: domain.SnippetID{},
			Order:     domain.SortOrderTitle,
		})
		screen.Send(pointer.Wheeled{At: screen.CellOf("Snippet 02"), Lines: 3})

		screen.Press(keypress.Letter('j'))

		assert.Equal(t, 1, lineIndexOf(screen, "Snippet 04"))
	})
}

func manySnippets(t *testing.T, count int) []domain.Snippet {
	t.Helper()

	ids := testkit.NewSequentialIDs()
	snippets := make([]domain.Snippet, 0, count)

	for number := 1; number <= count; number++ {
		snippets = append(snippets, testkit.Snippet(t, testkit.SnippetSpec{
			ID:       ids.NewSnippetID(),
			Title:    fmt.Sprintf("Snippet %02d", number),
			Fragment: testkit.FragmentSpec{ID: ids.NewFragmentID(), Language: "Go"},
		}))
	}

	return snippets
}
