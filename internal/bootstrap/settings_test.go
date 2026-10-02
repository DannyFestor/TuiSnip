package bootstrap_test

import (
	"log/slog"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/config"
	"github.com/DannyFestor/TuiSnip/internal/bootstrap"
)

var wordBoundary = regexp.MustCompile(`([a-z0-9])([A-Z])`)

func TestSettingsFrom(t *testing.T) {
	t.Parallel()

	t.Run("gives every KeyMap field the keys of the Binding it is named after", func(t *testing.T) {
		t.Parallel()

		cfg := defaultConfig(t)

		settings := reflect.ValueOf(bootstrap.SettingsFrom(cfg, time.UTC))

		for keyMap := range settings.Type().Fields() {
			if keyMap.Type.Kind() != reflect.Struct {
				continue
			}

			scope, err := config.ParseScope(snakeCase(keyMap.Name))
			require.NoError(t, err)

			for field := range keyMap.Type.Fields() {
				binding, err := config.ParseBinding(snakeCase(field.Name))
				require.NoError(t, err, "%s.%s", keyMap.Name, field.Name)

				got := settings.FieldByIndex(keyMap.Index).FieldByIndex(field.Index).Interface()
				assert.Equal(t, written(cfg.Bindings[scope][binding]), got, "%s.%s", keyMap.Name, field.Name)
			}
		}
	})

	t.Run("passes the Location through", func(t *testing.T) {
		t.Parallel()

		tokyo := time.FixedZone("JST", 9*60*60)

		assert.Same(t, tokyo, bootstrap.SettingsFrom(defaultConfig(t), tokyo).Location)
	})
}

func defaultConfig(t *testing.T) config.Config {
	t.Helper()

	cfg, err := config.Load(t.Context(), config.Options{
		Path:   filepath.Join(t.TempDir(), "config.toml"),
		Logger: slog.New(slog.DiscardHandler),
	})
	require.NoError(t, err)

	return cfg
}

func snakeCase(name string) string {
	return strings.ToLower(wordBoundary.ReplaceAllString(name, "${1}_${2}"))
}

func written(keys []config.Key) []string {
	strs := make([]string, 0, len(keys))
	for _, key := range keys {
		strs = append(strs, key.String())
	}

	return strs
}
