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

	pane := tagpane.New(testsettings.Default(t).Keys, look.NewStyles())
	pane, _, _ = pane.Update(look.Resized{Box: box})

	return pane
}
