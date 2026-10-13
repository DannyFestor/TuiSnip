package atomicfile_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/atomicfile"
)

func TestReplaceOwnerOnly(t *testing.T) {
	t.Parallel()

	t.Run("replaces the file's content, readable by its owner only", func(t *testing.T) {
		t.Parallel()

		path := existing(t, "old")

		require.NoError(t, atomicfile.ReplaceOwnerOnly(path, []byte("new")))

		assert.Equal(t, "new", read(t, path))
		info, err := os.Stat(path)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
		assertOnlyFileIn(t, path)
	})

	t.Run("fails without creating a missing directory", func(t *testing.T) {
		t.Parallel()

		missing := filepath.Join(t.TempDir(), "missing")

		err := atomicfile.ReplaceOwnerOnly(filepath.Join(missing, fileName), []byte("new"))

		require.ErrorContains(t, err, "atomicfile.ReplaceOwnerOnly: ")
		assert.NoDirExists(t, missing)
	})

	t.Run("leaves the target and no temporary file behind when the rename fails", func(t *testing.T) {
		t.Parallel()

		target := filepath.Join(t.TempDir(), fileName)
		require.NoError(t, os.Mkdir(target, 0o700))
		kept := filepath.Join(target, "kept")
		require.NoError(t, os.WriteFile(kept, []byte("kept"), filePermissions))

		err := atomicfile.ReplaceOwnerOnly(target, []byte("new"))

		require.ErrorContains(t, err, "atomicfile.ReplaceOwnerOnly: ")
		assert.Equal(t, "kept", read(t, kept))
		assertOnlyFileIn(t, target)
	})
}
