//go:build feature

package feature_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/app/folder"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/testapp"
)

func TestFolderCreatedAtRootGetsPlainText(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)

	created, err := app.CreateFolder.Run(t.Context(), folder.CreateInput{Name: "docker", ParentID: domain.FolderID{}})

	require.NoError(t, err)
	stored, err := app.FolderRepository.Find(t.Context(), created.ID())
	require.NoError(t, err)
	assert.True(t, stored.AtRoot())
	assert.Equal(t, value.PlainText(), stored.DefaultLanguage())
}

func TestFolderCreatedInsideAnotherCopiesItsDefaultLanguage(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	golang := seededFolder(
		t,
		app,
		testkit.FolderSpec{ID: testkit.NewSequentialIDs().NewFolderID(), Name: "go", DefaultLanguage: "Go"},
	)

	created, err := app.CreateFolder.Run(t.Context(), folder.CreateInput{Name: "testing", ParentID: golang.ID()})

	require.NoError(t, err)
	stored, err := app.FolderRepository.Find(t.Context(), created.ID())
	require.NoError(t, err)
	assert.Equal(t, golang.ID(), stored.ParentID())
	assert.Equal(t, "Go", stored.DefaultLanguage().String())
}

func TestRenamedFolderShowsInTree(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	golang := seededFolder(t, app, testkit.FolderSpec{ID: testkit.NewSequentialIDs().NewFolderID(), Name: "go"})

	_, err := app.RenameFolder.Run(t.Context(), folder.RenameInput{FolderID: golang.ID(), Name: "golang"})

	require.NoError(t, err)
	tree, err := app.FolderTree.Run(t.Context(), browse.FolderTreeInput{})
	require.NoError(t, err)
	require.Len(t, tree.Folders, 1)
	assert.Equal(t, "golang", tree.Folders[0].Folder.Name().String())
}
