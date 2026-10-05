//go:build e2e

package e2e_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testapp"
)

const (
	tabbedContent  = "if x {\n\treturn\n}\n"
	readOnlyNotice = "Contains tabs: read-only here, edit with ctrl+e ($EDITOR)"
	listHintsShown = "y Copy · e edit · n new"
)

func TestEditSnippetKeepsTabbedContent(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	ids := testkit.NewSequentialIDs()
	stored := testkit.Snippet(t, testkit.SnippetSpec{
		ID:       ids.NewSnippetID(),
		Title:    rootTitle,
		FolderID: domain.FolderID{},
		Fragment: testkit.FragmentSpec{ID: ids.NewFragmentID(), Language: "Go", Content: tabbedContent},
	})
	testapp.SeedSnippet(t, app, stored)
	screen := open(t, app)
	screen.waitForFrame(rootTitle)

	screen.press(keypress.Letter('3'), keypress.Letter('e'))
	screen.waitForFrame(readOnlyNotice)
	screen.press(keypress.Typed(" edited")...)
	screen.press(keypress.Ctrl('s'))
	screen.waitForFrame(listHintsShown)
	screen.waitForFrame(rootTitle + " edited")

	saved, err := app.SnippetRepository.Find(t.Context(), stored.ID())
	require.NoError(t, err)
	assert.Equal(t, rootTitle+" edited", saved.Title().String())
	assert.Equal(t, tabbedContent, saved.FirstFragment().Content().String())
	assert.True(t, saved.UpdatedAt().After(stored.UpdatedAt()))
}
