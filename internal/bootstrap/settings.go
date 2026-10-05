package bootstrap

import (
	"time"

	"github.com/DannyFestor/TuiSnip/internal/adapters/config"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/mainscreen"
)

func SettingsFrom(cfg config.Config, location *time.Location, remembered mainscreen.Remembered) tui.Settings {
	return tui.Settings{
		Keys:          keysFrom(cfg.Bindings),
		ForcedQuitKey: config.ForcedQuitKey().String(),
		Location:      location,
		Remembered:    remembered,
	}
}

func keysFrom(bindings map[config.Scope]map[config.Binding][]config.Key) binding.Keys {
	keys := make(binding.Keys, len(bindings))

	for scope, named := range bindings {
		written := make(map[string][]string, len(named))
		for name, bound := range named {
			written[name.String()] = keyStrings(bound)
		}

		keys[binding.Scope(scope.String())] = written
	}

	return keys
}

func keyStrings(keys []config.Key) []string {
	written := make([]string, 0, len(keys))
	for _, key := range keys {
		written = append(written, key.String())
	}

	return written
}
