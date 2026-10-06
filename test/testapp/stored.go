package testapp

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/bootstrap"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

func StoredFolder(t *testing.T, app *bootstrap.App, id domain.FolderID) domain.Folder {
	t.Helper()

	found, stored := withID(storedFolders(t, app), id)
	require.True(t, stored, "Folder %v is not stored", id)

	return found
}

func StoredTag(t *testing.T, app *bootstrap.App, id domain.TagID) domain.Tag {
	t.Helper()

	found, stored := withID(storedTags(t, app), id)
	require.True(t, stored, "Tag %v is not stored", id)

	return found
}

func StoredSnippet(t *testing.T, app *bootstrap.App, id domain.SnippetID) domain.Snippet {
	t.Helper()

	found, stored := FindSnippet(t, app, id)
	require.True(t, stored, "Snippet %v is not stored", id)

	return found
}

// FindSnippet searches the Root and every Folder, because no Action reads one Snippet by its ID.
func FindSnippet(t *testing.T, app *bootstrap.App, id domain.SnippetID) (domain.Snippet, bool) {
	t.Helper()

	for _, folderID := range storedFolderIDs(t, app) {
		found, stored := withID(snippetsIn(t, app, folderID), id)
		if stored {
			return found, true
		}
	}

	return domain.Snippet{}, false
}

func withID[E interface{ ID() ID }, ID comparable](items []E, id ID) (E, bool) {
	index := slices.IndexFunc(items, func(item E) bool { return item.ID() == id })
	if index < 0 {
		var missing E

		return missing, false
	}

	return items[index], true
}

func storedFolderIDs(t *testing.T, app *bootstrap.App) []domain.FolderID {
	t.Helper()

	folders := storedFolders(t, app)

	ids := make([]domain.FolderID, 0, len(folders)+1)
	ids = append(ids, domain.FolderID{})

	for _, stored := range folders {
		ids = append(ids, stored.ID())
	}

	return ids
}

func storedFolders(t *testing.T, app *bootstrap.App) []domain.Folder {
	t.Helper()

	tree, err := app.FolderTree.Run(t.Context(), browse.FolderTreeInput{})
	require.NoError(t, err)

	return flattened(tree.Folders)
}

func flattened(nodes []browse.FolderNode) []domain.Folder {
	folders := make([]domain.Folder, 0, len(nodes))
	for index := range nodes {
		folders = append(folders, nodes[index].Folder)
		folders = append(folders, flattened(nodes[index].Children)...)
	}

	return folders
}

func storedTags(t *testing.T, app *bootstrap.App) []domain.Tag {
	t.Helper()

	counts, err := app.TagList.Run(t.Context(), browse.TagListInput{})
	require.NoError(t, err)

	tags := make([]domain.Tag, 0, len(counts))
	for _, count := range counts {
		tags = append(tags, count.Tag)
	}

	return tags
}

func snippetsIn(t *testing.T, app *bootstrap.App, folderID domain.FolderID) []domain.Snippet {
	t.Helper()

	listed, err := app.SnippetsInFolder.Run(t.Context(), browse.SnippetsInFolderInput{
		FolderID: folderID,
		Order:    domain.SortOrderTitle,
	})
	require.NoError(t, err)

	return listed
}
