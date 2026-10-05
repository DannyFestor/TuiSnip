//go:build feature

package feature_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/app/folder"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/bootstrap"
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

func TestFolderDeletedWithItsSubtreeAfterPreviewCountsIt(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	ids := testkit.NewSequentialIDs()
	golang := seededFolder(t, app, testkit.FolderSpec{ID: ids.NewFolderID(), Name: "go"})
	nested := seededFolder(t, app, testkit.FolderSpec{ID: ids.NewFolderID(), Name: "testing", ParentID: golang.ID()})
	docker := seededFolder(t, app, testkit.FolderSpec{ID: ids.NewFolderID(), Name: "docker"})
	seededSnippet(t, app, ids, golang.ID())
	deleted := seededSnippet(t, app, ids, nested.ID())
	kept := seededSnippet(t, app, ids, docker.ID())

	preview, err := app.PreviewDeleteFolder.Run(t.Context(), folder.PreviewDeleteInput{FolderID: golang.ID()})
	require.NoError(t, err)
	assert.Equal(t, 1, preview.SubfolderCount)
	assert.Equal(t, 2, preview.SnippetCount)

	require.NoError(t, app.DeleteFolder.Run(t.Context(), folder.DeleteInput{FolderID: golang.ID()}))

	tree, err := app.FolderTree.Run(t.Context(), browse.FolderTreeInput{})
	require.NoError(t, err)
	require.Len(t, tree.Folders, 1)
	assert.Equal(t, docker.ID(), tree.Folders[0].Folder.ID())
	_, err = app.SnippetRepository.Find(t.Context(), deleted.ID())
	require.ErrorIs(t, err, domain.ErrNotFound)
	_, err = app.SnippetRepository.Find(t.Context(), kept.ID())
	assert.NoError(t, err)
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

func TestFolderDefaultLanguageChangeLeavesItsSnippetsAlone(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	ids := testkit.NewSequentialIDs()
	golang := seededFolder(t, app, testkit.FolderSpec{ID: ids.NewFolderID(), Name: "go"})
	filed := seededTaggedSnippet(t, app, ids, testkit.SnippetSpec{
		FolderID: golang.ID(),
		Fragment: testkit.FragmentSpec{Language: "Bash"},
	})

	_, err := app.SetFolderDefaultLanguage.Run(
		t.Context(), folder.SetDefaultLanguageInput{FolderID: golang.ID(), Language: "Go"},
	)

	require.NoError(t, err)
	tree, err := app.FolderTree.Run(t.Context(), browse.FolderTreeInput{})
	require.NoError(t, err)
	require.Len(t, tree.Folders, 1)
	assert.Equal(t, "Go", tree.Folders[0].Folder.DefaultLanguage().String())
	kept, err := app.SnippetRepository.Find(t.Context(), filed.ID())
	require.NoError(t, err)
	assert.Equal(t, filed, kept)
}

func TestNewSnippetTakesTheChangedFolderDefaultLanguage(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	ids := testkit.NewSequentialIDs()
	golang := seededFolder(t, app, testkit.FolderSpec{ID: ids.NewFolderID(), Name: "go"})
	existing := seededTaggedSnippet(t, app, ids, testkit.SnippetSpec{
		FolderID: golang.ID(),
		Fragment: testkit.FragmentSpec{Language: "Bash"},
	})
	_, err := app.SetFolderDefaultLanguage.Run(
		t.Context(), folder.SetDefaultLanguageInput{FolderID: golang.ID(), Language: "Go"},
	)
	require.NoError(t, err)

	created := create(t, app, snippet.CreateInput{
		Title:       "table test",
		Description: "",
		Language:    defaultLanguageOf(t, app, golang.ID()),
		Content:     "for _, tt := range tests {}\n",
		FolderID:    golang.ID(),
		Tags:        nil,
	})

	assert.Equal(t, "Go", created.FirstFragment().Language().String())
	kept, err := app.SnippetRepository.Find(t.Context(), existing.ID())
	require.NoError(t, err)
	assert.Equal(t, existing, kept)
}

func defaultLanguageOf(t *testing.T, app *bootstrap.App, folderID domain.FolderID) string {
	t.Helper()

	tree, err := app.FolderTree.Run(t.Context(), browse.FolderTreeInput{})
	require.NoError(t, err)

	for index := range tree.Folders {
		if filed := tree.Folders[index].Folder; filed.ID() == folderID {
			return filed.DefaultLanguage().String()
		}
	}

	require.FailNow(t, "Folder not in the tree", folderID)

	return ""
}

func TestFolderCreatedInsideCopiesTheChangedDefaultLanguage(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	golang := seededFolder(t, app, testkit.FolderSpec{ID: testkit.NewSequentialIDs().NewFolderID(), Name: "go"})
	_, err := app.SetFolderDefaultLanguage.Run(
		t.Context(), folder.SetDefaultLanguageInput{FolderID: golang.ID(), Language: "Go"},
	)
	require.NoError(t, err)

	created, err := app.CreateFolder.Run(t.Context(), folder.CreateInput{Name: "testing", ParentID: golang.ID()})

	require.NoError(t, err)
	assert.Equal(t, "Go", created.DefaultLanguage().String())
}

func TestFolderDefaultLanguageRefusesAnUnknownLanguage(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	golang := seededFolder(t, app, testkit.FolderSpec{ID: testkit.NewSequentialIDs().NewFolderID(), Name: "go"})

	_, err := app.SetFolderDefaultLanguage.Run(
		t.Context(), folder.SetDefaultLanguageInput{FolderID: golang.ID(), Language: "golang"},
	)

	require.ErrorIs(t, err, value.ErrUnknownLanguage)
	stored, err := app.FolderRepository.Find(t.Context(), golang.ID())
	require.NoError(t, err)
	assert.Equal(t, value.PlainText(), stored.DefaultLanguage())
}
