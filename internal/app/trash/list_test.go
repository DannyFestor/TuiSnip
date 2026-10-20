package trash_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/trash"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

var errDatabaseLocked = errors.New("database is locked")

func TestNewList(t *testing.T) {
	t.Parallel()

	t.Run("names the missing dependency", func(t *testing.T) {
		t.Parallel()

		_, err := trash.NewList(nil)

		require.ErrorIs(t, err, domain.ErrMissingDependency)
		assert.ErrorContains(t, err, "trash.NewList: loader")
	})
}

func TestList_Run(t *testing.T) {
	t.Parallel()

	t.Run("lists the Trash items newest deleted first, then by id descending", func(t *testing.T) {
		t.Parallel()

		shelved := newShelved(t)
		older := shelved.snippet(t, "older", domain.FolderID{})
		first := shelved.snippet(t, "first", domain.FolderID{})
		second := shelved.snippet(t, "second", domain.FolderID{})
		shelved.trashSnippet(t, older, deletedAt().Add(-time.Hour), nil)
		shelved.trashSnippet(t, first, deletedAt(), nil)
		shelved.trashSnippet(t, second, deletedAt(), nil)

		got := listed(t, shelved.library)

		assert.Equal(t, []string{"second", "first", "older"}, namesOf(got))
	})

	t.Run("gives a Snippet the current path of its Folder, a trashed one included", func(t *testing.T) {
		t.Parallel()

		shelved := newShelved(t)
		golang := shelved.folder(t, "go", domain.FolderID{})
		nested := shelved.folder(t, "testing", golang.ID())
		shelved.trashFolder(t, golang, deletedAt().Add(-time.Hour))
		trashed := shelved.snippet(t, "Table test", nested.ID())
		shelved.trashSnippet(t, trashed, deletedAt(), nil)

		got := listed(t, shelved.library)

		require.Len(t, got, 2)
		assert.Equal(t, trashed, got[0].Snippet)
		assert.Equal(t, []string{"go", "testing"}, got[0].Path)
		assert.Empty(t, got[1].Path)
	})

	t.Run("marks each item as a Snippet or a Folder", func(t *testing.T) {
		t.Parallel()

		shelved := newShelved(t)
		docker := shelved.folder(t, "docker", domain.FolderID{})
		prune := shelved.snippet(t, "Prune everything", domain.FolderID{})
		shelved.trashFolder(t, docker, deletedAt().Add(-time.Hour))
		shelved.trashSnippet(t, prune, deletedAt(), nil)

		got := listed(t, shelved.library)

		require.Len(t, got, 2)
		assert.Equal(t, trash.ItemKindSnippet, got[0].Kind)
		assert.Equal(t, trash.ItemKindFolder, got[1].Kind)
	})

	t.Run("gives a detached item its former path", func(t *testing.T) {
		t.Parallel()

		shelved := newShelved(t)
		detached := shelved.snippet(t, "Prune everything", domain.FolderID{})
		shelved.trashSnippet(t, detached, deletedAt(), []string{"docker", "compose"})

		got := listed(t, shelved.library)

		require.Len(t, got, 1)
		assert.Equal(t, []string{"docker", "compose"}, got[0].Path)
	})

	t.Run("counts and outlines a Folder's subtree without its separate Trash items", func(t *testing.T) {
		t.Parallel()

		shelved := newShelved(t)
		docker := shelved.folder(t, "docker", domain.FolderID{})
		compose := shelved.folder(t, "compose", docker.ID())
		separate := shelved.folder(t, "separate", docker.ID())
		shelved.snippet(t, "zap images", docker.ID())
		shelved.snippet(t, "Build", docker.ID())
		shelved.snippet(t, "Up", compose.ID())
		shelved.snippet(t, "inside separate", separate.ID())
		trashedInside := shelved.snippet(t, "trashed inside", docker.ID())
		shelved.trashSnippet(t, trashedInside, deletedAt().Add(-2*time.Hour), nil)
		shelved.trashFolder(t, separate, deletedAt().Add(-time.Hour))
		shelved.trashFolder(t, docker, deletedAt())

		got := listed(t, shelved.library)

		require.Len(t, got, 3)
		assert.Equal(t, docker, got[0].Folder)
		assert.Equal(t, trash.Contents{
			SubfolderCount: 1,
			SnippetCount:   3,
			Outline: trash.Outline{
				Name:          "docker",
				SnippetTitles: []string{"Build", "zap images"},
				Subfolders: []trash.Outline{
					{Name: "compose", SnippetTitles: []string{"Up"}, Subfolders: []trash.Outline{}},
				},
			},
		}, got[0].Contents)
		assert.Equal(t, []string{"docker"}, got[1].Path)
	})

	t.Run("lists an empty Trash as no items", func(t *testing.T) {
		t.Parallel()

		shelved := newShelved(t)
		shelved.snippet(t, "live", domain.FolderID{})

		assert.Empty(t, listed(t, shelved.library))
	})

	t.Run("passes up a failed load", func(t *testing.T) {
		t.Parallel()

		loader := NewMockLoader(t)
		loader.EXPECT().Load(mock.Anything).Return(domain.Library{}, errDatabaseLocked)
		list, err := trash.NewList(loader)
		require.NoError(t, err)

		_, err = list.Run(t.Context(), trash.ListInput{})

		require.ErrorIs(t, err, errDatabaseLocked)
		assert.ErrorContains(t, err, "trash.List: ")
	})
}

func listed(t *testing.T, library domain.Library) []trash.Item {
	t.Helper()

	loader := NewMockLoader(t)
	loader.EXPECT().Load(mock.Anything).Return(library, nil)

	list, err := trash.NewList(loader)
	require.NoError(t, err)

	items, err := list.Run(t.Context(), trash.ListInput{})
	require.NoError(t, err)

	return items
}

func namesOf(items []trash.Item) []string {
	names := make([]string, 0, len(items))
	for index := range items {
		names = append(names, items[index].Name())
	}

	return names
}
