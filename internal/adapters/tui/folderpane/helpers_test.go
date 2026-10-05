package folderpane_test

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/folderpane"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/test/foldertree"
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

	pane := folderpane.New(testsettings.Default(t).Keys, nil)
	pane, _, _ = pane.Update(look.Resized{Box: box})

	return pane
}

func withTree(pane folderpane.Pane, tree browse.Tree) folderpane.Pane {
	pane, _ = pane.WithTree(tree)

	return pane
}

func withCursorOn(pane folderpane.Pane, id domain.FolderID) folderpane.Pane {
	pane, _ = pane.WithCursorOn(id)

	return pane
}

func samplePane(t *testing.T, sample foldertree.Sample) folderpane.Pane {
	t.Helper()

	return withTree(paneIn(t, look.Size{Width: boxWidth, Height: boxHeight}), sample.Tree)
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
