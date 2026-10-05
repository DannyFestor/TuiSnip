package sqlite_test

import (
	"bytes"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
)

const rawTagID = "0190a000-0000-7000-8000-0000000000aa"

type subtreeFixture struct {
	path       string
	folders    *sqlite.FolderRepository
	snippets   *sqlite.SnippetRepository
	snippetIDs *testkit.SequentialIDs
	goID       domain.FolderID
	testingID  domain.FolderID
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
}

func TestFolderRepository_Insert(t *testing.T) {
	t.Parallel()

	t.Run("refuses a Folder under a parent that does not exist", func(t *testing.T) {
		t.Parallel()

		repository := newFolderRepository(t, openDatabase(t, newDatabasePath(t)), slog.New(slog.DiscardHandler))
		ids := testkit.NewSequentialIDs()
		orphan := testkit.Folder(t, testkit.FolderSpec{ID: ids.NewFolderID(), ParentID: ids.NewFolderID()})

		require.Error(t, repository.Insert(t.Context(), orphan))
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

func TestFolderRepository_CountSubfolders(t *testing.T) {
	t.Parallel()

	t.Run("counts the subfolders at every depth", func(t *testing.T) {
		t.Parallel()

		fixture := newSubtreeFixture(t)

		got, err := fixture.folders.CountSubfolders(t.Context(), fixture.goID)

		require.NoError(t, err)
		assert.Equal(t, 2, got)
	})

	t.Run("counts no subfolders under a Folder without children", func(t *testing.T) {
		t.Parallel()

		fixture := newSubtreeFixture(t)

		got, err := fixture.folders.CountSubfolders(t.Context(), fixture.dockerID)

		require.NoError(t, err)
		assert.Zero(t, got)
	})
}

func TestFolderRepository_CountSnippetsInSubtree(t *testing.T) {
	t.Parallel()

	t.Run("counts the Snippets in the Folder and every subfolder", func(t *testing.T) {
		t.Parallel()

		fixture := newSubtreeFixture(t)

		got, err := fixture.folders.CountSnippetsInSubtree(t.Context(), fixture.goID)

		require.NoError(t, err)
		assert.Equal(t, 4, got)
	})

	t.Run("counts only the subtree below a nested Folder", func(t *testing.T) {
		t.Parallel()

		fixture := newSubtreeFixture(t)

		got, err := fixture.folders.CountSnippetsInSubtree(t.Context(), fixture.testingID)

		require.NoError(t, err)
		assert.Equal(t, 3, got)
	})
}

func TestFolderRepository_Delete(t *testing.T) {
	t.Parallel()

	t.Run("removes every subfolder and Snippet in the subtree and nothing else", func(t *testing.T) {
		t.Parallel()

		fixture := newSubtreeFixture(t)

		require.NoError(t, fixture.folders.Delete(t.Context(), fixture.goID))

		folders, err := fixture.folders.List(t.Context())
		require.NoError(t, err)
		require.Len(t, folders, 1)
		assert.Equal(t, fixture.dockerID, folders[0].ID())

		snippets, err := fixture.snippets.List(t.Context())
		require.NoError(t, err)
		assert.ElementsMatch(t, fixture.kept, snippets)
	})

	t.Run("deletes Snippets that carry Tags", func(t *testing.T) {
		t.Parallel()

		fixture := newSubtreeFixture(t)
		tagged := insertSnippet(
			t,
			fixture.snippets,
			fixture.snippetIDs,
			testkit.SnippetSpec{FolderID: fixture.testingID},
		)
		insertRawTag(t, fixture.path, tagged.ID())

		require.NoError(t, fixture.folders.Delete(t.Context(), fixture.goID))

		_, err := fixture.snippets.Find(t.Context(), tagged.ID())
		assert.ErrorIs(t, err, domain.ErrNotFound)
	})

	t.Run("reports a missing Folder as not found", func(t *testing.T) {
		t.Parallel()

		repository := newFolderRepository(t, openDatabase(t, newDatabasePath(t)), slog.New(slog.DiscardHandler))

		err := repository.Delete(t.Context(), testkit.NewSequentialIDs().NewFolderID())

		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.ErrorContains(t, err, "sqlite.FolderRepository.Delete")
	})
}

func folderName(t *testing.T, raw string) value.FolderName {
	t.Helper()

	name, err := value.NewFolderName(raw)
	require.NoError(t, err)

	return name
}

func newFolderRepository(t *testing.T, database *sqlite.Database, logger *slog.Logger) *sqlite.FolderRepository {
	t.Helper()

	return sqlite.NewFolderRepository(database, logger)
}

func insertFolder(t *testing.T, repository *sqlite.FolderRepository, spec testkit.FolderSpec) domain.Folder {
	t.Helper()

	folder := testkit.Folder(t, spec)
	require.NoError(t, repository.Insert(t.Context(), folder))

	return folder
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
		folders:    folders,
		snippets:   snippets,
		snippetIDs: snippetIDs,
		goID:       goFolder.ID(),
		testingID:  testingFolder.ID(),
		dockerID:   dockerFolder.ID(),
		kept:       kept,
	}
}

func insertRawTag(t *testing.T, path string, snippetID domain.SnippetID) {
	t.Helper()

	execRaw(
		t,
		path,
		"INSERT INTO tags (id, name, name_key, created_at, updated_at) VALUES (?, 'Go', 'go', 1, 1)",
		rawTagID,
	)
	execRaw(t, path, "INSERT INTO snippet_tag (snippet_id, tag_id) VALUES (?, ?)", snippetID.String(), rawTagID)
}
