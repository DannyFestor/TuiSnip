//go:build feature

package feature_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/app/folder"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/test/testapp"
)

func TestSeededFoldersListAsTreeWithSnippetCounts(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	golang := testapp.SeedFolder(t, app, folder.CreateInput{Name: "go"})
	docker := testapp.SeedFolder(t, app, folder.CreateInput{Name: "docker"})
	tests := testapp.SeedFolder(t, app, folder.CreateInput{Name: "testing", ParentID: golang.ID()})
	testapp.SeedSnippet(t, app, snippet.CreateInput{FolderID: golang.ID()})
	testapp.SeedSnippet(t, app, snippet.CreateInput{FolderID: tests.ID()})
	testapp.SeedSnippet(t, app, snippet.CreateInput{Title: "at the Root"})

	tree, err := app.FolderTree.Run(t.Context(), browse.FolderTreeInput{})

	require.NoError(t, err)
	assert.Equal(t, browse.Tree{
		RootSnippetCount: 1,
		Folders: []browse.FolderNode{
			{Folder: docker, SnippetCount: 0, Children: nil},
			{Folder: golang, SnippetCount: 1, Children: []browse.FolderNode{
				{Folder: tests, SnippetCount: 1, Children: nil},
			}},
		},
	}, tree)
}

func TestSnippetsInFolderListOnlyThatFolder(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	golang := testapp.SeedFolder(t, app, folder.CreateInput{Name: "go"})
	filed := testapp.SeedSnippet(t, app, snippet.CreateInput{FolderID: golang.ID()})
	testapp.SeedSnippet(t, app, snippet.CreateInput{Title: "at the Root"})

	listed, err := app.SnippetsInFolder.Run(t.Context(), browse.SnippetsInFolderInput{
		FolderID: golang.ID(),
		Order:    domain.SortOrderTitle,
	})

	require.NoError(t, err)
	require.Len(t, listed, 1)
	assert.Equal(t, filed.ID(), listed[0].ID())
}
