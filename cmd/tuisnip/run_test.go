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

	t.Run("prints usage with the flags and commands for help", func(t *testing.T) {
		t.Parallel()

		got := runWith(t, t.TempDir(), "-h")

		assert.Equal(t, exitOK, got.code)
		assert.Contains(t, got.stderr, "-paths")
		assert.Contains(t, got.stderr, "export <file>")
	})

	t.Run("refuses an unknown flag", func(t *testing.T) {
		t.Parallel()

		got := runWith(t, t.TempDir(), "--sync")

		assert.Equal(t, exitUsage, got.code)
		assert.Contains(t, got.stderr, "flag provided but not defined: -sync")
	})

	t.Run("refuses an unknown command and prints usage", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()

		got := runWith(t, home, "get", "curl")

		assert.Equal(t, exitUsage, got.code)
		assert.Contains(t, got.stderr, "tuisnip: unknown command: get\n")
		assert.Contains(t, got.stderr, "export <file>")
		assert.Empty(t, entries(t, home))
	})

	t.Run("exports the Library and reports the path and counts", func(t *testing.T) {
		t.Parallel()

		target := filepath.Join(t.TempDir(), "backup.json")

		got := runWith(t, t.TempDir(), "export", target)

		assert.Equal(t, exitOK, got.code, got.stderr)
		assert.Equal(t, "Exported 0 Folders, 0 Tags, 0 Snippets to "+target+"\n", got.stdout)
		assert.Contains(t, read(t, target), `"format": "tuisnip-export"`)
	})

	refusals := []struct {
		name       string
		args       []string
		wantStderr []string
	}{
		{
			name:       "refuses an export without a file",
			args:       []string{"export"},
			wantStderr: []string{"tuisnip: export: wrong number of arguments\n", "usage: tuisnip export <file>\n"},
		},
		{
			name:       "refuses an export to two files",
			args:       []string{"export", "one.json", "two.json"},
			wantStderr: []string{"tuisnip: export: wrong number of arguments\n", "usage: tuisnip export <file>\n"},
		},
		{
			name:       "refuses - as the file to export to",
			args:       []string{"export", "-"},
			wantStderr: []string{"tuisnip: export: - is not supported\n", "usage: tuisnip export <file>\n"},
		},
		{
			name:       "refuses --paths together with a command",
			args:       []string{"--paths", "export", "backup.json"},
			wantStderr: []string{"tuisnip: -version and -paths take no command\n", "export <file>"},
		},
		{
			name:       "refuses --version together with a command",
			args:       []string{"--version", "export", "backup.json"},
			wantStderr: []string{"tuisnip: -version and -paths take no command\n", "export <file>"},
		},
	}
	for _, tt := range refusals {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			home := t.TempDir()

			got := runWith(t, home, tt.args...)

			assert.Equal(t, exitUsage, got.code)
			assert.Empty(t, got.stdout)
			assert.Empty(t, entries(t, home))

			for _, want := range tt.wantStderr {
				assert.Contains(t, got.stderr, want)
			}
		})
	}

	t.Run("leaves an existing file untouched when the export fails", func(t *testing.T) {
		t.Parallel()

		if os.Geteuid() == 0 {
			t.Skip("root writes into a read-only directory")
		}

		dir := t.TempDir()
		target := filepath.Join(dir, "backup.json")
		require.NoError(t, os.WriteFile(target, []byte("good backup"), 0o600))
		require.NoError(t, os.Chmod(dir, 0o500)) //nolint:gosec // G302: a directory needs its execute bit
		t.Cleanup(func() {
			require.NoError(t, os.Chmod(dir, 0o700)) //nolint:gosec // G302: a directory needs its execute bit
		})

		got := runWith(t, t.TempDir(), "export", target)

		assert.Equal(t, exitFailure, got.code)
		assert.True(t, strings.HasPrefix(got.stderr, "tuisnip: export: "), got.stderr)
		assert.Empty(t, got.stdout)
		assert.Equal(t, "good backup", read(t, target))
		assert.Len(t, entries(t, dir), 1)
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

func read(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path) //nolint:gosec // G304: a path under t.TempDir
	require.NoError(t, err)

	return string(data)
}

func entries(t *testing.T, dir string) []os.DirEntry {
	t.Helper()

	found, err := os.ReadDir(dir)
	require.NoError(t, err)

	return found
}
