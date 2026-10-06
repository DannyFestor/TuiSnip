//go:build e2e

package e2e_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/bootstrap"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testapp"
)

const (
	storedCommand  = "echo stored\n"
	editedCommand  = "echo edited\n"
	editorFailed   = "Editor exited with an error; changes discarded"
	externalEditor = 'E'
)

func TestExternalEditThenSave(t *testing.T) {
	t.Parallel()

	app, stored := startWithEditor(t, testapp.ReplacingEditor)
	screen := open(t, app)
	screen.waitForFrame(rootTitle)

	screen.press(keypress.Letter('3'), keypress.Letter(externalEditor))
	screen.waitForFrame(unsavedOverlay)
	screen.waitForFrame("echo edited")
	screen.press(keypress.Ctrl('s'))
	screen.waitForFrame(listHintsShown)

	saved := testapp.StoredSnippet(t, app, stored.ID())
	assert.Equal(t, editedCommand, saved.FirstFragment().Content().String())
}

func TestExternalEditFromEditOverlayKeepsTabs(t *testing.T) {
	t.Parallel()

	app, stored := startWithEditor(t, testapp.TabbingEditor)
	screen := open(t, app)
	screen.waitForFrame(rootTitle)

	screen.press(keypress.Letter('3'), keypress.Letter('e'), keypress.Ctrl('e'))
	screen.waitForFrame(readOnlyNotice)
	screen.press(keypress.Ctrl('s'))
	screen.waitForFrame(listHintsShown)

	saved := testapp.StoredSnippet(t, app, stored.ID())
	assert.Equal(t, tabbedContent, saved.FirstFragment().Content().String())
}

func TestFailingExternalEditorDiscardsChanges(t *testing.T) {
	t.Parallel()

	app, stored := startWithEditor(t, testapp.FailingEditor)
	screen := open(t, app)
	screen.waitForFrame(rootTitle)

	screen.press(keypress.Letter('3'), keypress.Letter(externalEditor))
	screen.waitForFrame(editorFailed)

	assert.NotContains(t, screen.frame.get(), "Editing")

	kept := testapp.StoredSnippet(t, app, stored.ID())
	assert.Equal(t, storedCommand, kept.FirstFragment().Content().String())
}

func startWithEditor(t *testing.T, editor testapp.Editor) (*bootstrap.App, domain.Snippet) {
	t.Helper()

	home := testapp.NewHome(t)
	home.UseEditor(t, editor)
	app := home.Start(t, testapp.RecordingTool)

	stored := testapp.SeedSnippet(t, app, snippet.CreateInput{
		Title:    rootTitle,
		Language: "Bash",
		Content:  storedCommand,
	})

	return app, stored
}
