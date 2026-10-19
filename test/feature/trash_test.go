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
	"github.com/DannyFestor/TuiSnip/internal/app/trash"
	"github.com/DannyFestor/TuiSnip/internal/bootstrap"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/test/testapp"
)

func TestATagCarriedOnlyByTrashedSnippetsCountsNone(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	golang := testapp.SeedTag(t, app, tag.CreateInput{Name: "go"})
	docker := testapp.SeedFolder(t, app, folder.CreateInput{Name: "docker"})
	trashed := testapp.SeedSnippet(t, app, snippet.CreateInput{Tags: []domain.Tag{golang}})
	testapp.SeedSnippet(t, app, snippet.CreateInput{FolderID: docker.ID(), Tags: []domain.Tag{golang}})
	require.NoError(t, app.Delete.Run(t.Context(), snippet.DeleteInput{SnippetID: trashed.ID()}))
	require.NoError(t, app.DeleteFolder.Run(t.Context(), folder.DeleteInput{FolderID: docker.ID()}))

	preview, err := app.PreviewDeleteTag.Run(t.Context(), tag.PreviewDeleteInput{TagID: golang.ID()})

	require.NoError(t, err)
	assert.Zero(t, preview.SnippetCount)
}

func TestDeletingAFolderHidesItsSubtreeUntilItIsRestored(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	golang := testapp.SeedTag(t, app, tag.CreateInput{Name: "go"})
	docker := testapp.SeedFolder(t, app, folder.CreateInput{Name: "docker"})
	compose := testapp.SeedFolder(t, app, folder.CreateInput{Name: "compose", ParentID: docker.ID()})
	nested := testapp.SeedSnippet(t, app, snippet.CreateInput{
		Title: "Compose up", FolderID: compose.ID(), Tags: []domain.Tag{golang},
	})
	treeBefore := folderTree(t, app)
	require.NoError(t, app.DeleteFolder.Run(t.Context(), folder.DeleteInput{FolderID: docker.ID()}))

	hidden := folderTree(t, app)
	assert.Empty(t, hidden.Folders)
	assert.Equal(t, 1, hidden.TrashCount)
	assert.Equal(t, []browse.TagCount{{Tag: golang, SnippetCount: 0}}, listedTags(t, app))
	assert.Empty(t, searched(t, app, "Compose"))
	_, found := testapp.FindSnippet(t, app, nested.ID())
	assert.False(t, found)

	items := trashItems(t, app)
	require.Len(t, items, 1)
	assert.Equal(t, docker, items[0].Folder)
	assert.Equal(t, 1, items[0].Contents.SubfolderCount)
	assert.Equal(t, 1, items[0].Contents.SnippetCount)

	restored, err := app.Restore.Run(t.Context(), trash.RestoreInput{Item: trash.FolderItem(docker.ID())})

	require.NoError(t, err)
	assert.Equal(t, domain.Restoration{FolderID: domain.FolderID{}, Fallback: domain.RestoreFallbackNone}, restored)
	assert.Equal(t, treeBefore, folderTree(t, app))
	assert.Equal(t, nested, testapp.StoredSnippet(t, app, nested.ID()))
	assert.Empty(t, trashItems(t, app))
}

func TestASnippetWhoseFolderIsInTheTrashRestoresToTheRoot(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	docker := testapp.SeedFolder(t, app, folder.CreateInput{Name: "docker"})
	prune := testapp.SeedSnippet(t, app, snippet.CreateInput{Title: "Prune everything", FolderID: docker.ID()})
	require.NoError(t, app.Delete.Run(t.Context(), snippet.DeleteInput{SnippetID: prune.ID()}))
	require.NoError(t, app.DeleteFolder.Run(t.Context(), folder.DeleteInput{FolderID: docker.ID()}))

	restored, err := app.Restore.Run(t.Context(), trash.RestoreInput{Item: trash.SnippetItem(prune.ID())})

	require.NoError(t, err)
	assert.Equal(t, domain.RestoreFallbackFolderInTrash, restored.Fallback)
	assert.True(t, testapp.StoredSnippet(t, app, prune.ID()).AtRoot())
	items := trashItems(t, app)
	require.Len(t, items, 1)
	assert.Equal(t, docker, items[0].Folder)
}

func TestRestoringAnItemNoLongerInTheTrashIsReportedNotFound(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	prune := testapp.SeedSnippet(t, app, snippet.CreateInput{Title: "Prune everything"})
	require.NoError(t, app.Delete.Run(t.Context(), snippet.DeleteInput{SnippetID: prune.ID()}))
	restore := trash.RestoreInput{Item: trash.SnippetItem(prune.ID())}
	_, err := app.Restore.Run(t.Context(), restore)
	require.NoError(t, err)

	_, err = app.Restore.Run(t.Context(), restore)

	require.ErrorIs(t, err, domain.ErrNotFound)
}

func TestCopyingFromTheTrashCopiesOnlyATrashItem(t *testing.T) {
	t.Parallel()

	home, app := testapp.Start(t, testapp.RecordingTool)
	prune := testapp.SeedSnippet(
		t,
		app,
		snippet.CreateInput{Title: "Prune everything", Fragments: plainFragment("docker system prune")},
	)
	copyInput := snippet.CopyInput{SnippetID: prune.ID(), FragmentID: prune.FirstFragment().ID()}

	_, err := app.CopyFromTrash.Run(t.Context(), copyInput)
	require.ErrorIs(t, err, domain.ErrNotFound)

	require.NoError(t, app.Delete.Run(t.Context(), snippet.DeleteInput{SnippetID: prune.ID()}))
	_, err = app.CopyFromTrash.Run(t.Context(), copyInput)

	require.NoError(t, err)
	assert.Equal(t, "docker system prune", home.Copied(t))
}

func folderTree(t *testing.T, app *bootstrap.App) browse.Tree {
	t.Helper()

	tree, err := app.FolderTree.Run(t.Context(), browse.FolderTreeInput{})
	require.NoError(t, err)

	return tree
}

func searched(t *testing.T, app *bootstrap.App, text string) []domain.SearchHit {
	t.Helper()

	hits, err := app.Query.Run(t.Context(), search.QueryInput{Text: text})
	require.NoError(t, err)

	return hits
}

func trashItems(t *testing.T, app *bootstrap.App) []trash.Item {
	t.Helper()

	items, err := app.ListTrash.Run(t.Context(), trash.ListInput{})
	require.NoError(t, err)

	return items
}

func TestMovingASnippetIntoAFolderInTheTrashFails(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	docker := testapp.SeedFolder(t, app, folder.CreateInput{Name: "docker"})
	moving := testapp.SeedSnippet(t, app, snippet.CreateInput{})
	require.NoError(t, app.DeleteFolder.Run(t.Context(), folder.DeleteInput{FolderID: docker.ID()}))

	_, err := app.Move.Run(t.Context(), snippet.MoveInput{SnippetID: moving.ID(), FolderID: docker.ID()})

	require.ErrorIs(t, err, domain.ErrFolderGone)
	assert.True(t, testapp.StoredSnippet(t, app, moving.ID()).AtRoot())
}
