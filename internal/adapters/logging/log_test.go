package logging_test

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/logging"
)

const (
	rotationThreshold = 5 << 20
	logFileName       = "tuisnip.log"
)

var sessionPattern = regexp.MustCompile(`session=([0-9a-f-]{36})`)

func TestOpen(t *testing.T) {
	t.Parallel()

	t.Run("creates the state directory and log file owner-only", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), "state", "tuisnip", logFileName)

		openAndLogStart(t, path)

		assert.Equal(t, os.FileMode(0o700), permissions(t, filepath.Dir(path)))
		assert.Equal(t, os.FileMode(0o600), permissions(t, path))
	})

	t.Run("writes text lines with the source and a session", func(t *testing.T) {
		t.Parallel()

		path := logPath(t)

		openAndLogStart(t, path)

		written := readFile(t, path)
		assert.Contains(t, written, "level=INFO")
		assert.Contains(t, written, "source=")
		assert.Contains(t, written, `msg=started`)
		assert.Regexp(t, sessionPattern, written)
	})

	t.Run("gives every start its own session and appends to the log", func(t *testing.T) {
		t.Parallel()

		path := logPath(t)

		openAndLogStart(t, path)
		openAndLogStart(t, path)

		sessions := sessionPattern.FindAllStringSubmatch(readFile(t, path), -1)
		require.Len(t, sessions, 2)
		assert.NotEqual(t, sessions[0][1], sessions[1][1])
	})

	t.Run("leaves out debug lines", func(t *testing.T) {
		t.Parallel()

		path := logPath(t)
		log, err := logging.Open(path)
		require.NoError(t, err)

		log.Logger().DebugContext(t.Context(), "clipboard tool run")
		require.NoError(t, log.Close())

		assert.Empty(t, readFile(t, path))
	})

	t.Run("keeps a log at the threshold", func(t *testing.T) {
		t.Parallel()

		path := logPath(t)
		writeSized(t, path, rotationThreshold)

		openAndLogStart(t, path)

		assert.NoFileExists(t, path+".1")
	})

	t.Run("rotates a log over the threshold", func(t *testing.T) {
		t.Parallel()

		path := logPath(t)
		writeSized(t, path, rotationThreshold+1)

		openAndLogStart(t, path)

		assert.Equal(t, int64(rotationThreshold+1), size(t, path+".1"))
		assert.Contains(t, readFile(t, path), "msg=started")
	})

	t.Run("keeps two rotated logs and drops the oldest", func(t *testing.T) {
		t.Parallel()

		path := logPath(t)
		writeSized(t, path+".2", 2)
		writeSized(t, path+".1", 1)
		writeSized(t, path, rotationThreshold+1)

		openAndLogStart(t, path)

		assert.Equal(t, int64(rotationThreshold+1), size(t, path+".1"))
		assert.Equal(t, int64(1), size(t, path+".2"))
		assert.NoFileExists(t, path+".3")
	})

	t.Run("refuses a log path whose directory is a file", func(t *testing.T) {
		t.Parallel()

		blocker := filepath.Join(t.TempDir(), "state")
		writeSized(t, blocker, 0)

		_, err := logging.Open(filepath.Join(blocker, logFileName))

		require.Error(t, err)
	})
}

func logPath(t *testing.T) string {
	t.Helper()

	return filepath.Join(t.TempDir(), logFileName)
}

func openAndLogStart(t *testing.T, path string) {
	t.Helper()

	log, err := logging.Open(path)
	require.NoError(t, err)

	log.Logger().InfoContext(t.Context(), "started")
	require.NoError(t, log.Close())
}

func writeSized(t *testing.T, path string, bytes int64) {
	t.Helper()

	require.NoError(t, os.WriteFile(path, nil, 0o600))
	require.NoError(t, os.Truncate(path, bytes))
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path) //nolint:gosec // G304: a path under t.TempDir
	require.NoError(t, err)

	return string(data)
}

func size(t *testing.T, path string) int64 {
	t.Helper()

	info, err := os.Stat(path)
	require.NoError(t, err)

	return info.Size()
}

func permissions(t *testing.T, path string) os.FileMode {
	t.Helper()

	info, err := os.Stat(path)
	require.NoError(t, err)

	return info.Mode().Perm()
}
