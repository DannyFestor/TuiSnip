//go:build e2e

package e2e_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testapp"
)

func TestSortOrderSurvivesRestart(t *testing.T) {
	t.Parallel()

	home, first := testapp.Start(t, testapp.RecordingTool)
	screen := open(t, first)
	screen.waitForFrame("3 Root · by title")

	screen.press(keypress.Typed("3s")...)
	screen.waitForFrame("3 Root · by last updated")
	waitForState(t, home, `sort = "updated"`)

	restarted := open(t, home.Start(t, testapp.RecordingTool))

	restarted.waitForFrame("3 Root · by last updated")
}

func TestBrokenStateFileStartsInTitleOrder(t *testing.T) {
	t.Parallel()

	home := testapp.NewHome(t)
	home.WriteState(t, "[snippet_list\nsort = ")

	screen := open(t, home.Start(t, testapp.RecordingTool))

	screen.waitForFrame("3 Root · by title")
	assert.Regexp(t, `level=WARN .*msg="state file replaced by defaults"`, home.Log(t))
}
