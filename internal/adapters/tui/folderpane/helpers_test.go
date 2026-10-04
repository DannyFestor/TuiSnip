package folderpane_test

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/folderpane"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

const (
	boxWidth  = 23
	boxHeight = 5
)

func upperCursor() look.FrameStyle {
	return look.FrameStyle{
		Border: lipgloss.NewStyle(),
		Title:  lipgloss.NewStyle(),
		Cursor: lipgloss.NewStyle().Transform(strings.ToUpper),
	}
}

func paneIn(t *testing.T, box look.Size) folderpane.Pane {
	t.Helper()

	pane := folderpane.New(testsettings.Default(t).Keys)
	pane, _, _ = pane.Update(look.Resized{Box: box})

	return pane
}

func pressed(pane folderpane.Pane, keys ...tea.KeyPressMsg) folderpane.Pane {
	for _, key := range keys {
		pane, _, _ = pane.Update(key)
	}

	return pane
}

func viewLines(pane folderpane.Pane) []string {
	return strings.Split(ansi.Strip(pane.View(upperCursor())), "\n")
}
