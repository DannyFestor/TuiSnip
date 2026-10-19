package sqlite_test

import (
	"bytes"
	"log/slog"
	"maps"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
)

type subtreeFixture struct {
	path       string
	database   *sqlite.Database
	folders    *sqlite.FolderRepository
	snippets   *sqlite.SnippetRepository
	snippetIDs *testkit.SequentialIDs
	goID       domain.FolderID
	testingID  domain.FolderID
	mocksID    domain.FolderID
	dockerID   domain.FolderID
	kept       []domain.Snippet
}

func TestFolderRepository_List(t *testing.T) {
	t.Parallel()

	t.Run("returns every inserted Folder", func(t *testing.T) {
		t.Parallel()

		repository := newFolderRepository(t, openDatabase(t, newDatabasePath(t)), slog.New(slog.DiscardHandler))
		ids := testkit.NewSequentialIDs()
		parent := insertFolder(t, repository, testkit.FolderSpec{ID: ids.NewFolderID(), Name: "go"})
		child := insertFolder(t, repository, testkit.FolderSpec{
			ID:              ids.NewFolderID(),
			Name:            "testing",
			ParentID:        parent.ID(),
			DefaultLanguage: "Go",
		})

		got, err := repository.List(t.Context())

		require.NoError(t, err)
		assert.ElementsMatch(t, []domain.Folder{parent, child}, got)
	})

	t.Run("returns no Folders from an empty database", func(t *testing.T) {
		t.Parallel()

		repository := newFolderRepository(t, openDatabase(t, newDatabasePath(t)), slog.New(slog.DiscardHandler))

		got, err := repository.List(t.Context())

		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("skips a Folder whose Default Language is unknown", func(t *testing.T) {
		t.Parallel()

		var logged bytes.Buffer

		path := newDatabasePath(t)
		repository := newFolderRepository(t, openDatabase(t, path), slog.New(slog.NewJSONHandler(&logged, nil)))
		ids := testkit.NewSequentialIDs()
		healthy := insertFolder(t, repository, testkit.FolderSpec{ID: ids.NewFolderID(), Name: "healthy"})
		corrupt := insertFolder(t, repository, testkit.FolderSpec{ID: ids.NewFolderID(), Name: "corrupt"})
		execRaw(t, path, "UPDATE folders SET default_language = 'Klingon' WHERE id = ?", corrupt.ID().String())

		got, err := repository.List(t.Context())

		require.NoError(t, err)
		assert.Equal(t, []domain.Folder{healthy}, got)
		assert.Contains(t, logged.String(), `"level":"WARN"`)
		assert.Contains(t, logged.String(), `"folder_id":"`+corrupt.ID().String()+`"`)
		assert.Contains(t, logged.String(), `"error":"`+domain.ErrCorruptRecord.Error())
	})

	t.Run("leaves out a Folder in the Trash and every Folder under it", func(t *testing.T) {
		t.Parallel()

		fixture := newTrashFixture(t)

		got, err := fixture.folders(t).List(t.Context())

		require.NoError(t, err)
		assert.Equal(t, []domain.Folder{fixture.liveFolder}, got)
	})
}

func TestFolderRepository_Insert(t *testing.T) {
	t.Parallel()

	t.Run("refuses a Folder under a parent that does not exist", func(t *testing.T) {
		t.Parallel()

		repository := newFolderRepository(t, openDatabase(t, newDatabasePath(t)), slog.New(slog.DiscardHandler))
		ids := testkit.NewSequentialIDs()
		orphan := testkit.Folder(t, testkit.FolderSpec{ID: ids.NewFolderID(), ParentID: ids.NewFolderID()})

		require.ErrorIs(t, repository.Insert(t.Context(), orphan), domain.ErrFolderGone)
	})

	t.Run("refuses a Folder under a parent in the Trash", func(t *testing.T) {
		t.Parallel()

		fixture := newTrashFixture(t)
		placed := testkit.Folder(
			t,
			testkit.FolderSpec{ID: fixture.ids.NewFolderID(), ParentID: fixture.trashedFolder.ID()},
		)

		err := fixture.folders(t).Insert(t.Context(), placed)

		require.ErrorIs(t, err, domain.ErrFolderGone)
		assert.ErrorContains(t, err, "sqlite.FolderRepository.Insert")
	})

	t.Run("refuses a Folder under a parent inside a Folder in the Trash", func(t *testing.T) {
		t.Parallel()

		fixture := newTrashFixture(t)
		placed := testkit.Folder(
			t,
			testkit.FolderSpec{ID: fixture.ids.NewFolderID(), ParentID: fixture.nestedFolder.ID()},
		)

		require.ErrorIs(t, fixture.folders(t).Insert(t.Context(), placed), domain.ErrFolderGone)
	})
}

func TestFolderRepository_Find(t *testing.T) {
	t.Parallel()

	t.Run("returns the Folder", func(t *testing.T) {
		t.Parallel()

		repository := newFolderRepository(t, openDatabase(t, newDatabasePath(t)), slog.New(slog.DiscardHandler))
		stored := insertFolder(t, repository, testkit.FolderSpec{ID: testkit.NewSequentialIDs().NewFolderID()})

		got, err := repository.Find(t.Context(), stored.ID())

		require.NoError(t, err)
		assert.Equal(t, stored, got)
	})

	t.Run("reports a missing Folder as not found", func(t *testing.T) {
		t.Parallel()

		repository := newFolderRepository(t, openDatabase(t, newDatabasePath(t)), slog.New(slog.DiscardHandler))

		_, err := repository.Find(t.Context(), testkit.NewSequentialIDs().NewFolderID())

		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.ErrorContains(t, err, "sqlite.FolderRepository.Find")
	})

	inTrash := []struct {
		name   string
		folder func(fixture trashFixture) domain.Folder
	}{
		{
			name:   "reports a Folder in the Trash as not found",
			folder: func(fixture trashFixture) domain.Folder { return fixture.trashedFolder },
		},
		{
			name:   "reports a Folder under a Folder in the Trash as not found",
			folder: func(fixture trashFixture) domain.Folder { return fixture.nestedFolder },
		},
	}
	for _, tt := range inTrash {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fixture := newTrashFixture(t)

			_, err := fixture.folders(t).Find(t.Context(), tt.folder(fixture).ID())

			require.ErrorIs(t, err, domain.ErrNotFound)
		})
	}

	t.Run("reports and logs a corrupt Folder", func(t *testing.T) {
		t.Parallel()

		var logged bytes.Buffer

		path := newDatabasePath(t)
		repository := newFolderRepository(t, openDatabase(t, path), slog.New(slog.NewJSONHandler(&logged, nil)))
		corrupt := insertFolder(t, repository, testkit.FolderSpec{ID: testkit.NewSequentialIDs().NewFolderID()})
		execRaw(t, path, "UPDATE folders SET default_language = 'Klingon' WHERE id = ?", corrupt.ID().String())

		_, err := repository.Find(t.Context(), corrupt.ID())

		require.ErrorIs(t, err, domain.ErrCorruptRecord)
		assert.Contains(t, logged.String(), `"level":"ERROR"`)
		assert.Contains(t, logged.String(), `"folder_id":"`+corrupt.ID().String()+`"`)
	})
}

func TestFolderRepository_Update(t *testing.T) {
	t.Parallel()

	t.Run("stores the new name and time", func(t *testing.T) {
		t.Parallel()

		repository := newFolderRepository(t, openDatabase(t, newDatabasePath(t)), slog.New(slog.DiscardHandler))
		stored := insertFolder(
			t,
			repository,
			testkit.FolderSpec{ID: testkit.NewSequentialIDs().NewFolderID(), Name: "go"},
		)
		renamed, err := stored.Rename(folderName(t, "golang"), stored.UpdatedAt().Add(time.Minute))
		require.NoError(t, err)

		require.NoError(t, repository.Update(t.Context(), renamed))

		got, err := repository.Find(t.Context(), stored.ID())
		require.NoError(t, err)
		assert.Equal(t, renamed, got)
	})

	t.Run("reports a missing Folder as not found", func(t *testing.T) {
		t.Parallel()

		repository := newFolderRepository(t, openDatabase(t, newDatabasePath(t)), slog.New(slog.DiscardHandler))
		missing := testkit.Folder(t, testkit.FolderSpec{ID: testkit.NewSequentialIDs().NewFolderID()})

		err := repository.Update(t.Context(), missing)

		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.ErrorContains(t, err, "sqlite.FolderRepository.Update")
	})
}

func TestFolderRepository_Trash(t *testing.T) {
	t.Parallel()

	t.Run("hides the Folder's whole subtree and marks only the Folder", func(t *testing.T) {
		t.Parallel()

		fixture := newSubtreeFixture(t)

		require.NoError(t, fixture.folders.Trash(t.Context(), fixture.goID, trashedAt()))

		folders, err := fixture.folders.List(t.Context())
		require.NoError(t, err)
		require.Len(t, folders, 1)
		assert.Equal(t, fixture.dockerID, folders[0].ID())

		snippets, err := fixture.snippets.List(t.Context())
		require.NoError(t, err)
		assert.ElementsMatch(t, fixture.kept, snippets)

		library := loadedLibrary(t, fixture.database)
		assert.Len(t, library.Folders, 4)
		assert.Len(t, library.Snippets, 6)
		assert.Equal(t, []domain.FolderID{fixture.goID}, slices.Collect(maps.Keys(library.Trash.Folders)))
		assert.WithinDuration(t, trashedAt(), library.Trash.Folders[fixture.goID].DeletedAt(), 0)
		assert.Empty(t, library.Trash.Snippets)
	})

	tests := []struct {
		name    string
		missing func(fixture trashFixture) domain.FolderID
	}{
		{
			name: "reports a Folder already in the Trash as not found",
			missing: func(fixture trashFixture) domain.FolderID {
				return fixture.trashedFolder.ID()
			},
		},
		{
			name: "reports a Folder under one in the Trash as not found",
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

			err := fixture.folders(t).Trash(t.Context(), tt.missing(fixture), trashedAt())

			require.ErrorIs(t, err, domain.ErrNotFound)
			assert.ErrorContains(t, err, "sqlite.FolderRepository.Trash")
		})
	}
}

func TestFolderRepository_ListDescendantIDs(t *testing.T) {
	t.Parallel()

	t.Run("lists the subfolders at every depth", func(t *testing.T) {
		t.Parallel()

		fixture := newSubtreeFixture(t)

		got, err := fixture.folders.ListDescendantIDs(t.Context(), fixture.goID)

		require.NoError(t, err)
		assert.ElementsMatch(t, []domain.FolderID{fixture.testingID, fixture.mocksID}, got)
	})

	t.Run("lists nothing under a Folder without children", func(t *testing.T) {
		t.Parallel()

		fixture := newSubtreeFixture(t)

		got, err := fixture.folders.ListDescendantIDs(t.Context(), fixture.dockerID)

		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

func TestFolderRepository_Move(t *testing.T) {
	t.Parallel()

	t.Run("moves the Folder and its subtree under another Folder", func(t *testing.T) {
		t.Parallel()

		fixture := newSubtreeFixture(t)
		moved := movedUnder(t, fixture.find(t, fixture.testingID), fixture.dockerID)

		require.NoError(t, fixture.folders.Move(t.Context(), moved))

		assert.Equal(t, moved, fixture.find(t, fixture.testingID))
		assert.Equal(t, fixture.testingID, fixture.find(t, fixture.mocksID).ParentID())
	})

	t.Run("refuses a move under a Folder in the Trash", func(t *testing.T) {
		t.Parallel()

		fixture := newTrashFixture(t)
		root := insertFolder(t, fixture.folders(t), testkit.FolderSpec{ID: fixture.ids.NewFolderID(), Name: "root"})

		err := fixture.folders(t).Move(t.Context(), movedUnder(t, root, fixture.trashedFolder.ID()))

		require.ErrorIs(t, err, domain.ErrFolderGone)
		assert.ErrorContains(t, err, "sqlite.FolderRepository.Move")
	})

	t.Run("moves the Folder to the Root", func(t *testing.T) {
		t.Parallel()

		fixture := newSubtreeFixture(t)
		moved := fixture.find(t, fixture.testingID).MoveToRoot()

		require.NoError(t, fixture.folders.Move(t.Context(), moved))

		assert.True(t, fixture.find(t, fixture.testingID).AtRoot())
	})

	t.Run("refuses a move into the Folder's own subtree the caller did not see", func(t *testing.T) {
		t.Parallel()

		fixture := newSubtreeFixture(t)
		moved := movedUnder(t, fixture.find(t, fixture.goID), fixture.mocksID)

		err := fixture.folders.Move(t.Context(), moved)

		require.ErrorIs(t, err, domain.ErrFolderCycle)
		require.ErrorContains(t, err, "sqlite.FolderRepository.Move")
		assert.True(t, fixture.find(t, fixture.goID).AtRoot())
	})

	t.Run("reports a missing Folder as not found", func(t *testing.T) {
		t.Parallel()

		repository := newFolderRepository(t, openDatabase(t, newDatabasePath(t)), slog.New(slog.DiscardHandler))
		missing := testkit.Folder(t, testkit.FolderSpec{ID: testkit.NewSequentialIDs().NewFolderID()})

		err := repository.Move(t.Context(), missing)

		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.ErrorContains(t, err, "sqlite.FolderRepository.Move")
	})
}

func (f subtreeFixture) find(t *testing.T, id domain.FolderID) domain.Folder {
	t.Helper()

	found, err := f.folders.Find(t.Context(), id)
	require.NoError(t, err)

	return found
}

func movedUnder(t *testing.T, folder domain.Folder, parentID domain.FolderID) domain.Folder {
	t.Helper()

	moved, err := folder.MoveUnder(parentID, nil)
	require.NoError(t, err)

	return moved
}

func folderName(t *testing.T, raw string) value.FolderName {
	t.Helper()

	name, err := value.NewFolderName(raw)
	require.NoError(t, err)

	return name
}

func newSubtreeFixture(t *testing.T) subtreeFixture {
	t.Helper()

	path := newDatabasePath(t)
	database := openDatabase(t, path)
	folders := newFolderRepository(t, database, slog.New(slog.DiscardHandler))
	snippets := newSnippetRepository(t, database)
	folderIDs := testkit.NewSequentialIDs()
	goFolder := insertFolder(t, folders, testkit.FolderSpec{ID: folderIDs.NewFolderID(), Name: "go"})
	testingFolder := insertFolder(t, folders, testkit.FolderSpec{
		ID: folderIDs.NewFolderID(), Name: "testing", ParentID: goFolder.ID(),
	})
	mocksFolder := insertFolder(t, folders, testkit.FolderSpec{
		ID: folderIDs.NewFolderID(), Name: "mocks", ParentID: testingFolder.ID(),
	})
	dockerFolder := insertFolder(t, folders, testkit.FolderSpec{ID: folderIDs.NewFolderID(), Name: "docker"})

	snippetIDs := testkit.NewSequentialIDs()
	for _, folderID := range []domain.FolderID{goFolder.ID(), testingFolder.ID(), testingFolder.ID(), mocksFolder.ID()} {
		insertSnippet(t, snippets, snippetIDs, testkit.SnippetSpec{FolderID: folderID})
	}

	kept := []domain.Snippet{
		insertSnippet(t, snippets, snippetIDs, testkit.SnippetSpec{FolderID: dockerFolder.ID()}),
		insertSnippet(t, snippets, snippetIDs, testkit.SnippetSpec{}),
	}

	return subtreeFixture{
		path:       path,
		database:   database,
		folders:    folders,
		snippets:   snippets,
		snippetIDs: snippetIDs,
		goID:       goFolder.ID(),
		testingID:  testingFolder.ID(),
		mocksID:    mocksFolder.ID(),
		dockerID:   dockerFolder.ID(),
		kept:       kept,
	}
}
