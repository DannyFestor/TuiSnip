package tagpane_test

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

	t.Run("says there are no Tags yet", func(t *testing.T) {
		t.Parallel()

		pane := paneIn(t, look.Size{Width: boxWidth, Height: boxHeight})

		assert.Equal(t, "No Tags yet.    ", ansi.Strip(pane.View()))
	})

	t.Run("cuts the text to a narrow box", func(t *testing.T) {
		t.Parallel()

		pane := paneIn(t, look.Size{Width: 6, Height: boxHeight})

		assert.Equal(t, "No Ta…", ansi.Strip(pane.View()))
	})
}

func TestPane_ShortHelp(t *testing.T) {
	t.Parallel()

	pane := paneIn(t, look.Size{Width: boxWidth, Height: boxHeight})

	assert.Equal(t, testsettings.Default(t).Keys.For(binding.ScopeTags).ShortHelp(), pane.ShortHelp())
	assert.Equal(t, [][]key.Binding{pane.ShortHelp()}, pane.FullHelp())
}
