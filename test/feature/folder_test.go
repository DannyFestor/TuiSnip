//go:build feature

package feature_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/app/folder"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
	"github.com/DannyFestor/TuiSnip/test/testapp"
)

func TestFolderCreatedAtRootGetsPlainText(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)

	created, err := app.CreateFolder.Run(t.Context(), folder.CreateInput{Name: "docker"})

	require.NoError(t, err)
	stored := testapp.StoredFolder(t, app, created.ID())
	assert.True(t, stored.AtRoot())
	assert.Equal(t, value.PlainText(), stored.DefaultLanguage())
}

func TestFolderCreatedInsideAnotherCopiesItsDefaultLanguage(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	golang := testapp.SeedFolder(t, app, folder.CreateInput{Name: "go"})
	_, err := app.SetFolderDefaultLanguage.Run(
		t.Context(), folder.SetDefaultLanguageInput{FolderID: golang.ID(), Language: "Go"},
	)
	require.NoError(t, err)

	created, err := app.CreateFolder.Run(t.Context(), folder.CreateInput{Name: "testing", ParentID: golang.ID()})

	require.NoError(t, err)
	stored := testapp.StoredFolder(t, app, created.ID())
	assert.Equal(t, golang.ID(), stored.ParentID())
	assert.Equal(t, "Go", stored.DefaultLanguage().String())
}

func TestFolderDeletedWithItsSubtreeAfterPreviewCountsIt(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	golang := testapp.SeedFolder(t, app, folder.CreateInput{Name: "go"})
	nested := testapp.SeedFolder(t, app, folder.CreateInput{Name: "testing", ParentID: golang.ID()})
	docker := testapp.SeedFolder(t, app, folder.CreateInput{Name: "docker"})
	testapp.SeedSnippet(t, app, snippet.CreateInput{FolderID: golang.ID()})
	deleted := testapp.SeedSnippet(t, app, snippet.CreateInput{FolderID: nested.ID()})
	kept := testapp.SeedSnippet(t, app, snippet.CreateInput{FolderID: docker.ID()})

	preview, err := app.PreviewDeleteFolder.Run(t.Context(), folder.PreviewDeleteInput{FolderID: golang.ID()})
	require.NoError(t, err)
	assert.Equal(t, 1, preview.SubfolderCount)
	assert.Equal(t, 2, preview.SnippetCount)

	require.NoError(t, app.DeleteFolder.Run(t.Context(), folder.DeleteInput{FolderID: golang.ID()}))

	tree, err := app.FolderTree.Run(t.Context(), browse.FolderTreeInput{})
	require.NoError(t, err)
	require.Len(t, tree.Folders, 1)
	assert.Equal(t, docker.ID(), tree.Folders[0].Folder.ID())
	_, deletedStored := testapp.FindSnippet(t, app, deleted.ID())
	assert.False(t, deletedStored)
	_, keptStored := testapp.FindSnippet(t, app, kept.ID())
	assert.True(t, keptStored)
}

func TestRenamedFolderShowsInTree(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	golang := testapp.SeedFolder(t, app, folder.CreateInput{Name: "go"})

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
	golang := testapp.SeedFolder(t, app, folder.CreateInput{Name: "go"})
	filed := testapp.SeedSnippet(t, app, snippet.CreateInput{FolderID: golang.ID(), Language: "Bash"})

	_, err := app.SetFolderDefaultLanguage.Run(
		t.Context(), folder.SetDefaultLanguageInput{FolderID: golang.ID(), Language: "Go"},
	)

	require.NoError(t, err)
	assert.Equal(t, "Go", testapp.StoredFolder(t, app, golang.ID()).DefaultLanguage().String())
	assert.Equal(t, filed, testapp.StoredSnippet(t, app, filed.ID()))
}

func TestNewSnippetTakesTheChangedFolderDefaultLanguage(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	golang := testapp.SeedFolder(t, app, folder.CreateInput{Name: "go"})
	existing := testapp.SeedSnippet(t, app, snippet.CreateInput{FolderID: golang.ID(), Language: "Bash"})
	_, err := app.SetFolderDefaultLanguage.Run(
		t.Context(), folder.SetDefaultLanguageInput{FolderID: golang.ID(), Language: "Go"},
	)
	require.NoError(t, err)

	created := create(t, app, snippet.CreateInput{
		Title:       "table test",
		Description: "",
		Language:    testapp.StoredFolder(t, app, golang.ID()).DefaultLanguage().String(),
		Content:     "for _, tt := range tests {}\n",
		FolderID:    golang.ID(),
		Tags:        nil,
	})

	assert.Equal(t, "Go", created.FirstFragment().Language().String())
	assert.Equal(t, existing, testapp.StoredSnippet(t, app, existing.ID()))
}

func TestFolderDefaultLanguageRefusesAnUnknownLanguage(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	golang := testapp.SeedFolder(t, app, folder.CreateInput{Name: "go"})

	_, err := app.SetFolderDefaultLanguage.Run(
		t.Context(), folder.SetDefaultLanguageInput{FolderID: golang.ID(), Language: "golang"},
	)

	require.ErrorIs(t, err, value.ErrUnknownLanguage)
	assert.Equal(t, value.PlainText(), testapp.StoredFolder(t, app, golang.ID()).DefaultLanguage())
}
