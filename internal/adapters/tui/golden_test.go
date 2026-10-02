package tui_test

import (
	"testing"

	"github.com/charmbracelet/x/exp/golden"
)

func TestMainScreenLayout(t *testing.T) {
	t.Parallel()

	screen := start(t, newModel(t, listerOf(t, sampleSnippets(t)...), NewMockSnippetCopier(t)), wideWidth, wideHeight)
	screen.press(letter('3'))

	golden.RequireEqual(t, []byte(screen.screen()))
}

func TestEmptyMainScreenLayout(t *testing.T) {
	t.Parallel()

	screen := start(t, newModel(t, listerOf(t), NewMockSnippetCopier(t)), wideWidth, wideHeight)

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
	screen.press(letter('4'))

	golden.RequireEqual(t, []byte(screen.screen()))
}
