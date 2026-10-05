package config_test

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/config"
)

const (
	configFileName     = "config.toml"
	defaultPermissions = 0o644
)

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

func keyNames(cfg config.Config, scope config.Scope, binding config.Binding) []string {
	keys := cfg.Bindings[scope][binding]

	names := make([]string, 0, len(keys))
	for _, key := range keys {
		names = append(names, key.String())
	}

	return names
}
