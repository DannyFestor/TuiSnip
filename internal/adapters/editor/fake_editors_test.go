package editor_test

import (
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/editor"
)

const (
	signingScript           = "testdata/signing-editor.sh"
	appendingScript         = "testdata/appending-editor.sh"
	failingScript           = "testdata/failing-editor.sh"
	permissionCheckerScript = "testdata/permission-checking-editor.sh"

	recordedPathFile = "edited-path"
)

type fakeEditors struct {
	dir      string
	recorder string
	paths    map[string]string
}

func newFakeEditors(t *testing.T) *fakeEditors {
	t.Helper()

	recorder := filepath.Join(t.TempDir(), recordedPathFile)

	return &fakeEditors{dir: t.TempDir(), recorder: recorder, paths: map[string]string{}}
}

func (f *fakeEditors) install(t *testing.T, name, script string) string {
	t.Helper()

	target, err := filepath.Abs(script)
	require.NoError(t, err)

	path := filepath.Join(f.dir, name)
	require.NoError(t, os.Symlink(target, path))
	f.paths[name] = path

	return path
}

func (f *fakeEditors) launcher(configured string, env ...string) *editor.Launcher {
	environ := append([]string{"PATH=" + f.dir + ":/usr/bin:/bin", "RECORD_PATH=" + f.recorder}, env...)

	return editor.New(editor.Options{
		Configured: configured,
		Environ:    func() []string { return environ },
		LookPath:   f.lookPath,
		Logger:     slog.New(slog.DiscardHandler),
	})
}

func (f *fakeEditors) lookPath(name string) (string, error) {
	path, found := f.paths[name]
	if !found {
		return "", exec.ErrNotFound
	}

	return path, nil
}

func (f *fakeEditors) editedPath(t *testing.T) string {
	t.Helper()

	data, err := os.ReadFile(f.recorder)
	require.NoError(t, err)

	return string(data)
}
