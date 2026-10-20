//go:build e2e

package e2e_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/folder"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/app/tag"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testapp"
)

const (
	goneElsewhereTitle = "Gone elsewhere"
	rootListTitle      = "3 Root · by title"
)

func TestEditOverSnippetTrashedElsewhereThenSaveAtRoot(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	docker := testapp.SeedFolder(t, app, folder.CreateInput{Name: "docker"})
	prune := testapp.SeedTag(t, app, tag.CreateInput{Name: "prune"})
	stored := testapp.SeedSnippet(t, app, snippet.CreateInput{
		Title:    "Prune everything",
		FolderID: docker.ID(),
		Fragments: []snippet.FragmentInput{
			{Title: "prune.sh", Language: "Bash", Content: "docker system prune\n"},
			{Language: "YAML", Content: "volumes: true\n"},
		},
		Tags: []domain.Tag{prune},
	})
	screen := open(t, app)
	screen.waitForFrame("docker")

	screen.press(keypress.Letter('j'))
	screen.waitForFrame("Prune everything")
	screen.press(keypress.Letter('3'), keypress.Letter('e'))
	screen.waitForFrame("Editing")
	require.NoError(t, app.Delete.Run(t.Context(), snippet.DeleteInput{SnippetID: stored.ID()}))
	require.NoError(t, app.DeleteTag.Run(t.Context(), tag.DeleteInput{TagID: prune.ID()}))
	screen.press(keypress.Typed(" now")...)
	screen.press(keypress.Ctrl('s'))
	screen.waitForFrame(goneElsewhereTitle)
	screen.press(keypress.Letter('y'))
	screen.waitForFrame(`Saved "Prune everything now" as a new Snippet at the Root`)
	screen.waitForFrame(rootListTitle)

	trashed := testapp.TrashItems(t, app)
	require.Len(t, trashed, 1)
	assert.Equal(t, stored.ID(), trashed[0].Snippet.ID())

	savedAtRoot := testapp.StoredSnippets(t, app)
	require.Len(t, savedAtRoot, 1)
	assert.True(t, savedAtRoot[0].AtRoot())
	assert.Equal(t, "Prune everything now", savedAtRoot[0].Title().String())
	assert.Equal(t, fragmentTextsOf(stored), fragmentTextsOf(savedAtRoot[0]))
	assert.Equal(t, []domain.Tag{prune}, savedAtRoot[0].Tags())
}

func TestCaptureIntoFolderTrashedElsewhereThenSaveAtRoot(t *testing.T) {
	t.Parallel()

	home, app := testapp.Start(t, testapp.RecordingTool)
	docker := testapp.SeedFolder(t, app, folder.CreateInput{Name: "docker"})
	home.PutOnClipboard(t, clipboardText)
	screen := open(t, app)
	screen.waitForFrame("docker")

	screen.press(keypress.Letter('j'), keypress.Letter('p'))
	screen.waitForFrame(unsavedOverlay)
	require.NoError(t, app.DeleteFolder.Run(t.Context(), folder.DeleteInput{FolderID: docker.ID()}))
	screen.press(keypress.Typed(capturedTitle)...)
	screen.press(keypress.Ctrl('s'))
	screen.waitForFrame(`Folder "docker" was moved to the Trash or deleted in another TuiSnip.`)
	screen.press(keypress.Letter('y'))
	screen.waitForFrame(`Saved "` + capturedTitle + `" as a new Snippet at the Root`)
	screen.waitForFrame(rootListTitle)

	captured := testapp.StoredSnippets(t, app)
	require.Len(t, captured, 1)
	assert.True(t, captured[0].AtRoot())
	assert.Equal(t, clipboardText, captured[0].FirstFragment().Content().String())
}

type fragmentText struct {
	title    string
	language string
	content  string
}

func fragmentTextsOf(stored domain.Snippet) []fragmentText {
	texts := make([]fragmentText, 0, len(stored.Fragments()))
	for _, fragment := range stored.Fragments() {
		texts = append(texts, fragmentText{
			title:    fragment.Title().String(),
			language: fragment.Language().String(),
			content:  fragment.Content().String(),
		})
	}

	return texts
}
