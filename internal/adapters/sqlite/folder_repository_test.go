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
