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
	"github.com/DannyFestor/TuiSnip/internal/app/trash"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/test/testapp"
)

func TestPermanentlyDeletingAFolderKeepsASeparatelyTrashedSnippetInTheTrash(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	docker := testapp.SeedFolder(t, app, folder.CreateInput{Name: "docker"})
	prune := testapp.SeedSnippet(t, app, snippet.CreateInput{Title: "Prune everything", FolderID: docker.ID()})
	require.NoError(t, app.Delete.Run(t.Context(), snippet.DeleteInput{SnippetID: prune.ID()}))
	require.NoError(t, app.DeleteFolder.Run(t.Context(), folder.DeleteInput{FolderID: docker.ID()}))

	err := app.DeletePermanently.Run(t.Context(), trash.DeletePermanentlyInput{Item: trash.FolderItem(docker.ID())})

	require.NoError(t, err)
	listed := trashItems(t, app)
	require.Len(t, listed, 1)
	assert.Equal(t, prune.ID(), listed[0].Snippet.ID())
	assert.Equal(t, []string{"docker"}, listed[0].Path)

	restored, err := app.Restore.Run(t.Context(), trash.RestoreInput{Item: trash.SnippetItem(prune.ID())})

	require.NoError(t, err)
	assert.Equal(
		t,
		domain.Restoration{FolderID: domain.FolderID{}, Fallback: domain.RestoreFallbackFolderDeleted},
		restored,
	)
	assert.True(t, testapp.StoredSnippet(t, app, prune.ID()).AtRoot())
}

func TestEmptyingTheTrashSparesASnippetTrashedAfterThePromptOpened(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	docker := testapp.SeedFolder(t, app, folder.CreateInput{Name: "docker"})
	testapp.SeedSnippet(t, app, snippet.CreateInput{Title: "Compose up", FolderID: docker.ID()})
	prune := testapp.SeedSnippet(t, app, snippet.CreateInput{Title: "Prune everything"})
	late := testapp.SeedSnippet(t, app, snippet.CreateInput{Title: "Trashed later"})
	require.NoError(t, app.DeleteFolder.Run(t.Context(), folder.DeleteInput{FolderID: docker.ID()}))
	require.NoError(t, app.Delete.Run(t.Context(), snippet.DeleteInput{SnippetID: prune.ID()}))
	counted := itemIDs(trashItems(t, app))
	require.NoError(t, app.Delete.Run(t.Context(), snippet.DeleteInput{SnippetID: late.ID()}))

	emptied, err := app.EmptyTrash.Run(t.Context(), trash.EmptyInput{Items: counted})

	require.NoError(t, err)
	assert.Equal(t, domain.TrashEmptied{Folders: 1, Snippets: 2}, emptied)
	listed := trashItems(t, app)
	require.Len(t, listed, 1)
	assert.Equal(t, late.ID(), listed[0].Snippet.ID())
}

func TestPermanentDeletionNeverDeletesTags(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	golang := testapp.SeedTag(t, app, tag.CreateInput{Name: "go"})
	shell := testapp.SeedTag(t, app, tag.CreateInput{Name: "shell"})
	deleted := testapp.SeedSnippet(t, app, snippet.CreateInput{Tags: []domain.Tag{golang}})
	emptied := testapp.SeedSnippet(t, app, snippet.CreateInput{Tags: []domain.Tag{shell}})
	require.NoError(t, app.Delete.Run(t.Context(), snippet.DeleteInput{SnippetID: deleted.ID()}))
	require.NoError(t, app.Delete.Run(t.Context(), snippet.DeleteInput{SnippetID: emptied.ID()}))

	require.NoError(t, app.DeletePermanently.Run(
		t.Context(), trash.DeletePermanentlyInput{Item: trash.SnippetItem(deleted.ID())},
	))
	_, err := app.EmptyTrash.Run(t.Context(), trash.EmptyInput{Items: itemIDs(trashItems(t, app))})

	require.NoError(t, err)
	assert.Empty(t, trashItems(t, app))
	assert.Equal(
		t, []browse.TagCount{{Tag: golang, SnippetCount: 0}, {Tag: shell, SnippetCount: 0}}, listedTags(t, app),
	)
}

func itemIDs(items []trash.Item) []trash.ItemID {
	ids := make([]trash.ItemID, 0, len(items))
	for index := range items {
		ids = append(ids, items[index].ID())
	}

	return ids
}
