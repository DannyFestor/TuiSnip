package sqlite_test

import (
	"bytes"
	"database/sql"
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/db/migrations"
	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
)

const (
	databaseFileName = "tuisnip.db"
	rawDriverName    = "sqlite"
)

func newDatabasePath(t *testing.T) string {
	t.Helper()

	return filepath.Join(t.TempDir(), databaseFileName)
}

func openDatabase(t *testing.T, path string) *sqlite.Database {
	t.Helper()

	database, err := sqlite.Open(t.Context(), testOptions(path))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, database.Close()) })

	return database
}

func openAndClose(t *testing.T, path string) error {
	t.Helper()

	database, err := sqlite.Open(t.Context(), testOptions(path))
	if err != nil {
		return err
	}

	return database.Close()
}

func testOptions(path string) sqlite.Options {
	return sqlite.Options{
		Path:   path,
		Logger: slog.New(slog.DiscardHandler),
		Clock:  testkit.NewFixedClock(time.Date(2026, time.October, 1, 12, 0, 0, 0, time.UTC)),
	}
}

func execRaw(t *testing.T, path, statement string, args ...any) {
	t.Helper()

	db, err := sql.Open(rawDriverName, path)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	_, err = db.ExecContext(t.Context(), statement, args...)
	require.NoError(t, err)
}

func insertRawFolder(t *testing.T, path string, folderID domain.FolderID) {
	t.Helper()

	execRaw(
		t,
		path,
		"INSERT INTO folders (id, name, default_language, created_at, updated_at) VALUES (?, 'scripts', 'Bash', 1, 1)",
		folderID.String(),
	)
}

func migrateRawToPreviousSchema(t *testing.T, path string) {
	t.Helper()

	db, err := sql.Open(rawDriverName, path)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	migrator, err := goose.NewProvider(goose.DialectSQLite3, db, migrations.FS(), goose.WithLogger(goose.NopLogger()))
	require.NoError(t, err)

	_, err = migrator.UpTo(t.Context(), knownSchemaVersion-1)
	require.NoError(t, err)
}

func holdStartupLock(t *testing.T, databasePath string) {
	t.Helper()

	file, err := os.OpenFile(filepath.Clean(databasePath+".lock"), os.O_CREATE|os.O_RDWR, 0o600)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, file.Close()) })

	require.NoError(t, syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB))
}

func plantBackupDirFiles(t *testing.T, databasePath string, names ...string) {
	t.Helper()

	dir := sqlite.BackupDir(databasePath)
	require.NoError(t, os.MkdirAll(dir, 0o700))

	for _, name := range names {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), nil, 0o600))
	}
}

func backupNames(t *testing.T, path string) []string {
	t.Helper()

	entries, err := os.ReadDir(sqlite.BackupDir(path))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}

	require.NoError(t, err)

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}

	return names
}

func newSnippetRepository(t *testing.T, database *sqlite.Database) *sqlite.SnippetRepository {
	t.Helper()

	return sqlite.NewSnippetRepository(database, slog.New(slog.DiscardHandler))
}

func newLoggingSnippetRepository(t *testing.T, path string, logged *bytes.Buffer) *sqlite.SnippetRepository {
	t.Helper()

	return sqlite.NewSnippetRepository(openDatabase(t, path), slog.New(slog.NewJSONHandler(logged, nil)))
}

func newTagRepository(t *testing.T, database *sqlite.Database) *sqlite.TagRepository {
	t.Helper()

	return sqlite.NewTagRepository(database, slog.New(slog.DiscardHandler))
}

func newLoggingTagRepository(t *testing.T, path string, logged *bytes.Buffer) *sqlite.TagRepository {
	t.Helper()

	return sqlite.NewTagRepository(openDatabase(t, path), slog.New(slog.NewJSONHandler(logged, nil)))
}

func insertTag(t *testing.T, repository *sqlite.TagRepository, ids *testkit.SequentialIDs, name string) domain.Tag {
	t.Helper()

	tag := testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: name})
	require.NoError(t, repository.Insert(t.Context(), tag))

	return tag
}

func insertSnippet(
	t *testing.T,
	repository *sqlite.SnippetRepository,
	ids *testkit.SequentialIDs,
	spec testkit.SnippetSpec,
) domain.Snippet {
	t.Helper()

	spec.ID = ids.NewSnippetID()
	spec.Fragment.ID = ids.NewFragmentID()
	snippet := testkit.Snippet(t, spec)
	require.NoError(t, repository.Insert(t.Context(), snippet))

	return snippet
}

func editedSnippet(t *testing.T, stored domain.Snippet, content string) domain.Snippet {
	t.Helper()

	title, err := value.NewTitle("edited")
	require.NoError(t, err)

	edited, err := stored.Edit(title, stored.Description(), mustContent(t, content), stored.UpdatedAt().Add(time.Hour))
	require.NoError(t, err)

	return edited
}

func mustContent(t *testing.T, raw string) value.Content {
	t.Helper()

	content, err := value.NewContent(raw)
	require.NoError(t, err)

	return content
}

func insertCorruptSnippet(
	t *testing.T,
	repository *sqlite.SnippetRepository,
	path string,
	ids *testkit.SequentialIDs,
) domain.Snippet {
	t.Helper()

	snippet := insertSnippet(t, repository, ids, testkit.SnippetSpec{Title: "corrupt"})
	execRaw(t, path, "UPDATE fragments SET language = 'Klingon' WHERE snippet_id = ?", snippet.ID().String())

	return snippet
}
