package atomicfile_test

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/atomicfile"
)

func TestReplace(t *testing.T) {
	t.Parallel()

	t.Run("replaces the file's content", func(t *testing.T) {
		t.Parallel()

		path := existing(t, "old")

		require.NoError(t, atomicfile.Replace(path, []byte("new")))

		assert.Equal(t, "new", read(t, path))
	})

	t.Run("creates the file and its directory with readable permissions", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), "tuisnip", fileName)

		require.NoError(t, atomicfile.Replace(path, []byte("new")))

		info, err := os.Stat(path)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(filePermissions), info.Mode().Perm())
	})

	t.Run("leaves only the file in its directory", func(t *testing.T) {
		t.Parallel()

		path := existing(t, "old")

		require.NoError(t, atomicfile.Replace(path, []byte("new")))

		assertOnlyFileIn(t, path)
	})

	t.Run("fails when the directory can't be created", func(t *testing.T) {
		t.Parallel()

		blocker := existing(t, "")

		err := atomicfile.Replace(filepath.Join(blocker, fileName), []byte("new"))

		require.ErrorContains(t, err, "atomicfile.Replace: ")
	})

	t.Run("leaves one complete file after concurrent writes", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), fileName)
		contents := []string{"first", "second"}

		var writes sync.WaitGroup
		for writer := range concurrentWrites {
			writes.Go(func() {
				assert.NoError(t, atomicfile.Replace(path, []byte(contents[writer%len(contents)])))
			})
		}

		writes.Wait()

		assert.Contains(t, contents, read(t, path))
		assertOnlyFileIn(t, path)
	})
}
