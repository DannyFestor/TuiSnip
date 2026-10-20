package sqlite_test

import (
	"log/slog"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
)

func TestTrashRepository_CountItems(t *testing.T) {
	t.Parallel()

	t.Run("counts each Trash item once, a trashed Folder as one", func(t *testing.T) {
		t.Parallel()

		fixture := newTrashFixture(t)

		got, err := fixture.trash(t).CountItems(t.Context())

		require.NoError(t, err)
		assert.Equal(t, 2, got)
	})

	t.Run("counts nothing in an empty database", func(t *testing.T) {
		t.Parallel()

		got, err := newTrashRepository(t, openDatabase(t, newDatabasePath(t))).CountItems(t.Context())

		require.NoError(t, err)
		assert.Zero(t, got)
	})
}

func TestTrashRepository_Load(t *testing.T) {
	t.Parallel()

	t.Run("returns every Folder and only the Snippets in the Trash, with the Trash marks", func(t *testing.T) {
		t.Parallel()

		fixture := newTrashFixture(t)

		got, err := fixture.trash(t).Load(t.Context())

		require.NoError(t, err)
		assert.ElementsMatch(
			t, []domain.Folder{fixture.liveFolder, fixture.trashedFolder, fixture.nestedFolder}, got.Folders,
		)
		assert.ElementsMatch(t, []domain.Snippet{fixture.trashedSnippet, fixture.nestedSnippet}, got.Snippets)
		assert.Contains(t, got.Trash.Folders, fixture.trashedFolder.ID())
		assert.Contains(t, got.Trash.Snippets, fixture.trashedSnippet.ID())
		assert.Len(t, got.Trash.Folders, 1)
		assert.Len(t, got.Trash.Snippets, 1)
	})
}

func TestTrashRepository_Find(t *testing.T) {
	t.Parallel()

	t.Run("returns a Snippet Trash item with its Fragments and Tags", func(t *testing.T) {
		t.Parallel()

		fixture := newTrashFixture(t)

		got, err := fixture.trash(t).Find(t.Context(), fixture.trashedSnippet.ID())

		require.NoError(t, err)
		assert.Equal(t, fixture.trashedSnippet, got)
	})

	tests := []struct {
		name    string
		missing func(fixture trashFixture) domain.SnippetID
	}{
		{name: "reports a live Snippet as not found", missing: func(fixture trashFixture) domain.SnippetID {
			return fixture.liveSnippet.ID()
		}},
		{
			name: "reports a Snippet inside a trashed Folder as not found",
			missing: func(fixture trashFixture) domain.SnippetID {
				return fixture.nestedSnippet.ID()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fixture := newTrashFixture(t)

			_, err := fixture.trash(t).Find(t.Context(), tt.missing(fixture))

			require.ErrorIs(t, err, domain.ErrNotFound)
			assert.ErrorContains(t, err, "sqlite.TrashRepository.Find")
		})
	}
}

func TestTrashRepository_RestoreSnippet(t *testing.T) {
	t.Parallel()

	t.Run("puts the Snippet back into its live Folder", func(t *testing.T) {
		t.Parallel()

		fixture := newTrashFixture(t)

		got, err := fixture.trash(t).RestoreSnippet(t.Context(), fixture.trashedSnippet.ID())

		require.NoError(t, err)
		assert.Equal(
			t,
			domain.Restoration{FolderID: fixture.liveFolder.ID(), Fallback: domain.RestoreFallbackNone},
			got,
		)
		assert.Equal(t, fixture.trashedSnippet, fixture.find(t, fixture.trashedSnippet.ID()))
		assert.Empty(t, fixture.library(t).Trash.Snippets)
	})

	t.Run("puts the Snippet at the Root when its Folder is in the Trash", func(t *testing.T) {
		t.Parallel()

		fixture := newTrashFixture(t)
		require.NoError(t, fixture.folders(t).Trash(t.Context(), fixture.liveFolder.ID(), trashedAt()))

		got, err := fixture.trash(t).RestoreSnippet(t.Context(), fixture.trashedSnippet.ID())

		require.NoError(t, err)
		assert.Equal(
			t,
			domain.Restoration{FolderID: domain.FolderID{}, Fallback: domain.RestoreFallbackFolderInTrash},
			got,
		)
		assert.True(t, fixture.find(t, fixture.trashedSnippet.ID()).AtRoot())
	})

	t.Run("puts a detached Snippet at the Root and clears its former path", func(t *testing.T) {
		t.Parallel()

		fixture := newTrashFixture(t)
		detachRawSnippet(t, fixture, fixture.trashedSnippet.ID())

		got, err := fixture.trash(t).RestoreSnippet(t.Context(), fixture.trashedSnippet.ID())

		require.NoError(t, err)
		assert.Equal(
			t,
			domain.Restoration{FolderID: domain.FolderID{}, Fallback: domain.RestoreFallbackFolderDeleted},
			got,
		)
		assert.True(t, fixture.find(t, fixture.trashedSnippet.ID()).AtRoot())
		assert.Empty(t, fixture.library(t).Trash.Snippets)
	})

	tests := []struct {
		name    string
		missing func(fixture trashFixture) domain.SnippetID
	}{
		{name: "reports a live Snippet as not found", missing: func(fixture trashFixture) domain.SnippetID {
			return fixture.liveSnippet.ID()
		}},
		{
			name: "reports a Snippet inside a trashed Folder as not found",
			missing: func(fixture trashFixture) domain.SnippetID {
				return fixture.nestedSnippet.ID()
			},
		},
		{name: "reports a missing Snippet as not found", missing: func(trashFixture) domain.SnippetID {
			return testkit.NewSequentialIDs().NewSnippetID()
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fixture := newTrashFixture(t)

			_, err := fixture.trash(t).RestoreSnippet(t.Context(), tt.missing(fixture))

			require.ErrorIs(t, err, domain.ErrNotFound)
			assert.ErrorContains(t, err, "sqlite.TrashRepository.RestoreSnippet")
		})
	}
}

func TestTrashRepository_RestoreFolder(t *testing.T) {
	t.Parallel()

	t.Run("puts the Folder back under its live parent with its whole subtree", func(t *testing.T) {
		t.Parallel()

		fixture := newTrashFixture(t)

		got, err := fixture.trash(t).RestoreFolder(t.Context(), fixture.trashedFolder.ID())

		require.NoError(t, err)
		assert.Equal(
			t,
			domain.Restoration{FolderID: fixture.liveFolder.ID(), Fallback: domain.RestoreFallbackNone},
			got,
		)
		folders, err := fixture.folders(t).List(t.Context())
		require.NoError(t, err)
		assert.ElementsMatch(
			t, []domain.Folder{fixture.liveFolder, fixture.trashedFolder, fixture.nestedFolder}, folders,
		)
		assert.Equal(t, fixture.nestedSnippet, fixture.find(t, fixture.nestedSnippet.ID()))
	})

	t.Run("leaves a separate Trash item inside the Folder in the Trash", func(t *testing.T) {
		t.Parallel()

		fixture := newTrashFixture(t)
		require.NoError(t, fixture.folders(t).Trash(t.Context(), fixture.liveFolder.ID(), trashedAt()))

		_, err := fixture.trash(t).RestoreFolder(t.Context(), fixture.liveFolder.ID())

		require.NoError(t, err)
		assert.Contains(t, fixture.library(t).Trash.Folders, fixture.trashedFolder.ID())
		assert.Contains(t, fixture.library(t).Trash.Snippets, fixture.trashedSnippet.ID())
	})

	t.Run("puts the Folder at the Root when its parent is in the Trash", func(t *testing.T) {
		t.Parallel()

		fixture := newTrashFixture(t)
		require.NoError(t, fixture.folders(t).Trash(t.Context(), fixture.liveFolder.ID(), trashedAt()))

		got, err := fixture.trash(t).RestoreFolder(t.Context(), fixture.trashedFolder.ID())

		require.NoError(t, err)
		assert.Equal(
			t,
			domain.Restoration{FolderID: domain.FolderID{}, Fallback: domain.RestoreFallbackFolderInTrash},
			got,
		)
		restored, err := fixture.folders(t).Find(t.Context(), fixture.trashedFolder.ID())
		require.NoError(t, err)
		assert.True(t, restored.AtRoot())
	})

	t.Run("reports a live Folder as not found", func(t *testing.T) {
		t.Parallel()

		fixture := newTrashFixture(t)

		_, err := fixture.trash(t).RestoreFolder(t.Context(), fixture.liveFolder.ID())

		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.ErrorContains(t, err, "sqlite.TrashRepository.RestoreFolder")
	})
}

func TestTrashRepository_DeleteSnippet(t *testing.T) {
	t.Parallel()

	t.Run("deletes the Snippet for good and keeps its Tags", func(t *testing.T) {
		t.Parallel()

		fixture := newTrashFixture(t)

		err := fixture.trash(t).DeleteSnippet(t.Context(), fixture.trashedSnippet.ID())

		require.NoError(t, err)
		library := fixture.library(t)
		assert.ElementsMatch(t, []domain.Snippet{fixture.liveSnippet, fixture.nestedSnippet}, library.Snippets)
		assert.Empty(t, library.Trash.Snippets)
		assert.Equal(t, []domain.Tag{fixture.tag}, library.Tags)
	})

	tests := []struct {
		name    string
		missing func(fixture trashFixture) domain.SnippetID
	}{
		{name: "reports a live Snippet as not found", missing: func(fixture trashFixture) domain.SnippetID {
			return fixture.liveSnippet.ID()
		}},
		{
			name: "reports a Snippet inside a trashed Folder as not found",
			missing: func(fixture trashFixture) domain.SnippetID {
				return fixture.nestedSnippet.ID()
			},
		},
		{name: "reports a missing Snippet as not found", missing: func(trashFixture) domain.SnippetID {
			return testkit.NewSequentialIDs().NewSnippetID()
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fixture := newTrashFixture(t)

			err := fixture.trash(t).DeleteSnippet(t.Context(), tt.missing(fixture))

			require.ErrorIs(t, err, domain.ErrNotFound)
			require.ErrorContains(t, err, "sqlite.TrashRepository.DeleteSnippet")
			assert.Len(t, fixture.library(t).Snippets, 3)
		})
	}
}

func TestTrashRepository_DeleteFolder(t *testing.T) {
	t.Parallel()

	t.Run("deletes the Folder with its whole subtree and keeps the Tags", func(t *testing.T) {
		t.Parallel()

		fixture := newTrashFixture(t)

		err := fixture.trash(t).DeleteFolder(t.Context(), fixture.trashedFolder.ID())

		require.NoError(t, err)
		library := fixture.library(t)
		assert.Equal(t, []domain.Folder{fixture.liveFolder}, library.Folders)
		assert.ElementsMatch(t, []domain.Snippet{fixture.liveSnippet, fixture.trashedSnippet}, library.Snippets)
		assert.Equal(t, []domain.Tag{fixture.tag}, library.Tags)
	})

	t.Run("detaches a separate Trash Snippet inside it with the path it was in", func(t *testing.T) {
		t.Parallel()

		fixture := newTrashFixture(t)
		trashRawSnippet(t, fixture.path, fixture.nestedSnippet.ID())

		err := fixture.trash(t).DeleteFolder(t.Context(), fixture.trashedFolder.ID())

		require.NoError(t, err)
		library := fixture.library(t)
		assert.Equal(
			t,
			[]string{"live", "trashed", "nested"},
			library.Trash.Snippets[fixture.nestedSnippet.ID()].FormerPath().Strings(),
		)
		assert.True(t, snippetIn(t, library, fixture.nestedSnippet.ID()).AtRoot())
	})

	t.Run("detaches a separate Trash Folder inside it, which keeps its own subtree", func(t *testing.T) {
		t.Parallel()

		fixture := newTrashFixture(t)
		trashRawFolder(t, fixture.path, fixture.nestedFolder.ID())

		err := fixture.trash(t).DeleteFolder(t.Context(), fixture.trashedFolder.ID())

		require.NoError(t, err)
		library := fixture.library(t)
		assert.Equal(
			t, []string{"live", "trashed"}, library.Trash.Folders[fixture.nestedFolder.ID()].FormerPath().Strings(),
		)
		assert.True(t, folderIn(t, library, fixture.nestedFolder.ID()).AtRoot())
		assert.Equal(t, fixture.nestedSnippet, snippetIn(t, library, fixture.nestedSnippet.ID()))
	})

	t.Run("passes a detached Folder's former path on to the items it detaches", func(t *testing.T) {
		t.Parallel()

		fixture := newTrashFixture(t)
		detachRawFolder(t, fixture, fixture.trashedFolder.ID())
		trashRawSnippet(t, fixture.path, fixture.nestedSnippet.ID())

		err := fixture.trash(t).DeleteFolder(t.Context(), fixture.trashedFolder.ID())

		require.NoError(t, err)
		assert.Equal(
			t,
			[]string{"gone", "trashed", "nested"},
			fixture.library(t).Trash.Snippets[fixture.nestedSnippet.ID()].FormerPath().Strings(),
		)
	})

	tests := []struct {
		name    string
		missing func(fixture trashFixture) domain.FolderID
	}{
		{name: "reports a live Folder as not found", missing: func(fixture trashFixture) domain.FolderID {
			return fixture.liveFolder.ID()
		}},
		{
			name: "reports a Folder inside a trashed Folder as not found",
			missing: func(fixture trashFixture) domain.FolderID {
				return fixture.nestedFolder.ID()
			},
		},
		{name: "reports a missing Folder as not found", missing: func(trashFixture) domain.FolderID {
			return testkit.NewSequentialIDs().NewFolderID()
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fixture := newTrashFixture(t)

			err := fixture.trash(t).DeleteFolder(t.Context(), tt.missing(fixture))

			require.ErrorIs(t, err, domain.ErrNotFound)
			require.ErrorContains(t, err, "sqlite.TrashRepository.DeleteFolder")
			assert.Len(t, fixture.library(t).Folders, 3)
		})
	}
}

func TestTrashRepository_Empty(t *testing.T) {
	t.Parallel()

	t.Run("deletes the given Trash items and counts every Folder and Snippet removed", func(t *testing.T) {
		t.Parallel()

		fixture := newTrashFixture(t)

		got, err := fixture.trash(t).Empty(t.Context(), fixture.trashItems())

		require.NoError(t, err)
		assert.Equal(t, domain.TrashEmptied{Folders: 2, Snippets: 2}, got)
		library := fixture.library(t)
		assert.Equal(t, []domain.Folder{fixture.liveFolder}, library.Folders)
		assert.Equal(t, []domain.Snippet{fixture.liveSnippet}, library.Snippets)
		assert.Equal(t, []domain.Tag{fixture.tag}, library.Tags)
	})

	t.Run("skips a given item that left the Trash", func(t *testing.T) {
		t.Parallel()

		fixture := newTrashFixture(t)
		_, err := fixture.trash(t).RestoreSnippet(t.Context(), fixture.trashedSnippet.ID())
		require.NoError(t, err)

		got, err := fixture.trash(t).Empty(t.Context(), fixture.trashItems())

		require.NoError(t, err)
		assert.Equal(t, domain.TrashEmptied{Folders: 2, Snippets: 1}, got)
		assert.ElementsMatch(
			t,
			[]domain.Snippet{fixture.liveSnippet, fixture.trashedSnippet},
			fixture.library(t).Snippets,
		)
	})

	t.Run("leaves a Trash item it wasn't given in the Trash", func(t *testing.T) {
		t.Parallel()

		fixture := newTrashFixture(t)
		trashRawSnippet(t, fixture.path, fixture.liveSnippet.ID())

		_, err := fixture.trash(t).Empty(t.Context(), fixture.trashItems())

		require.NoError(t, err)
		assert.Contains(t, fixture.library(t).Trash.Snippets, fixture.liveSnippet.ID())
	})

	t.Run("detaches a Trash item it wasn't given from a Folder it deletes", func(t *testing.T) {
		t.Parallel()

		fixture := newTrashFixture(t)
		trashRawSnippet(t, fixture.path, fixture.nestedSnippet.ID())

		_, err := fixture.trash(t).Empty(t.Context(), fixture.trashItems())

		require.NoError(t, err)
		library := fixture.library(t)
		assert.Equal(t, []domain.Folder{fixture.liveFolder}, library.Folders)
		assert.Equal(
			t,
			[]string{"live", "trashed", "nested"},
			library.Trash.Snippets[fixture.nestedSnippet.ID()].FormerPath().Strings(),
		)
	})
}

func (f trashFixture) trashItems() domain.TrashItemIDs {
	return domain.TrashItemIDs{
		Folders:  []domain.FolderID{f.trashedFolder.ID()},
		Snippets: []domain.SnippetID{f.trashedSnippet.ID()},
	}
}

func snippetIn(t *testing.T, library domain.Library, id domain.SnippetID) domain.Snippet {
	t.Helper()

	index := slices.IndexFunc(library.Snippets, func(snippet domain.Snippet) bool { return snippet.ID() == id })
	require.NotEqual(t, -1, index, "snippet %s", id)

	return library.Snippets[index]
}

func folderIn(t *testing.T, library domain.Library, id domain.FolderID) domain.Folder {
	t.Helper()

	index := slices.IndexFunc(library.Folders, func(folder domain.Folder) bool { return folder.ID() == id })
	require.NotEqual(t, -1, index, "folder %s", id)

	return library.Folders[index]
}

func detachRawFolder(t *testing.T, fixture trashFixture, id domain.FolderID) {
	t.Helper()

	execRaw(
		t,
		fixture.path,
		`UPDATE folders SET parent_id = NULL, former_parent_path = '["gone"]' WHERE id = ?`,
		id.String(),
	)
}

func (f trashFixture) trash(t *testing.T) *sqlite.TrashRepository {
	t.Helper()

	return newTrashRepository(t, f.database)
}

func (f trashFixture) find(t *testing.T, id domain.SnippetID) domain.Snippet {
	t.Helper()

	found, err := f.snippets(t).Find(t.Context(), id)
	require.NoError(t, err)

	return found
}

func newTrashRepository(t *testing.T, database *sqlite.Database) *sqlite.TrashRepository {
	t.Helper()

	return sqlite.NewTrashRepository(database, slog.New(slog.DiscardHandler))
}

func detachRawSnippet(t *testing.T, fixture trashFixture, id domain.SnippetID) {
	t.Helper()

	execRaw(
		t,
		fixture.path,
		`UPDATE snippets SET folder_id = NULL, former_folder_path = '["live"]' WHERE id = ?`,
		id.String(),
	)
}
