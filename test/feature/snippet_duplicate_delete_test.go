//go:build feature

package feature_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/app/folder"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/app/tag"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/testapp"
)

func TestDuplicateMatchesTheOriginalInEveryFieldButIDsAndTimestamps(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	scripts := testapp.SeedFolder(t, app, folder.CreateInput{Name: "scripts"})
	shell := testapp.SeedTag(t, app, tag.CreateInput{Name: "shell"})
	original := testapp.SeedSnippet(t, app, snippet.CreateInput{
		Title:       "Prune everything",
		Description: "Reclaim disk space",
		Language:    "Bash",
		Content:     "docker system prune\n",
		FolderID:    scripts.ID(),
		Tags:        []domain.Tag{shell},
	})

	duplicate, err := app.Duplicate.Run(t.Context(), snippet.DuplicateInput{SnippetID: original.ID()})
	require.NoError(t, err)

	stored := testapp.StoredSnippet(t, app, duplicate.ID())
	assert.NotEqual(t, original.ID(), stored.ID())
	assert.NotEqual(t, original.FirstFragment().ID(), stored.FirstFragment().ID())
	assert.NotEqual(t, original.CreatedAt(), stored.CreatedAt())
	assert.Equal(t, original.Title(), stored.Title())
	assert.Equal(t, original.Description(), stored.Description())
	assert.Equal(t, original.FolderID(), stored.FolderID())
	assert.Equal(t, original.Tags(), stored.Tags())
	assert.Equal(t, original.FirstFragment().Language(), stored.FirstFragment().Language())
	assert.Equal(t, original.FirstFragment().Content(), stored.FirstFragment().Content())
	assert.Equal(t, original, testapp.StoredSnippet(t, app, original.ID()))
}

func TestDeletingASnippetLeavesItsTagsInPlace(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	shared := testapp.SeedTag(t, app, tag.CreateInput{Name: "go"})
	only := testapp.SeedTag(t, app, tag.CreateInput{Name: "testing"})
	deleted := testapp.SeedSnippet(t, app, snippet.CreateInput{Title: "deleted", Tags: []domain.Tag{shared, only}})
	kept := testapp.SeedSnippet(t, app, snippet.CreateInput{Title: "kept", Tags: []domain.Tag{shared}})

	err := app.Delete.Run(t.Context(), snippet.DeleteInput{SnippetID: deleted.ID()})
	require.NoError(t, err)

	_, stored := testapp.FindSnippet(t, app, deleted.ID())
	assert.False(t, stored)
	assert.Equal(t, []domain.Snippet{kept}, listAtRoot(t, app))
	assert.Equal(t, []browse.TagCount{
		{Tag: shared, SnippetCount: 1},
		{Tag: only, SnippetCount: 0},
	}, listedTags(t, app))
}

func TestDeletingASnippetThatIsGoneIsReportedNotFound(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)

	err := app.Delete.Run(t.Context(), snippet.DeleteInput{SnippetID: testkit.NewSequentialIDs().NewSnippetID()})

	require.ErrorIs(t, err, domain.ErrNotFound)
}
