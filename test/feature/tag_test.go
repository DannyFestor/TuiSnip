//go:build feature

package feature_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/app/search"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/testapp"
)

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
