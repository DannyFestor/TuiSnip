//go:build e2e

package e2e_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testapp"
)

const (
	emptyList       = "No Snippets here."
	savedPreview    = "Root · plaintext"
	searchOpened    = "Search · "
	remappedSearch  = "[bindings.global]\nsearch = [\"ctrl+f\"]\n"
	noClipboardTool = "No clipboard tool found (pbcopy, wl-copy, xclip, xsel)"
	nativeOnly      = "[copy]\nclipboard = \"native\"\n"
)

func TestCreateSearchThenCopy(t *testing.T) {
	t.Parallel()

	home, app := testapp.Start(t, testapp.RecordingTool)
	screen := open(t, app)
	screen.waitForFrame(emptyList)

	createPruneSnippet(screen)
	screen.waitForFrame(savedPreview)

	screen.press(keypress.Letter('/'))
	screen.press(keypress.Typed("prune")...)
	screen.waitForFrame("Search · 1 result")
	screen.press(keypress.Ctrl('y'))
	screen.waitForFrame("Copied")

	assert.Equal(t, "docker\nprune", home.Copied(t))
}

func TestSaveWithBlankTitleShowsError(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	screen := open(t, app)
	screen.waitForFrame(emptyList)

	screen.press(keypress.Letter('n'), keypress.Ctrl('s'))

	screen.waitForFrame("Title is blank")
}

func TestCopyWithoutToolShowsMessage(t *testing.T) {
	t.Parallel()

	home := testapp.NewHome(t)
	home.WriteConfig(t, nativeOnly)
	screen := open(t, home.Start(t, testapp.NoTool))
	screen.waitForFrame(emptyList)
	createPruneSnippet(screen)
	screen.waitForFrame(savedPreview)

	screen.press(keypress.Letter('3'), keypress.Letter('y'))

	screen.waitForFrame(noClipboardTool)
}

func TestRemappedBindingOpensSearch(t *testing.T) {
	t.Parallel()

	home := testapp.NewHome(t)
	home.WriteConfig(t, remappedSearch)
	screen := open(t, home.Start(t, testapp.RecordingTool))
	screen.waitForFrame(emptyList)

	screen.press(keypress.Ctrl('f'))

	screen.waitForFrame(searchOpened)
}

func createPruneSnippet(screen *session) {
	screen.press(keypress.Letter('n'))
	screen.press(keypress.Typed("Prune")...)
	screen.press(keypress.Special(tea.KeyTab))
	screen.press(keypress.Typed("Reclaim")...)
	screen.press(keypress.Special(tea.KeyEnter), keypress.Special(tea.KeyEnter))
	screen.press(keypress.Typed("docker")...)
	screen.press(keypress.Special(tea.KeyEnter))
	screen.press(keypress.Typed("prune")...)
	screen.press(keypress.Ctrl('s'))
}
