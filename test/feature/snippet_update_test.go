//go:build feature

package feature_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/app/tag"
	"github.com/DannyFestor/TuiSnip/internal/bootstrap"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/test/testapp"
)

func TestEditedSnippetPersistsWithNewUpdatedTime(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	created := create(
		t,
		app,
		snippet.CreateInput{Title: "curl", Description: "", Language: plainText, Content: "curl\n"},
	)

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
	created := create(
		t,
		first,
		snippet.CreateInput{Title: "curl", Description: "", Language: plainText, Content: "curl\n"},
	)

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
	created := create(t, app, snippet.CreateInput{Title: "go", Description: "", Language: plainText, Content: tabbed})

	_, err := app.Update.Run(t.Context(), editOf(created, "go return", tabbed))

	require.NoError(t, err)
	assert.Equal(t, tabbed, testapp.StoredSnippet(t, app, created.ID()).FirstFragment().Content().String())
}

func TestEditChangesTheLanguage(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	created := create(t, app, snippet.CreateInput{Title: "go", Description: "", Language: plainText, Content: "go"})
	edit := editOf(created, "go", "go")
	edit.Language = "Go"

	_, err := app.Update.Run(t.Context(), edit)

	require.NoError(t, err)
	assert.Equal(t, "Go", testapp.StoredSnippet(t, app, created.ID()).FirstFragment().Language().String())
}

func TestEditKeepsALanguageTheConfigNoLongerOffers(t *testing.T) {
	t.Parallel()

	home := testapp.NewHome(t)
	home.WriteConfig(t, "languages = [\"YAML\"]\n")
	app := home.Start(t, testapp.RecordingTool)
	created := create(t, app, snippet.CreateInput{Title: "go", Description: "", Language: "Go", Content: "go"})

	_, err := app.Update.Run(t.Context(), editOf(created, "go edited", "go"))

	require.NoError(t, err)
	assert.Equal(t, "Go", testapp.StoredSnippet(t, app, created.ID()).FirstFragment().Language().String())
}

func TestEditKeepsTheSnippetsTags(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	golang := testapp.SeedTag(t, app, tag.CreateInput{Name: "go"})
	tagged := testapp.SeedSnippet(t, app, snippet.CreateInput{Tags: []domain.Tag{golang}})

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

func TestEditRetagsTheSnippetCreatingItsNewTags(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	golang := testapp.SeedTag(t, app, tag.CreateInput{Name: "go"})
	shell := testapp.SeedTag(t, app, tag.CreateInput{Name: "shell"})
	tagged := testapp.SeedSnippet(t, app, snippet.CreateInput{Tags: []domain.Tag{golang}})
	edit := editOf(tagged, "retagged", "echo\n")
	edit.Tags = []domain.Tag{shell}
	edit.NewTags = []string{"Docker"}

	_, err := app.Update.Run(t.Context(), edit)

	require.NoError(t, err)
	carried := testapp.StoredSnippet(t, app, tagged.ID()).Tags()
	require.Len(t, carried, 2)
	assert.Equal(t, "Docker", carried[0].Name().String())
	assert.Equal(t, shell, carried[1])
	assert.Equal(t, map[string]int{"go": 0, "shell": 1, "Docker": 1}, tagCounts(t, app))
}

func TestEditCreatingATagNamedLikeAnotherIgnoringCaseIsRefused(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	golang := testapp.SeedTag(t, app, tag.CreateInput{Name: "go"})
	tagged := testapp.SeedSnippet(t, app, snippet.CreateInput{Title: "untouched", Tags: []domain.Tag{golang}})
	edit := editOf(tagged, "renamed", "")
	edit.Tags = nil
	edit.NewTags = []string{"GO"}

	_, err := app.Update.Run(t.Context(), edit)

	require.ErrorIs(t, err, domain.ErrTagNameTaken)
	stored := testapp.StoredSnippet(t, app, tagged.ID())
	assert.Equal(t, "untouched", stored.Title().String())
	assert.Equal(t, []domain.Tag{golang}, stored.Tags())
	assert.Equal(t, map[string]int{"go": 1}, tagCounts(t, app))
}

func editOf(loaded domain.Snippet, title, content string) snippet.UpdateInput {
	return snippet.UpdateInput{
		SnippetID:       loaded.ID(),
		LoadedUpdatedAt: loaded.UpdatedAt(),
		Title:           title,
		Description:     loaded.Description().String(),
		Language:        loaded.FirstFragment().Language().String(),
		Content:         content,
		Tags:            loaded.Tags(),
		NewTags:         nil,
	}
}

func tagCounts(t *testing.T, app *bootstrap.App) map[string]int {
	t.Helper()

	counts := make(map[string]int)
	for _, listed := range listedTags(t, app) {
		counts[listed.Tag.Name().String()] = listed.SnippetCount
	}

	return counts
}
