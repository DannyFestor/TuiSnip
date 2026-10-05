package tagpane_test

import (
	"testing"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/tagpane"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

const (
	boxWidth  = 16
	boxHeight = 3
)

func paneIn(t *testing.T, box look.Size) tagpane.Pane {
	t.Helper()

	return styledPaneIn(t, look.NewStyles(look.SchemeDark), box)
}

func styledPaneIn(t *testing.T, styles look.Styles, box look.Size) tagpane.Pane {
	t.Helper()

	pane := tagpane.New(testsettings.Default(t).Keys, styles)
	pane, _, _ = pane.Update(look.Resized{Box: box})

	return pane
}
