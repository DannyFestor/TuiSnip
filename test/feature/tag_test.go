//go:build feature

package feature_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/app/folder"
	"github.com/DannyFestor/TuiSnip/internal/app/search"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/app/tag"
	"github.com/DannyFestor/TuiSnip/internal/bootstrap"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/test/testapp"
)

func TestCreatedTagExistsWithACountOfZero(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)

	created, err := app.CreateTag.Run(t.Context(), tag.CreateInput{Name: " awk "})

	require.NoError(t, err)

	listed := listedTags(t, app)
	require.Len(t, listed, 1)
	assert.Equal(t, created.ID(), listed[0].Tag.ID())
	assert.Equal(t, "awk", listed[0].Tag.Name().String())
	assert.Zero(t, listed[0].SnippetCount)
}

func TestCreatingATagWhoseNameExistsIgnoringCaseIsRefused(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	existing := testapp.SeedTag(t, app, tag.CreateInput{Name: "go"})

	_, err := app.CreateTag.Run(t.Context(), tag.CreateInput{Name: "Go"})

	require.ErrorIs(t, err, domain.ErrTagNameTaken)
	assert.Equal(t, []browse.TagCount{{Tag: existing, SnippetCount: 0}}, listedTags(t, app))
}

func TestRenamingATagOntoAnotherMergesThemLeavingEachSnippetWithTheSurvivorOnce(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	golang := testapp.SeedTag(t, app, tag.CreateInput{Name: "go"})
	merged := testapp.SeedTag(t, app, tag.CreateInput{Name: "golang"})
	both := testapp.SeedSnippet(t, app, snippet.CreateInput{Tags: []domain.Tag{golang, merged}})
	onlyMerged := testapp.SeedSnippet(t, app, snippet.CreateInput{Tags: []domain.Tag{merged}})

	survivor, err := app.RenameTag.Run(t.Context(), tag.RenameInput{TagID: merged.ID(), Name: "Go"})

	require.NoError(t, err)
	assert.Equal(t, golang.ID(), survivor.ID())
	assert.Equal(t, "Go", survivor.Name().String())

	listed := listedTags(t, app)
	require.Len(t, listed, 1)
	assert.Equal(t, survivor.ID(), listed[0].Tag.ID())
	assert.Equal(t, 2, listed[0].SnippetCount)

	for _, carrier := range []domain.Snippet{both, onlyMerged} {
		assert.Equal(t, []domain.TagID{survivor.ID()}, tagIDsOf(testapp.StoredSnippet(t, app, carrier.ID())))
	}
}

func tagIDsOf(carrier domain.Snippet) []domain.TagID {
	ids := make([]domain.TagID, 0, len(carrier.Tags()))
	for _, carried := range carrier.Tags() {
		ids = append(ids, carried.ID())
	}

	return ids
}

func TestTagDeletedFromEverySnippetAfterPreviewCountsThem(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	golang := testapp.SeedTag(t, app, tag.CreateInput{Name: "go"})
	docker := testapp.SeedTag(t, app, tag.CreateInput{Name: "docker"})
	carrier := testapp.SeedSnippet(t, app, snippet.CreateInput{Tags: []domain.Tag{golang, docker}})
	testapp.SeedSnippet(t, app, snippet.CreateInput{Tags: []domain.Tag{golang}})

	preview, err := app.PreviewDeleteTag.Run(t.Context(), tag.PreviewDeleteInput{TagID: golang.ID()})
	require.NoError(t, err)
	assert.Equal(t, tag.DeletePreview{Tag: golang, SnippetCount: 2}, preview)

	require.NoError(t, app.DeleteTag.Run(t.Context(), tag.DeleteInput{TagID: golang.ID()}))

	assert.Equal(t, []browse.TagCount{{Tag: docker, SnippetCount: 1}}, listedTags(t, app))
	assert.Equal(t, []domain.Tag{docker}, testapp.StoredSnippet(t, app, carrier.ID()).Tags())
}

func listedTags(t *testing.T, app *bootstrap.App) []browse.TagCount {
	t.Helper()

	listed, err := app.TagList.Run(t.Context(), browse.TagListInput{})
	require.NoError(t, err)

	return listed
}

func TestTagListCountsTheSnippetsCarryingEachTag(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	golang := testapp.SeedTag(t, app, tag.CreateInput{Name: "go"})
	docker := testapp.SeedTag(t, app, tag.CreateInput{Name: "docker"})
	unused := testapp.SeedTag(t, app, tag.CreateInput{Name: "unused"})
	testapp.SeedSnippet(t, app, snippet.CreateInput{Tags: []domain.Tag{golang, docker}})
	testapp.SeedSnippet(t, app, snippet.CreateInput{Tags: []domain.Tag{golang}})

	listed, err := app.TagList.Run(t.Context(), browse.TagListInput{})

	require.NoError(t, err)
	assert.Equal(t, []browse.TagCount{
		{Tag: docker, SnippetCount: 1},
		{Tag: golang, SnippetCount: 2},
		{Tag: unused, SnippetCount: 0},
	}, listed)
}

func TestSelectingATagListsEverySnippetCarryingItAcrossFolders(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	scripts := testapp.SeedFolder(t, app, folder.CreateInput{Name: "scripts"})
	golang := testapp.SeedTag(t, app, tag.CreateInput{Name: "go"})
	atRoot := testapp.SeedSnippet(t, app, snippet.CreateInput{Title: "at the Root", Tags: []domain.Tag{golang}})
	filed := testapp.SeedSnippet(t, app, snippet.CreateInput{
		Title: "filed", FolderID: scripts.ID(), Tags: []domain.Tag{golang},
	})
	testapp.SeedSnippet(t, app, snippet.CreateInput{Title: "untagged", FolderID: scripts.ID()})

	listed, err := app.SnippetsWithTag.Run(
		t.Context(),
		browse.SnippetsWithTagInput{TagID: golang.ID(), Order: domain.SortOrderTitle},
	)

	require.NoError(t, err)
	assert.Equal(t, []domain.Snippet{atRoot, filed}, listed)
}

func TestSearchFindsASnippetByATagAlone(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	kubernetes := testapp.SeedTag(t, app, tag.CreateInput{Name: "kubernetes"})
	tagged := testapp.SeedSnippet(t, app, snippet.CreateInput{Title: "notes", Tags: []domain.Tag{kubernetes}})
	testapp.SeedSnippet(t, app, snippet.CreateInput{Title: "other"})

	hits, err := app.Query.Run(t.Context(), search.QueryInput{Text: "kubernetes"})

	require.NoError(t, err)
	require.Len(t, hits, 1)
	assert.Equal(t, tagged.ID(), hits[0].Snippet().ID())
}
