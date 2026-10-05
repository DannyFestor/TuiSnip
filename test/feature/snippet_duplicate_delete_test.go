//go:build feature

package feature_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/testapp"
)

func TestDuplicateMatchesTheOriginalInEveryFieldButIDsAndTimestamps(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	ids := testkit.NewSequentialIDs()
	scripts := seededFolder(t, app, testkit.FolderSpec{ID: ids.NewFolderID(), Name: "scripts"})
	shell := seededTag(t, app, testkit.TagSpec{ID: ids.NewTagID(), Name: "shell"})
	original := seededTaggedSnippet(t, app, ids, testkit.SnippetSpec{
		Title:       "Prune everything",
		Description: "Reclaim disk space",
		FolderID:    scripts.ID(),
		Fragment:    testkit.FragmentSpec{Language: "Bash", Content: "docker system prune\n"},
		Tags:        []domain.Tag{shell},
	})

	duplicate, err := app.Duplicate.Run(t.Context(), snippet.DuplicateInput{SnippetID: original.ID()})
	require.NoError(t, err)

	stored, err := app.SnippetRepository.Find(t.Context(), duplicate.ID())
	require.NoError(t, err)
	assert.NotEqual(t, original.ID(), stored.ID())
	assert.NotEqual(t, original.FirstFragment().ID(), stored.FirstFragment().ID())
	assert.NotEqual(t, original.CreatedAt(), stored.CreatedAt())
	assert.Equal(t, original.Title(), stored.Title())
	assert.Equal(t, original.Description(), stored.Description())
	assert.Equal(t, original.FolderID(), stored.FolderID())
	assert.Equal(t, original.Tags(), stored.Tags())
	assert.Equal(t, original.FirstFragment().Language(), stored.FirstFragment().Language())
	assert.Equal(t, original.FirstFragment().Content(), stored.FirstFragment().Content())

	kept, err := app.SnippetRepository.Find(t.Context(), original.ID())
	require.NoError(t, err)
	assert.Equal(t, original, kept)
}

func TestDeletingASnippetLeavesItsTagsInPlace(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	ids := testkit.NewSequentialIDs()
	shared := seededTag(t, app, testkit.TagSpec{ID: ids.NewTagID(), Name: "go"})
	only := seededTag(t, app, testkit.TagSpec{ID: ids.NewTagID(), Name: "testing"})
	deleted := seededTaggedSnippet(t, app, ids, testkit.SnippetSpec{Title: "deleted", Tags: []domain.Tag{shared, only}})
	kept := seededTaggedSnippet(t, app, ids, testkit.SnippetSpec{Title: "kept", Tags: []domain.Tag{shared}})

	err := app.Delete.Run(t.Context(), snippet.DeleteInput{SnippetID: deleted.ID()})
	require.NoError(t, err)

	_, err = app.SnippetRepository.Find(t.Context(), deleted.ID())
	require.ErrorIs(t, err, domain.ErrNotFound)
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
