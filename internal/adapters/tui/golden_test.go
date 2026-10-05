package tui_test

import (
	"testing"

	"github.com/charmbracelet/x/exp/golden"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui"
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

func TestSearchPopupLayout(t *testing.T) {
	t.Parallel()

	screen := start(t, layoutModel(t, listerOf(t, sampleSnippets(t)...)), wideWidth, wideHeight)
	screen.press(keypress.Letter('/'))

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

func layoutModel(t *testing.T, lister *MockFolderSnippetsLister) tui.Model {
	t.Helper()

	return browsingModel(t, foldertree.New(t), lister)
}
