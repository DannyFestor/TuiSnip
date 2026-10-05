package atomicfile_test

import (
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/atomicfile"
)

func TestCreateIfMissing(t *testing.T) {
	t.Parallel()

	t.Run("creates a missing file and its directory", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), "tuisnip", fileName)

		created, err := atomicfile.CreateIfMissing(path, []byte("default"))

		require.NoError(t, err)
		assert.True(t, created)
		assert.Equal(t, "default", read(t, path))
		assertOnlyFileIn(t, path)
	})

	t.Run("leaves an existing file untouched", func(t *testing.T) {
		t.Parallel()

		path := existing(t, "edited")

		created, err := atomicfile.CreateIfMissing(path, []byte("default"))

		require.NoError(t, err)
		assert.False(t, created)
		assert.Equal(t, "edited", read(t, path))
	})

	t.Run("creates the file once for concurrent callers", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), fileName)
		created := make([]bool, concurrentWrites)

		var writes sync.WaitGroup
		for writer := range concurrentWrites {
			writes.Go(func() {
				var err error

				created[writer], err = atomicfile.CreateIfMissing(path, []byte("default"))
				assert.NoError(t, err)
			})
		}

		writes.Wait()

		assert.Equal(t, 1, countTrue(created))
		assert.Equal(t, "default", read(t, path))
		assertOnlyFileIn(t, path)
	})

	t.Run("fails when the directory can't be created", func(t *testing.T) {
		t.Parallel()

		blocker := existing(t, "")

		_, err := atomicfile.CreateIfMissing(filepath.Join(blocker, fileName), []byte("default"))

		require.ErrorContains(t, err, "atomicfile.CreateIfMissing: ")
	})
}

func countTrue(values []bool) int {
	count := 0

	for _, value := range values {
		if value {
			count++
		}
	}

	return count
}
