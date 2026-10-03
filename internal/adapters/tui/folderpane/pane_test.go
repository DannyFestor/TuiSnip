package folderpane_test

import (
	"testing"

	"charm.land/bubbles/v2/key"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

func TestPane_View(t *testing.T) {
	t.Parallel()

	t.Run("shows the Root with no Snippets before a count arrives", func(t *testing.T) {
		t.Parallel()

		pane := paneIn(t, look.Size{Width: boxWidth, Height: boxHeight})

		assert.Equal(t, "◆ ROOT         0", ansi.Strip(pane.View(upperCursor())))
	})

	t.Run("shows how many Snippets the Root holds", func(t *testing.T) {
		t.Parallel()

		pane := paneIn(t, look.Size{Width: boxWidth, Height: boxHeight}).WithRootSnippetCount(42)

		assert.Equal(t, "◆ ROOT        42", ansi.Strip(pane.View(upperCursor())))
	})

	t.Run("cuts the Root row to a narrow box", func(t *testing.T) {
		t.Parallel()

		pane := paneIn(t, look.Size{Width: 6, Height: boxHeight}).WithRootSnippetCount(42)

		assert.Equal(t, "◆ … 42", ansi.Strip(pane.View(upperCursor())))
	})
}

func TestPane_ShortHelp(t *testing.T) {
	t.Parallel()

	pane := paneIn(t, look.Size{Width: boxWidth, Height: boxHeight})

	assert.Equal(t, testsettings.Default(t).Keys.For(binding.ScopeFolders).ShortHelp(), pane.ShortHelp())
	assert.Equal(t, [][]key.Binding{pane.ShortHelp()}, pane.FullHelp())
}
