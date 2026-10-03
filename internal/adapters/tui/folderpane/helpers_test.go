package folderpane_test

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/folderpane"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

const (
	boxWidth  = 16
	boxHeight = 3
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
