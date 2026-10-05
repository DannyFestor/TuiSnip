//go:build e2e

package e2e_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testapp"
)

const filedHeader = "Root / go / testing · plaintext"

func TestBrowseIntoNestedFolder(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	seedNestedFolders(t, app)
	screen := open(t, app)
	screen.waitForFrame("    testing")

	screen.press(keypress.Letter('G'))

	screen.waitForFrame(filedHeader)
	screen.waitForFrame("3 Root / go / testing")
}

func TestSearchRevealsSnippetInItsFolder(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	seedNestedFolders(t, app)
	screen := open(t, app)
	screen.waitForFrame(rootTitle)

	screen.press(keypress.Letter('/'))
	screen.press(keypress.Typed("skeleton")...)
	screen.waitForFrame("go / testing")
	screen.press(keypress.Special(tea.KeyEnter))

	screen.waitForFrame(filedHeader)
	screen.waitForFrame("3 Root / go / testing")
}

func TestCreateThenRenameFolder(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	seedNestedFolders(t, app)
	screen := open(t, app)
	screen.waitForFrame("    testing")

	screen.press(keypress.Letter('j'), keypress.Letter('N'))
	screen.press(keypress.Typed("errors")...)
	screen.press(keypress.Special(tea.KeyEnter))
	screen.waitForFrame("3 Root / go / errors")

	screen.press(keypress.Letter('r'))
	screen.press(keypress.Typed("-wrapping")...)
	screen.press(keypress.Special(tea.KeyEnter))

	screen.waitForFrame("3 Root / go / errors-wrapping")
}
