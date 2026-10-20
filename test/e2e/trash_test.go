//go:build e2e

package e2e_test

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/app/folder"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testapp"
)

var trashCountedOnce = regexp.MustCompile(`▪ Trash +1`)

func TestDeleteASnippetThenRestoreItFromTheTrash(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	docker := testapp.SeedFolder(t, app, folder.CreateInput{Name: "docker"})
	prune := testapp.SeedSnippet(t, app, snippet.CreateInput{Title: "Prune everything", FolderID: docker.ID()})
	screen := open(t, app)
	screen.waitForFrame("docker")

	screen.press(keypress.Letter('j'))
	screen.waitForFrame("Prune everything")
	screen.press(keypress.Letter('3'), keypress.Letter('d'))
	screen.waitForFrame(`Moved Snippet "Prune everything" to the Trash`)
	screen.waitForMatch(trashCountedOnce)

	screen.press(keypress.Letter('1'), keypress.Letter('G'))
	screen.waitForFrame("3 Trash")
	screen.waitForFrame("Root / docker")

	screen.press(keypress.Letter('3'), keypress.Letter('r'))
	screen.waitForFrame(`Restored "Prune everything" to docker`)
	screen.waitForFrame("The Trash is empty.")

	assert.Equal(t, docker.ID(), testapp.StoredSnippet(t, app, prune.ID()).FolderID())
}

func TestDeleteAFolderThenSeeItsSubtreeInTheTrash(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	golang := seedNestedFolders(t, app)
	testapp.SeedSnippet(t, app, snippet.CreateInput{Title: "Table test", FolderID: golang.ID()})
	screen := open(t, app)
	screen.waitForFrame("testing")

	screen.press(keypress.Letter('j'), keypress.Letter('d'))
	screen.waitForFrame(`Moved Folder "go" to the Trash`)
	screen.waitForMatch(trashCountedOnce)

	screen.press(keypress.Letter('G'))
	screen.waitForFrame("▸ go")
	screen.waitForFrame("1 subfolder · 2 Snippets")
	screen.waitForFrame("▾ testing")
}
