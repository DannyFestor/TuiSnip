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

func TestMovedSnippetKeepsItsLanguageInItsNewFolder(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	golang := testapp.SeedFolder(t, app, folder.CreateInput{Name: "go"})
	docker := testapp.SeedFolder(t, app, folder.CreateInput{Name: "docker"})
	filed := testapp.SeedSnippet(t, app, snippet.CreateInput{FolderID: golang.ID(), Language: "Bash"})

	_, err := app.Move.Run(t.Context(), snippet.MoveInput{SnippetID: filed.ID(), FolderID: docker.ID()})

	require.NoError(t, err)
	stored := testapp.StoredSnippet(t, app, filed.ID())
	assert.Equal(t, docker.ID(), stored.FolderID())
	assert.Equal(t, "Bash", stored.FirstFragment().Language().String())
	assert.Equal(t, filed.UpdatedAt(), stored.UpdatedAt())
}

func TestSnippetMovedToTheRoot(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	golang := testapp.SeedFolder(t, app, folder.CreateInput{Name: "go"})
	filed := testapp.SeedSnippet(t, app, snippet.CreateInput{FolderID: golang.ID()})

	_, err := app.Move.Run(t.Context(), snippet.MoveInput{SnippetID: filed.ID(), FolderID: domain.FolderID{}})

	require.NoError(t, err)
	assert.True(t, testapp.StoredSnippet(t, app, filed.ID()).AtRoot())
}

func TestMovedFolderTakesItsSubtreeAlong(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	golang := testapp.SeedFolder(t, app, folder.CreateInput{Name: "go"})
	nested := testapp.SeedFolder(t, app, folder.CreateInput{Name: "testing", ParentID: golang.ID()})
	languages := testapp.SeedFolder(t, app, folder.CreateInput{Name: "languages"})
	testapp.SeedSnippet(t, app, snippet.CreateInput{FolderID: nested.ID()})

	_, err := app.MoveFolder.Run(t.Context(), folder.MoveInput{FolderID: golang.ID(), ParentID: languages.ID()})

	require.NoError(t, err)
	tree, err := app.FolderTree.Run(t.Context(), browse.FolderTreeInput{})
	require.NoError(t, err)
	require.Len(t, tree.Folders, 1)
	moved := tree.Folders[0].Children
	require.Len(t, moved, 1)
	assert.Equal(t, golang.ID(), moved[0].Folder.ID())
	require.Len(t, moved[0].Children, 1)
	assert.Equal(t, nested.ID(), moved[0].Children[0].Folder.ID())
	assert.Equal(t, 1, moved[0].Children[0].SnippetCount)
}

func TestFolderMovedToTheRoot(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	golang := testapp.SeedFolder(t, app, folder.CreateInput{Name: "go"})
	nested := testapp.SeedFolder(t, app, folder.CreateInput{Name: "testing", ParentID: golang.ID()})

	_, err := app.MoveFolder.Run(t.Context(), folder.MoveInput{FolderID: nested.ID(), ParentID: domain.FolderID{}})

	require.NoError(t, err)
	assert.True(t, testapp.StoredFolder(t, app, nested.ID()).AtRoot())
}

func TestFolderMoveIntoItsOwnSubtreeIsRefused(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	golang := testapp.SeedFolder(t, app, folder.CreateInput{Name: "go"})
	nested := testapp.SeedFolder(t, app, folder.CreateInput{Name: "testing", ParentID: golang.ID()})
	deeper := testapp.SeedFolder(t, app, folder.CreateInput{Name: "mocks", ParentID: nested.ID()})

	for _, parentID := range []domain.FolderID{golang.ID(), nested.ID(), deeper.ID()} {
		_, err := app.MoveFolder.Run(t.Context(), folder.MoveInput{FolderID: golang.ID(), ParentID: parentID})

		require.ErrorIs(t, err, domain.ErrFolderCycle)
	}

	assert.True(t, testapp.StoredFolder(t, app, golang.ID()).AtRoot())
}
