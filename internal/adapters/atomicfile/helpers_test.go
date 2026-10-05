package atomicfile_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	fileName         = "state.toml"
	filePermissions  = 0o644
	concurrentWrites = 8
)

func existing(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), fileName)
	require.NoError(t, os.WriteFile(path, []byte(content), filePermissions))

	return path
}

func read(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path) //nolint:gosec // G304: a path under t.TempDir
	require.NoError(t, err)

	return string(data)
}

func assertOnlyFileIn(t *testing.T, path string) {
	t.Helper()

	entries, err := os.ReadDir(filepath.Dir(path))
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, filepath.Base(path), entries[0].Name())
}
