package testsettings

import (
	"log/slog"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/config"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/mainscreen"
	"github.com/DannyFestor/TuiSnip/internal/bootstrap"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

func Default(t *testing.T) tui.Settings {
	t.Helper()

	cfg, err := config.Load(t.Context(), config.Options{
		Path:   filepath.Join(t.TempDir(), "config.toml"),
		Logger: slog.New(slog.DiscardHandler),
	})
	require.NoError(t, err)

	return bootstrap.SettingsFrom(
		cfg,
		time.UTC,
		mainscreen.Remembered{SortOrder: domain.SortOrderTitle, CollapsedFolders: nil},
	)
}
