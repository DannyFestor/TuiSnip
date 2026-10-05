//go:build feature

package feature_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/bootstrap"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/testapp"
)

func TestEditedSnippetPersistsWithNewUpdatedTime(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	created := create(t, app, snippet.CreateInput{Title: "curl", Description: "", Content: "curl\n"})

	updated, err := app.Update.Run(t.Context(), editOf(created, "curl json", "curl -d @-\n"))

	require.NoError(t, err)
	listed := listAtRoot(t, app)
	require.Len(t, listed, 1)
	assert.Equal(t, "curl json", listed[0].Title().String())
	assert.Equal(t, "curl -d @-\n", listed[0].FirstFragment().Content().String())
	assert.True(t, updated.UpdatedAt().Equal(listed[0].UpdatedAt()))
	assert.True(t, listed[0].UpdatedAt().After(created.UpdatedAt()))
}

func TestEditOverSnippetChangedElsewhereIsRefused(t *testing.T) {
	t.Parallel()

	home := testapp.NewHome(t)
	first := home.Start(t, testapp.RecordingTool)
	second := home.Start(t, testapp.RecordingTool)
	created := create(t, first, snippet.CreateInput{Title: "curl", Description: "", Content: "curl\n"})

	_, err := second.Update.Run(t.Context(), editOf(created, "changed elsewhere", "curl\n"))
	require.NoError(t, err)

	_, err = first.Update.Run(t.Context(), editOf(created, "stale", "curl\n"))

	require.ErrorIs(t, err, domain.ErrConflict)
	assert.Equal(t, "changed elsewhere", listAtRoot(t, first)[0].Title().String())
}

func TestEditKeepsContentWithTabsByteForByte(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	tabbed := "if x {\n\treturn\n}\n"
	created := create(t, app, snippet.CreateInput{Title: "go", Description: "", Content: tabbed})

	_, err := app.Update.Run(t.Context(), editOf(created, "go return", tabbed))

	require.NoError(t, err)
	assert.Equal(t, tabbed, findStored(t, app, created.ID()).FirstFragment().Content().String())
}

func TestEditKeepsTheSnippetsTags(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	ids := testkit.NewSequentialIDs()
	golang := seededTag(t, app, testkit.TagSpec{ID: ids.NewTagID(), Name: "go"})
	tagged := seededTaggedSnippet(t, app, ids, testkit.SnippetSpec{Tags: []domain.Tag{golang}})

	_, err := app.Update.Run(t.Context(), editOf(tagged, "go edited", "package main\n"))

	require.NoError(t, err)
	listed, err := app.SnippetsWithTag.Run(t.Context(), browse.SnippetsWithTagInput{
		TagID: golang.ID(),
		Order: domain.SortOrderTitle,
	})
	require.NoError(t, err)
	require.Len(t, listed, 1)
	assert.Equal(t, "go edited", listed[0].Title().String())
	assert.Equal(t, []domain.Tag{golang}, listed[0].Tags())
}

func editOf(loaded domain.Snippet, title, content string) snippet.UpdateInput {
	return snippet.UpdateInput{
		SnippetID:       loaded.ID(),
		LoadedUpdatedAt: loaded.UpdatedAt(),
		Title:           title,
		Description:     loaded.Description().String(),
		Content:         content,
	}
}

func findStored(t *testing.T, app *bootstrap.App, id domain.SnippetID) domain.Snippet {
	t.Helper()

	found, err := app.SnippetRepository.Find(t.Context(), id)
	require.NoError(t, err)

	return found
}
