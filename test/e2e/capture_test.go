//go:build e2e

package e2e_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testapp"
)

const (
	clipboardText    = "docker ps --all\n"
	capturedTitle    = "List containers"
	unsavedOverlay   = "Editing •"
	folderHintsShown = "enter open · N new Folder"
)

func TestCaptureIntoFolder(t *testing.T) {
	t.Parallel()

	home, app := testapp.Start(t, testapp.RecordingTool)
	docker := testkit.Folder(t, testkit.FolderSpec{ID: testkit.NewSequentialIDs().NewFolderID(), Name: "docker"})
	testapp.SeedFolder(t, app, docker)
	home.PutOnClipboard(t, clipboardText)
	screen := open(t, app)
	screen.waitForFrame("docker")

	screen.press(keypress.Letter('j'), keypress.Letter('p'))
	screen.waitForFrame(unsavedOverlay)
	screen.press(keypress.Typed(capturedTitle)...)
	screen.press(keypress.Ctrl('s'))
	screen.waitForFrame(folderHintsShown)
	screen.waitForFrame("3 Root / docker")
	screen.waitForFrame(capturedTitle)

	listed, err := app.SnippetsInFolder.Run(t.Context(), browse.SnippetsInFolderInput{
		FolderID: docker.ID(), Order: domain.SortOrderTitle,
	})
	require.NoError(t, err)
	require.Len(t, listed, 1)
	assert.Equal(t, capturedTitle, listed[0].Title().String())
	assert.Equal(t, clipboardText, listed[0].FirstFragment().Content().String())
}

func TestCaptureFromEmptyClipboardOpensNothing(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	screen := open(t, app)
	screen.waitForFrame("1 Folders")

	screen.press(keypress.Letter('p'))
	screen.waitForFrame("Clipboard is empty")

	assert.NotContains(t, screen.frame.get(), "Editing")
}

func TestNewSnippetForTag(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	oneliner := testkit.Tag(t, testkit.TagSpec{ID: testkit.NewSequentialIDs().NewTagID(), Name: "oneliner"})
	testapp.SeedTag(t, app, oneliner)
	screen := open(t, app)
	screen.waitForFrame("# oneliner")

	screen.press(keypress.Letter('2'), keypress.Special(tea.KeyEnter), keypress.Letter('n'))
	screen.waitForFrame("Editing")
	screen.press(keypress.Typed(capturedTitle)...)
	screen.press(keypress.Ctrl('s'))
	screen.waitForFrame(listHintsShown)
	screen.waitForFrame(capturedTitle)

	assert.Contains(t, screen.frame.get(), "3 # oneliner")

	listed, err := app.SnippetsWithTag.Run(t.Context(), browse.SnippetsWithTagInput{
		TagID: oneliner.ID(), Order: domain.SortOrderTitle,
	})
	require.NoError(t, err)
	require.Len(t, listed, 1)
	assert.True(t, listed[0].AtRoot())
}
