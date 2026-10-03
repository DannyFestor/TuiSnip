package tui_test

import (
	"testing"

	"github.com/charmbracelet/x/exp/golden"

	"github.com/DannyFestor/TuiSnip/test/keypress"
)

func TestMainScreenLayout(t *testing.T) {
	t.Parallel()

	screen := start(t, newModel(t, listerOf(t, sampleSnippets(t)...), NewMockSnippetCopier(t)), wideWidth, wideHeight)
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

	screen := start(t, newModel(t, listerOf(t, sampleSnippets(t)...), NewMockSnippetCopier(t)), wideWidth, wideHeight)
	screen.press(keypress.Letter('n'))
	screen.press(keypress.Typed("Prune everything")...)

	golden.RequireEqual(t, []byte(screen.screen()))
}

func TestSearchPopupLayout(t *testing.T) {
	t.Parallel()

	screen := start(t, newModel(t, listerOf(t, sampleSnippets(t)...), NewMockSnippetCopier(t)), wideWidth, wideHeight)
	screen.press(keypress.Letter('/'))

	golden.RequireEqual(t, []byte(screen.screen()))
}

func TestSmallTerminalLayout(t *testing.T) {
	t.Parallel()

	screen := start(
		t,
		newModel(t, listerOf(t, sampleSnippets(t)...), NewMockSnippetCopier(t)),
		narrowWidth,
		narrowHeight,
	)
	screen.press(keypress.Letter('4'))

	golden.RequireEqual(t, []byte(screen.screen()))
}
