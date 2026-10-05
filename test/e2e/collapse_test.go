//go:build e2e

package e2e_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testapp"
)

const (
	nestedFolderRow = "    testing"
	staleFolderID   = "0194c3a0-dead-7000-8000-000000000000"
)

func TestCollapsedFolderSurvivesRestart(t *testing.T) {
	t.Parallel()

	home, first := testapp.Start(t, testapp.RecordingTool)
	golang := seedNestedFolders(t, first)
	screen := open(t, first)
	screen.waitForFrame(nestedFolderRow)

	screen.press(keypress.Letter('j'), keypress.Letter(' '))
	screen.waitForFrame("▸ go")
	waitForState(t, home, golang.ID().String())

	restarted := open(t, home.Start(t, testapp.RecordingTool))

	restarted.waitForFrame("▸ go")
	assert.NotContains(t, restarted.frame.get(), nestedFolderRow)
}

func TestStaleCollapsedFolderIsIgnoredThenDropped(t *testing.T) {
	t.Parallel()

	home := testapp.NewHome(t)
	home.WriteState(t, "[folders]\ncollapsed = [\""+staleFolderID+"\"]\n")
	app := home.Start(t, testapp.RecordingTool)
	seedNestedFolders(t, app)

	screen := open(t, app)

	screen.waitForFrame(nestedFolderRow)

	assert.NotContains(t, waitForState(t, home, "collapsed = []"), staleFolderID)
}
