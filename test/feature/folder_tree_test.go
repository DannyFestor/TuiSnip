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

func TestSeededFoldersListAsTreeWithSnippetCounts(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	ids := testkit.NewSequentialIDs()
	golang := seededFolder(t, app, testkit.FolderSpec{ID: ids.NewFolderID(), Name: "go"})
	docker := seededFolder(t, app, testkit.FolderSpec{ID: ids.NewFolderID(), Name: "docker"})
	tests := seededFolder(t, app, testkit.FolderSpec{ID: ids.NewFolderID(), Name: "testing", ParentID: golang.ID()})
	seededSnippet(t, app, ids, golang.ID())
	seededSnippet(t, app, ids, tests.ID())
	create(t, app, snippet.CreateInput{Title: "at the Root", Description: "", Content: "ls"})

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
	ids := testkit.NewSequentialIDs()
	golang := seededFolder(t, app, testkit.FolderSpec{ID: ids.NewFolderID(), Name: "go"})
	filed := seededSnippet(t, app, ids, golang.ID())
	create(t, app, snippet.CreateInput{Title: "at the Root", Description: "", Content: "ls"})

	listed, err := app.SnippetsInFolder.Run(t.Context(), browse.SnippetsInFolderInput{
		FolderID: golang.ID(),
		Order:    domain.SortOrderTitle,
	})

	require.NoError(t, err)
	require.Len(t, listed, 1)
	assert.Equal(t, filed.ID(), listed[0].ID())
}
