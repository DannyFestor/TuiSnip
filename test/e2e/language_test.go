//go:build e2e

package e2e_test

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/app/folder"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/bootstrap"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testapp"
)

const (
	curatedLanguages = "languages = [\"YAML\", \"Go\"]\n"
	languagePicker   = "Pick a Language"
	noLanguageMatch  = "No Language matches."
	overlayTitle     = "Editing"
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

	assert.Equal(t, "Bash", testapp.StoredSnippet(t, app, stored.ID()).FirstFragment().Language().String())
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
	golang := testapp.SeedFolder(t, app, folder.CreateInput{Name: "go"})
	filed := seededSnippetIn(t, app, golang.ID(), "Bash")
	screen := open(t, app)
	screen.waitForFrame("go")

	screen.press(keypress.Letter('j'), keypress.Letter('L'))
	screen.waitForFrame("Default Language of go")
	screen.press(keypress.Typed("go")...)
	screen.press(keypress.Special(tea.KeyEnter))

	require.Eventually(t, func() bool {
		return testapp.StoredFolder(t, app, golang.ID()).DefaultLanguage().String() == "Go"
	}, waitTimeout, pollInterval)

	kept := testapp.StoredSnippet(t, app, filed.ID())
	assert.Equal(t, "Bash", kept.FirstFragment().Language().String())
	assert.Equal(t, filed.UpdatedAt(), kept.UpdatedAt())
}

func TestNewSnippetStartsInFolderDefaultLanguage(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	golang := testapp.SeedFolder(t, app, folder.CreateInput{Name: "go"})
	screen := open(t, app)
	screen.waitForFrame("go")

	screen.press(keypress.Letter('j'), keypress.Letter('L'))
	screen.waitForFrame("Default Language of go")
	screen.press(keypress.Typed("go")...)
	screen.press(keypress.Special(tea.KeyEnter))
	screen.waitForFrame(folderHintsShown)

	screen.openNewSnippetShowing("Language    Go")
	screen.press(keypress.Typed(filedTitle)...)
	screen.press(keypress.Ctrl('s'))
	screen.waitForFrame("go · Go")

	listed, err := app.SnippetsInFolder.Run(t.Context(), browse.SnippetsInFolderInput{
		FolderID: golang.ID(), Order: domain.SortOrderTitle,
	})
	require.NoError(t, err)
	require.Len(t, listed, 1)
	assert.Equal(t, filedTitle, listed[0].Title().String())
	assert.Equal(t, "Go", listed[0].FirstFragment().Language().String())
}

// The Folder tree reloads after the Default Language is stored, and nothing on screen shows when it has.
func (s *session) openNewSnippetShowing(text string) {
	s.t.Helper()

	deadline := time.Now().Add(waitTimeout)

	for {
		s.press(keypress.Letter('n'))
		s.waitForFrame(overlayTitle)

		if strings.Contains(s.frame.get(), text) {
			return
		}

		if time.Now().After(deadline) {
			s.t.Fatalf("%q never appeared in a new Snippet; last frame:\n%s", text, s.frame.get())
		}

		s.press(keypress.Special(tea.KeyEscape))
		s.waitForFrame(folderHintsShown)
	}
}

func seededSnippetIn(t *testing.T, app *bootstrap.App, folderID domain.FolderID, language string) domain.Snippet {
	t.Helper()

	return testapp.SeedSnippet(t, app, snippet.CreateInput{
		Title:    rootTitle,
		Language: language,
		Content:  "echo hi\n",
		FolderID: folderID,
	})
}
