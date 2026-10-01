package config_test

import (
	"bytes"
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	defaultconfig "github.com/DannyFestor/TuiSnip/embeds/config"
	"github.com/DannyFestor/TuiSnip/internal/adapters/config"
)

const (
	configFileName     = "config.toml"
	defaultPermissions = 0o644
	concurrentStarts   = 8
	maxFileBytes       = 1 << 20
	oversizedBytes     = maxFileBytes + 1
)

func TestLoad(t *testing.T) {
	t.Parallel()

	t.Run("writes the default file when missing", func(t *testing.T) {
		t.Parallel()
		path := filepath.Join(t.TempDir(), "tuisnip", configFileName)

		var logs bytes.Buffer

		_, err := config.Load(t.Context(), testOptions(path, &logs))

		require.NoError(t, err)
		assert.Contains(t, logs.String(), "default config written")

		written, err := os.ReadFile(path) //nolint:gosec // G304: a path under t.TempDir
		require.NoError(t, err)
		assert.Equal(t, defaultconfig.Default(), written)

		info, err := os.Stat(path)
		require.NoError(t, err)
		assert.Equal(t, fs.FileMode(defaultPermissions), info.Mode().Perm())
	})

	t.Run("leaves an existing file untouched", func(t *testing.T) {
		t.Parallel()
		path := writeConfig(t, "theme = \"dark\"\n")

		var logs bytes.Buffer

		_, err := config.Load(t.Context(), testOptions(path, &logs))

		require.NoError(t, err)
		assert.NotContains(t, logs.String(), "default config written")

		written, err := os.ReadFile(path) //nolint:gosec // G304: a path under t.TempDir
		require.NoError(t, err)
		assert.Equal(t, "theme = \"dark\"\n", string(written))
	})

	t.Run("validates the embedded defaults with no warnings", func(t *testing.T) {
		t.Parallel()

		var logs bytes.Buffer

		got, err := config.Load(t.Context(), testOptions(filepath.Join(t.TempDir(), configFileName), &logs))

		require.NoError(t, err)
		assert.NotContains(t, logs.String(), "level=WARN")
		assert.Equal(t, config.ThemeAuto, got.Theme)
		assert.Equal(
			t,
			config.Copy{Clipboard: config.ClipboardBackendAuto, TrimTrailingNewline: true, QuitAfter: false},
			got.Copy,
		)
		assert.True(t, got.Mouse)
		assert.Empty(t, got.Editor)
		assert.Empty(t, got.Languages)
		assert.Equal(t, []string{"q"}, keyNames(got, config.ScopeGlobal, config.BindingQuit))
	})

	t.Run("matches the default file to every Scope and Binding", func(t *testing.T) {
		t.Parallel()

		got := loadValid(t, "")

		seen := map[string]bool{}

		for _, scopeName := range config.ScopeNames() {
			scope, err := config.ParseScope(scopeName)
			require.NoError(t, err)
			require.Contains(t, got.Bindings, scope)

			for binding := range got.Bindings[scope] {
				seen[binding.String()] = true
			}
		}

		assert.Len(t, got.Bindings, len(config.ScopeNames()))
		assert.ElementsMatch(t, config.BindingNames(), keysOf(seen))
	})

	t.Run("layers the user's values over the defaults", func(t *testing.T) {
		t.Parallel()

		got := loadValid(t, `
editor = "code --wait"
theme = "dark"
languages = ["Go", "Bash"]
mouse = false

[copy]
clipboard = "osc52"

[bindings.global]
search = ["S", "ctrl+f"]
zoom = []
`)

		assert.Equal(t, "code --wait", got.Editor)
		assert.Equal(t, config.ThemeDark, got.Theme)
		assert.Equal(t, []string{"Go", "Bash"}, languageNames(got))
		assert.False(t, got.Mouse)
		assert.Equal(
			t,
			config.Copy{Clipboard: config.ClipboardBackendOsc52, TrimTrailingNewline: true, QuitAfter: false},
			got.Copy,
		)
		assert.Equal(t, []string{"S", "ctrl+f"}, keyNames(got, config.ScopeGlobal, config.BindingSearch))
		assert.Empty(t, got.Bindings[config.ScopeGlobal][config.BindingZoom])
		assert.Equal(t, []string{"q"}, keyNames(got, config.ScopeGlobal, config.BindingQuit))
		assert.Equal(t, []string{"y"}, keyNames(got, config.ScopeSnippetList, config.BindingCopy))
	})

	t.Run("warns about unknown keys outside the Bindings and ignores them", func(t *testing.T) {
		t.Parallel()

		var logs bytes.Buffer

		_, err := config.Load(
			t.Context(),
			testOptions(writeConfig(t, "colour = \"red\"\n[copy]\nformat = \"html\"\n"), &logs),
		)

		require.NoError(t, err)
		assert.Contains(t, logs.String(), `config_key=colour`)
		assert.Contains(t, logs.String(), `config_key=copy.format`)
		assert.Equal(t, 2, strings.Count(logs.String(), "level=WARN"))
	})

	t.Run("reports every invalid value by key path", func(t *testing.T) {
		t.Parallel()
		path := writeConfig(t, `
theme = "blue"
languages = ["Go", "golang"]

[copy]
clipboard = "x11"

[bindings.sidebar]
quit = ["q"]

[bindings.global]
serch = ["/"]
`)

		_, err := config.Load(t.Context(), testOptions(path, &bytes.Buffer{}))

		invalid, ok := errors.AsType[config.InvalidError](err)
		require.True(t, ok, "want InvalidError, got %v", err)
		assert.Equal(t, path, invalid.Path)
		assert.Equal(t, "invalid config "+path+":\n"+
			`  theme: "blue" is not one of auto, light, dark`+"\n"+
			`  languages[1]: "golang" is not a Language`+"\n"+
			`  copy.clipboard: "x11" is not one of auto, native, osc52`+"\n"+
			"  bindings.global.serch: not a Binding in global; use one of "+globalBindingNames+"\n"+
			"  bindings.sidebar: not a Scope; use one of "+strings.Join(config.ScopeNames(), ", "),
			err.Error())
	})

	t.Run("stops at a decode error with its line", func(t *testing.T) {
		t.Parallel()
		path := writeConfig(t, "mouse = true\ntheme = 3\n")

		_, err := config.Load(t.Context(), testOptions(path, &bytes.Buffer{}))

		invalid, ok := errors.AsType[config.InvalidError](err)
		require.True(t, ok, "want InvalidError, got %v", err)
		assert.Equal(t, path, invalid.Path)
		assert.Contains(t, err.Error(), "line 2")
	})

	t.Run("rejects a bare string as a Binding's keys", func(t *testing.T) {
		t.Parallel()

		path := writeConfig(t, "[bindings.global]\nquit = \"q\"\n")

		_, err := config.Load(t.Context(), testOptions(path, &bytes.Buffer{}))

		invalid, ok := errors.AsType[config.InvalidError](err)
		require.True(t, ok, "want InvalidError, got %v", err)
		assert.Equal(t, path, invalid.Path)
	})

	t.Run("refuses a file over the size limit", func(t *testing.T) {
		t.Parallel()

		_, err := config.Load(
			t.Context(),
			testOptions(writeConfig(t, strings.Repeat("#", oversizedBytes)), &bytes.Buffer{}),
		)

		require.ErrorIs(t, err, config.ErrFileTooLarge)
	})

	t.Run("accepts a file at the size limit", func(t *testing.T) {
		t.Parallel()

		_, err := config.Load(
			t.Context(),
			testOptions(writeConfig(t, strings.Repeat("#", maxFileBytes)), &bytes.Buffer{}),
		)

		require.NoError(t, err)
	})

	t.Run("reports a directory at the config path", func(t *testing.T) {
		t.Parallel()
		path := filepath.Join(t.TempDir(), configFileName)
		require.NoError(t, os.Mkdir(path, 0o750))

		_, err := config.Load(t.Context(), testOptions(path, &bytes.Buffer{}))

		require.Error(t, err)
		assert.Contains(t, err.Error(), path)
	})

	t.Run("reports a broken symlink instead of replacing it", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		path := filepath.Join(dir, configFileName)
		require.NoError(t, os.Symlink(filepath.Join(dir, "missing.toml"), path))

		_, err := config.Load(t.Context(), testOptions(path, &bytes.Buffer{}))

		require.ErrorIs(t, err, fs.ErrNotExist)
		target, err := os.Readlink(path)
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(dir, "missing.toml"), target)
	})

	t.Run("follows a symlink to the user's file", func(t *testing.T) {
		t.Parallel()
		path := filepath.Join(t.TempDir(), configFileName)
		require.NoError(t, os.Symlink(writeConfig(t, "theme = \"light\"\n"), path))

		got, err := config.Load(t.Context(), testOptions(path, &bytes.Buffer{}))

		require.NoError(t, err)
		assert.Equal(t, config.ThemeLight, got.Theme)
	})

	t.Run("gives every concurrent first start the complete default file", func(t *testing.T) {
		t.Parallel()
		dir := filepath.Join(t.TempDir(), "tuisnip")
		path := filepath.Join(dir, configFileName)
		errs := make([]error, concurrentStarts)

		var starts sync.WaitGroup
		for i := range concurrentStarts {
			starts.Go(func() {
				_, errs[i] = config.Load(t.Context(), testOptions(path, &bytes.Buffer{}))
			})
		}

		starts.Wait()

		for _, err := range errs {
			require.NoError(t, err)
		}

		entries, err := os.ReadDir(dir)
		require.NoError(t, err)
		assert.Len(t, entries, 1, "temporary files left behind")
	})
}

func FuzzLoad(f *testing.F) {
	f.Add(defaultconfig.Default())
	f.Add([]byte(""))
	f.Add([]byte("theme = 3\n"))
	f.Add([]byte("[bindings.global]\nquit = \"q\"\n"))
	f.Add([]byte("\xef\xbb\xbftheme = \"dark\"\n[bindings.nope.x]\ny = 1\n"))

	f.Fuzz(func(t *testing.T, contents []byte) {
		path := filepath.Join(t.TempDir(), configFileName)
		require.NoError(t, os.WriteFile(path, contents, defaultPermissions))

		got, err := config.Load(t.Context(), testOptions(path, &bytes.Buffer{}))
		if err != nil {
			return
		}

		assert.Len(t, got.Bindings, len(config.ScopeNames()))
	})
}

const globalBindingNames = "back, bottom, capture, down, focus_folders, focus_left, focus_list, focus_next, " +
	"focus_prev, focus_right, focus_snippet, focus_tags, help, new_snippet, open, page_down, page_up, " +
	"quit, search, top, up, zoom"

func testOptions(path string, logs *bytes.Buffer) config.Options {
	return config.Options{
		Path:   path,
		Logger: slog.New(slog.NewTextHandler(logs, &slog.HandlerOptions{Level: slog.LevelInfo})),
	}
}

func writeConfig(t *testing.T, contents string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), configFileName)
	require.NoError(t, os.WriteFile(path, []byte(contents), defaultPermissions))

	return path
}

func loadValid(t *testing.T, contents string) config.Config {
	t.Helper()

	got, err := config.Load(t.Context(), testOptions(writeConfig(t, contents), &bytes.Buffer{}))
	require.NoError(t, err)

	return got
}

func languageNames(cfg config.Config) []string {
	names := make([]string, 0, len(cfg.Languages))
	for _, language := range cfg.Languages {
		names = append(names, language.String())
	}

	return names
}

func keyNames(cfg config.Config, scope config.Scope, binding config.Binding) []string {
	keys := cfg.Bindings[scope][binding]

	names := make([]string, 0, len(keys))
	for _, key := range keys {
		names = append(names, key.String())
	}

	return names
}

func keysOf(set map[string]bool) []string {
	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}

	return keys
}
