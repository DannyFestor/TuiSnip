//go:build e2e

package e2e_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/folder"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testapp"
)

func TestMovedSnippetLeavesTheListAndKeepsItsLanguage(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	docker := testapp.SeedFolder(t, app, folder.CreateInput{Name: "docker"})
	moving := seededSnippetIn(t, app, domain.FolderID{}, "Bash")
	screen := open(t, app)
	screen.waitForFrame(rootTitle)

	screen.press(keypress.Letter('3'), keypress.Letter('m'))
	screen.waitForFrame("Move " + rootTitle + " to")
	screen.press(keypress.Typed("docker")...)
	screen.press(keypress.Special(tea.KeyEnter))
	screen.waitForFrame(emptyList)

	moved := testapp.StoredSnippet(t, app, moving.ID())
	assert.Equal(t, docker.ID(), moved.FolderID())
	assert.Equal(t, "Bash", moved.FirstFragment().Language().String())
}

func TestMovedFolderStaysSelectedAndCannotMoveIntoItsOwnSubtree(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	golang := seedNestedFolders(t, app)
	docker := testapp.SeedFolder(t, app, folder.CreateInput{Name: "docker"})
	screen := open(t, app)
	screen.waitForFrame("testing")

	screen.press(keypress.Letter('j'), keypress.Letter('j'), keypress.Letter('m'))
	screen.waitForFrame("Move go to")
	screen.press(keypress.Typed("go / testing")...)
	screen.press(keypress.Special(tea.KeyEnter), keypress.Special(tea.KeyEscape))
	screen.waitForFrame(folderHintsShown)

	screen.press(keypress.Letter('m'))
	screen.waitForFrame("Move go to")
	screen.press(keypress.Typed("docker")...)
	screen.press(keypress.Special(tea.KeyEnter))
	screen.waitForFrame("3 Root / docker / go · by title")

	require.Equal(t, docker.ID(), testapp.StoredFolder(t, app, golang.ID()).ParentID())
}
