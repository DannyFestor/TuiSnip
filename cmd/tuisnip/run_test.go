package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type outcome struct {
	code   int
	stdout string
	stderr string
}

func TestRun(t *testing.T) {
	t.Parallel()

	t.Run("prints the paths and creates nothing", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()

		got := runWith(t, home, "--paths")

		assert.Equal(t, exitOK, got.code)
		assert.Equal(t, ""+
			"config   "+home+"/.config/tuisnip/config.toml\n"+
			"data     "+home+"/.local/share/tuisnip/tuisnip.db\n"+
			"backup   "+home+"/.local/share/tuisnip/backups\n"+
			"state    "+home+"/.local/state/tuisnip/state.toml\n"+
			"log      "+home+"/.local/state/tuisnip/tuisnip.log\n",
			got.stdout)
		assert.Empty(t, entries(t, home))
	})

	t.Run("prints the version", func(t *testing.T) {
		t.Parallel()

		got := runWith(t, t.TempDir(), "--version")

		assert.Equal(t, exitOK, got.code)
		assert.Regexp(t, `^tuisnip \S+\n$`, got.stdout)
	})

	t.Run("prints usage for help", func(t *testing.T) {
		t.Parallel()

		got := runWith(t, t.TempDir(), "-h")

		assert.Equal(t, exitOK, got.code)
		assert.Contains(t, got.stderr, "-paths")
	})

	t.Run("refuses an unknown flag", func(t *testing.T) {
		t.Parallel()

		got := runWith(t, t.TempDir(), "--sync")

		assert.Equal(t, exitUsage, got.code)
		assert.Contains(t, got.stderr, "flag provided but not defined: -sync")
	})

	t.Run("refuses positional arguments", func(t *testing.T) {
		t.Parallel()

		got := runWith(t, t.TempDir(), "get", "curl")

		assert.Equal(t, exitUsage, got.code)
		assert.Contains(t, got.stderr, "tuisnip: unexpected arguments: [get curl]")
	})

	t.Run("fails without a HOME", func(t *testing.T) {
		t.Parallel()

		got := runWith(t, "", "--paths")

		assert.Equal(t, exitFailure, got.code)
		assert.Contains(t, got.stderr, "tuisnip: xdg: HOME is not set")
	})

	t.Run("refuses to start with an invalid config and names the key", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		configFile := home + "/.config/tuisnip/config.toml"
		require.NoError(t, os.MkdirAll(filepath.Dir(configFile), 0o700))
		require.NoError(t, os.WriteFile(configFile, []byte("theme = \"purple\"\n"), 0o600))

		got := runWith(t, home)

		assert.Equal(t, exitFailure, got.code)
		assert.True(t, strings.HasPrefix(got.stderr, "tuisnip: invalid config "+configFile+":\n"), got.stderr)
		assert.Contains(t, got.stderr, "theme")
	})
}

func runWith(t *testing.T, home string, args ...string) outcome {
	t.Helper()

	var stdout, stderr bytes.Buffer

	code := run(t.Context(), args, environment{
		stdout:   &stdout,
		stderr:   &stderr,
		getenv:   homeOnly(home),
		environ:  func() []string { return nil },
		lookPath: func(string) (string, error) { return "", exec.ErrNotFound },
		goos:     "darwin",
	})

	return outcome{code: code, stdout: stdout.String(), stderr: stderr.String()}
}

func homeOnly(home string) func(string) string {
	return func(key string) string {
		if key == "HOME" {
			return home
		}

		return ""
	}
}

func entries(t *testing.T, dir string) []os.DirEntry {
	t.Helper()

	found, err := os.ReadDir(dir)
	require.NoError(t, err)

	return found
}
