//go:build e2e

package e2e_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/bootstrap"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testapp"
)

const deployStrip = " main.go │ YAML "

func TestSwitchFragmentThenCopy(t *testing.T) {
	t.Parallel()

	home, app := testapp.Start(t, testapp.RecordingTool)
	seedDeploySnippet(t, app)
	screen := open(t, app)
	screen.waitForFrame(deployStrip)

	screen.press(keypress.Letter('4'), keypress.Letter(']'))
	screen.waitForFrame("Root · YAML")
	screen.press(keypress.Letter('y'))
	screen.waitForFrame("Copied")

	assert.Equal(t, "on: push", home.Copied(t))
}

func TestAddTitleAndMoveFragmentThenSave(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	stored := seedDeploySnippet(t, app)
	screen := open(t, app)
	screen.waitForFrame(deployStrip)

	screen.press(keypress.Letter('3'), keypress.Letter('e'))
	screen.waitForFrame("Fragment title main.go")
	screen.press(keypress.Ctrl('n'))
	screen.waitForFrame(" main.go │ plaintext │ YAML ")
	screen.press(keypress.Special(tea.KeyDown), keypress.Special(tea.KeyDown), keypress.Special(tea.KeyDown))
	screen.press(keypress.Typed("notes.txt")...)
	screen.waitForFrame(" main.go │ notes.txt │ YAML ")
	screen.press(tea.KeyPressMsg{Code: tea.KeyLeft, Mod: tea.ModAlt | tea.ModShift})
	screen.waitForFrame(" notes.txt │ main.go │ YAML ")
	screen.press(keypress.Ctrl('s'))
	screen.waitForFrame(listHintsShown)
	screen.waitForFrame(" notes.txt │ main.go │ YAML ")

	assert.Equal(t, []string{"notes.txt", "main.go", ""}, fragmentTitles(testapp.StoredSnippet(t, app, stored.ID())))
}

func seedDeploySnippet(t *testing.T, app *bootstrap.App) domain.Snippet {
	t.Helper()

	return testapp.SeedSnippet(t, app, snippet.CreateInput{
		Title: "Deploy",
		Fragments: []snippet.FragmentInput{
			{Title: "main.go", Language: "Go", Content: "package main\n"},
			{Language: "YAML", Content: "on: push\n"},
		},
	})
}

func fragmentTitles(stored domain.Snippet) []string {
	titles := make([]string, 0, len(stored.Fragments()))
	for _, fragment := range stored.Fragments() {
		titles = append(titles, fragment.Title().String())
	}

	return titles
}
