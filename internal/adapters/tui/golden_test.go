package tui_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/exp/golden"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/test/foldertree"
	"github.com/DannyFestor/TuiSnip/test/keypress"
)

func TestMainScreenLayout(t *testing.T) {
	t.Parallel()

	screen := start(t, layoutModel(t, listerOf(t, sampleSnippets(t)...)), wideWidth, wideHeight)
	screen.press(keypress.Letter('3'))

	golden.RequireEqual(t, []byte(screen.screen()))
}

func TestEmptyMainScreenLayout(t *testing.T) {
	t.Parallel()

	screen := start(t, newModel(t, listerOf(t), NewMockSnippetCopier(t)), wideWidth, wideHeight)

	golden.RequireEqual(t, []byte(screen.screen()))
}

func TestEditOverlayLayout(t *testing.T) {
	t.Parallel()

	screen := start(t, layoutModel(t, listerOf(t, sampleSnippets(t)...)), wideWidth, wideHeight)
	screen.press(keypress.Letter('n'))
	screen.press(keypress.Typed("Prune everything")...)

	golden.RequireEqual(t, []byte(screen.screen()))
}

func TestTagEditorLayout(t *testing.T) {
	t.Parallel()

	screen := start(t, layoutModel(t, listerOf(t, sampleSnippets(t)...)), wideWidth, wideHeight)
	screen.press(keypress.Letter('n'), keypress.Ctrl('t'), keypress.Special(tea.KeyEnter))
	screen.press(keypress.Typed("o")...)

	golden.RequireEqual(t, []byte(screen.screen()))
}

func TestSearchPopupLayout(t *testing.T) {
	t.Parallel()

	screen := start(t, layoutModel(t, listerOf(t, sampleSnippets(t)...)), wideWidth, wideHeight)
	screen.press(keypress.Letter('/'))

	golden.RequireEqual(t, []byte(screen.screen()))
}

func TestHelpOverlayLayout(t *testing.T) {
	t.Parallel()

	screen := start(t, layoutModel(t, listerOf(t, sampleSnippets(t)...)), wideWidth, wideHeight)
	screen.press(keypress.Letter('3'))
	screen.press(keypress.Letter('?'))

	golden.RequireEqual(t, []byte(screen.screen()))
}

func TestSmallTerminalLayout(t *testing.T) {
	t.Parallel()

	screen := start(t, layoutModel(t, listerOf(t, sampleSnippets(t)...)), narrowWidth, narrowHeight)
	screen.press(keypress.Letter('4'))

	golden.RequireEqual(t, []byte(screen.screen()))
}

func TestZoomedPaneLayout(t *testing.T) {
	t.Parallel()

	screen := start(t, layoutModel(t, listerOf(t, sampleSnippets(t)...)), wideWidth, wideHeight)
	screen.press(keypress.Letter('3'), keypress.Letter('z'))

	golden.RequireEqual(t, []byte(screen.screen()))
}

func TestWrappedSnippetPaneLayout(t *testing.T) {
	t.Parallel()

	screen := start(t, layoutModel(t, listerOf(t, longLineSnippet(t))), wideWidth, wideHeight)
	screen.press(keypress.Letter('4'), keypress.Letter('w'))

	golden.RequireEqual(t, []byte(screen.screen()))
}

func TestDarkSchemeLayout(t *testing.T) {
	t.Parallel()

	golden.RequireEqual(t, []byte(themed(t, look.ThemeAuto).styledScreen()))
}

func TestLightSchemeLayout(t *testing.T) {
	t.Parallel()

	golden.RequireEqual(t, []byte(themed(t, look.ThemeLight).styledScreen()))
}

func layoutModel(t *testing.T, lister *MockFolderSnippetsLister) tui.Model {
	t.Helper()

	with := browsingActions(t, foldertree.New(t), lister)
	with.tagLister = tagsOf(t, sampleTagCounts(t)...)

	return modelWith(t, with)
}
