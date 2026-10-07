//go:build e2e

package e2e_test

import (
	"testing"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/pointer"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testapp"
)

const (
	folderHints      = "N new Folder"
	snippetPaneHints = "w wrap"
	searchOpenTitle  = "Search · "
	mouseOff         = "mouse = false\n"
)

func TestClickFolderRowBrowsesIt(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	seedNestedFolders(t, app)
	screen := open(t, app)
	screen.waitForFrame("    testing")

	screen.click("testing")

	screen.waitForFrame(filedHeader)
	screen.waitForFrame("3 Root / go / testing")
}

func TestDoubleClickSnippetOpensSnippetPane(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	seedNestedFolders(t, app)
	screen := open(t, app)
	screen.waitForFrame(rootTitle)

	screen.doubleClick(rootTitle)

	screen.waitForFrame(snippetPaneHints)
}

func TestClickOutsideOverlayClosesIt(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	seedNestedFolders(t, app)
	screen := open(t, app)
	screen.waitForFrame(rootTitle)
	screen.press(keypress.Letter('/'))
	screen.waitForFrame(searchOpenTitle)

	screen.clickAt(pointer.Point{X: 0, Y: 0})

	screen.waitForFrame(folderHints)
}

func TestMouseOffIgnoresClicks(t *testing.T) {
	t.Parallel()

	home := testapp.NewHome(t)
	home.WriteConfig(t, mouseOff)
	app := home.Start(t, testapp.RecordingTool)
	seedNestedFolders(t, app)
	screen := open(t, app)
	screen.waitForFrame("    testing")

	screen.click("testing")
	screen.press(keypress.Letter('j'))

	screen.waitForFrame("3 Root / go · by title")
}
