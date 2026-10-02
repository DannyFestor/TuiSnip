package testapp

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/xdg"
	"github.com/DannyFestor/TuiSnip/internal/bootstrap"
)

const (
	platformGOOS      = "darwin"
	platformTool      = "pbcopy"
	toolSearchPath    = "PATH=/usr/bin:/bin"
	recordedStdin     = ".stdin"
	scriptPermissions = 0o700
	filePermissions   = 0o600
	dirPermissions    = 0o700
)

type Home struct {
	Paths    xdg.Paths
	toolsDir string
}

func NewHome(t *testing.T) *Home {
	t.Helper()

	home := t.TempDir()

	paths, err := xdg.Resolve(func(key string) string {
		if key == "HOME" {
			return home
		}

		return ""
	})
	require.NoError(t, err)

	return &Home{Paths: paths, toolsDir: t.TempDir()}
}

func Start(t *testing.T, tool ClipboardTool) (*Home, *bootstrap.App) {
	t.Helper()

	home := NewHome(t)

	app := home.Start(t, tool)

	return home, app
}

func (h *Home) Start(t *testing.T, tool ClipboardTool) *bootstrap.App {
	t.Helper()

	app, err := h.Open(t, tool)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, app.Close()) })

	return app
}

func (h *Home) Open(t *testing.T, tool ClipboardTool) (*bootstrap.App, error) {
	t.Helper()

	app, err := bootstrap.New(t.Context(), bootstrap.Options{
		Paths:    h.Paths,
		Environ:  func() []string { return []string{toolSearchPath} },
		LookPath: h.installed(t, tool),
		GOOS:     platformGOOS,
	})
	if err != nil {
		return nil, fmt.Errorf("testapp.Home.Open: %w", err)
	}

	return app, nil
}

func (h *Home) WriteConfig(t *testing.T, toml string) {
	t.Helper()

	h.writeFile(t, h.Paths.ConfigFile, toml)
}

func (h *Home) Block(t *testing.T, dir string) {
	t.Helper()

	h.writeFile(t, dir, "")
}

func (h *Home) Copied(t *testing.T) string {
	t.Helper()

	data, err := os.ReadFile(h.ToolRecording())
	require.NoError(t, err)

	return string(data)
}

func (h *Home) ToolRecording() string {
	return filepath.Join(h.toolsDir, platformTool+recordedStdin)
}

func (h *Home) Log(t *testing.T) string {
	t.Helper()

	data, err := os.ReadFile(h.Paths.LogFile)
	require.NoError(t, err)

	return string(data)
}

func (h *Home) installed(t *testing.T, tool ClipboardTool) func(string) (string, error) {
	t.Helper()

	if tool == NoTool {
		return func(string) (string, error) { return "", exec.ErrNotFound }
	}

	path := filepath.Join(h.toolsDir, platformTool)
	require.NoError(t, os.WriteFile(path, tool.script(), scriptPermissions))

	return func(name string) (string, error) {
		if name != platformTool {
			return "", exec.ErrNotFound
		}

		return path, nil
	}
}

func (*Home) writeFile(t *testing.T, path, content string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Dir(path), dirPermissions))
	require.NoError(t, os.WriteFile(path, []byte(content), filePermissions))
}
