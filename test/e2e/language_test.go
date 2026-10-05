//go:build e2e

package e2e_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/bootstrap"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testapp"
)

const (
	curatedLanguages = "languages = [\"YAML\", \"Go\"]\n"
	languagePicker   = "Pick a Language"
	noLanguageMatch  = "No Language matches."
)

func TestPickLanguageForEditedSnippet(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	stored := seededSnippetIn(t, app, domain.FolderID{}, "Go")
	screen := open(t, app)
	screen.waitForFrame(rootTitle)

	screen.press(keypress.Letter('3'), keypress.Letter('e'), keypress.Ctrl('l'))
	screen.waitForFrame(languagePicker)
	screen.press(keypress.Typed("bash")...)
	screen.press(keypress.Special(tea.KeyEnter), keypress.Ctrl('s'))
	screen.waitForFrame("Root · Bash")

	saved, err := app.SnippetRepository.Find(t.Context(), stored.ID())
	require.NoError(t, err)
	assert.Equal(t, "Bash", saved.FirstFragment().Language().String())
}

func TestCuratedLanguagesRevealEveryLanguage(t *testing.T) {
	t.Parallel()

	home := testapp.NewHome(t)
	home.WriteConfig(t, curatedLanguages)
	screen := open(t, home.Start(t, testapp.RecordingTool))
	screen.waitForFrame(emptyList)

	screen.press(keypress.Letter('n'), keypress.Ctrl('l'))
	screen.press(keypress.Typed("bas")...)
	screen.waitForFrame(noLanguageMatch)
	screen.press(keypress.Ctrl('a'))
	screen.waitForFrame("Bash")
	screen.press(keypress.Special(tea.KeyEnter))
	screen.waitForFrame("Language    Bash")

	screen.press(keypress.Special(tea.KeyEscape), keypress.Letter('y'))
	screen.waitForFrame("enter open · N new Folder")
}

func TestFolderDefaultLanguageLeavesItsSnippetsAlone(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	golang := testkit.Folder(t, testkit.FolderSpec{ID: testkit.NewSequentialIDs().NewFolderID(), Name: "go"})
	testapp.SeedFolder(t, app, golang)
	filed := seededSnippetIn(t, app, golang.ID(), "Bash")
	screen := open(t, app)
	screen.waitForFrame("go")

	screen.press(keypress.Letter('j'), keypress.Letter('L'))
	screen.waitForFrame("Default Language of go")
	screen.press(keypress.Typed("go")...)
	screen.press(keypress.Special(tea.KeyEnter))

	require.Eventually(t, func() bool {
		stored, err := app.FolderRepository.Find(t.Context(), golang.ID())

		return err == nil && stored.DefaultLanguage().String() == "Go"
	}, waitTimeout, pollInterval)

	kept, err := app.SnippetRepository.Find(t.Context(), filed.ID())
	require.NoError(t, err)
	assert.Equal(t, "Bash", kept.FirstFragment().Language().String())
	assert.Equal(t, filed.UpdatedAt(), kept.UpdatedAt())
}

func seededSnippetIn(t *testing.T, app *bootstrap.App, folderID domain.FolderID, language string) domain.Snippet {
	t.Helper()

	ids := testkit.NewSequentialIDs()
	stored := snippetWithIDs(t, ids, testkit.SnippetSpec{
		Title:    rootTitle,
		FolderID: folderID,
		Fragment: testkit.FragmentSpec{Language: language, Content: "echo hi\n"},
	})
	testapp.SeedSnippet(t, app, stored)

	return stored
}
