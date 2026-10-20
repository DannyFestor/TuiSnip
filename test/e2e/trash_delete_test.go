//go:build e2e

package e2e_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/folder"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testapp"
)

func TestDeleteAFolderPermanentlyThenRestoreASnippetTrashedFromIt(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	docker := testapp.SeedFolder(t, app, folder.CreateInput{Name: "docker"})
	testapp.SeedSnippet(t, app, snippet.CreateInput{Title: "Compose up", FolderID: docker.ID()})
	prune := testapp.SeedSnippet(t, app, snippet.CreateInput{Title: "Prune everything", FolderID: docker.ID()})
	require.NoError(t, app.Delete.Run(t.Context(), snippet.DeleteInput{SnippetID: prune.ID()}))
	require.NoError(t, app.DeleteFolder.Run(t.Context(), folder.DeleteInput{FolderID: docker.ID()}))
	screen := open(t, app)
	screen.waitForFrame("▪ Trash")

	screen.press(keypress.Letter('G'))
	screen.waitForFrame("▸ docker")
	screen.press(keypress.Letter('3'), keypress.Letter('d'))
	screen.waitForFrame(`Permanently delete Folder "docker" and 1 Snippet? This cannot be undone. [y/N]`)

	screen.press(keypress.Letter('y'))
	screen.waitForFrame(`Permanently deleted Folder "docker"`)
	screen.waitForMatch(trashCountedOnce)
	screen.waitForFrame("Root / docker")

	screen.press(keypress.Letter('r'))
	screen.waitForFrame(`Restored "Prune everything" to the Root; its Folder was deleted`)
	screen.waitForFrame("The Trash is empty.")

	assert.True(t, testapp.StoredSnippet(t, app, prune.ID()).AtRoot())
}

func TestEmptyTheTrashAfterConfirming(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	for _, title := range []string{"Prune everything", "Compose up"} {
		trashed := testapp.SeedSnippet(t, app, snippet.CreateInput{Title: title})
		require.NoError(t, app.Delete.Run(t.Context(), snippet.DeleteInput{SnippetID: trashed.ID()}))
	}

	screen := open(t, app)
	screen.waitForFrame("▪ Trash")

	screen.press(keypress.Letter('G'))
	screen.waitForFrame("Compose up")
	screen.press(keypress.Letter('3'), keypress.Letter('D'))
	screen.waitForFrame("Permanently delete 2 Snippets in the Trash? This cannot be undone. [y/N]")

	screen.press(keypress.Letter('y'))
	screen.waitForFrame("Emptied the Trash: 2 Snippets")
	screen.waitForFrame("The Trash is empty.")
}
