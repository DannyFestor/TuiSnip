//go:build feature

package feature_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/app/search"
	"github.com/DannyFestor/TuiSnip/internal/app/tag"
	"github.com/DannyFestor/TuiSnip/internal/bootstrap"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
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
	existing := seededTag(t, app, testkit.TagSpec{ID: testkit.NewSequentialIDs().NewTagID(), Name: "go"})

	_, err := app.CreateTag.Run(t.Context(), tag.CreateInput{Name: "Go"})

	require.ErrorIs(t, err, domain.ErrTagNameTaken)
	assert.Equal(t, []browse.TagCount{{Tag: existing, SnippetCount: 0}}, listedTags(t, app))
}

func TestRenamingATagOntoAnotherMergesThemLeavingEachSnippetWithTheSurvivorOnce(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	ids := testkit.NewSequentialIDs()
	golang := seededTag(t, app, testkit.TagSpec{ID: ids.NewTagID(), Name: "go"})
	merged := seededTag(t, app, testkit.TagSpec{ID: ids.NewTagID(), Name: "golang"})
	both := seededTaggedSnippet(t, app, ids, testkit.SnippetSpec{Tags: []domain.Tag{golang, merged}})
	onlyMerged := seededTaggedSnippet(t, app, ids, testkit.SnippetSpec{Tags: []domain.Tag{merged}})

	survivor, err := app.RenameTag.Run(t.Context(), tag.RenameInput{TagID: merged.ID(), Name: "Go"})

	require.NoError(t, err)
	assert.Equal(t, golang.ID(), survivor.ID())
	assert.Equal(t, "Go", survivor.Name().String())

	listed := listedTags(t, app)
	require.Len(t, listed, 1)
	assert.Equal(t, survivor.ID(), listed[0].Tag.ID())
	assert.Equal(t, 2, listed[0].SnippetCount)

	for _, carrier := range []domain.Snippet{both, onlyMerged} {
		stored, findErr := app.SnippetRepository.Find(t.Context(), carrier.ID())
		require.NoError(t, findErr)
		assert.Equal(t, []domain.TagID{survivor.ID()}, tagIDsOf(stored))
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
	ids := testkit.NewSequentialIDs()
	golang := seededTag(t, app, testkit.TagSpec{ID: ids.NewTagID(), Name: "go"})
	docker := seededTag(t, app, testkit.TagSpec{ID: ids.NewTagID(), Name: "docker"})
	carrier := seededTaggedSnippet(t, app, ids, testkit.SnippetSpec{Tags: []domain.Tag{golang, docker}})
	seededTaggedSnippet(t, app, ids, testkit.SnippetSpec{Tags: []domain.Tag{golang}})

	preview, err := app.PreviewDeleteTag.Run(t.Context(), tag.PreviewDeleteInput{TagID: golang.ID()})
	require.NoError(t, err)
	assert.Equal(t, tag.DeletePreview{Tag: golang, SnippetCount: 2}, preview)

	require.NoError(t, app.DeleteTag.Run(t.Context(), tag.DeleteInput{TagID: golang.ID()}))

	assert.Equal(t, []browse.TagCount{{Tag: docker, SnippetCount: 1}}, listedTags(t, app))
	stored, err := app.SnippetRepository.Find(t.Context(), carrier.ID())
	require.NoError(t, err)
	assert.Equal(t, []domain.Tag{docker}, stored.Tags())
}

func listedTags(t *testing.T, app *bootstrap.App) []browse.TagCount {
	t.Helper()

	listed, err := app.TagList.Run(t.Context(), browse.TagListInput{})
	require.NoError(t, err)

	return listed
}

func TestTagNamesDifferingOnlyInCaseAreOneTagKeepingTheFirstSpelling(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	ids := testkit.NewSequentialIDs()
	first := seededTag(t, app, testkit.TagSpec{ID: ids.NewTagID(), Name: "Go"})

	err := app.TagRepository.Insert(t.Context(), testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "go"}))

	require.ErrorIs(t, err, domain.ErrTagNameTaken)

	listed, err := app.TagList.Run(t.Context(), browse.TagListInput{})
	require.NoError(t, err)
	assert.Equal(t, []browse.TagCount{{Tag: first, SnippetCount: 0}}, listed)
}

func TestTagListCountsTheSnippetsCarryingEachTag(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	ids := testkit.NewSequentialIDs()
	golang := seededTag(t, app, testkit.TagSpec{ID: ids.NewTagID(), Name: "go"})
	docker := seededTag(t, app, testkit.TagSpec{ID: ids.NewTagID(), Name: "docker"})
	unused := seededTag(t, app, testkit.TagSpec{ID: ids.NewTagID(), Name: "unused"})
	seededTaggedSnippet(t, app, ids, testkit.SnippetSpec{Tags: []domain.Tag{golang, docker}})
	seededTaggedSnippet(t, app, ids, testkit.SnippetSpec{Tags: []domain.Tag{golang}})

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
	ids := testkit.NewSequentialIDs()
	folder := seededFolder(t, app, testkit.FolderSpec{ID: ids.NewFolderID(), Name: "scripts"})
	golang := seededTag(t, app, testkit.TagSpec{ID: ids.NewTagID(), Name: "go"})
	atRoot := seededTaggedSnippet(t, app, ids, testkit.SnippetSpec{Title: "at the Root", Tags: []domain.Tag{golang}})
	filed := seededTaggedSnippet(t, app, ids, testkit.SnippetSpec{
		Title: "filed", FolderID: folder.ID(), Tags: []domain.Tag{golang},
	})
	seededTaggedSnippet(t, app, ids, testkit.SnippetSpec{Title: "untagged", FolderID: folder.ID()})

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
	ids := testkit.NewSequentialIDs()
	kubernetes := seededTag(t, app, testkit.TagSpec{ID: ids.NewTagID(), Name: "kubernetes"})
	tagged := seededTaggedSnippet(t, app, ids, testkit.SnippetSpec{Title: "notes", Tags: []domain.Tag{kubernetes}})
	seededTaggedSnippet(t, app, ids, testkit.SnippetSpec{Title: "other"})

	hits, err := app.Query.Run(t.Context(), search.QueryInput{Text: "kubernetes"})

	require.NoError(t, err)
	require.Len(t, hits, 1)
	assert.Equal(t, tagged.ID(), hits[0].Snippet().ID())
}
