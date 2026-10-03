package bootstrap_test

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/config"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/bootstrap"
)

func TestSettingsFrom(t *testing.T) {
	t.Parallel()

	t.Run("gives each Binding its keys as Bubble Tea writes them", func(t *testing.T) {
		t.Parallel()

		cfg := loadConfig(t, "[bindings.snippet_list]\ncopy = [\"shift+ctrl+y\", \"Y\"]\n")

		keys := bootstrap.SettingsFrom(cfg, time.UTC).Keys

		assert.Equal(t, []string{"ctrl+shift+y", "Y"}, keys[binding.ScopeSnippetList][binding.Copy])
	})

	t.Run("keeps every Scope and Binding of the config", func(t *testing.T) {
		t.Parallel()

		cfg := loadConfig(t, "")

		keys := bootstrap.SettingsFrom(cfg, time.UTC).Keys

		for scope, bindings := range cfg.Bindings {
			assert.Len(t, keys[binding.Scope(scope.String())], len(bindings), "%s", scope)
		}
	})

	t.Run("keeps an unbound Binding with no keys", func(t *testing.T) {
		t.Parallel()

		cfg := loadConfig(t, "[bindings.confirm]\nno = []\n")

		no, ok := bootstrap.SettingsFrom(cfg, time.UTC).Keys[binding.ScopeConfirm][binding.No]

		assert.True(t, ok)
		assert.Empty(t, no)
	})

	t.Run("passes the forced quit key through", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "ctrl+c", bootstrap.SettingsFrom(loadConfig(t, ""), time.UTC).ForcedQuitKey)
	})

	t.Run("passes the Location through", func(t *testing.T) {
		t.Parallel()

		tokyo := time.FixedZone("JST", 9*60*60)

		assert.Same(t, tokyo, bootstrap.SettingsFrom(loadConfig(t, ""), tokyo).Location)
	})
}

func loadConfig(t *testing.T, contents string) config.Config {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.toml")
	if contents != "" {
		require.NoError(t, os.WriteFile(path, []byte(contents), 0o600))
	}

	cfg, err := config.Load(t.Context(), config.Options{Path: path, Logger: slog.New(slog.DiscardHandler)})
	require.NoError(t, err)

	return cfg
}
