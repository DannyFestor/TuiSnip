package state_test

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/state"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

const (
	stateFileName     = "state.toml"
	filePermissions   = 0o644
	concurrentWriters = 8

	dockerID = "0194c3a0-f01d-7000-8000-000000000001"
	goID     = "0194c3a0-f01d-7000-8000-000000000002"
)

func TestLoad(t *testing.T) {
	t.Parallel()

	t.Run("reads the remembered sort order", func(t *testing.T) {
		t.Parallel()

		path := writeState(t, "[snippet_list]\nsort = \"updated\"\n")

		file, logs := load(t, path)

		assert.Equal(t, domain.SortOrderUpdated, file.SortOrder())
		assert.Empty(t, logs.String())
	})

	t.Run("ignores unknown keys", func(t *testing.T) {
		t.Parallel()

		path := writeState(t, "future = 1\n[snippet_list]\nsort = \"created\"\nwrap = true\n")

		file, logs := load(t, path)

		assert.Equal(t, domain.SortOrderCreated, file.SortOrder())
		assert.Empty(t, logs.String())
	})

	t.Run("defaults a missing sort key to title order", func(t *testing.T) {
		t.Parallel()

		path := writeState(t, "[snippet_list]\n")

		file, _ := load(t, path)

		assert.Equal(t, domain.SortOrderTitle, file.SortOrder())
	})

	fallbacks := []struct {
		name string
		path func(t *testing.T) string
	}{
		{
			name: "falls back to title order with a Warn when the file is missing",
			path: func(t *testing.T) string {
				t.Helper()

				return filepath.Join(t.TempDir(), stateFileName)
			},
		},
		{
			name: "falls back to title order with a Warn when the file is not TOML",
			path: func(t *testing.T) string {
				t.Helper()

				return writeState(t, "[snippet_list\nsort = ")
			},
		},
		{
			name: "falls back to title order with a Warn when the sort order is unknown",
			path: func(t *testing.T) string {
				t.Helper()

				return writeState(t, "[snippet_list]\nsort = \"language\"\n")
			},
		},
		{
			name: "falls back to title order with a Warn when the file is unreadable",
			path: func(t *testing.T) string {
				t.Helper()

				path := filepath.Join(t.TempDir(), stateFileName)
				require.NoError(t, os.Mkdir(path, 0o750))

				return path
			},
		},
	}
	for _, tt := range fallbacks {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := tt.path(t)

			file, logs := load(t, path)

			assert.Equal(t, domain.SortOrderTitle, file.SortOrder())
			assert.Contains(t, logs.String(), "level=WARN")
			assert.Contains(t, logs.String(), "path="+path)
		})
	}
}

func TestFile_SaveSortOrder(t *testing.T) {
	t.Parallel()

	t.Run("remembers the sort order for the next start", func(t *testing.T) {
		t.Parallel()

		path := writeState(t, "[snippet_list]\nsort = \"title\"\n")
		file, _ := load(t, path)

		require.NoError(t, file.SaveSortOrder(t.Context(), domain.SortOrderCreated))

		reloaded, logs := load(t, path)
		assert.Equal(t, domain.SortOrderCreated, reloaded.SortOrder())
		assert.Empty(t, logs.String())
	})

	t.Run("reports the saved sort order", func(t *testing.T) {
		t.Parallel()

		file, _ := load(t, filepath.Join(t.TempDir(), stateFileName))

		require.NoError(t, file.SaveSortOrder(t.Context(), domain.SortOrderUpdated))

		assert.Equal(t, domain.SortOrderUpdated, file.SortOrder())
	})

	t.Run("creates the file and its directory when missing", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), "tuisnip", stateFileName)
		file, _ := load(t, path)

		require.NoError(t, file.SaveSortOrder(t.Context(), domain.SortOrderUpdated))

		info, err := os.Stat(path)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(filePermissions), info.Mode().Perm())
	})

	t.Run("replaces a broken file", func(t *testing.T) {
		t.Parallel()

		path := writeState(t, "not toml at all [")
		file, _ := load(t, path)

		require.NoError(t, file.SaveSortOrder(t.Context(), domain.SortOrderCreated))

		reloaded, logs := load(t, path)
		assert.Equal(t, domain.SortOrderCreated, reloaded.SortOrder())
		assert.Empty(t, logs.String())
	})

	t.Run("leaves only the state file in its directory", func(t *testing.T) {
		t.Parallel()

		path := writeState(t, "[snippet_list]\nsort = \"title\"\n")
		file, _ := load(t, path)

		require.NoError(t, file.SaveSortOrder(t.Context(), domain.SortOrderUpdated))
		require.NoError(t, file.SaveSortOrder(t.Context(), domain.SortOrderCreated))

		entries, err := os.ReadDir(filepath.Dir(path))
		require.NoError(t, err)
		require.Len(t, entries, 1)
		assert.Equal(t, stateFileName, entries[0].Name())
	})

	t.Run("fails when the directory can't be created", func(t *testing.T) {
		t.Parallel()

		blocker := filepath.Join(t.TempDir(), "tuisnip")
		require.NoError(t, os.WriteFile(blocker, nil, filePermissions))
		file, _ := load(t, filepath.Join(blocker, stateFileName))

		err := file.SaveSortOrder(t.Context(), domain.SortOrderCreated)

		require.ErrorContains(t, err, "state.File.SaveSortOrder: ")
	})

	t.Run("survives concurrent saves with a valid file", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), stateFileName)
		file, _ := load(t, path)
		orders := []domain.SortOrder{domain.SortOrderUpdated, domain.SortOrderCreated}

		var group sync.WaitGroup
		for writer := range concurrentWriters {
			group.Go(func() {
				assert.NoError(t, file.SaveSortOrder(t.Context(), orders[writer%len(orders)]))
			})
		}

		group.Wait()

		reloaded, logs := load(t, path)
		assert.Contains(t, orders, reloaded.SortOrder())
		assert.Empty(t, logs.String())
	})
}

func TestFile_CollapsedFolders(t *testing.T) {
	t.Parallel()

	t.Run("reads the remembered collapsed Folders", func(t *testing.T) {
		t.Parallel()

		path := writeState(t, "[folders]\ncollapsed = [\""+dockerID+"\", \""+goID+"\"]\n")

		file, logs := load(t, path)

		assert.Equal(t, []domain.FolderID{folderID(t, dockerID), folderID(t, goID)}, file.CollapsedFolders())
		assert.Empty(t, logs.String())
	})

	t.Run("collapses no Folder when the key is missing", func(t *testing.T) {
		t.Parallel()

		file, _ := load(t, writeState(t, "[snippet_list]\nsort = \"created\"\n"))

		assert.Empty(t, file.CollapsedFolders())
	})

	t.Run("falls back to defaults with a Warn when an id is not a UUID", func(t *testing.T) {
		t.Parallel()

		path := writeState(t, "[snippet_list]\nsort = \"created\"\n[folders]\ncollapsed = [\"docker\"]\n")

		file, logs := load(t, path)

		assert.Empty(t, file.CollapsedFolders())
		assert.Equal(t, domain.SortOrderTitle, file.SortOrder())
		assert.Contains(t, logs.String(), "level=WARN")
	})
}

func TestFile_SaveCollapsedFolders(t *testing.T) {
	t.Parallel()

	t.Run("remembers the collapsed Folders for the next start, keeping the sort order", func(t *testing.T) {
		t.Parallel()

		path := writeState(t, "[snippet_list]\nsort = \"created\"\n")
		file, _ := load(t, path)
		collapsed := []domain.FolderID{folderID(t, goID)}

		require.NoError(t, file.SaveCollapsedFolders(t.Context(), collapsed))

		reloaded, logs := load(t, path)
		assert.Equal(t, collapsed, reloaded.CollapsedFolders())
		assert.Equal(t, domain.SortOrderCreated, reloaded.SortOrder())
		assert.Empty(t, logs.String())
	})

	t.Run("forgets every collapsed Folder when none is left", func(t *testing.T) {
		t.Parallel()

		path := writeState(t, "[folders]\ncollapsed = [\""+goID+"\"]\n")
		file, _ := load(t, path)

		require.NoError(t, file.SaveCollapsedFolders(t.Context(), nil))

		reloaded, _ := load(t, path)
		assert.Empty(t, reloaded.CollapsedFolders())
	})

	t.Run("keeps the collapsed Folders when the sort order is saved", func(t *testing.T) {
		t.Parallel()

		path := writeState(t, "[folders]\ncollapsed = [\""+goID+"\"]\n")
		file, _ := load(t, path)

		require.NoError(t, file.SaveSortOrder(t.Context(), domain.SortOrderUpdated))

		reloaded, _ := load(t, path)
		assert.Equal(t, []domain.FolderID{folderID(t, goID)}, reloaded.CollapsedFolders())
	})

	t.Run("fails and keeps the old Folders when the directory can't be created", func(t *testing.T) {
		t.Parallel()

		blocker := filepath.Join(t.TempDir(), "tuisnip")
		require.NoError(t, os.WriteFile(blocker, nil, filePermissions))
		file, _ := load(t, filepath.Join(blocker, stateFileName))

		err := file.SaveCollapsedFolders(t.Context(), []domain.FolderID{folderID(t, goID)})

		require.ErrorContains(t, err, "state.File.SaveCollapsedFolders: ")
		assert.Empty(t, file.CollapsedFolders())
	})
}

func folderID(t *testing.T, raw string) domain.FolderID {
	t.Helper()

	parsed, err := uuid.Parse(raw)
	require.NoError(t, err)

	return domain.FolderID(parsed)
}

func load(t *testing.T, path string) (*state.File, *bytes.Buffer) {
	t.Helper()

	var logs bytes.Buffer

	logger := slog.New(slog.NewTextHandler(&logs, nil))

	return state.Load(t.Context(), state.Options{Path: path, Logger: logger}), &logs
}

func writeState(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), stateFileName)
	require.NoError(t, os.WriteFile(path, []byte(content), filePermissions))

	return path
}
