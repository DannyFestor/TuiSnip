//go:build e2e

package e2e_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testapp"
)

func TestSwitchFragmentThenCopy(t *testing.T) {
	t.Parallel()

	home, app := testapp.Start(t, testapp.RecordingTool)
	ids := testkit.NewSequentialIDs()
	home.InsertSnippet(t, app, testkit.Snippet(t, testkit.SnippetSpec{
		ID:    ids.NewSnippetID(),
		Title: "Deploy",
		Fragments: []testkit.FragmentSpec{
			{ID: ids.NewFragmentID(), Title: "main.go", Language: "Go", Content: "package main\n"},
			{ID: ids.NewFragmentID(), Language: "YAML", Content: "on: push\n"},
		},
	}))
	screen := open(t, app)
	screen.waitForFrame(" main.go │ YAML ")

	screen.press(keypress.Letter('4'), keypress.Letter(']'))
	screen.waitForFrame("Root · YAML")
	screen.press(keypress.Letter('y'))
	screen.waitForFrame("Copied")

	assert.Equal(t, "on: push", home.Copied(t))
}
