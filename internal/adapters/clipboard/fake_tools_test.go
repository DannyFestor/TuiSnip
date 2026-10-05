package clipboard_test

import (
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/clipboard"
)

const (
	recordingScript = "testdata/recording-tool.sh"
	failingScript   = "testdata/failing-tool.sh"
	hangingScript   = "testdata/hanging-tool.sh"

	searchPath = "PATH=/usr/bin:/bin"
	linuxGOOS  = "linux"
	darwinGOOS = "darwin"

	unicodeText = "Grüße 日本語 한국어\n"
)

type fakeTools struct {
	dir   string
	paths map[string]string
}

func newFakeTools(t *testing.T) *fakeTools {
	t.Helper()

	return &fakeTools{dir: t.TempDir(), paths: map[string]string{}}
}

func (f *fakeTools) install(t *testing.T, name, script string) {
	t.Helper()

	target, err := filepath.Abs(script)
	require.NoError(t, err)

	path := filepath.Join(f.dir, name)
	require.NoError(t, os.Symlink(target, path))
	f.paths[name] = path
}

func (f *fakeTools) lookPath(name string) (string, error) {
	path, found := f.paths[name]
	if !found {
		return "", exec.ErrNotFound
	}

	return path, nil
}

func (f *fakeTools) recorded(t *testing.T, name, suffix string) string {
	t.Helper()

	data, err := os.ReadFile(f.paths[name] + suffix) //nolint:gosec // G304: a path under t.TempDir
	if os.IsNotExist(err) {
		return ""
	}

	require.NoError(t, err)

	return string(data)
}

func (f *fakeTools) received(t *testing.T, name string) string {
	t.Helper()

	return f.recorded(t, name, ".stdin")
}

func (f *fakeTools) options(goos string, env ...string) clipboard.Options {
	return clipboard.Options{
		Environ:  func() []string { return append([]string{searchPath}, env...) },
		LookPath: f.lookPath,
		GOOS:     goos,
		Logger:   slog.New(slog.DiscardHandler),
	}
}
