//go:build e2e

package e2e_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/app/tag"
	"github.com/DannyFestor/TuiSnip/internal/bootstrap"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testapp"
)

const tagEditorFilter = "filter or new Tag:"

func TestToggleAndCreateTagsInTheTagEditor(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	golang := testapp.SeedTag(t, app, tag.CreateInput{Name: "go"})
	testapp.SeedTag(t, app, tag.CreateInput{Name: "docker"})
	stored := testapp.SeedSnippet(t, app, snippet.CreateInput{Title: rootTitle, Tags: []domain.Tag{golang}})
	screen := open(t, app)
	screen.waitForFrame(rootTitle)
	screen.waitForFrame("# docker")

	screen.press(keypress.Letter('3'), keypress.Letter('e'), keypress.Ctrl('t'))
	screen.waitForFrame(tagEditorFilter)
	screen.press(keypress.Special(tea.KeyEnter), keypress.Special(tea.KeyDown), keypress.Special(tea.KeyEnter))
	screen.press(keypress.Typed("API")...)
	screen.waitForFrame(`+ create "API"`)
	screen.press(keypress.Special(tea.KeyEnter), keypress.Special(tea.KeyEscape))
	screen.waitForFrame("Tags        #API #docker")
	screen.press(keypress.Ctrl('s'))
	screen.waitForFrame("Root · plaintext · #API #docker")

	assert.Equal(t, []string{"API", "docker"}, tagNames(testapp.StoredSnippet(t, app, stored.ID()).Tags()))
	assert.ElementsMatch(t, []string{"API", "docker", "go"}, listedTagNames(t, app))
}

func TestCancellingTheEditOverlayDiscardsCreatedTags(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	stored := testapp.SeedSnippet(t, app, snippet.CreateInput{Title: rootTitle})
	screen := open(t, app)
	screen.waitForFrame(rootTitle)

	screen.press(keypress.Letter('3'), keypress.Letter('e'), keypress.Ctrl('t'))
	screen.press(keypress.Typed("scratch")...)
	screen.press(keypress.Special(tea.KeyEnter), keypress.Special(tea.KeyEscape))
	screen.waitForFrame("Tags        #scratch")
	screen.press(keypress.Special(tea.KeyEscape))
	screen.waitForFrame("Discard the unsaved changes?")
	screen.press(keypress.Letter('y'))
	screen.waitForFrame(listHintsShown)

	assert.Empty(t, testapp.StoredSnippet(t, app, stored.ID()).Tags())
	assert.Empty(t, listedTagNames(t, app))
}

func TestTagNameTypedInAnotherCaseOffersTheTag(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	docker := testapp.SeedTag(t, app, tag.CreateInput{Name: "docker"})
	screen := open(t, app)
	screen.waitForFrame("# docker")

	screen.press(keypress.Letter('n'))
	screen.press(keypress.Typed("Pods")...)
	screen.press(keypress.Ctrl('t'))
	screen.press(keypress.Typed("DOCKER")...)
	screen.waitForFrame(tagEditorFilter + " DOCKER")
	assert.NotContains(t, screen.frame.get(), "+ create")
	screen.press(keypress.Special(tea.KeyEnter), keypress.Special(tea.KeyEscape), keypress.Ctrl('s'))
	screen.waitForFrame("Root · plaintext · #docker")

	listed, err := app.SnippetsWithTag.Run(t.Context(), browse.SnippetsWithTagInput{
		TagID: docker.ID(),
		Order: domain.SortOrderTitle,
	})
	require.NoError(t, err)
	require.Len(t, listed, 1)
	assert.Equal(t, "Pods", listed[0].Title().String())
	assert.Equal(t, []string{"docker"}, listedTagNames(t, app))
}

func tagNames(tags []domain.Tag) []string {
	names := make([]string, 0, len(tags))
	for _, carried := range tags {
		names = append(names, carried.Name().String())
	}

	return names
}

func listedTagNames(t *testing.T, app *bootstrap.App) []string {
	t.Helper()

	listed, err := app.TagList.Run(t.Context(), browse.TagListInput{})
	require.NoError(t, err)

	names := make([]string, 0, len(listed))
	for _, counted := range listed {
		names = append(names, counted.Tag.Name().String())
	}

	return names
}
