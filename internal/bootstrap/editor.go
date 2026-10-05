package bootstrap

import (
	"log/slog"

	"github.com/DannyFestor/TuiSnip/internal/adapters/config"
	"github.com/DannyFestor/TuiSnip/internal/adapters/editor"
)

func newExternalEditor(cfg config.Config, options Options, logger *slog.Logger) *editor.Launcher {
	return editor.New(editor.Options{
		Configured: cfg.Editor,
		Environ:    options.Environ,
		LookPath:   options.LookPath,
		Logger:     logger,
	})
}
